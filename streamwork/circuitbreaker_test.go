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

	output, err := Stream2(
		ctx, ReadSeq(
			func(yield func(int) bool) {
				for i := range 10 {
					if !yield(i) {
						return
					}
					runtime.Gosched() // allows the goroutines to receive all values already sent
					if i >= 5 {
						cb.Cut() // cutting the circuit breaker should close the stream and make the next yield return false
					}
					require.LessOrEqual(t, i, 5) // we should never get to the next value
				}
			},
		), CircuitBreakerWorker[int](cb),
	)
	require.NoError(t, err)
	require.Equal(t, []int{0, 1, 2, 3, 4, 5}, output)
}
