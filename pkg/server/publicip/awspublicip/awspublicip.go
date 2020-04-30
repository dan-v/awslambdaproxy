package awspublicip

import (
	"bytes"
	"fmt"
	"github.com/dan-v/awslambdaproxy/pkg/server/publicip"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	DefaultHTTPTimeout = time.Second * 10
	AWSProviderURL     = "http://checkip.amazonaws.com/"
)

type PublicIPClient struct {
	providerURL string
	httpClient  *http.Client
}

func New() *PublicIPClient {
	return &PublicIPClient{
		providerURL: AWSProviderURL,
		httpClient: &http.Client{
			Timeout: DefaultHTTPTimeout,
		},
	}
}

func (p *PublicIPClient) GetIP() (string, error) {
	resp, err := p.httpClient.Get(p.providerURL)
	if err != nil {
		return "", fmt.Errorf("http request to get ip address from %v failed: %w", p.providerURL, err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body from %v failed: %w", p.providerURL, err)
	}

	ip := string(bytes.TrimSpace(buf))
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("unable to parse ip %v: %w",
			publicip.ErrInvalidIPAddress, publicip.ErrInvalidIPAddress)
	}
	return ip, nil
}

func (p *PublicIPClient) ProviderURL() string {
	return p.providerURL
}
