package awspublicip

import (
	"context"
	"errors"
	"fmt"
	"github.com/dan-v/awslambdaproxy/pkg/server/publicip"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublicIPClient_GetIP(t *testing.T) {
	tests := []struct {
		name     string
		handler  func(w http.ResponseWriter, r *http.Request)
		expected string
		error    error
	}{
		{
			name: "valid ip address returns as expected",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "1.1.1.1")
			},
			expected: "1.1.1.1",
			error:    nil,
		},
		{
			name: "valid ip address with padding returns trimmed",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "    1.1.1.1     ")
			},
			expected: "1.1.1.1",
			error:    nil,
		},
		{
			name: "invalid ip results in error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "invalid")
			},
			expected: "",
			error:    publicip.ErrInvalidIPAddress,
		},
		{
			name: "empty response results in error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "")
			},
			expected: "",
			error:    publicip.ErrInvalidIPAddress,
		},
		{
			name: "server timeout causes deadline exceeded error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(20 * time.Millisecond)
			},
			expected: "",
			error:    context.DeadlineExceeded,
		},
		{
			name: "invalid content length causes body read error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", "1")
			},
			expected: "",
			error:    errors.New("unexpected EOF"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer ts.Close()

			p := New()
			p.providerURL = ts.URL
			p.httpClient.Timeout = time.Millisecond * 10
			ip, err := p.GetIP()
			if tt.error != nil {
				assert.Contains(t, errors.Unwrap(err).Error(), tt.error.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, ip)
		})
	}
}

func TestPublicIPClient_ProviderURL(t *testing.T) {
	p := New()
	assert.Equal(t, AWSProviderURL, p.ProviderURL())
}
