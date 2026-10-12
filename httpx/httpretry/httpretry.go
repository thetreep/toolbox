// Package httpretry retries failed HTTP calls, only when repeating them cannot duplicate a side effect.
package httpretry

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultMaxTries      = 3
	defaultMaxElapsed    = 10 * time.Second
	defaultInitialDelay  = 200 * time.Millisecond
	defaultMaxDelay      = 2 * time.Second
	defaultMaxBufferSize = 10 << 20
	drainLimit           = 4 << 10
)

// ShouldRetry decides whether a finished attempt must be retried; exactly one of resp and err is non-nil.
type ShouldRetry func(req *http.Request, resp *http.Response, err error) bool

// Option customizes NewTransport.
type Option func(*Transport)

// WithMaxTries sets the total number of attempts, the first one included (default 3).
func WithMaxTries(n int) Option { return func(t *Transport) { t.maxTries = max(n, 1) } }

// WithMaxElapsed gives up once the next wait would exceed this duration since the first attempt (default 10s).
func WithMaxElapsed(d time.Duration) Option { return func(t *Transport) { t.maxElapsed = d } }

// WithBackoff sets the exponential wait between attempts (default 200ms doubling up to 2s, jittered).
func WithBackoff(initial, maxDelay time.Duration) Option {
	return func(t *Transport) { t.initialDelay, t.maxDelay = max(initial, 0), max(maxDelay, 0) }
}

// WithShouldRetry replaces DefaultShouldRetry, which also removes the idempotency guard.
func WithShouldRetry(fn ShouldRetry) Option {
	return func(t *Transport) { t.shouldRetry, t.customPolicy = fn, true }
}

// WithMaxBufferedBody bounds the body buffered to rewind requests without GetBody (default 10 MiB).
func WithMaxBufferedBody(bytes int) Option { return func(t *Transport) { t.maxBuffered = bytes } }

// Transport is an opt-in retrying http.RoundTripper, meant to wrap the logging transport so each try is logged.
type Transport struct {
	base         http.RoundTripper
	maxTries     int
	maxElapsed   time.Duration
	initialDelay time.Duration
	maxDelay     time.Duration
	maxBuffered  int
	shouldRetry  ShouldRetry
	customPolicy bool

	now    func() time.Time
	sleep  func(ctx context.Context, d time.Duration) error
	jitter func(d time.Duration) time.Duration
}

// NewTransport wraps base (http.DefaultTransport when nil).
func NewTransport(base http.RoundTripper, opts ...Option) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}

	t := &Transport{
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
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// DefaultShouldRetry retries network errors, 5xx and 429, but only idempotent methods or calls carrying an Idempotency-Key.
func DefaultShouldRetry(req *http.Request, resp *http.Response, err error) bool {
	if !isIdempotent(req) {
		return false
	}

	if err != nil {
		return retryableError(req.Context(), err)
	}

	return resp.StatusCode >= http.StatusInternalServerError || resp.StatusCode == http.StatusTooManyRequests
}

func isIdempotent(req *http.Request) bool {
	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	}

	return req.Header.Get("Idempotency-Key") != ""
}

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
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.maxTries == 1 || !t.mayRetry(req) {
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
			if wait, ok := retryAfter(resp, t.now()); ok {
				delay = max(delay, wait)
			}
		}

		if t.now().Sub(start)+delay > t.maxElapsed {
			return resp, roundTripErr
		}

		if resp != nil {
			drainAndClose(resp.Body)
		}

		if sleepErr := t.sleep(req.Context(), delay); sleepErr != nil {
			return nil, sleepErr
		}
	}
}

// mayRetry skips the body handling for calls that the default policy would never retry.
func (t *Transport) mayRetry(req *http.Request) bool {
	return t.customPolicy || isIdempotent(req)
}

// rewindable returns req, or a clone with a replayable body when the original cannot be re-read.
func (t *Transport) rewindable(req *http.Request) (*http.Request, error) {
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

func (t *Transport) backoff(attempt int) time.Duration {
	d := t.initialDelay

	for i := 1; i < attempt && d < t.maxDelay; i++ {
		d *= 2
	}

	return t.jitter(min(d, t.maxDelay))
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
