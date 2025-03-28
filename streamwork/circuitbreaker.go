package streamwork

import (
	"context"
	"sync"
)

// CircuitBreaker allows you to control the flow of a stream.
// Use NewCircuitBreaker to create one.
// Then add it to your stream with CircuitBreakerWorker.
// You can then call the Cut method to stop values from flowing down the stream once and for all.
// Or the Pause and Resume methods to only stop them temporarily.
type CircuitBreaker = *circuitBreaker

type circuitBreaker struct {
	cancelFns   []context.CancelFunc
	canceled    bool
	paused      bool
	pauseChans  []chan struct{}
	resumeChans []chan struct{}
	lock        sync.Mutex
}

func (cb *circuitBreaker) Cut() {
	cb.lock.Lock()
	defer cb.lock.Unlock()
	cb.canceled = true
	for _, cancelFn := range cb.cancelFns {
		cancelFn()
	}
}

func (cb *circuitBreaker) IsPaused() bool {
	cb.lock.Lock()
	defer cb.lock.Unlock()
	return cb.paused
}

func (cb *circuitBreaker) Pause() {
	cb.lock.Lock()
	defer cb.lock.Unlock()
	cb.paused = true
	for _, pChan := range cb.pauseChans {
		pChan <- struct{}{}
		close(pChan)
	}
	cb.pauseChans = nil
}

func (cb *circuitBreaker) Resume() {
	cb.lock.Lock()
	defer cb.lock.Unlock()
	cb.paused = false
	for _, rChan := range cb.resumeChans {
		rChan <- struct{}{}
		close(rChan)
	}
	cb.resumeChans = nil
}

func (cb *circuitBreaker) pauseChan() <-chan struct{} {
	cb.lock.Lock()
	defer cb.lock.Unlock()
	pChan := make(chan struct{}, 1)
	if cb.paused {
		pChan <- struct{}{}
		return pChan
	}
	cb.pauseChans = append(cb.pauseChans, pChan)
	return pChan
}

func (cb *circuitBreaker) resumeChan() <-chan struct{} {
	cb.lock.Lock()
	defer cb.lock.Unlock()
	rChan := make(chan struct{}, 1)
	if !cb.paused {
		rChan <- struct{}{}
		return rChan
	}
	cb.resumeChans = append(cb.resumeChans, rChan)
	return rChan
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
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-source:
					if !ok {
						return
					}
					if cb.IsPaused() {
						select {
						case <-ctx.Done():
							return
						case <-cb.resumeChan():
						}
					}
					for {
						select {
						case <-ctx.Done():
							return
						case chanOut <- v:
							break
						case <-cb.pauseChan():
							select {
							case <-ctx.Done():
								return
							case <-cb.resumeChan():
							}
						}
					}
				}
			}
		}()
		return chanOut
	}
}
