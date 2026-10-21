// Package httpretry retries failed HTTP calls, only when repeating them cannot duplicate a side effect.
package httpretry

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	defaultMaxTries      = 3
	defaultMaxElapsed    = 10 * time.Second
	defaultInitialDelay  = 200 * time.Millisecond
	defaultMaxDelay      = 2 * time.Second
	defaultMaxBufferSize = 10 << 20
	drainLimit           = 4 << 10
	maxRateLimitJitter   = time.Second
	// Rate-limit reset values from this magnitude up are Unix timestamps, not delays.
	unixTimestampFloor = 1e9
)

// ShouldRetry decides whether a finished attempt must be retried; exactly one of resp and err is non-nil.
type ShouldRetry func(req *http.Request, resp *http.Response, err error) bool

// Option customizes NewTransport.
type Option func(*transport)

// WithMaxTries sets the total number of attempts, the first one included (default 3).
func WithMaxTries(n int) Option { return func(t *transport) { t.maxTries = max(n, 1) } }

// WithMaxElapsed gives up once the next wait would exceed this duration since the first attempt (default 10s).
func WithMaxElapsed(d time.Duration) Option { return func(t *transport) { t.maxElapsed = d } }

// WithBackoff sets the exponential wait between attempts (default 200ms doubling up to 2s, jittered).
func WithBackoff(initial, maxDelay time.Duration) Option {
	return func(t *transport) { t.initialDelay, t.maxDelay = max(initial, 0), max(maxDelay, 0) }
}

// WithShouldRetry replaces DefaultShouldRetry; it is only asked about idempotent requests (see IsIdempotent and WithIdempotent).
func WithShouldRetry(fn ShouldRetry) Option {
	return func(t *transport) { t.shouldRetry = fn }
}

// WithIdempotent declares the requests fn accepts safe to replay on top of IsIdempotent (e.g. an OAuth token POST); fn must not read the body.
func WithIdempotent(fn func(*http.Request) bool) Option {
	return func(t *transport) { t.idempotent = fn }
}

// WithRateLimitReset makes a 429 without Retry-After wait for X-Ratelimit-Reset, else RateLimit-Reset, plus up to 1s of jitter.
// A value below 1e9 is delta-seconds, anything above a Unix timestamp; zero, negative or unreadable values are ignored.
func WithRateLimitReset() Option { return func(t *transport) { t.rateLimitReset = true } }

// WithMaxBufferedBody bounds the body buffered to rewind requests without GetBody (default 10 MiB).
func WithMaxBufferedBody(bytes int) Option { return func(t *transport) { t.maxBuffered = bytes } }

// IdempotentPaths reports requests using method whose URL path ends with one of suffixes, for WithIdempotent.
func IdempotentPaths(method string, suffixes ...string) func(*http.Request) bool {
	return func(req *http.Request) bool {
		return req.Method == method && slices.ContainsFunc(suffixes, func(suffix string) bool {
			return strings.HasSuffix(req.URL.Path, suffix)
		})
	}
}

// Transport is an opt-in retrying http.RoundTripper, meant to wrap the logging transport; the pointer alias leaves no usable zero value.
type Transport = *transport

type transport struct {
	base         http.RoundTripper
	maxTries     int
	maxElapsed   time.Duration
	initialDelay time.Duration
	maxDelay     time.Duration
	maxBuffered  int
	shouldRetry  ShouldRetry
	idempotent   func(*http.Request) bool

	rateLimitReset bool

	now             func() time.Time
	sleep           func(ctx context.Context, d time.Duration) error
	jitter          func(d time.Duration) time.Duration
	rateLimitJitter func() time.Duration
}

// NewTransport wraps base (http.DefaultTransport when nil).
func NewTransport(base http.RoundTripper, opts ...Option) Transport {
	if base == nil {
		base = http.DefaultTransport
	}

	t := &transport{
		base:         base,
		maxTries:     defaultMaxTries,
		maxElapsed:   defaultMaxElapsed,
		initialDelay: defaultInitialDelay,
		maxDelay:     defaultMaxDelay,
		maxBuffered:  defaultMaxBufferSize,
		shouldRetry:  DefaultShouldRetry,
		now:          time.Now,
		sleep:        sleepContext,
		jitter:       func(d time.Duration) time.Duration { return d/2 + rand.N(d/2+1) },

		rateLimitJitter: func() time.Duration { return rand.N(maxRateLimitJitter + 1) },
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// DefaultShouldRetry retries network errors, 5xx and 429; idempotency is checked by Transport, not here.
func DefaultShouldRetry(req *http.Request, resp *http.Response, err error) bool {
	if err != nil {
		return retryableError(req.Context(), err)
	}

	return resp.StatusCode >= http.StatusInternalServerError || resp.StatusCode == http.StatusTooManyRequests
}

// IsIdempotent reports whether repeating req is safe: an idempotent method or an Idempotency-Key header.
func IsIdempotent(req *http.Request) bool {
	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	}

	return req.Header.Get("Idempotency-Key") != ""
}

// FIXME: every non-TLS transport error is retried, it should be restricted to transient ones.
func retryableError(ctx context.Context, err error) bool {
	if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var (
		unknownAuthority x509.UnknownAuthorityError
		certificate      *tls.CertificateVerificationError
	)

	return !errors.As(err, &unknownAuthority) && !errors.As(err, &certificate)
}

// RoundTrip implements http.RoundTripper.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.maxTries == 1 || !t.replayable(req) {
		return t.base.RoundTrip(req)
	}

	req, err := t.rewindable(req)
	if err != nil {
		return nil, err
	}

	if req.Body != nil && req.Body != http.NoBody && req.GetBody == nil {
		return t.base.RoundTrip(req)
	}

	start := t.now()

	for attempt := 1; ; attempt++ {
		attemptReq := req
		if attempt > 1 && req.GetBody != nil {
			attemptReq = req.Clone(req.Context())

			if attemptReq.Body, err = req.GetBody(); err != nil {
				return nil, err
			}
		}

		resp, roundTripErr := t.base.RoundTrip(attemptReq)
		if attempt >= t.maxTries || !t.shouldRetry(attemptReq, resp, roundTripErr) {
			return resp, roundTripErr
		}

		delay := t.backoff(attempt)

		if resp != nil {
			if wait, ok := t.serverAdvisedWait(resp); ok {
				delay = max(delay, wait)
			}
		}

		if t.now().Sub(start)+delay > t.maxElapsed {
			return resp, roundTripErr
		}

		if resp != nil {
			drainAndClose(resp.Body)
		}

		// FIXME: cancelling during the backoff hides the last real error behind the context error.
		if sleepErr := t.sleep(req.Context(), delay); sleepErr != nil {
			return nil, sleepErr
		}
	}
}

func (t *transport) replayable(req *http.Request) bool {
	return IsIdempotent(req) || (t.idempotent != nil && t.idempotent(req))
}

// rewindable returns req, or a clone with a replayable body when the original cannot be re-read.
func (t *transport) rewindable(req *http.Request) (*http.Request, error) {
	if req.Body == nil || req.Body == http.NoBody || req.GetBody != nil {
		return req, nil
	}

	buffered, err := io.ReadAll(io.LimitReader(req.Body, int64(t.maxBuffered)+1))
	if err != nil {
		_ = req.Body.Close()

		return nil, err
	}

	clone := req.Clone(req.Context())

	if len(buffered) > t.maxBuffered {
		// Too large to hold: send it once, as-is.
		clone.Body = &joinedBody{Reader: io.MultiReader(bytes.NewReader(buffered), req.Body), closer: req.Body}
		clone.GetBody = nil

		return clone, nil
	}

	_ = req.Body.Close()

	clone.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(buffered)), nil }
	clone.Body, _ = clone.GetBody()

	return clone, nil
}

type joinedBody struct {
	io.Reader
	closer io.Closer
}

func (b *joinedBody) Close() error { return b.closer.Close() }

func (t *transport) backoff(attempt int) time.Duration {
	d := t.initialDelay

	for i := 1; i < attempt && d < t.maxDelay; i++ {
		d *= 2
	}

	return t.jitter(min(d, t.maxDelay))
}

func (t *transport) serverAdvisedWait(resp *http.Response) (time.Duration, bool) {
	if wait, ok := retryAfter(resp, t.now()); ok {
		return wait, true
	}

	if !t.rateLimitReset || resp.StatusCode != http.StatusTooManyRequests {
		return 0, false
	}

	wait, ok := rateLimitReset(resp, t.now())
	if !ok {
		return 0, false
	}

	// Spreads concurrent callers that were all told the same reset, so they don't re-burst together.
	return wait + t.rateLimitJitter(), true
}

func rateLimitReset(resp *http.Response, now time.Time) (time.Duration, bool) {
	for _, name := range []string{"X-Ratelimit-Reset", "RateLimit-Reset"} {
		value, err := strconv.ParseFloat(resp.Header.Get(name), 64)
		if err != nil || math.IsNaN(value) || value <= 0 {
			continue
		}

		if value >= unixTimestampFloor {
			return max(time.Unix(int64(min(value, 1<<40)), 0).Sub(now), 0), true
		}

		return time.Duration(value * float64(time.Second)), true
	}

	return 0, false
}

func retryAfter(resp *http.Response, now time.Time) (time.Duration, bool) {
	value := resp.Header.Get("Retry-After")
	if value == "" {
		return 0, false
	}

	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second, true
	}

	if when, err := http.ParseTime(value); err == nil {
		return max(when.Sub(now), 0), true
	}

	return 0, false
}

func drainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}

	_, _ = io.CopyN(io.Discard, body, drainLimit)
	_ = body.Close()
}

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
