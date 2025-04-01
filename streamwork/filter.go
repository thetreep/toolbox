package streamwork

import "context"

// Filter defines a worker which keeps or rejects its input values.
func Filter[TIN any](f func(ctx context.Context, v TIN) bool) Worker[TIN, TIN] {
	return func(ctx context.Context, source <-chan TIN, cfg streamConfig) <-chan TIN {
		chanOut := make(chan TIN, cfg.bufferSize)
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
					if !f(ctx, v) {
						continue
					}
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
		}()

		return chanOut
	}
}
