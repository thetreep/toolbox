package threadpool_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thetreep/toolbox/tests"
	"github.com/thetreep/toolbox/threadpool"
)

func Test_Threadpool_Impl(t *testing.T) {
	tests.Setup(t, func(ctx context.Context) {
		requests := []string{"alpha", "beta", "gamma", "delta", "epsilon"}

		pool, err := threadpool.NewThreadPool(1)
		assert.NotNil(t, pool)
		assert.NoError(t, err)

		for _, r := range requests {
			r := r
			pool.Run(func() {
				fmt.Println("Handling request:", r)
			})
		}
		size := pool.QueueSize()
		assert.NotEqual(t, 0, size)

		pool.Run(func() {
			panic("panic handling")
		})

		pool.Stop()
	})
}
