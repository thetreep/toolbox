package httpretry_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thetreep/toolbox/httpx"
	"github.com/thetreep/toolbox/httpx/httpretry"
	"github.com/thetreep/toolbox/httpx/redact"
	"github.com/thetreep/toolbox/logger"
)

type logEntry struct {
	Event          string `json:"event"`
	ChildRequestID string `json:"child-request-id"`
	Request        struct {
		Method string          `json:"method"`
		Body   json.RawMessage `json:"body"`
	} `json:"request"`
	Response struct {
		StatusCode int `json:"statusCode"`
	} `json:"response"`
}

// flaky answers 503 until the failures are used up, recording every request body.
type flaky struct {
	mu       sync.Mutex
	failures int
	bodies   []string
}

func (f *flaky) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	f.mu.Lock()
	f.bodies = append(f.bodies, string(body))
	fail := len(f.bodies) <= f.failures
	f.mu.Unlock()

	if fail {
		w.WriteHeader(http.StatusServiceUnavailable)

		return
	}

	_, _ = io.WriteString(w, "ok")
}

// do sends req through retry wrapping logging and returns the response body and the logged entries.
func do(t *testing.T, handler http.Handler, req func(url string) *http.Request, opts ...httpretry.Option) (string, []logEntry) {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	logging := httpx.NewLoggingTransport("demo", nil, redact.Policy{})
	client := &http.Client{Transport: httpretry.NewTransport(logging, append([]httpretry.Option{httpretry.WithBackoff(time.Millisecond, time.Millisecond)}, opts...)...)}

	var out bytes.Buffer

	ctx := logger.CtxWithLogger(context.Background(), slog.New(slog.NewJSONHandler(&out, nil)))

	resp, err := client.Do(req(srv.URL).WithContext(ctx))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	var entries []logEntry

	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		var e logEntry

		require.NoError(t, json.Unmarshal([]byte(line), &e), line)

		if e.Event == "outgoing_http_call" {
			entries = append(entries, e)
		}
	}

	return string(body), entries
}

func statuses(entries []logEntry) []int {
	var out []int

	for _, e := range entries {
		out = append(out, e.Response.StatusCode)
	}

	return out
}

func TestEachAttemptIsLoggedOnceWhenRetryWrapsLogging(t *testing.T) {
	h := &flaky{failures: 1}

	body, entries := do(t, h, func(url string) *http.Request {
		req, _ := http.NewRequest(http.MethodGet, url, nil)

		return req
	})

	assert.Equal(t, "ok", body)
	assert.Equal(t, []int{503, 200}, statuses(entries))
	assert.NotEmpty(t, entries[0].ChildRequestID)
	assert.NotEqual(t, entries[0].ChildRequestID, entries[1].ChildRequestID)
}

func TestPostIsNeitherRetriedNorLoggedTwice(t *testing.T) {
	h := &flaky{failures: 5}

	_, entries := do(t, h, func(url string) *http.Request {
		req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(`{"n":1}`))
		req.Header.Set("Content-Type", "application/json")

		return req
	})

	assert.Equal(t, []string{`{"n":1}`}, h.bodies)
	assert.Equal(t, []int{503}, statuses(entries))
	assert.JSONEq(t, `{"n":1}`, string(entries[0].Request.Body))
}

func TestStreamedBodyIsReplayedAndLoggedOnEachAttempt(t *testing.T) {
	h := &flaky{failures: 2}

	// NopCloser hides the length and GetBody, so the retry transport has to buffer the body.
	body, entries := do(t, h, func(url string) *http.Request {
		req, _ := http.NewRequest(http.MethodPut, url, io.NopCloser(strings.NewReader(`{"n":1}`)))
		req.Header.Set("Content-Type", "application/json")

		return req
	})

	assert.Equal(t, "ok", body)
	assert.Equal(t, []string{`{"n":1}`, `{"n":1}`, `{"n":1}`}, h.bodies)
	assert.Equal(t, []int{503, 503, 200}, statuses(entries))

	for _, e := range entries {
		assert.JSONEq(t, `{"n":1}`, string(e.Request.Body))
	}
}

func TestIdempotentPostIsReplayedWithItsBody(t *testing.T) {
	tests := []struct {
		name string
		opts []httpretry.Option
		set  func(*http.Request)
	}{
		{"idempotency key", nil, func(r *http.Request) { r.Header.Set("Idempotency-Key", "k1") }},
		{"WithIdempotent", []httpretry.Option{httpretry.WithIdempotent(func(r *http.Request) bool { return r.URL.Path == "/token" })}, func(*http.Request) {}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &flaky{failures: 1}

			_, entries := do(t, h, func(url string) *http.Request {
				req, _ := http.NewRequest(http.MethodPost, url+"/token", io.NopCloser(strings.NewReader(`{"n":1}`)))
				req.Header.Set("Content-Type", "application/json")
				tt.set(req)

				return req
			}, tt.opts...)

			assert.Equal(t, []string{`{"n":1}`, `{"n":1}`}, h.bodies)
			assert.Equal(t, []int{503, 200}, statuses(entries))
			assert.Equal(t, http.MethodPost, entries[1].Request.Method)
			assert.JSONEq(t, `{"n":1}`, string(entries[1].Request.Body))
		})
	}
}
