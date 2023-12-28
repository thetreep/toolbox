package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	t.Run("SuccessOnFirstAttempt", func(t *testing.T) {
		retryable := func() error {
			return nil
		}

		err := Retry(context.Background(), retryable)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("RetryAttempts", func(t *testing.T) {
		var attempts uint = 3
		retryable := func() error {
			return errors.New("temporary error")
		}

		err := Retry(context.Background(), retryable, RetryAttempts(attempts))
		if err == nil {
			t.Errorf("Expected an error, got nil")
		}

		assert.Error(t, err, ErrMaxAttemptsReached)
	})

	t.Run("MaxGlobalDelayReached", func(t *testing.T) {
		retryable := func() error {
			time.Sleep(600 * time.Millisecond)
			return errors.New("temporary error")
		}

		maxGlobalDelay := 500 * time.Millisecond

		err := Retry(context.Background(), retryable, RetryMaxDelay(maxGlobalDelay))
		if err == nil {
			t.Errorf("Expected an error, got nil")
		}

		assert.Error(t, err, ErrMaxDelayReached)
	})
}
