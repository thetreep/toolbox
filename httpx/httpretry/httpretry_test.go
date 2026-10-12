package httpretry

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type step struct {
	status  int
	err     error
	headers http.Header
}

type recorded struct {
	method string
	body   string
	header http.Header
}

type script struct {
	steps    []step
	calls    []recorded
	closed   int
	bodyRead int
}

func (s *script) RoundTrip(req *http.Request) (*http.Response, error) {
	var body string

	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		_ = req.Body.Close()
		body = string(b)
	}

	s.calls = append(s.calls, recorded{method: req.Method, body: body, header: req.Header.Clone()})

	st := s.steps[min(len(s.calls)-1, len(s.steps)-1)]
	if st.err != nil {
		return nil, st.err
	}

	return &http.Response{
		StatusCode: st.status,
		Header:     st.headers,
		Body:       &trackedBody{Reader: strings.NewReader("resp"), script: s},
		Request:    req,
	}, nil
}

type trackedBody struct {
	io.Reader
	script *script
}

func (b *trackedBody) Close() error { b.script.closed++; return nil }

func newTestTransport(base http.RoundTripper, opts ...Option) (Transport, *[]time.Duration) {
	tr := NewTransport(base, opts...)

	var (
		slept []time.Duration
		clock = time.Date(2026, 4, 21, 8, 0, 0, 0, time.UTC)
	)

	tr.now = func() time.Time { return clock }
	tr.jitter = func(d time.Duration) time.Duration { return d }
	tr.sleep = func(_ context.Context, d time.Duration) error {
		slept = append(slept, d)
		clock = clock.Add(d)

		return nil
	}

	return tr, &slept
}

func do(t *testing.T, tr http.RoundTripper, method, body string, headers ...string) (*http.Response, error) {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, "https://api.example.test/x", reader)
	require.NoError(t, err)

	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}

	return tr.RoundTrip(req)
}

func TestIdempotentRequestsAreRetriedOn5xx(t *testing.T) {
	for _, method := range []string{"GET", "HEAD", "OPTIONS", "PUT", "DELETE"} {
		base := &script{steps: []step{{status: 503}, {status: 502}, {status: 200}}}
		tr, slept := newTestTransport(base)

		resp, err := do(t, tr, method, "")
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode, method)
		assert.Len(t, base.calls, 3, method)
		assert.Equal(t, []time.Duration{200 * time.Millisecond, 400 * time.Millisecond}, *slept)
	}
}

func TestPostIsNotRetriedWithoutIdempotencyKey(t *testing.T) {
	base := &script{steps: []step{{status: 503}, {status: 200}}}
	tr, slept := newTestTransport(base)

	resp, err := do(t, tr, "POST", `{"book":true}`)
	require.NoError(t, err)
	assert.Equal(t, 503, resp.StatusCode)
	assert.Len(t, base.calls, 1)
	assert.Empty(t, *slept)

	base = &script{steps: []step{{err: errors.New("connection reset")}, {status: 200}}}
	tr, _ = newTestTransport(base)
	_, err = do(t, tr, "POST", `{"book":true}`)
	assert.Error(t, err)
	assert.Len(t, base.calls, 1)
}

func TestPostWithIdempotencyKeyIsRetriedWithItsBody(t *testing.T) {
	base := &script{steps: []step{{status: 500}, {status: 201}}}
	tr, _ := newTestTransport(base)

	resp, err := do(t, tr, "POST", `{"book":true}`, "Idempotency-Key", "abc")
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
	require.Len(t, base.calls, 2)
	assert.Equal(t, `{"book":true}`, base.calls[0].body)
	assert.Equal(t, `{"book":true}`, base.calls[1].body, "the body is rewound for the second try")
}

func TestNetworkErrorsAreRetriedForIdempotentRequests(t *testing.T) {
	base := &script{steps: []step{{err: errors.New("connection reset")}, {status: 200}}}
	tr, _ := newTestTransport(base)

	resp, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, base.calls, 2)
}

func TestTooManyRequestsHonorsRetryAfterSeconds(t *testing.T) {
	base := &script{steps: []step{{status: 429, headers: http.Header{"Retry-After": {"3"}}}, {status: 200}}}
	tr, slept := newTestTransport(base)

	resp, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, []time.Duration{3 * time.Second}, *slept)
}

func TestRetryAfterHTTPDate(t *testing.T) {
	base := &script{steps: []step{{status: 503, headers: http.Header{"Retry-After": {"Tue, 21 Apr 2026 08:00:05 GMT"}}}, {status: 200}}}
	tr, slept := newTestTransport(base)

	_, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, []time.Duration{5 * time.Second}, *slept)
}

func TestRetryAfterBeyondMaxElapsedGivesUp(t *testing.T) {
	base := &script{steps: []step{{status: 503, headers: http.Header{"Retry-After": {"60"}}}, {status: 200}}}
	tr, slept := newTestTransport(base)

	resp, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, 503, resp.StatusCode, "the response is handed back untouched")
	assert.Len(t, base.calls, 1)
	assert.Empty(t, *slept)
	assert.Zero(t, base.closed, "a returned response must stay open")
}

func TestMaxTriesReturnsTheLastResponse(t *testing.T) {
	base := &script{steps: []step{{status: 503}}}
	tr, _ := newTestTransport(base, WithMaxTries(4), WithMaxElapsed(time.Minute))

	resp, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, 503, resp.StatusCode)
	assert.Len(t, base.calls, 4)
	assert.Equal(t, 3, base.closed, "intermediate responses are closed, the last one is the caller's")
}

func TestClientErrorsAreNotRetried(t *testing.T) {
	base := &script{steps: []step{{status: 404}, {status: 200}}}
	tr, _ := newTestTransport(base)

	resp, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	assert.Len(t, base.calls, 1)
}

func TestBodyWithoutGetBodyIsBufferedWithoutMutatingTheCaller(t *testing.T) {
	base := &script{steps: []step{{status: 500}, {status: 200}}}
	tr, _ := newTestTransport(base)

	req, err := http.NewRequest("PUT", "https://api.example.test/x", io.NopCloser(strings.NewReader("streamed")))
	require.NoError(t, err)
	require.Nil(t, req.GetBody)

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	require.Len(t, base.calls, 2)
	assert.Equal(t, "streamed", base.calls[0].body)
	assert.Equal(t, "streamed", base.calls[1].body)
	assert.Nil(t, req.GetBody, "the caller's request is not mutated")
}

func TestOversizedBodyWithoutGetBodyIsSentOnce(t *testing.T) {
	base := &script{steps: []step{{status: 500}, {status: 200}}}
	tr, _ := newTestTransport(base, WithMaxBufferedBody(8))

	req, _ := http.NewRequest("PUT", "https://api.example.test/x", io.NopCloser(strings.NewReader("this body is longer than the buffer")))

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	require.Len(t, base.calls, 1)
	assert.Equal(t, "this body is longer than the buffer", base.calls[0].body, "no truncated body is ever sent")
}

func TestContextCancelStopsTheBackoff(t *testing.T) {
	base := &script{steps: []step{{status: 503}}}
	tr := NewTransport(base, WithBackoff(time.Hour, time.Hour), WithMaxElapsed(48*time.Hour))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.test/x", nil)

	_, err := tr.RoundTrip(req)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Len(t, base.calls, 1)
}

func TestCanceledContextErrorsAreNotRetried(t *testing.T) {
	base := &script{steps: []step{{err: context.Canceled}, {status: 200}}}
	tr, _ := newTestTransport(base)

	_, err := do(t, tr, "GET", "")
	assert.ErrorIs(t, err, context.Canceled)
	assert.Len(t, base.calls, 1)
}

func TestCustomPolicyOwnsTheDecision(t *testing.T) {
	base := &script{steps: []step{{status: 409}, {status: 200}}}
	tr, _ := newTestTransport(base, WithShouldRetry(func(_ *http.Request, resp *http.Response, _ error) bool {
		return resp != nil && resp.StatusCode == 409
	}))

	resp, err := do(t, tr, "POST", "x")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, base.calls, 2)
}

func TestSingleTryIsAPassThrough(t *testing.T) {
	base := &script{steps: []step{{status: 503}}}
	tr, _ := newTestTransport(base, WithMaxTries(1))

	resp, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, 503, resp.StatusCode)
	assert.Len(t, base.calls, 1)
}
