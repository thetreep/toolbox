package httpx

import (
	"net"
	"net/http"
	"time"
)

// DefaultClient is a default http.Client with
// tracing enabled.
// Deprecated: use NewClient instead
// nolint
var DefaultClient = &http.Client{
	Timeout:   time.Minute * 15,
	Transport: NewTransport(),
}

//nolint:exhaustruct // we don't need all fields
func NewClient() *http.Client {
	return &http.Client{
		Timeout:   time.Minute * 15,
		Transport: NewTransport(),
	}
}

func NewTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout: 20 * time.Second,
	}).DialContext
	transport.TLSHandshakeTimeout = 15 * time.Second
	return transport
}
