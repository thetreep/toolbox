package streamwork

import (
	"context"
)

// Flatten takes the batched output of a worker and flattens it for the next one
func Flatten[T any]() Worker[[]T, T] {
	return func(ctx context.Context, source <-chan []T, cfg streamConfig) <-chan T {
		chanOut := make(chan T, cfg.bufferSize)
		go func() {
			defer close(chanOut)
			for {
				if ctx.Err() != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case batch, ok := <-source:
					if !ok {
						return
					}
					for _, v := range batch {
						if ctx.Err() != nil {
							return
						}
						select {
						case <-ctx.Done():
							return
						case chanOut <- v:
						}
					}
				}
			}
		}()
		return chanOut
	}
}
