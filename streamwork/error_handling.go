package streamwork

import (
	"context"
	"errors"
	"sync"
)

type ErrHandler func(ctx context.Context, err error) bool

// AlwaysContinue changes the default error-handling behavior of a worker: instead of stopping on the first error, it will keep going,
// allowing all results and errors to be collected.
func AlwaysContinue[TIN any, TOUT any](w Worker[TIN, TOUT]) Worker[TIN, TOUT] {
	return func(ctx context.Context, source <-chan TIN, cfg streamConfig) <-chan TOUT {
		baseErrHandler := cfg.errHandler
		cfg.errHandler = func(ctx context.Context, err error) bool {
			_ = baseErrHandler(ctx, err) // whatever the original errHandler says, we continue
			return true
		}
		return w(ctx, source, cfg)
	}
}

// errCollector is a simple concurrency-safe ErrHandler and ErrGetter which stores all errors it is given and `errors.Join` them if necessary.
type errCollector struct {
	lock sync.Mutex
	errs []error
}

func (c *errCollector) Handle(_ context.Context, err error) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.errs = append(c.errs, err)
	return false // fail fast by default
}

func (c *errCollector) Err() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	switch len(c.errs) {
	case 0:
		return nil
	case 1:
		return c.errs[0]
	default:
		return errors.Join(c.errs...)
	}
}

type ErrGetter func() error
