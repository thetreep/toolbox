package retry

import (
	"context"
	"errors"
	"time"
)

type Retryable func() error

func Retry(ctx context.Context, retryable Retryable, retryOptions ...RetryOption) error {
	config := NewDefaultRetryConfig()
	for _, opt := range retryOptions {
		opt(config)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	globalTimer := time.After(config.maxGlobalDelay)
	errs := []error{}

	for i := config.attempts; i > 0; i-- {
		err := retryable()
		if err == nil {
			return nil
		}

		errs = append(errs, err)

		select {
		case <-time.After(config.waitBetween):
		case <-globalTimer:
			errs = append(errs, ErrMaxDelayReached)
			return errors.Join(errs...)
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	errs = append(errs, ErrMaxAttemptsReached)

	return errors.Join(errs...)
}

type RetryConfig struct {
	attempts       uint
	waitBetween    time.Duration
	maxGlobalDelay time.Duration
}

func NewDefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		attempts:       1,
		waitBetween:    200 * time.Millisecond,
		maxGlobalDelay: 5 * time.Second,
	}
}

type RetryOption func(*RetryConfig)

var ErrMaxAttemptsReached = errors.New("max attempts reached")

func RetryAttempts(attempts uint) RetryOption {
	return func(c *RetryConfig) {
		if attempts == 0 {
			attempts++
		}

		c.attempts = attempts
	}
}

var ErrMaxDelayReached = errors.New("max global delay reached")

func RetryMaxDelay(delay time.Duration) RetryOption {
	return func(c *RetryConfig) {
		if delay > 0 {
			c.maxGlobalDelay = delay
		}
	}
}

func RetryDelayBetween(waitBetween time.Duration) RetryOption {
	return func(c *RetryConfig) {
		if waitBetween > 0 {
			c.waitBetween = waitBetween
		}
	}
}
