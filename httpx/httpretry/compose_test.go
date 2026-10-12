package httpretry_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thetreep/toolbox/httpx"
	"github.com/thetreep/toolbox/httpx/httpretry"
	"github.com/thetreep/toolbox/httpx/redact"
	"github.com/thetreep/toolbox/logger"
)

func TestEachAttemptIsLoggedWhenRetryWrapsLogging(t *testing.T) {
	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	logging := httpx.NewLoggingTransport("demo", nil, redact.Policy{})
	client := &http.Client{Transport: httpretry.NewTransport(logging, httpretry.WithBackoff(time.Millisecond, time.Millisecond))}

	var out bytes.Buffer

	ctx := logger.CtxWithLogger(context.Background(), slog.New(slog.NewJSONHandler(&out, nil)))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)

	resp, err := client.Do(req)
	require.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	assert.Equal(t, "ok", string(body))
	assert.Equal(t, 2, strings.Count(out.String(), `"event":"outgoing_http_call"`))
	assert.Equal(t, 2, strings.Count(out.String(), `"child-request-id"`))
	assert.Contains(t, out.String(), `"statusCode":503`)
}
