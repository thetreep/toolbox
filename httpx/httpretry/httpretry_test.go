package httpretry

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
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
	tr.rateLimitJitter = func() time.Duration { return 0 }
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

func TestCustomPolicyDecidesWhichFailuresAreRetried(t *testing.T) {
	base := &script{steps: []step{{status: 409}, {status: 200}}}
	tr, _ := newTestTransport(base, WithShouldRetry(func(_ *http.Request, resp *http.Response, _ error) bool {
		return resp != nil && resp.StatusCode == 409
	}))

	resp, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, base.calls, 2)

	base = &script{steps: []step{{status: 404}, {status: 200}}}
	tr, _ = newTestTransport(base, WithShouldRetry(func(_ *http.Request, resp *http.Response, _ error) bool {
		return resp != nil && resp.StatusCode == 404
	}))

	resp, err = do(t, tr, "PATCH", "", "Idempotency-Key", "k")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, base.calls, 2)

	base = &script{steps: []step{{status: 409}, {status: 200}}}
	tr, _ = newTestTransport(base, WithShouldRetry(func(*http.Request, *http.Response, error) bool { return false }))

	resp, err = do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode, "a predicate can narrow the defaults")
	assert.Len(t, base.calls, 1)
}

func TestCustomPolicyCannotBypassTheIdempotencyGuard(t *testing.T) {
	always := WithShouldRetry(func(*http.Request, *http.Response, error) bool { return true })

	base := &script{steps: []step{{status: 409}, {status: 200}}}
	tr, slept := newTestTransport(base, always)

	resp, err := do(t, tr, "POST", `{"book":true}`)
	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
	assert.Len(t, base.calls, 1)
	assert.Empty(t, *slept)

	base = &script{steps: []step{{err: errors.New("connection reset")}, {status: 200}}}
	tr, _ = newTestTransport(base, always)

	_, err = do(t, tr, "POST", `{"book":true}`)
	assert.Error(t, err)
	assert.Len(t, base.calls, 1)

	base = &script{steps: []step{{status: 409}, {status: 200}}}
	tr, _ = newTestTransport(base, always)

	resp, err = do(t, tr, "POST", `{"book":true}`, "Idempotency-Key", "k")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode, "an Idempotency-Key makes the request eligible")
}

func isTokenRequest(req *http.Request) bool { return req.URL.Path == "/oauth/token" }

func TestDeclaredIdempotentPostIsRetriedWithItsBody(t *testing.T) {
	base := &script{steps: []step{{status: 503}, {status: 200}}}
	tr, _ := newTestTransport(base, WithIdempotent(isTokenRequest))

	req, err := http.NewRequest(http.MethodPost, "https://auth.example.test/oauth/token", io.NopCloser(strings.NewReader("grant_type=client_credentials")))
	require.NoError(t, err)
	require.Nil(t, req.GetBody)

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	require.Len(t, base.calls, 2)
	assert.Equal(t, "grant_type=client_credentials", base.calls[0].body)
	assert.Equal(t, "grant_type=client_credentials", base.calls[1].body)
}

func TestPostRejectedByTheDeclaredPredicateIsSentOnce(t *testing.T) {
	base := &script{steps: []step{{status: 503}, {status: 200}}}
	tr, slept := newTestTransport(base, WithIdempotent(isTokenRequest))

	resp, err := do(t, tr, "POST", `{"book":true}`)
	require.NoError(t, err)
	assert.Equal(t, 503, resp.StatusCode)
	assert.Len(t, base.calls, 1)
	assert.Empty(t, *slept)
}

func TestDeclaredIdempotentStillGoesThroughTheFailurePolicy(t *testing.T) {
	base := &script{steps: []step{{status: 409}, {status: 200}}}
	tr, _ := newTestTransport(base, WithIdempotent(func(*http.Request) bool { return true }))

	resp, err := do(t, tr, "POST", `{"book":true}`)
	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode, "WithIdempotent does not widen which failures are retried")
	assert.Len(t, base.calls, 1)
}

func TestDeclaredIdempotentIsNotAskedAboutIdempotentRequests(t *testing.T) {
	base := &script{steps: []step{{status: 503}, {status: 200}}}
	tr, _ := newTestTransport(base, WithIdempotent(func(*http.Request) bool {
		t.Error("the predicate must not be consulted for an already idempotent request")
		return false
	}))

	resp, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, base.calls, 2)
}

func TestNonIdempotentRequestSkipsPolicyAndBuffering(t *testing.T) {
	base := &script{steps: []step{{status: 503}}}
	tr, _ := newTestTransport(base, WithShouldRetry(func(*http.Request, *http.Response, error) bool {
		t.Error("the policy must not be consulted for a non-idempotent request")
		return true
	}))

	body := io.NopCloser(strings.NewReader(`{"book":true}`))
	req, _ := http.NewRequest(http.MethodPost, "https://api.example.test/x", body)
	req.GetBody = nil

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Len(t, base.calls, 1)
	assert.Same(t, req, resp.Request, "the original request is passed through untouched")
	assert.Equal(t, `{"book":true}`, base.calls[0].body)
}

func TestIsIdempotent(t *testing.T) {
	for method, want := range map[string]bool{"GET": true, "HEAD": true, "OPTIONS": true, "PUT": true, "DELETE": true, "POST": false, "PATCH": false} {
		req, _ := http.NewRequest(method, "https://api.example.test/x", nil)
		assert.Equal(t, want, IsIdempotent(req), method)
	}

	req, _ := http.NewRequest("PATCH", "https://api.example.test/x", nil)
	req.Header.Set("Idempotency-Key", "k")
	assert.True(t, IsIdempotent(req))
}

func TestSingleTryIsAPassThrough(t *testing.T) {
	base := &script{steps: []step{{status: 503}}}
	tr, _ := newTestTransport(base, WithMaxTries(1))

	resp, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, 503, resp.StatusCode)
	assert.Len(t, base.calls, 1)
}

func TestRateLimitResetWait(t *testing.T) {
	t.Parallel()

	clock := time.Date(2026, 4, 21, 8, 0, 0, 0, time.UTC)
	unixIn := func(d time.Duration) string { return strconv.FormatInt(clock.Add(d).Unix(), 10) }
	backoff := 200 * time.Millisecond

	tests := []struct {
		name    string
		status  int
		headers http.Header
		jitter  time.Duration
		want    time.Duration
	}{
		{name: "delta seconds", status: 429, headers: http.Header{"X-Ratelimit-Reset": {"2"}}, want: 2 * time.Second},
		{name: "fractional seconds", status: 429, headers: http.Header{"X-Ratelimit-Reset": {"1.5"}}, want: 1500 * time.Millisecond},
		{name: "IETF draft header", status: 429, headers: http.Header{"Ratelimit-Reset": {"3"}}, want: 3 * time.Second},
		{name: "X-Ratelimit-Reset wins over the draft header", status: 429, headers: http.Header{"X-Ratelimit-Reset": {"1"}, "Ratelimit-Reset": {"3"}}, want: time.Second},
		{name: "unreadable first header falls through", status: 429, headers: http.Header{"X-Ratelimit-Reset": {"soon"}, "Ratelimit-Reset": {"3"}}, want: 3 * time.Second},
		{name: "unix timestamp", status: 429, headers: http.Header{"X-Ratelimit-Reset": {unixIn(7 * time.Second)}}, want: 7 * time.Second},
		{name: "unix timestamp in the past", status: 429, headers: http.Header{"X-Ratelimit-Reset": {unixIn(-time.Minute)}}, want: backoff},
		{name: "jitter is added", status: 429, headers: http.Header{"X-Ratelimit-Reset": {"1"}}, jitter: 700 * time.Millisecond, want: 1700 * time.Millisecond},
		{name: "reset shorter than the backoff keeps the backoff", status: 429, headers: http.Header{"X-Ratelimit-Reset": {"0.05"}}, want: backoff},
		{name: "zero", status: 429, headers: http.Header{"X-Ratelimit-Reset": {"0"}}, jitter: time.Second, want: backoff},
		{name: "negative", status: 429, headers: http.Header{"X-Ratelimit-Reset": {"-4"}}, jitter: time.Second, want: backoff},
		{name: "unreadable", status: 429, headers: http.Header{"X-Ratelimit-Reset": {"soon"}}, jitter: time.Second, want: backoff},
		{name: "not a number", status: 429, headers: http.Header{"X-Ratelimit-Reset": {"NaN"}}, want: backoff},
		{name: "absent", status: 429, want: backoff},
		{name: "Retry-After wins", status: 429, headers: http.Header{"X-Ratelimit-Reset": {"1"}, "Retry-After": {"4"}}, jitter: time.Second, want: 4 * time.Second},
		{name: "only 429 is concerned", status: 503, headers: http.Header{"X-Ratelimit-Reset": {"2"}}, want: backoff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			base := &script{steps: []step{{status: tt.status, headers: tt.headers}, {status: 200}}}
			tr, slept := newTestTransport(base, WithRateLimitReset())
			tr.rateLimitJitter = func() time.Duration { return tt.jitter }

			resp, err := do(t, tr, "GET", "")
			require.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)
			assert.Equal(t, []time.Duration{tt.want}, *slept)
		})
	}
}

func TestRateLimitResetIsIgnoredByDefault(t *testing.T) {
	t.Parallel()

	base := &script{steps: []step{{status: 429, headers: http.Header{"X-Ratelimit-Reset": {"5"}}}, {status: 200}}}
	tr, slept := newTestTransport(base)

	_, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, []time.Duration{200 * time.Millisecond}, *slept)
}

func TestRateLimitJitterStaysWithinOneSecond(t *testing.T) {
	t.Parallel()

	tr := NewTransport(nil)

	var maxSeen time.Duration

	for range 1000 {
		jitter := tr.rateLimitJitter()
		require.GreaterOrEqual(t, jitter, time.Duration(0))
		require.LessOrEqual(t, jitter, time.Second)

		maxSeen = max(maxSeen, jitter)
	}

	assert.Greater(t, maxSeen, 500*time.Millisecond, "the jitter is spread, not constant")
}

func TestRateLimitResetBeyondMaxElapsedGivesUp(t *testing.T) {
	t.Parallel()

	base := &script{steps: []step{{status: 429, headers: http.Header{"X-Ratelimit-Reset": {"60"}}}, {status: 200}}}
	tr, slept := newTestTransport(base, WithRateLimitReset())

	resp, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode)
	assert.Len(t, base.calls, 1)
	assert.Empty(t, *slept)
}

func TestRateLimitResetPlusJitterCountsAgainstMaxElapsed(t *testing.T) {
	t.Parallel()

	base := &script{steps: []step{{status: 429, headers: http.Header{"X-Ratelimit-Reset": {"9.5"}}}, {status: 200}}}
	tr, _ := newTestTransport(base, WithRateLimitReset())
	tr.rateLimitJitter = func() time.Duration { return time.Second }

	resp, err := do(t, tr, "GET", "")
	require.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode, "9.5s + 1s exceeds the 10s budget")
	assert.Len(t, base.calls, 1)
}

func TestContextCancelStopsTheRateLimitWait(t *testing.T) {
	t.Parallel()

	base := &script{steps: []step{{status: 429, headers: http.Header{"X-Ratelimit-Reset": {"3600"}}}}}
	tr := NewTransport(base, WithRateLimitReset(), WithMaxElapsed(48*time.Hour))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.test/x", nil)

	_, err := tr.RoundTrip(req)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Len(t, base.calls, 1)
}

func TestIdempotentPaths(t *testing.T) {
	t.Parallel()

	matches := IdempotentPaths(http.MethodPost, "/search", "/v2/lookup")

	tests := []struct {
		method string
		path   string
		want   bool
	}{
		{http.MethodPost, "/search", true},
		{http.MethodPost, "/api/prefix/search", true},
		{http.MethodPost, "/api/v2/lookup", true},
		{http.MethodPost, "/searchX", false},
		{http.MethodPost, "/search/extra", false},
		{http.MethodPost, "/research", false},
		{http.MethodPost, "/lookup", false},
		{http.MethodPatch, "/search", false},
		{http.MethodGet, "/search", false},
		{http.MethodPost, "/", false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, matches(mustRequest(t, tt.method, tt.path)), "%s %s", tt.method, tt.path)
	}

	assert.False(t, IdempotentPaths(http.MethodPost)(mustRequest(t, http.MethodPost, "/search")), "no suffix matches nothing")
}

func mustRequest(t *testing.T, method, path string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, "https://api.example.test"+path, nil)
	require.NoError(t, err)

	return req
}

func TestIdempotentPathsDeclarePostsRetryableThroughTheTransport(t *testing.T) {
	t.Parallel()

	base := &script{steps: []step{{status: 503}, {status: 200}}}
	tr, _ := newTestTransport(base, WithIdempotent(IdempotentPaths(http.MethodPost, "/search")))

	req, err := http.NewRequest(http.MethodPost, "https://api.example.test/v1/search", io.NopCloser(strings.NewReader(`{"q":1}`)))
	require.NoError(t, err)

	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	require.Len(t, base.calls, 2)
	assert.Equal(t, `{"q":1}`, base.calls[1].body)

	base = &script{steps: []step{{status: 503}, {status: 200}}}
	tr, _ = newTestTransport(base, WithIdempotent(IdempotentPaths(http.MethodPost, "/search")))

	req, err = http.NewRequest(http.MethodPost, "https://api.example.test/v1/book", strings.NewReader(`{"q":1}`))
	require.NoError(t, err)

	resp, err = tr.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, 503, resp.StatusCode)
	assert.Len(t, base.calls, 1)
}
