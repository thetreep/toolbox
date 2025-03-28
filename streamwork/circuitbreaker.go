package streamwork

import (
	"context"
	"sync"
)

// CircuitBreaker allows you to control the flow of a stream.
// Use NewCircuitBreaker to create one.
// Then add it to your stream with CircuitBreakerWorker.
// You can then call the Cut method to stop values from flowing down the stream once and for all.
type CircuitBreaker = *circuitBreaker

type circuitBreaker struct {
	cancelFns []context.CancelFunc
	canceled  bool
	lock      sync.Mutex
}

func (cb *circuitBreaker) Cut() {
	cb.lock.Lock()
	defer cb.lock.Unlock()
	cb.canceled = true
	for _, cancelFn := range cb.cancelFns {
		cancelFn()
	}
}

func NewCircuitBreaker() CircuitBreaker {
	return &circuitBreaker{}
}

func CircuitBreakerWorker[T any](cb CircuitBreaker) Worker[T, T] {
	return func(ctx context.Context, source <-chan T, cfg streamConfig) <-chan T {
		ctx, cancelFn := context.WithCancel(ctx)
		func() {
			cb.lock.Lock()
			defer cb.lock.Unlock()
			if cb.canceled {
				cancelFn()
			} else {
				cb.cancelFns = append(cb.cancelFns, cancelFn)
			}
		}()
		chanOut := make(chan T, cfg.bufferSize)
		go func() {
			defer close(chanOut)
			pipeChannels(ctx, source, chanOut)
		}()
		return chanOut
	}
}
