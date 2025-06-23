package streamwork

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker(t *testing.T) {
	runtime.GOMAXPROCS(2)
	ctx := context.Background()

	cb := NewCircuitBreaker()

	output, err := Stream3(
		ctx, ReadSeq(
			func(yield func(int) bool) {
				for i := range 10 {
					if !yield(i) {
						return
					}
				}
			},
		), CircuitBreakerWorker[int](cb),
		WorkerFunc(
			func(ctx context.Context, v int) int {
				if v >= 5 {
					cb.Cut() // cutting the circuit breaker should close the stream and make the next yield return false
				}
				return v
			},
		),
	)
	require.NoError(t, err)
	require.Equal(t, []int{0, 1, 2, 3, 4, 5}, output)
}
