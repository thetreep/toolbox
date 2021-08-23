package httpx

import (
	"net/http"
	"time"

	"go.opencensus.io/plugin/ochttp"
	propagation "go.opencensus.io/plugin/ochttp/propagation/b3"
)

// DefaultClient is a default http.Client with
// tracing enabled.
var DefaultClient = http.Client{
	Timeout: time.Duration(30 * time.Second),
	Transport: &ochttp.Transport{
		Propagation: &propagation.HTTPFormat{},
	},
}
