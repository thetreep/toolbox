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
			_ = baseErrHandler(ctx, WrapErrNotFatal(err)) // whatever the original errHandler says, we continue
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
	if errors.Is(err, ErrNoValue) {
		return true
	}
	mustContinue := false // fail fast by default
	if errors.Is(err, ErrNotFatal) {
		mustContinue = true
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.errs = append(c.errs, err)
	return mustContinue
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

func WrapErrNotFatal(err error) error {
	if err == nil {
		return nil
	}
	return errNotFatal{err}
}

func WrapErrNotFatal2[T any](v T, err error) (T, error) {
	if err == nil {
		return v, nil
	}
	return v, errNotFatal{err}
}

type constantError string

func (e constantError) Error() string {
	return string(e)
}

const ErrNotFatal = constantError("not fatal for streamwork")

type errNotFatal struct {
	err error
}

func (e errNotFatal) Is(err error) bool {
	return err == ErrNotFatal
}

func (e errNotFatal) Error() string {
	return e.err.Error()
}

func (e errNotFatal) Unwrap() error {
	return e.err
}

const ErrNoValue = constantError("no value to stream, but not an error")

func ExtractNonFatalErrs(err error) ([]error, error) {
	joinedErr, isJoined := err.(interface{ Unwrap() []error })
	if isJoined {
		var nonFatalErrs []error
		var fatalErr error
		for _, e := range joinedErr.Unwrap() {
			nf, f := ExtractNonFatalErrs(e)
			nonFatalErrs = append(nonFatalErrs, nf...)
			if f != nil {
				fatalErr = f
			}
		}
		return nonFatalErrs, fatalErr
	} else {
		wrappedErr, isWrapped := err.(interface{ Unwrap() error })
		if !isWrapped {
			if errors.Is(err, ErrNotFatal) {
				return []error{err}, nil
			} else {
				return nil, err
			}
		}
		return ExtractNonFatalErrs(wrappedErr.Unwrap())
	}
}
