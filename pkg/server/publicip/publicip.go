package publicip

import "errors"

type Client interface {
	GetIP() (string, error)
	ProviderURL() string
}

var (
	ErrInvalidIPAddress = errors.New("invalid ip address")
)
