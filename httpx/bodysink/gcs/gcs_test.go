package gcs

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memWriter struct {
	bytes.Buffer
	name    string
	closeFn func(name string, body []byte)
	failOn  error
}

func (w *memWriter) Write(p []byte) (int, error) {
	if w.failOn != nil {
		return 0, w.failOn
	}

	return w.Buffer.Write(p)
}

func (w *memWriter) Close() error {
	w.closeFn(w.name, w.Bytes())

	return nil
}

func TestKeyAndRefLayout(t *testing.T) {
	var (
		mu      sync.Mutex
		objects = map[string]string{}
	)

	open := func(_ context.Context, name string) io.WriteCloser {
		return &memWriter{name: name, closeFn: func(name string, body []byte) {
			mu.Lock()
			defer mu.Unlock()

			objects[name] = string(body)
		}}
	}

	sink := newWithWriter(open, "bucket", "outgoing-http")

	key := "2026-04-21/pao/11111111-2222-4333-8444-555555555555/response"
	ref, err := sink.Put(context.Background(), key, []byte(`{"redacted":true}`))
	require.NoError(t, err)
	require.NoError(t, sink.Close(context.Background()))

	assert.Equal(t, "gs://bucket/outgoing-http/"+key, ref)
	assert.Equal(t, map[string]string{"outgoing-http/" + key: `{"redacted":true}`}, objects)
}

func TestWriteErrorIsReturnedByStore(t *testing.T) {
	boom := errors.New("boom")
	open := func(_ context.Context, name string) io.WriteCloser {
		return &memWriter{name: name, failOn: boom, closeFn: func(string, []byte) {}}
	}

	err := store{open: open, object: func(k string) string { return k }}.Store(context.Background(), "k", []byte("x"))
	assert.ErrorIs(t, err, boom)
}
