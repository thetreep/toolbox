package streamwork

import (
	"context"
	"iter"
)

type Source[T any] func(ctx context.Context, cfg streamConfig) <-chan T

// ReadSlice provides a Source from any slice, to be used with the Stream2, Stream3, etc. functions.
func ReadSlice[T any](s []T) Source[T] {
	return func(ctx context.Context, cfg streamConfig) <-chan T {
		c := make(chan T) // we don't need to buffer here, reading a slice is basically instant
		go func() {
			defer close(c)
			for _, v := range s {
				if ctx.Err() != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case c <- v:
				}
			}
		}()
		return c
	}
}

type readSeqConfig struct {
	bufferSize int
}

type ReadSeqOption func(*readSeqConfig)

func ReadSeqOptionBufferSize(bufferSize int) ReadSeqOption {
	return func(cfg *readSeqConfig) {
		cfg.bufferSize = bufferSize
	}
}

// ReadSeq provides a Source from an iter.Seq, to be used with the Stream2, Stream3, etc. functions.
// You can configure a buffer size like this `ReadSeq(iter, ReadSeqOptionBufferSize(1))`
func ReadSeq[T any](s iter.Seq[T], options ...StreamOption) Source[T] {
	return func(ctx context.Context, cfg streamConfig) <-chan T {
		for _, option := range options {
			option(&cfg)
		}
		c := make(chan T, cfg.bufferSize)
		go func() {
			defer close(c)
			for v := range s {
				if ctx.Err() != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case c <- v:
				}
			}
		}()
		return c
	}
}

// ReadSeqErr provides a Source from an iter.Seq2 which can provides errors, to be used with the Stream2, Stream3, etc. functions.
// You can configure a buffer size like this `ReadSeq(iter, ReadSeqOptionBufferSize(1))`
func ReadSeqErr[T any](s iter.Seq2[T, error], options ...StreamOption) Source[T] {
	return func(ctx context.Context, cfg streamConfig) <-chan T {
		for _, option := range options {
			option(&cfg)
		}
		c := make(chan T, cfg.bufferSize)
		go func() {
			defer close(c)
			for v, err := range s {
				if err != nil {
					if !cfg.errHandler(ctx, err) {
						return
					}
					continue
				}
				if ctx.Err() != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case c <- v:
				}
			}
		}()
		return c
	}
}
