package server

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ginuerzh/gost"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// the cli for gost has a bunch of unexported functions
// to help setup chains and listeners (https://github.com/ginuerzh/gost/tree/8ab2fe6f77d43fdd5613569377f9e853a84bcc4d)
// rather than recreate this, just copying as is for now

type stringList []string

func (l *stringList) String() string {
	return fmt.Sprintf("%s", *l)
}

func (l *stringList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

type route struct {
	ServeNodes stringList
	ChainNodes stringList
	Retries    int
}

func (r *route) parseChain() (*gost.Chain, error) {
	chain := gost.NewChain()
	chain.Retries = r.Retries
	gid := 1 // group ID

	for _, ns := range r.ChainNodes {
		ngroup := gost.NewNodeGroup()
		ngroup.ID = gid
		gid++

		// parse the base nodes
		nodes, err := parseChainNode(ns)
		if err != nil {
			return nil, err
		}

		nid := 1 // node ID
		for i := range nodes {
			nodes[i].ID = nid
			nid++
		}
		ngroup.AddNode(nodes...)

		ngroup.SetSelector(nil,
			gost.WithFilter(
				&gost.FailFilter{
					MaxFails:    nodes[0].GetInt("max_fails"),
					FailTimeout: nodes[0].GetDuration("fail_timeout"),
				},
				&gost.InvalidFilter{},
			),
			gost.WithStrategy(gost.NewStrategy(nodes[0].Get("strategy"))),
		)

		if cfg := nodes[0].Get("peer"); cfg != "" {
			f, err := os.Open(cfg)
			if err != nil {
				return nil, err
			}

			peerCfg := newPeerConfig()
			peerCfg.group = ngroup
			peerCfg.baseNodes = nodes
			peerCfg.Reload(f)
			f.Close()

			go gost.PeriodReload(peerCfg, cfg)
		}

		chain.AddNodeGroup(ngroup)
	}

	return chain, nil
}

func parseChainNode(ns string) (nodes []gost.Node, err error) {
	node, err := gost.ParseNode(ns)
	if err != nil {
		return
	}

	if auth := node.Get("auth"); auth != "" && node.User == nil {
		c, err := base64.StdEncoding.DecodeString(auth)
		if err != nil {
			return nil, err
		}
		cs := string(c)
		s := strings.IndexByte(cs, ':')
		if s < 0 {
			node.User = url.User(cs)
		} else {
			node.User = url.UserPassword(cs[:s], cs[s+1:])
		}
	}
	if node.User == nil {
		users, err := parseUsers(node.Get("secrets"))
		if err != nil {
			return nil, err
		}
		if len(users) > 0 {
			node.User = users[0]
		}
	}

	serverName, sport, _ := net.SplitHostPort(node.Addr)
	if serverName == "" {
		serverName = "localhost" // default server name
	}

	rootCAs, err := loadCA(node.Get("ca"))
	if err != nil {
		return
	}
	tlsCfg := &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: !node.GetBool("secure"),
		RootCAs:            rootCAs,
	}
	wsOpts := &gost.WSOptions{}
	wsOpts.EnableCompression = node.GetBool("compression")
	wsOpts.ReadBufferSize = node.GetInt("rbuf")
	wsOpts.WriteBufferSize = node.GetInt("wbuf")
	wsOpts.UserAgent = node.Get("agent")
	wsOpts.Path = node.Get("path")

	timeout := node.GetDuration("timeout")

	var tr gost.Transporter
	switch node.Transport {
	case "tls":
		tr = gost.TLSTransporter()
	case "mtls":
		tr = gost.MTLSTransporter()
	case "ws":
		tr = gost.WSTransporter(wsOpts)
	case "mws":
		tr = gost.MWSTransporter(wsOpts)
	case "wss":
		tr = gost.WSSTransporter(wsOpts)
	case "mwss":
		tr = gost.MWSSTransporter(wsOpts)
	case "kcp":
		config, err := parseKCPConfig(node.Get("c"))
		if err != nil {
			return nil, err
		}
		if config == nil {
			conf := gost.DefaultKCPConfig
			if node.GetBool("tcp") {
				conf.TCP = true
			}
			config = &conf
		}
		tr = gost.KCPTransporter(config)
	case "ssh":
		if node.Protocol == "direct" || node.Protocol == "remote" {
			tr = gost.SSHForwardTransporter()
		} else {
			tr = gost.SSHTunnelTransporter()
		}
	case "quic":
		config := &gost.QUICConfig{
			TLSConfig:   tlsCfg,
			KeepAlive:   node.GetBool("keepalive"),
			Timeout:     timeout,
			IdleTimeout: node.GetDuration("idle"),
		}

		if cipher := node.Get("cipher"); cipher != "" {
			sum := sha256.Sum256([]byte(cipher))
			config.Key = sum[:]
		}

		tr = gost.QUICTransporter(config)
	case "http2":
		tr = gost.HTTP2Transporter(tlsCfg)
	case "h2":
		tr = gost.H2Transporter(tlsCfg, node.Get("path"))
	case "h2c":
		tr = gost.H2CTransporter(node.Get("path"))
	case "obfs4":
		tr = gost.Obfs4Transporter()
	case "ohttp":
		tr = gost.ObfsHTTPTransporter()
	case "otls":
		tr = gost.ObfsTLSTransporter()
	case "ftcp":
		tr = gost.FakeTCPTransporter()
	case "udp":
		tr = gost.UDPTransporter()
	default:
		tr = gost.TCPTransporter()
	}

	var connector gost.Connector
	switch node.Protocol {
	case "http2":
		connector = gost.HTTP2Connector(node.User)
	case "socks", "socks5":
		connector = gost.SOCKS5Connector(node.User)
	case "socks4":
		connector = gost.SOCKS4Connector()
	case "socks4a":
		connector = gost.SOCKS4AConnector()
	case "ss":
		connector = gost.ShadowConnector(node.User)
	case "ssu":
		connector = gost.ShadowUDPConnector(node.User)
	case "direct":
		connector = gost.SSHDirectForwardConnector()
	case "remote":
		connector = gost.SSHRemoteForwardConnector()
	case "forward":
		connector = gost.ForwardConnector()
	case "sni":
		connector = gost.SNIConnector(node.Get("host"))
	case "http":
		connector = gost.HTTPConnector(node.User)
	case "relay":
		connector = gost.RelayConnector(node.User)
	default:
		connector = gost.AutoConnector(node.User)
	}

	node.DialOptions = append(node.DialOptions,
		gost.TimeoutDialOption(timeout),
	)

	node.ConnectOptions = []gost.ConnectOption{
		gost.UserAgentConnectOption(node.Get("agent")),
		gost.NoTLSConnectOption(node.GetBool("notls")),
		gost.NoDelayConnectOption(node.GetBool("nodelay")),
	}

	host := node.Get("host")
	if host == "" {
		host = node.Host
	}

	sshConfig := &gost.SSHConfig{}
	if s := node.Get("ssh_key"); s != "" {
		key, err := gost.ParseSSHKeyFile(s)
		if err != nil {
			return nil, err
		}
		sshConfig.Key = key
	}
	handshakeOptions := []gost.HandshakeOption{
		gost.AddrHandshakeOption(node.Addr),
		gost.HostHandshakeOption(host),
		gost.UserHandshakeOption(node.User),
		gost.TLSConfigHandshakeOption(tlsCfg),
		gost.IntervalHandshakeOption(node.GetDuration("ping")),
		gost.TimeoutHandshakeOption(timeout),
		gost.RetryHandshakeOption(node.GetInt("retry")),
		gost.SSHConfigHandshakeOption(sshConfig),
	}

	node.Client = &gost.Client{
		Connector:   connector,
		Transporter: tr,
	}

	node.Bypass = parseBypass(node.Get("bypass"))

	ips := parseIP(node.Get("ip"), sport)
	for _, ip := range ips {
		nd := node.Clone()
		nd.Addr = ip
		// override the default node address
		nd.HandshakeOptions = append(handshakeOptions, gost.AddrHandshakeOption(ip))
		// One node per IP
		nodes = append(nodes, nd)
	}
	if len(ips) == 0 {
		node.HandshakeOptions = handshakeOptions
		nodes = []gost.Node{node}
	}

	if node.Transport == "obfs4" {
		for i := range nodes {
			if err := gost.Obfs4Init(nodes[i], false); err != nil {
				return nil, err
			}
		}
	}

	return
}

func (r *route) GenRouters() ([]router, error) {
	chain, err := r.parseChain()
	if err != nil {
		return nil, err
	}

	var rts []router

	for _, ns := range r.ServeNodes {
		node, err := gost.ParseNode(ns)
		if err != nil {
			return nil, err
		}

		if auth := node.Get("auth"); auth != "" && node.User == nil {
			c, err := base64.StdEncoding.DecodeString(auth)
			if err != nil {
				return nil, err
			}
			cs := string(c)
			s := strings.IndexByte(cs, ':')
			if s < 0 {
				node.User = url.User(cs)
			} else {
				node.User = url.UserPassword(cs[:s], cs[s+1:])
			}
		}
		authenticator, err := parseAuthenticator(node.Get("secrets"))
		if err != nil {
			return nil, err
		}
		if authenticator == nil && node.User != nil {
			kvs := make(map[string]string)
			kvs[node.User.Username()], _ = node.User.Password()
			authenticator = gost.NewLocalAuthenticator(kvs)
		}
		if node.User == nil {
			if users, _ := parseUsers(node.Get("secrets")); len(users) > 0 {
				node.User = users[0]
			}
		}
		certFile, keyFile := node.Get("cert"), node.Get("key")
		tlsCfg, err := tlsConfig(certFile, keyFile)
		if err != nil && certFile != "" && keyFile != "" {
			return nil, err
		}

		wsOpts := &gost.WSOptions{}
		wsOpts.EnableCompression = node.GetBool("compression")
		wsOpts.ReadBufferSize = node.GetInt("rbuf")
		wsOpts.WriteBufferSize = node.GetInt("wbuf")
		wsOpts.Path = node.Get("path")

		ttl := node.GetDuration("ttl")
		timeout := node.GetDuration("timeout")

		tunRoutes := parseIPRoutes(node.Get("route"))
		gw := net.ParseIP(node.Get("gw")) // default gateway
		for i := range tunRoutes {
			if tunRoutes[i].Gateway == nil {
				tunRoutes[i].Gateway = gw
			}
		}

		var ln gost.Listener
		switch node.Transport {
		case "tls":
			ln, err = gost.TLSListener(node.Addr, tlsCfg)
		case "mtls":
			ln, err = gost.MTLSListener(node.Addr, tlsCfg)
		case "ws":
			ln, err = gost.WSListener(node.Addr, wsOpts)
		case "mws":
			ln, err = gost.MWSListener(node.Addr, wsOpts)
		case "wss":
			ln, err = gost.WSSListener(node.Addr, tlsCfg, wsOpts)
		case "mwss":
			ln, err = gost.MWSSListener(node.Addr, tlsCfg, wsOpts)
		case "kcp":
			config, er := parseKCPConfig(node.Get("c"))
			if er != nil {
				return nil, er
			}
			if config == nil {
				conf := gost.DefaultKCPConfig
				if node.GetBool("tcp") {
					conf.TCP = true
				}
				config = &conf
			}
			ln, err = gost.KCPListener(node.Addr, config)
		case "ssh":
			config := &gost.SSHConfig{
				Authenticator: authenticator,
				TLSConfig:     tlsCfg,
			}
			if s := node.Get("ssh_key"); s != "" {
				key, err := gost.ParseSSHKeyFile(s)
				if err != nil {
					return nil, err
				}
				config.Key = key
			}
			if s := node.Get("ssh_authorized_keys"); s != "" {
				keys, err := gost.ParseSSHAuthorizedKeysFile(s)
				if err != nil {
					return nil, err
				}
				config.AuthorizedKeys = keys
			}
			if node.Protocol == "forward" {
				ln, err = gost.TCPListener(node.Addr)
			} else {
				ln, err = gost.SSHTunnelListener(node.Addr, config)
			}
		case "quic":
			config := &gost.QUICConfig{
				TLSConfig:   tlsCfg,
				KeepAlive:   node.GetBool("keepalive"),
				Timeout:     timeout,
				IdleTimeout: node.GetDuration("idle"),
			}
			if cipher := node.Get("cipher"); cipher != "" {
				sum := sha256.Sum256([]byte(cipher))
				config.Key = sum[:]
			}

			ln, err = gost.QUICListener(node.Addr, config)
		case "http2":
			ln, err = gost.HTTP2Listener(node.Addr, tlsCfg)
		case "h2":
			ln, err = gost.H2Listener(node.Addr, tlsCfg, node.Get("path"))
		case "h2c":
			ln, err = gost.H2CListener(node.Addr, node.Get("path"))
		case "tcp":
			// Directly use SSH port forwarding if the last chain node is forward+ssh
			if chain.LastNode().Protocol == "forward" && chain.LastNode().Transport == "ssh" {
				chain.Nodes()[len(chain.Nodes())-1].Client.Connector = gost.SSHDirectForwardConnector()
				chain.Nodes()[len(chain.Nodes())-1].Client.Transporter = gost.SSHForwardTransporter()
			}
			ln, err = gost.TCPListener(node.Addr)
		case "udp":
			ln, err = gost.UDPListener(node.Addr, &gost.UDPListenConfig{
				TTL:       ttl,
				Backlog:   node.GetInt("backlog"),
				QueueSize: node.GetInt("queue"),
			})
		case "rtcp":
			// Directly use SSH port forwarding if the last chain node is forward+ssh
			if chain.LastNode().Protocol == "forward" && chain.LastNode().Transport == "ssh" {
				chain.Nodes()[len(chain.Nodes())-1].Client.Connector = gost.SSHRemoteForwardConnector()
				chain.Nodes()[len(chain.Nodes())-1].Client.Transporter = gost.SSHForwardTransporter()
			}
			ln, err = gost.TCPRemoteForwardListener(node.Addr, chain)
		case "rudp":
			ln, err = gost.UDPRemoteForwardListener(node.Addr,
				chain,
				&gost.UDPListenConfig{
					TTL:       ttl,
					Backlog:   node.GetInt("backlog"),
					QueueSize: node.GetInt("queue"),
				})
		case "obfs4":
			if err = gost.Obfs4Init(node, true); err != nil {
				return nil, err
			}
			ln, err = gost.Obfs4Listener(node.Addr)
		case "ohttp":
			ln, err = gost.ObfsHTTPListener(node.Addr)
		case "otls":
			ln, err = gost.ObfsTLSListener(node.Addr)
		case "tun":
			cfg := gost.TunConfig{
				Name:    node.Get("name"),
				Addr:    node.Get("net"),
				Peer:    node.Get("peer"),
				MTU:     node.GetInt("mtu"),
				Routes:  tunRoutes,
				Gateway: node.Get("gw"),
			}
			ln, err = gost.TunListener(cfg)
		case "tap":
			cfg := gost.TapConfig{
				Name:    node.Get("name"),
				Addr:    node.Get("net"),
				MTU:     node.GetInt("mtu"),
				Routes:  strings.Split(node.Get("route"), ","),
				Gateway: node.Get("gw"),
			}
			ln, err = gost.TapListener(cfg)
		case "ftcp":
			ln, err = gost.FakeTCPListener(
				node.Addr,
				&gost.FakeTCPListenConfig{
					TTL:       ttl,
					Backlog:   node.GetInt("backlog"),
					QueueSize: node.GetInt("queue"),
				},
			)
		case "dns":
			ln, err = gost.DNSListener(
				node.Addr,
				&gost.DNSOptions{
					Mode:      node.Get("mode"),
					TLSConfig: tlsCfg,
				},
			)
		case "redu", "redirectu":
			ln, err = gost.UDPRedirectListener(node.Addr, &gost.UDPListenConfig{
				TTL:       ttl,
				Backlog:   node.GetInt("backlog"),
				QueueSize: node.GetInt("queue"),
			})
		default:
			ln, err = gost.TCPListener(node.Addr)
		}
		if err != nil {
			return nil, err
		}

		var handler gost.Handler
		switch node.Protocol {
		case "http2":
			handler = gost.HTTP2Handler()
		case "socks", "socks5":
			handler = gost.SOCKS5Handler()
		case "socks4", "socks4a":
			handler = gost.SOCKS4Handler()
		case "ss":
			handler = gost.ShadowHandler()
		case "http":
			handler = gost.HTTPHandler()
		case "tcp":
			handler = gost.TCPDirectForwardHandler(node.Remote)
		case "rtcp":
			handler = gost.TCPRemoteForwardHandler(node.Remote)
		case "udp":
			handler = gost.UDPDirectForwardHandler(node.Remote)
		case "rudp":
			handler = gost.UDPRemoteForwardHandler(node.Remote)
		case "forward":
			handler = gost.SSHForwardHandler()
		case "red", "redirect":
			handler = gost.TCPRedirectHandler()
		case "redu", "redirectu":
			handler = gost.UDPRedirectHandler()
		case "ssu":
			handler = gost.ShadowUDPHandler()
		case "sni":
			handler = gost.SNIHandler()
		case "tun":
			handler = gost.TunHandler()
		case "tap":
			handler = gost.TapHandler()
		case "dns":
			handler = gost.DNSHandler(node.Remote)
		case "relay":
			handler = gost.RelayHandler(node.Remote)
		default:
			// start from 2.5, if remote is not empty, then we assume that it is a forward tunnel.
			if node.Remote != "" {
				handler = gost.TCPDirectForwardHandler(node.Remote)
			} else {
				handler = gost.AutoHandler()
			}
		}

		var whitelist, blacklist *gost.Permissions
		if node.Values.Get("whitelist") != "" {
			if whitelist, err = gost.ParsePermissions(node.Get("whitelist")); err != nil {
				return nil, err
			}
		}
		if node.Values.Get("blacklist") != "" {
			if blacklist, err = gost.ParsePermissions(node.Get("blacklist")); err != nil {
				return nil, err
			}
		}

		node.Bypass = parseBypass(node.Get("bypass"))
		hosts := parseHosts(node.Get("hosts"))
		ips := parseIP(node.Get("ip"), "")

		resolver := parseResolver(node.Get("dns"))
		if resolver != nil {
			resolver.Init(
				gost.ChainResolverOption(chain),
				gost.TimeoutResolverOption(timeout),
				gost.TTLResolverOption(ttl),
				gost.PreferResolverOption(node.Get("prefer")),
				gost.SrcIPResolverOption(net.ParseIP(node.Get("ip"))),
			)
		}

		handler.Init(
			gost.AddrHandlerOption(ln.Addr().String()),
			gost.ChainHandlerOption(chain),
			gost.UsersHandlerOption(node.User),
			gost.AuthenticatorHandlerOption(authenticator),
			gost.TLSConfigHandlerOption(tlsCfg),
			gost.WhitelistHandlerOption(whitelist),
			gost.BlacklistHandlerOption(blacklist),
			gost.StrategyHandlerOption(gost.NewStrategy(node.Get("strategy"))),
			gost.MaxFailsHandlerOption(node.GetInt("max_fails")),
			gost.FailTimeoutHandlerOption(node.GetDuration("fail_timeout")),
			gost.BypassHandlerOption(node.Bypass),
			gost.ResolverHandlerOption(resolver),
			gost.HostsHandlerOption(hosts),
			gost.RetryHandlerOption(node.GetInt("retry")), // override the global retry option.
			gost.TimeoutHandlerOption(timeout),
			gost.ProbeResistHandlerOption(node.Get("probe_resist")),
			gost.KnockingHandlerOption(node.Get("knock")),
			gost.NodeHandlerOption(node),
			gost.IPsHandlerOption(ips),
			gost.TCPModeHandlerOption(node.GetBool("tcp")),
			gost.IPRoutesHandlerOption(tunRoutes...),
		)

		rt := router{
			node:     node,
			server:   &gost.Server{Listener: ln},
			handler:  handler,
			chain:    chain,
			resolver: resolver,
			hosts:    hosts,
		}
		rts = append(rts, rt)
	}

	return rts, nil
}

type router struct {
	node     gost.Node
	server   *gost.Server
	handler  gost.Handler
	chain    *gost.Chain
	resolver gost.Resolver
	hosts    *gost.Hosts
}

func (r *router) Serve() error {
	log.Printf("%s on %s", r.node.String(), r.server.Addr())
	return r.server.Serve(r.handler)
}

func (r *router) Close() error {
	if r == nil || r.server == nil {
		return nil
	}
	return r.server.Close()
}

var (
	routers []router
)

type baseConfig struct {
	route
	Routes []route
	Debug  bool
}

var (
	defaultCertFile = "cert.pem"
	defaultKeyFile  = "key.pem"
)

// Load the certificate from cert and key files, will use the default certificate if the provided info are invalid.
func tlsConfig(certFile, keyFile string) (*tls.Config, error) {
	if certFile == "" || keyFile == "" {
		certFile, keyFile = defaultCertFile, defaultKeyFile
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}

func loadCA(caFile string) (cp *x509.CertPool, err error) {
	if caFile == "" {
		return
	}
	cp = x509.NewCertPool()
	data, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	if !cp.AppendCertsFromPEM(data) {
		return nil, errors.New("AppendCertsFromPEM failed")
	}
	return
}

func parseKCPConfig(configFile string) (*gost.KCPConfig, error) {
	if configFile == "" {
		return nil, nil
	}
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &gost.KCPConfig{}
	if err = json.NewDecoder(file).Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}

func parseUsers(authFile string) (users []*url.Userinfo, err error) {
	if authFile == "" {
		return
	}

	file, err := os.Open(authFile)
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		s := strings.SplitN(line, " ", 2)
		if len(s) == 1 {
			users = append(users, url.User(strings.TrimSpace(s[0])))
		} else if len(s) == 2 {
			users = append(users, url.UserPassword(strings.TrimSpace(s[0]), strings.TrimSpace(s[1])))
		}
	}

	err = scanner.Err()
	return
}

func parseAuthenticator(s string) (gost.Authenticator, error) {
	if s == "" {
		return nil, nil
	}
	f, err := os.Open(s)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	au := gost.NewLocalAuthenticator(nil)
	au.Reload(f)

	go gost.PeriodReload(au, s)

	return au, nil
}

func parseIP(s string, port string) (ips []string) {
	if s == "" {
		return
	}
	if port == "" {
		port = "8080" // default port
	}

	file, err := os.Open(s)
	if err != nil {
		ss := strings.Split(s, ",")
		for _, s := range ss {
			s = strings.TrimSpace(s)
			if s != "" {
				// TODO: support IPv6
				if !strings.Contains(s, ":") {
					s = s + ":" + port
				}
				ips = append(ips, s)
			}

		}
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.Contains(line, ":") {
			line = line + ":" + port
		}
		ips = append(ips, line)
	}
	return
}

func parseBypass(s string) *gost.Bypass {
	if s == "" {
		return nil
	}
	var matchers []gost.Matcher
	var reversed bool
	if strings.HasPrefix(s, "~") {
		reversed = true
		s = strings.TrimLeft(s, "~")
	}

	f, err := os.Open(s)
	if err != nil {
		for _, s := range strings.Split(s, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			matchers = append(matchers, gost.NewMatcher(s))
		}
		return gost.NewBypass(reversed, matchers...)
	}
	defer f.Close()

	bp := gost.NewBypass(reversed)
	bp.Reload(f)
	go gost.PeriodReload(bp, s)

	return bp
}

func parseResolver(cfg string) gost.Resolver {
	if cfg == "" {
		return nil
	}
	var nss []gost.NameServer

	f, err := os.Open(cfg)
	if err != nil {
		for _, s := range strings.Split(cfg, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if strings.HasPrefix(s, "https") {
				p := "https"
				u, _ := url.Parse(s)
				if u == nil || u.Scheme == "" {
					continue
				}
				if u.Scheme == "https-chain" {
					p = u.Scheme
				}
				ns := gost.NameServer{
					Addr:     s,
					Protocol: p,
				}
				nss = append(nss, ns)
				continue
			}

			ss := strings.Split(s, "/")
			if len(ss) == 1 {
				ns := gost.NameServer{
					Addr: ss[0],
				}
				nss = append(nss, ns)
			}
			if len(ss) == 2 {
				ns := gost.NameServer{
					Addr:     ss[0],
					Protocol: ss[1],
				}
				nss = append(nss, ns)
			}
		}
		return gost.NewResolver(0, nss...)
	}
	defer f.Close()

	resolver := gost.NewResolver(0)
	resolver.Reload(f)

	go gost.PeriodReload(resolver, cfg)

	return resolver
}

func parseHosts(s string) *gost.Hosts {
	f, err := os.Open(s)
	if err != nil {
		return nil
	}
	defer f.Close()

	hosts := gost.NewHosts()
	hosts.Reload(f)

	go gost.PeriodReload(hosts, s)

	return hosts
}

func parseIPRoutes(s string) (routes []gost.IPRoute) {
	if s == "" {
		return
	}

	file, err := os.Open(s)
	if err != nil {
		ss := strings.Split(s, ",")
		for _, s := range ss {
			if _, inet, _ := net.ParseCIDR(strings.TrimSpace(s)); inet != nil {
				routes = append(routes, gost.IPRoute{Dest: inet})
			}
		}
		return
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Replace(scanner.Text(), "\t", " ", -1)
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var route gost.IPRoute
		var ss []string
		for _, s := range strings.Split(line, " ") {
			if s = strings.TrimSpace(s); s != "" {
				ss = append(ss, s)
			}
		}
		if len(ss) > 0 && ss[0] != "" {
			_, route.Dest, _ = net.ParseCIDR(strings.TrimSpace(ss[0]))
			if route.Dest == nil {
				continue
			}
		}
		if len(ss) > 1 && ss[1] != "" {
			route.Gateway = net.ParseIP(ss[1])
		}
		routes = append(routes, route)
	}
	return routes
}

type peerConfig struct {
	Strategy    string `json:"strategy"`
	MaxFails    int    `json:"max_fails"`
	FailTimeout time.Duration
	period      time.Duration // the period for live reloading
	Nodes       []string      `json:"nodes"`
	group       *gost.NodeGroup
	baseNodes   []gost.Node
	stopped     chan struct{}
}

func newPeerConfig() *peerConfig {
	return &peerConfig{
		stopped: make(chan struct{}),
	}
}

func (cfg *peerConfig) Validate() {
}

func (cfg *peerConfig) Reload(r io.Reader) error {
	if cfg.Stopped() {
		return nil
	}

	if err := cfg.parse(r); err != nil {
		return err
	}
	cfg.Validate()

	group := cfg.group
	group.SetSelector(
		nil,
		gost.WithFilter(
			&gost.FailFilter{
				MaxFails:    cfg.MaxFails,
				FailTimeout: cfg.FailTimeout,
			},
			&gost.InvalidFilter{},
		),
		gost.WithStrategy(gost.NewStrategy(cfg.Strategy)),
	)

	gNodes := cfg.baseNodes
	nid := len(gNodes) + 1
	for _, s := range cfg.Nodes {
		nodes, err := parseChainNode(s)
		if err != nil {
			return err
		}

		for i := range nodes {
			nodes[i].ID = nid
			nid++
		}

		gNodes = append(gNodes, nodes...)
	}

	nodes := group.SetNodes(gNodes...)
	for _, node := range nodes[len(cfg.baseNodes):] {
		if node.Bypass != nil {
			node.Bypass.Stop() // clear the old nodes
		}
	}

	return nil
}

func (cfg *peerConfig) parse(r io.Reader) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	// compatible with JSON format
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(cfg); err == nil {
		return nil
	}

	split := func(line string) []string {
		if line == "" {
			return nil
		}
		if n := strings.IndexByte(line, '#'); n >= 0 {
			line = line[:n]
		}
		line = strings.Replace(line, "\t", " ", -1)
		line = strings.TrimSpace(line)

		var ss []string
		for _, s := range strings.Split(line, " ") {
			if s = strings.TrimSpace(s); s != "" {
				ss = append(ss, s)
			}
		}
		return ss
	}

	cfg.Nodes = nil
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		ss := split(line)
		if len(ss) < 2 {
			continue
		}

		switch ss[0] {
		case "strategy":
			cfg.Strategy = ss[1]
		case "max_fails":
			cfg.MaxFails, _ = strconv.Atoi(ss[1])
		case "fail_timeout":
			cfg.FailTimeout, _ = time.ParseDuration(ss[1])
		case "reload":
			cfg.period, _ = time.ParseDuration(ss[1])
		case "peer":
			cfg.Nodes = append(cfg.Nodes, ss[1])
		}
	}

	return scanner.Err()
}

func (cfg *peerConfig) Period() time.Duration {
	if cfg.Stopped() {
		return -1
	}
	return cfg.period
}

// Stop stops reloading.
func (cfg *peerConfig) Stop() {
	select {
	case <-cfg.stopped:
	default:
		close(cfg.stopped)
	}
}

// Stopped checks whether the reloader is stopped.
func (cfg *peerConfig) Stopped() bool {
	select {
	case <-cfg.stopped:
		return true
	default:
		return false
	}
}
