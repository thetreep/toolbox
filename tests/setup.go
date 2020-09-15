package tests

import (
	"context"
	"testing"
	"time"
)

const timeout = 4 * time.Minute

// Setup sets the context for testing.
func Setup(t *testing.T, callback func(context.Context)) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	callback(ctx)
}
