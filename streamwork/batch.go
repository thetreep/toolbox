package streamwork

import (
	"context"
	"time"
)

// Batch takes the output of a worker and groups it into batches of up-to `maxCount` values for bulk processing.
func Batch[T any](maxCount int, options ...BatchOption) Worker[T, []T] {
	if maxCount <= 1 {
		panic("maxCount must be > 1")
	}

	return func(ctx context.Context, source <-chan T, streamCfg streamConfig) <-chan []T {
		cfg := &batchConfig{
			maxCount:  maxCount,
			streamCfg: streamCfg,
		}
		for _, opt := range options {
			opt(cfg)
		}
		chanOut := make(chan []T, cfg.streamCfg.bufferSize)
		go func() {
			defer close(chanOut)
			batch := make([]T, 0, cfg.maxCount)

			sendBatch := func() {
				if len(batch) == 0 {
					return
				}
				select {
				case <-ctx.Done():
					return
				case chanOut <- batch:
				}
				batch = make([]T, 0, cfg.maxCount)
			}

			for {
				var timeoutChan <-chan time.Time
				if cfg.maxWait > 0 {
					timeoutChan = time.After(cfg.maxWait)
				}
				select {
				case <-ctx.Done():
					if len(batch) > 0 { // try sending the last batch but don't block on it
						select {
						case chanOut <- batch:
						default:
						}
					}
					return
				case <-timeoutChan:
					sendBatch()
				case v, ok := <-source:
					if !ok {
						sendBatch()
						return
					}
					batch = append(batch, v)
					if len(batch) >= cfg.maxCount {
						sendBatch()
					}
				}
			}
		}()
		return chanOut
	}
}

type batchConfig struct {
	maxCount  int
	maxWait   time.Duration
	streamCfg streamConfig
}

type BatchOption func(*batchConfig)

// BatchOptionMaxWait makes Batch wait at most the given time.Duration for a new value.
// If no new value is provided during this time, the current batch is sent even if incomplete.
func BatchOptionMaxWait(d time.Duration) BatchOption {
	return func(c *batchConfig) {
		c.maxWait = d
	}
}

func BatchOptionFromStreamOption(opt StreamOption) BatchOption {
	return func(c *batchConfig) {
		opt(&c.streamCfg)
	}
}
