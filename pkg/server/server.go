package server

import (
	"fmt"
	"runtime"
	"time"

	"github.com/dan-v/awslambdaproxy/pkg/server/publicip"
	"github.com/dan-v/awslambdaproxy/pkg/server/publicip/awspublicip"
	"github.com/sirupsen/logrus"
)

const (
	// LambdaMinMemorySize is the minimum memory size for a Lambda function in MB
	LambdaMinMemorySize = 128
	// LambdaMaxMemorySize is the maximum memory size for a Lambda function in MB
	LambdaMaxMemorySize = 3008

	// LambdaDelayedCleanupTime is the time to wait if active connections still exist
	// this should be never be higher than LambdaExecutionTimeoutBuffer or function timeout
	// will happen before cleanup occurs
	LambdaDelayedCleanupTime = time.Second * 20
	// LambdaExecutionTimeoutBuffer is the time added to user specified execution frequency
	// to get the overall function timeout value
	LambdaExecutionTimeoutBuffer = time.Second * 30

	// LambdaMinExecutionFrequency is the minimum frequency for function execution.
	LambdaMinExecutionFrequency = time.Second * 60
	// LambdaMaxExecutionFrequency is the maximum frequency for function execution.
	// The current max execution time is 900 seconds, but this takes into account
	// LambdaExecutionTimeoutBuffer + 10 seconds of leeway
	LambdaMaxExecutionFrequency = time.Second * 860
)

// Config is used to define the configuration for Server
type Config struct {
	// LambdaRegions is all regions to execute Lambda functions in
	LambdaRegions []string
	// LambdaMemory is the size of memory to assign Lambda function
	LambdaMemory int
	// LambdaExecutionFrequency is the frequency at which to execute Lambda functions
	LambdaExecutionFrequency time.Duration
	// ProxyListeners defines all listeners, protocol, and auth information
	// in format like this [scheme://][user:pass@host]:port.
	// see https://github.com/ginuerzh/gost/blob/master/README_en.md#getting-started
	ProxyListeners []string
	// ProxyDebug is whether debug logging should be shown for proxy traffic
	// note this will log all visited domains
	ProxyDebug bool
	// ReverseTunnelSSHUser is the ssh user to use for the lambda reverse ssh tunnel
	ReverseTunnelSSHUser string
	// ReverseTunnelSSHPort is the ssh port to use for the lambda reverse ssh tunnel
	ReverseTunnelSSHPort string
	// Debug enables general debug logging
	Debug bool
	// Bypass is a comma separated list of ips/domains to bypass proxy
	Bypass string
}

// Server is the long running server component of awslambdaproxy
type Server struct {
	publicIPClient           publicip.Client
	lambdaRegions            []string
	lambdaMemory             int64
	lambdaExecutionFrequency time.Duration
	lambdaTimeoutSeconds     int64
	proxyListeners           []string
	proxyDebug               bool
	reverseTunnelSSHUser     string
	reverseTunnelSSHPort     string
	debug                    bool
	bypass                   string
	logger                   *logrus.Logger
}

func New(config Config) (*Server, error) {
	logger := logrus.New()
	if config.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	err := validateConfig(config)
	if err != nil {
		return nil, err
	}

	functionTimeout := int(config.LambdaExecutionFrequency.Seconds()) + int(LambdaExecutionTimeoutBuffer.Seconds())
	s := &Server{
		publicIPClient:           awspublicip.New(),
		lambdaRegions:            config.LambdaRegions,
		lambdaMemory:             int64(config.LambdaMemory),
		lambdaExecutionFrequency: config.LambdaExecutionFrequency,
		lambdaTimeoutSeconds:     int64(functionTimeout),
		proxyListeners:           config.ProxyListeners,
		proxyDebug:               config.ProxyDebug,
		reverseTunnelSSHUser:     config.ReverseTunnelSSHUser,
		reverseTunnelSSHPort:     config.ReverseTunnelSSHPort,
		debug:                    config.Debug,
		bypass:                   config.Bypass,
		logger:                   logger,
	}

	logger.WithFields(logrus.Fields{
		"publicIPClient":           s.publicIPClient.ProviderURL(),
		"lambdaRegions":            s.lambdaRegions,
		"lambdaMemory":             s.lambdaMemory,
		"lambdaExecutionFrequency": s.lambdaExecutionFrequency,
		"lambdaTimeoutSeconds":     s.lambdaTimeoutSeconds,
		"proxyListeners":           s.proxyListeners,
		"proxyDebug":               s.proxyDebug,
		"reverseTunnelSSHUser":     s.reverseTunnelSSHUser,
		"reverseTunnelSSHPort":     s.reverseTunnelSSHPort,
		"debug":                    s.debug,
		"bypass":                   s.bypass,
	}).Info("server has been configured with the following values")

	return s, nil
}

func (s *Server) Run() {
	publicIP, err := s.publicIPClient.GetIP()
	if err != nil {
		s.logger.WithError(err).Fatalf("error getting public IP address")
	}

	s.logger.Infof("setting up lambda infrastructure")
	err = setupLambdaInfrastructure(s.lambdaRegions, s.lambdaMemory, s.lambdaTimeoutSeconds)
	if err != nil {
		s.logger.WithError(err).Fatalf("failed to setup lambda infrastructure")
	}

	s.logger.Infof("starting ssh tunnel manager")
	privateKey, err := NewSSHManager()
	if err != nil {
		s.logger.WithError(err).Fatalf("failed to setup ssh tunnel manager")
	}

	s.logger.Infof("starting local proxy")
	localProxy, err := NewLocalProxy(s.proxyListeners, s.proxyDebug, s.bypass)
	if err != nil {
		s.logger.WithError(err).Fatalf("failed to setup local proxy")
	}

	s.logger.Println("starting connection manager")
	tunnelConnectionManager, err := newTunnelConnectionManager(s.lambdaExecutionFrequency, localProxy)
	if err != nil {
		s.logger.WithError(err).Fatalf("failed to setup connection manager")
	}

	s.logger.Println("starting lambda execution manager")
	_, err = newLambdaExecutionManager(publicIP, s.lambdaRegions, s.lambdaExecutionFrequency,
		s.reverseTunnelSSHUser, s.reverseTunnelSSHPort, privateKey, tunnelConnectionManager.tunnelRedeployNeeded)
	if err != nil {
		s.logger.WithError(err).Fatalf("failed to setup lambda execution manager")
	}

	s.logger.Println("#######################################")
	s.logger.Println("proxy ip address: ", publicIP)
	s.logger.Println("listeners: ", s.proxyListeners)
	s.logger.Println("#######################################")

	runtime.Goexit()
}

func validateConfig(config Config) error {
	// validate memory
	if config.LambdaMemory < LambdaMinMemorySize || config.LambdaMemory > LambdaMaxMemorySize {
		return fmt.Errorf("invalid lambda memory size '%vMB' - should be between %v and %v",
			config.LambdaMemory, LambdaMinMemorySize, LambdaMaxMemorySize)
	}
	if config.LambdaMemory%64 != 0 {
		return fmt.Errorf("invalid lambda memory size '%vMB' - should be in increments of 64MB",
			config.LambdaMemory)
	}

	// validate frequency
	if config.LambdaExecutionFrequency < LambdaMinExecutionFrequency || config.LambdaExecutionFrequency > LambdaMaxExecutionFrequency {
		return fmt.Errorf("invalid lambda execution frequency '%v' - should be between %v and %v",
			config.LambdaExecutionFrequency, LambdaMinExecutionFrequency, LambdaMaxExecutionFrequency)
	}

	// validate ssh user and port
	if config.ReverseTunnelSSHUser == "" {
		return fmt.Errorf("need to specify ReverseTunnelSSHUser")
	}
	if config.ReverseTunnelSSHPort == "" {
		return fmt.Errorf("need to specify ReverseTunnelSSHPort")
	}

	// validate listeners
	if len(config.ProxyListeners) == 0 {
		return fmt.Errorf("no listener has been specified")
	}

	// validate regions
	validRegions := GetValidLambdaRegions()
	for _, region := range config.LambdaRegions {
		valid := false
		for _, validRegion := range validRegions {
			if region == validRegion {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid region '%v' specified. valid regions: %v",
				region, validRegions)
		}
	}
	return nil
}
