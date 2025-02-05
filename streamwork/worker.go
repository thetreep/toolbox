package streamwork

import (
	"context"
)

type Worker[TIN any, TOUT any] func(ctx context.Context, source <-chan TIN, cfg streamConfig) <-chan TOUT

// WorkerFunc defines a worker from a simple function which transforms a value into another with no possible error.
// See WorkerFuncErr for a variant with errors.
func WorkerFunc[TIN any, TOUT any](f func(ctx context.Context, v TIN) TOUT) Worker[TIN, TOUT] {
	return func(ctx context.Context, source <-chan TIN, cfg streamConfig) <-chan TOUT {
		chanOut := make(chan TOUT, cfg.bufferSize)
		go func() {
			defer close(chanOut)
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-source:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case chanOut <- f(ctx, v):
					}
				}
			}
		}()

		return chanOut
	}
}

// WorkerFuncErr defines a worker from a simple function which transforms a value into another with a possible error.
// See WorkerFunc for a variant without errors.
func WorkerFuncErr[TIN any, TOUT any](f func(ctx context.Context, v TIN) (TOUT, error)) Worker[TIN, TOUT] {
	return func(ctx context.Context, source <-chan TIN, cfg streamConfig) <-chan TOUT {
		chanOut := make(chan TOUT, cfg.bufferSize)
		go func() {
			defer close(chanOut)
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-source:
					if !ok {
						return
					}
					out, err := f(ctx, v)
					if err != nil {
						if !cfg.errHandler(ctx, err) {
							return
						}
						continue
					}
					select {
					case <-ctx.Done():
						return
					case chanOut <- out:
					}
				}
			}
		}()

		return chanOut
	}
}
