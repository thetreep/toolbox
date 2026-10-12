package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/jussi-kalliokoski/slogdriver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/thetreep/toolbox/httpx/redact"
	"github.com/thetreep/toolbox/logger"
)

var update = flag.Bool("update", false, "rewrite golden files")

func assertGolden(t *testing.T, path, got string) {
	t.Helper()

	// Round-tripping through a map sorts the keys: context attributes are appended in random order.
	var entry map[string]any
	require.NoError(t, json.Unmarshal([]byte(got), &entry))

	pretty, err := json.MarshalIndent(entry, "", "  ")
	require.NoError(t, err)

	pretty = append(pretty, '\n')

	if *update {
		require.NoError(t, os.WriteFile(path, pretty, 0o644))
	}

	want, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(want), string(pretty))
}

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *lockedBuffer) lines() []string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return strings.Split(strings.TrimSpace(b.buf.String()), "\n")
}

func jsonLogger(out io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		},
	}))
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func cannedResponse(status int, contentType, body string) roundTripFunc {
	return func(r *http.Request) (*http.Response, error) {
		if r.Body != nil {
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close()
		}

		return &http.Response{
			Status:     http.StatusText(status),
			StatusCode: status,
			Header:     http.Header{"Content-Type": {contentType}, "Set-Cookie": {"sid=abc"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	}
}

func newTestTransport(base http.RoundTripper, opts ...LoggingOption) *LoggingTransport {
	tr := NewLoggingTransport("demo", base, redact.Policy{}, opts...)

	var tick int64

	tr.now = func() time.Time {
		tick++

		return time.Date(2026, 4, 21, 8, 0, 0, 0, time.UTC).Add(time.Duration(tick) * 125 * time.Millisecond)
	}
	tr.newID = func() string { return "11111111-2222-4333-8444-555555555555" }

	return tr
}

func doRequest(t *testing.T, tr http.RoundTripper, out io.Writer, method, url, contentType, body string) (*http.Response, string) {
	t.Helper()

	ctx := logger.CtxWithLogger(context.Background(), jsonLogger(out))

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	require.NoError(t, err)

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	req.Header.Set("Authorization", "Bearer abcdef1234567890xyz")

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)

	got, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	return resp, string(got)
}

func parseEntry(t *testing.T, line string) map[string]any {
	t.Helper()

	var entry map[string]any
	require.NoError(t, json.Unmarshal([]byte(line), &entry), "entry must be valid JSON: %.200s", line)

	return entry
}

func TestLoggingEntryGolden(t *testing.T) {
	var out lockedBuffer

	tr := newTestTransport(
		cannedResponse(200, "application/json; charset=UTF-8", `{ "routes": [ {"distanceMeters": 1444, "token": "t0k" } ] }`),
		WithParentRequestID(func(context.Context) string { return "parent-1" }),
		WithUserAgent("imaas"),
	)

	_, body := doRequest(t, tr, &out, http.MethodPost, "https://api.example.test/v1/search?q=paris&key=SECRET",
		"application/json", `{"origin": {"city": "Paris"}, "password": "hunter2", "email": "jean@example.com"}`)
	assert.JSONEq(t, `{"routes":[{"distanceMeters":1444,"token":"t0k"}]}`, body, "the caller gets the real body")

	lines := out.lines()
	require.Len(t, lines, 1)

	assertGolden(t, "testdata/logging_entry.golden.json", lines[0])

	// Keys the reporting aggregator reads.
	entry := parseEntry(t, lines[0])
	assert.Equal(t, "outgoing_http_call", entry["event"])
	assert.Equal(t, "demo", entry["target-api"])
	assert.Equal(t, float64(200), entry["response"].(map[string]any)["statusCode"])
	assert.Equal(t, "/v1/search", entry["request"].(map[string]any)["path"])
	assert.Equal(t, float64(125000000), entry["durationNanoSeconds"])
}

func TestDoesNotMutateCallerRequest(t *testing.T) {
	var out lockedBuffer

	tr := newTestTransport(cannedResponse(200, "text/plain", "ok"),
		WithParentRequestID(func(context.Context) string { return "" }), WithUserAgent("imaas"))

	ctx := logger.CtxWithLogger(context.Background(), jsonLogger(&out))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.example.test/x", strings.NewReader("payload"))
	require.NoError(t, err)

	req.Header.Set("Accept", "text/plain")

	originalBody := req.Body

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)
	_ = resp.Body.Close()

	assert.Equal(t, http.Header{"Accept": {"text/plain"}}, req.Header)
	assert.True(t, originalBody == req.Body, "caller body is untouched")
	assert.Equal(t, ctx, req.Context())
	assert.Equal(t, "unknown", resp.Request.Header.Get("X-Parent-Request-Id"))
}

func TestResponseIsStreamedAndLoggedOnEOF(t *testing.T) {
	var out lockedBuffer

	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "first,")
		w.(http.Flusher).Flush()
		<-release
		_, _ = io.WriteString(w, "second")
	}))
	defer srv.Close()

	tr := NewLoggingTransport("demo", nil, redact.Policy{})
	ctx := logger.CtxWithLogger(context.Background(), jsonLogger(&out))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)

	buf := make([]byte, 6)
	_, err = io.ReadFull(resp.Body, buf)
	require.NoError(t, err)
	assert.Equal(t, "first,", string(buf))
	assert.Equal(t, []string{""}, out.lines(), "nothing is logged while the body is still streaming")

	close(release)

	rest, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "second", string(rest))

	lines := out.lines()
	require.Len(t, lines, 1)
	assert.Contains(t, parseEntry(t, lines[0])["response"].(map[string]any)["body"], "first,second")

	require.NoError(t, resp.Body.Close())
	assert.Len(t, out.lines(), 1, "closing after EOF does not log twice")
}

func TestLoggedOnCloseWithoutReading(t *testing.T) {
	var out lockedBuffer

	tr := newTestTransport(cannedResponse(200, "text/plain", "unread"))
	ctx := logger.CtxWithLogger(context.Background(), jsonLogger(&out))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.test/", nil)

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	require.Len(t, out.lines(), 1)
}

func TestDurationStopsAtResponseHeaders(t *testing.T) {
	var out lockedBuffer

	tr := newTestTransport(cannedResponse(200, "text/plain", "slow body"))

	clock := time.Date(2026, 4, 21, 8, 0, 0, 0, time.UTC)
	tr.now = func() time.Time { return clock }

	ctx := logger.CtxWithLogger(context.Background(), jsonLogger(&out))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.test/", nil)

	tr.base = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		clock = clock.Add(300 * time.Millisecond)

		return cannedResponse(200, "text/plain", "slow body")(r)
	})

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)

	clock = clock.Add(10 * time.Second)

	_, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	entry := parseEntry(t, out.lines()[0])
	assert.Equal(t, float64(300*time.Millisecond), entry["durationNanoSeconds"])
	assert.Equal(t, "300ms", entry["duration"])
	assert.Contains(t, entry["msg"], "(300ms)")
}

func TestPathIsRedactedInSummaryAndSpan(t *testing.T) {
	var out lockedBuffer

	tr := newTestTransport(cannedResponse(200, "text/plain", "ok"))
	_, _ = doRequest(t, tr, &out, http.MethodGet, "https://api.example.test/users/jean@example.com", "", "")

	line := out.lines()[0]
	assert.NotContains(t, line, "jean@example.com")

	_, _ = doRequest(t, tr, &out, http.MethodGet, "https://api.example.test/cards/4111%201111%201111%201111", "", "")
	assert.NotContains(t, out.lines()[1], "4111")
}

func TestBodyReadErrorIsLogged(t *testing.T) {
	var out lockedBuffer

	tr := newTestTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200, Status: "OK", Header: http.Header{"Content-Type": {"text/plain"}},
			Body: io.NopCloser(io.MultiReader(strings.NewReader("partial"), iotestErrReader{})), Request: r,
		}, nil
	}))
	ctx := logger.CtxWithLogger(context.Background(), jsonLogger(&out))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.test/", nil)

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)

	_, err = io.ReadAll(resp.Body)
	require.Error(t, err)

	entry := parseEntry(t, out.lines()[0])
	assert.Equal(t, "WARN", entry["level"])
	assert.Equal(t, "connection reset", entry["error"])
}

type iotestErrReader struct{}

func (iotestErrReader) Read([]byte) (int, error) { return 0, errors.New("connection reset") }

func TestBinaryBodiesAreLoggedAsSizeOnly(t *testing.T) {
	var out lockedBuffer

	pdf := "%PDF-1.7\x00\x01\x02" + strings.Repeat("\xff", 1000)
	tr := newTestTransport(cannedResponse(200, "application/pdf", pdf))

	doRequest(t, tr, &out, http.MethodGet, "https://api.example.test/ticket.pdf", "", "")

	entry := parseEntry(t, out.lines()[0])
	resp := entry["response"].(map[string]any)
	assert.Equal(t, "[binary: application/pdf, 1011 bytes]", resp["body"])
	assert.Equal(t, float64(1011), resp["bodySize"])
}

func TestTransportErrorIsLogged(t *testing.T) {
	var out lockedBuffer

	boom := errors.New("connection reset")
	tr := newTestTransport(roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, boom }))

	ctx := logger.CtxWithLogger(context.Background(), jsonLogger(&out))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.test/x", nil)

	resp, err := tr.RoundTrip(req)
	require.ErrorIs(t, err, boom)
	assert.Nil(t, resp)

	entry := parseEntry(t, out.lines()[0])
	assert.Contains(t, entry["msg"], "-> no response")
	assert.Nil(t, entry["response"])
	assert.Equal(t, "connection reset", entry["error"])
}

func TestRealServerWithRequestBodyStreaming(t *testing.T) {
	var out lockedBuffer

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(append([]byte(`{"echo":`), append(b, '}')...))
	}))
	defer srv.Close()

	tr := NewLoggingTransport("demo", nil, redact.Policy{})
	_, body := doRequest(t, tr, &out, http.MethodPost, srv.URL, "application/json", `{"password":"p","n":1}`)
	assert.JSONEq(t, `{"echo":{"password":"p","n":1}}`, body)

	entry := parseEntry(t, out.lines()[0])
	assert.Equal(t, map[string]any{"password": "REDACTED", "n": float64(1)}, entry["request"].(map[string]any)["body"])
	assert.Equal(t, map[string]any{"echo": map[string]any{"password": "REDACTED", "n": float64(1)}},
		entry["response"].(map[string]any)["body"])
}

func TestSpanAttributes(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	var out lockedBuffer

	tr := newTestTransport(cannedResponse(503, "text/plain", "down"))
	tr.tracer = provider.Tracer("test")

	doRequest(t, tr, &out, http.MethodGet, "https://api.example.test/v1/x", "", "")

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrs := map[string]any{}
	for _, kv := range spans[0].Attributes {
		attrs[string(kv.Key)] = kv.Value.AsInterface()
	}

	assert.Equal(t, "demo outgoing HTTP call", spans[0].Name)
	assert.Equal(t, "GET", attrs["http.request.method"])
	assert.Equal(t, "api.example.test", attrs["server.address"])
	assert.Equal(t, "/v1/x", attrs["url.path"])
	assert.Equal(t, int64(503), attrs["http.response.status_code"])
	assert.Equal(t, "demo", attrs["target-api"])
	assert.Equal(t, "Error", spans[0].Status.Code.String())
}

type fakeSink struct {
	mu   sync.Mutex
	puts map[string][]byte
	err  error
}

func (s *fakeSink) Put(_ context.Context, key string, body []byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.err != nil {
		return "", s.err
	}

	if s.puts == nil {
		s.puts = map[string][]byte{}
	}

	s.puts[key] = body

	return "fake://" + key, nil
}

func bigJSON(n int) string {
	var b strings.Builder

	b.WriteString(`{"password":"hunter2","items":[`)

	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}

		b.WriteString(`{"label":"café <b>&</b> \"quoted\"","mail":"jean@example.com","n":12345}`)
	}

	b.WriteString(`]}`)

	return b.String()
}

func TestEntryStaysUnderBudget(t *testing.T) {
	for _, status := range []int{200, 500} {
		var out lockedBuffer

		response := bigJSON(60000)
		request := bigJSON(60000)
		require.Greater(t, len(response), 3<<20)

		tr := newTestTransport(cannedResponse(status, "application/json", response))
		doRequest(t, tr, &out, http.MethodPost, "https://api.example.test/big", "application/json", request)

		line := out.lines()[0]
		t.Logf("status %d: entry is %d bytes for %d+%d body bytes", status, len(line), len(request), len(response))
		assert.LessOrEqual(t, len(line), DefaultEntryBudget, "status %d", status)

		entry := parseEntry(t, line)
		for _, side := range []string{"request", "response"} {
			part := entry[side].(map[string]any)
			assert.Equal(t, true, part["bodyTruncated"], side)
			assert.Equal(t, float64(len(response)), part["bodySize"], side)
			body, ok := part["body"].(string)
			require.True(t, ok, "%s body is a string preview", side)
			assert.Regexp(t, `…\[truncated \d+/\d+ bytes\]$`, body)
			assert.True(t, utf8.ValidString(body))
			assert.NotContains(t, body, "hunter2")
			assert.NotContains(t, body, "jean@example.com")
			assert.NotContains(t, part, "bodyRef", "no sink configured")
		}

		reqBody := len(entry["request"].(map[string]any)["body"].(string))
		respBody := len(entry["response"].(map[string]any)["body"].(string))

		if status >= 400 {
			assert.Greater(t, respBody, 2*reqBody, "response is prioritized on errors")
		} else {
			assert.InDelta(t, reqBody, respBody, float64(reqBody)/10, "bodies share the budget evenly")
		}
	}
}

func TestPreviewCutsOnRuneBoundary(t *testing.T) {
	var out lockedBuffer

	tr := newTestTransport(cannedResponse(200, "text/plain", strings.Repeat("é", 400000)), WithEntryBudget(20000))
	doRequest(t, tr, &out, http.MethodGet, "https://api.example.test/", "", "")

	line := out.lines()[0]
	assert.LessOrEqual(t, len(line), 20000)

	body := parseEntry(t, line)["response"].(map[string]any)["body"].(string)
	assert.True(t, utf8.ValidString(body))
	assert.NotContains(t, body, "�")
}

func TestOversizeBodiesGoToTheSink(t *testing.T) {
	var out lockedBuffer

	sink := &fakeSink{}
	response := bigJSON(20000)
	tr := newTestTransport(cannedResponse(200, "application/json", response), WithBodySink(sink))

	doRequest(t, tr, &out, http.MethodPost, "https://api.example.test/big", "application/json", `{"small":true,"password":"p"}`)

	entry := parseEntry(t, out.lines()[0])
	resp := entry["response"].(map[string]any)
	req := entry["request"].(map[string]any)

	assert.Equal(t, "fake://2026-04-21/demo/11111111-2222-4333-8444-555555555555/response", resp["bodyRef"])
	assert.Equal(t, true, resp["bodyTruncated"])
	assert.NotContains(t, req, "bodyRef", "a body that fits is not stored")
	assert.Equal(t, map[string]any{"small": true, "password": "REDACTED"}, req["body"])

	require.Len(t, sink.puts, 1)

	stored := sink.puts["2026-04-21/demo/11111111-2222-4333-8444-555555555555/response"]
	assert.True(t, json.Valid(stored))
	assert.NotContains(t, string(stored), "hunter2")
	assert.NotContains(t, string(stored), "jean@example.com")
	assert.Greater(t, len(stored), 1<<20/2)
}

func TestSinkFailureKeepsThePreview(t *testing.T) {
	var out lockedBuffer

	tr := newTestTransport(cannedResponse(200, "application/json", bigJSON(20000)), WithBodySink(&fakeSink{err: errors.New("queue full")}))
	doRequest(t, tr, &out, http.MethodGet, "https://api.example.test/big", "", "")

	lines := out.lines()
	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], `"level":"WARN"`)
	assert.Contains(t, lines[0], "queue full")

	resp := parseEntry(t, lines[1])["response"].(map[string]any)
	assert.Equal(t, true, resp["bodyTruncated"])
	assert.NotContains(t, resp, "bodyRef")
}

func TestCaptureIsBoundedWithoutSink(t *testing.T) {
	var out lockedBuffer

	tr := newTestTransport(cannedResponse(200, "text/plain", strings.Repeat("x", 5<<20)), WithEntryBudget(10000))
	_, body := doRequest(t, tr, &out, http.MethodGet, "https://api.example.test/", "", "")
	assert.Len(t, body, 5<<20, "the caller always gets the full body")

	resp := parseEntry(t, out.lines()[0])["response"].(map[string]any)
	assert.Equal(t, float64(5<<20), resp["bodySize"])
	assert.Contains(t, resp["body"], "bytes]")
}

func TestGCPHandlerEmitsJSONBodiesRaw(t *testing.T) {
	var out lockedBuffer

	tr := newTestTransport(cannedResponse(200, "application/json", `{"a":[1,{"b":"x<y&z"}],"token":"t"}`))

	ctx := logger.CtxWithLogger(context.Background(), slog.New(slogdriver.NewHandler(&out, slogdriver.Config{ProjectID: "p"})))
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.example.test/x", strings.NewReader(`{"password":"p","n":1}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	line := out.lines()[0]
	entry := parseEntry(t, line)

	assert.Equal(t, map[string]any{"password": "REDACTED", "n": float64(1)}, entry["request"].(map[string]any)["body"])
	assert.Equal(t, map[string]any{"a": []any{float64(1), map[string]any{"b": "x<y&z"}}, "token": "REDACTED"},
		entry["response"].(map[string]any)["body"])
	assert.Contains(t, line, `"body":{"a":[1,{"b":"x<y&z"}],"token":"REDACTED"}`, "emitted raw, not as a quoted string")
	assert.Equal(t, "demo", entry["target-api"])
}
