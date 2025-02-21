package streamwork

import (
	"context"
	"sync"
)

// Parallelize takes a worker and spawns `instancesCount` instances of it, distributing the inputs between all and combining the outputs.
func Parallelize[TIN any, TOUT any](instancesCount int, worker Worker[TIN, TOUT]) Worker[TIN, TOUT] {
	return func(ctx context.Context, source <-chan TIN, cfg streamConfig) <-chan TOUT {
		chanOut := make(chan TOUT, instancesCount)
		ctx, cfg.errHandler = wrapErrHandleWithCancel(ctx, cfg.errHandler)
		go func() {
			defer close(chanOut)
			wg := sync.WaitGroup{}
			wg.Add(instancesCount)
			for range instancesCount {
				go func() {
					defer wg.Done()
					chanOutWorker := worker(ctx, source, cfg)
					pipeChannels(ctx, chanOutWorker, chanOut)
				}()
			}
			wg.Wait()
		}()

		return chanOut
	}
}

func pipeChannels[T any](ctx context.Context, src <-chan T, dest chan T) {
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-src:
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				return
			case dest <- v:
			}
		}
	}
}

func wrapErrHandleWithCancel(ctx context.Context, baseErrHandler ErrHandler) (context.Context, func(ctx context.Context, err error) bool) {
	ctx, cancelFn := context.WithCancelCause(ctx)
	errHandler := func(ctx context.Context, err error) bool {
		mustContinue := baseErrHandler(ctx, err)
		if !mustContinue {
			cancelFn(err)
		}
		return mustContinue
	}
	return ctx, errHandler
}
