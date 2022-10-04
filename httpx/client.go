package httpx

import (
	"net"
	"net/http"
	"time"

	"go.opencensus.io/plugin/ochttp"
	propagation "go.opencensus.io/plugin/ochttp/propagation/b3"
)

// DefaultClient is a default http.Client with
// tracing enabled.
// nolint
var DefaultClient = &http.Client{
	Timeout: time.Minute * 15,
	Transport: &ochttp.Transport{
		Base: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 20 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 15 * time.Second,
		},
		Propagation: &propagation.HTTPFormat{},
	},
}
