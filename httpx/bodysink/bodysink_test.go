package bodysink_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thetreep/toolbox/httpx/bodysink"
	"github.com/thetreep/toolbox/logger"
)

type fakeStore struct {
	mu      sync.Mutex
	stored  map[string]string
	gate    chan struct{}
	failing bool
}

func (s *fakeStore) Store(ctx context.Context, key string, body []byte) error {
	if s.gate != nil {
		select {
		case <-s.gate:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.failing {
		return errors.New("gcs unavailable")
	}

	if s.stored == nil {
		s.stored = map[string]string{}
	}

	s.stored[key] = string(body)

	return nil
}

func (s *fakeStore) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.stored)
}

func ref(key string) string { return "mem://" + key }

func TestPutReturnsTheRefAndUploadsInBackground(t *testing.T) {
	store := &fakeStore{gate: make(chan struct{})}
	sink := bodysink.New(store, ref)

	got, err := sink.Put(context.Background(), "a/b", []byte("body"))
	require.NoError(t, err)
	assert.Equal(t, "mem://a/b", got)
	assert.Zero(t, store.count(), "Put does not wait for the upload")

	close(store.gate)
	require.NoError(t, sink.Close(context.Background()))
	assert.Equal(t, map[string]string{"a/b": "body"}, store.stored)
}

func TestQueueIsBoundedAndNeverBlocks(t *testing.T) {
	store := &fakeStore{gate: make(chan struct{})}
	sink := bodysink.New(store, ref, bodysink.WithWorkers(1), bodysink.WithQueueSize(2))

	var accepted, rejected int

	for i := 0; i < 20; i++ {
		_, err := sink.Put(context.Background(), "k", []byte("x"))

		switch {
		case err == nil:
			accepted++
		case errors.Is(err, bodysink.ErrQueueFull):
			rejected++
		default:
			t.Fatal(err)
		}
	}

	// One job held by the worker plus a queue of two.
	assert.LessOrEqual(t, accepted, 3)
	assert.Greater(t, rejected, 0)

	close(store.gate)
	require.NoError(t, sink.Close(context.Background()))
}

func TestCloseDrainsQueuedUploads(t *testing.T) {
	store := &fakeStore{}
	sink := bodysink.New(store, ref, bodysink.WithWorkers(2), bodysink.WithQueueSize(100))

	for i := 0; i < 50; i++ {
		_, err := sink.Put(context.Background(), string(rune('A'+i)), []byte("x"))
		require.NoError(t, err)
	}

	require.NoError(t, sink.Close(context.Background()))
	assert.Equal(t, 50, store.count())

	_, err := sink.Put(context.Background(), "late", nil)
	assert.ErrorIs(t, err, bodysink.ErrClosed)
	assert.NoError(t, sink.Close(context.Background()), "Close is idempotent")
}

func TestCloseHonorsItsContext(t *testing.T) {
	store := &fakeStore{gate: make(chan struct{})}
	sink := bodysink.New(store, ref, bodysink.WithWorkers(1))

	_, err := sink.Put(context.Background(), "stuck", []byte("x"))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	assert.ErrorIs(t, sink.Close(ctx), context.DeadlineExceeded)

	close(store.gate)
	require.NoError(t, sink.Close(context.Background()))
}

func TestUploadFailureIsLoggedAsWarning(t *testing.T) {
	var out bytes.Buffer

	ctx := logger.CtxWithLogger(context.Background(), slog.New(slog.NewJSONHandler(&out, nil)))
	sink := bodysink.New(&fakeStore{failing: true}, ref)

	_, err := sink.Put(ctx, "k", []byte("secret-free body"))
	require.NoError(t, err)
	require.NoError(t, sink.Close(context.Background()))

	assert.Contains(t, out.String(), `"level":"WARN"`)
	assert.Contains(t, out.String(), "gcs unavailable")
	assert.Contains(t, out.String(), `"bodyKey":"k"`)
	assert.False(t, strings.Contains(out.String(), "secret-free body"), "the body is never logged")
}

func TestUploadIsDetachedFromTheCallerContext(t *testing.T) {
	store := &fakeStore{gate: make(chan struct{})}
	sink := bodysink.New(store, ref)

	ctx, cancel := context.WithCancel(context.Background())

	_, err := sink.Put(ctx, "k", []byte("x"))
	require.NoError(t, err)
	cancel()

	close(store.gate)
	require.NoError(t, sink.Close(context.Background()))
	assert.Equal(t, 1, store.count(), "the HTTP call ending does not cancel the upload")
}

func TestConcurrentPutAndClose(t *testing.T) {
	sink := bodysink.New(&fakeStore{}, ref)

	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := 0; j < 200; j++ {
				_, _ = sink.Put(context.Background(), "k", []byte("x"))
			}
		}()
	}

	require.NoError(t, sink.Close(context.Background()))
	wg.Wait()
}
