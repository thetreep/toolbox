// Package gcs stores HTTP bodies in a Google Cloud Storage bucket.
package gcs

import (
	"context"
	"io"
	"path"

	"cloud.google.com/go/storage"

	"github.com/thetreep/toolbox/httpx/bodysink"
)

// New returns an asynchronous sink writing to gs://bucket/prefix/<key>; call Close on shutdown to flush it.
func New(client *storage.Client, bucket, prefix string, opts ...bodysink.Option) *bodysink.Async {
	open := func(ctx context.Context, name string) io.WriteCloser {
		return client.Bucket(bucket).Object(name).NewWriter(ctx)
	}

	return newWithWriter(open, bucket, prefix, opts...)
}

type writerFunc func(ctx context.Context, name string) io.WriteCloser

func newWithWriter(open writerFunc, bucket, prefix string, opts ...bodysink.Option) *bodysink.Async {
	object := func(key string) string { return path.Join(prefix, key) }
	ref := func(key string) string { return "gs://" + bucket + "/" + object(key) }

	return bodysink.New(store{open: open, object: object}, ref, opts...)
}

type store struct {
	open   writerFunc
	object func(key string) string
}

func (s store) Store(ctx context.Context, key string, body []byte) error {
	w := s.open(ctx, s.object(key))

	if _, err := w.Write(body); err != nil {
		_ = w.Close()

		return err
	}

	return w.Close()
}
