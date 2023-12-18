package httpx

import (
	"net"
	"net/http"
	"time"
)

// DefaultClient is a default http.Client with
// tracing enabled.
// nolint
var DefaultClient = &http.Client{
	Timeout: time.Minute * 15,
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 20 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 15 * time.Second,
	},
}
