package streamwork

import (
	"context"
	"iter"
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
				if ctx.Err() != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case v, ok := <-source:
					if !ok {
						return
					}
					if ctx.Err() != nil {
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

// WorkerSeq defines a worker from an iterating function which transforms a sequence of values into another with no possible error.
// See WorkerSeqErr for a variant with errors.
func WorkerSeq[TIN any, TOUT any](f func(ctx context.Context, v iter.Seq[TIN]) iter.Seq[TOUT]) Worker[TIN, TOUT] {
	return func(ctx context.Context, source <-chan TIN, cfg streamConfig) <-chan TOUT {
		chanOut := make(chan TOUT, cfg.bufferSize)
		seqOut := f(
			ctx, func(yield func(TIN) bool) {
				for {
					if ctx.Err() != nil {
						return
					}
					select {
					case <-ctx.Done():
						return
					case v, ok := <-source:
						if !ok {
							return
						}
						if !yield(v) {
							return
						}
					}
				}
			},
		)
		go func() {
			defer close(chanOut)
			for vOut := range seqOut {
				if ctx.Err() != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case chanOut <- vOut:
				}
			}
		}()

		return chanOut
	}
}

// WorkerSeqErr defines a worker from an iterating function which transforms a sequence of values into another with a possible error.
// See WorkerSeq for a variant without errors.
func WorkerSeqErr[TIN any, TOUT any](f func(ctx context.Context, v iter.Seq[TIN]) iter.Seq2[TOUT, error]) Worker[TIN, TOUT] {
	return func(ctx context.Context, source <-chan TIN, cfg streamConfig) <-chan TOUT {
		chanOut := make(chan TOUT, cfg.bufferSize)
		seqOut := f(
			ctx, func(yield func(TIN) bool) {
				for {
					if ctx.Err() != nil {
						return
					}
					select {
					case <-ctx.Done():
						return
					case v, ok := <-source:
						if !ok {
							return
						}
						if !yield(v) {
							return
						}
					}
				}
			},
		)
		go func() {
			defer close(chanOut)
			for vOut, err := range seqOut {
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
				case chanOut <- vOut:
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
				if ctx.Err() != nil {
					return
				}
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
					if ctx.Err() != nil {
						return
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
