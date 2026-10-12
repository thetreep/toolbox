// Package bodysink uploads redacted HTTP bodies asynchronously through a bounded worker pool.
package bodysink

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/thetreep/toolbox/logger"
)

var (
	// ErrQueueFull is returned by Put instead of blocking the HTTP call.
	ErrQueueFull = errors.New("bodysink: upload queue is full")
	// ErrClosed is returned by Put after Close.
	ErrClosed = errors.New("bodysink: closed")
)

// Store persists one object; implementations must be safe for concurrent use.
type Store interface {
	Store(ctx context.Context, key string, body []byte) error
}

// Option customizes New.
type Option func(*async)

// WithWorkers sets the number of concurrent uploads (default 4).
func WithWorkers(n int) Option { return func(a *async) { a.workers = max(n, 1) } }

// WithQueueSize sets how many uploads may wait for a worker (default 64).
func WithQueueSize(n int) Option { return func(a *async) { a.queueSize = max(n, 1) } }

// WithUploadTimeout bounds one upload (default 30s).
func WithUploadTimeout(d time.Duration) Option { return func(a *async) { a.timeout = d } }

type job struct {
	ctx  context.Context //nolint:containedctx // detached context carrying the caller's logger
	key  string
	body []byte
}

// Async implements httpx.BodySink: Put returns at once and the upload runs in the background; the pointer alias leaves no usable zero value.
type Async = *async

type async struct {
	store     Store
	ref       func(key string) string
	workers   int
	queueSize int
	timeout   time.Duration

	mu     sync.RWMutex
	closed bool
	queue  chan job
	wg     sync.WaitGroup
}

// New starts the workers; ref maps a key to the reference logged next to the preview.
func New(store Store, ref func(key string) string, opts ...Option) Async {
	a := &async{store: store, ref: ref, workers: 4, queueSize: 64, timeout: 30 * time.Second}

	for _, opt := range opts {
		opt(a)
	}

	a.queue = make(chan job, a.queueSize)

	for i := 0; i < a.workers; i++ {
		a.wg.Add(1)

		go a.work()
	}

	return a
}

// Put takes ownership of body, which the caller must not modify afterwards.
func (a *async) Put(ctx context.Context, key string, body []byte) (string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.closed {
		return "", ErrClosed
	}

	select {
	case a.queue <- job{ctx: context.WithoutCancel(ctx), key: key, body: body}:
		return a.ref(key), nil
	default:
		return "", ErrQueueFull
	}
}

// Close stops accepting bodies and waits for the queued uploads until ctx is done.
func (a *async) Close(ctx context.Context) error {
	a.mu.Lock()
	if !a.closed {
		a.closed = true
		close(a.queue)
	}
	a.mu.Unlock()

	drained := make(chan struct{})

	go func() {
		a.wg.Wait()
		close(drained)
	}()

	select {
	case <-drained:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *async) work() {
	defer a.wg.Done()

	for j := range a.queue {
		ctx, cancel := context.WithTimeout(j.ctx, a.timeout)

		if err := a.store.Store(ctx, j.key, j.body); err != nil {
			logger.Warn(ctx, "could not upload HTTP body",
				slog.String("error", err.Error()), slog.String("bodyKey", j.key))
		}

		cancel()
	}
}
