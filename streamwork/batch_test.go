package streamwork

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBatch(t *testing.T) {
	ctx := context.Background()

	output, err := Stream2(ctx, ReadSlice([]int{1, 2, 3, 4, 5}), Batch[int](2))
	require.NoError(t, err)
	require.Equal(t, [][]int{{1, 2}, {3, 4}, {5}}, output)

	ctxWithCancel, cancel := context.WithCancel(ctx)
	output, err = Stream2(
		ctxWithCancel, ReadSeq(
			func(yield func(int) bool) {
				for i := range 5 {
					if !yield(i + 1) {
						return
					}
				}
				runtime.Gosched() // allows the goroutines to receive all values already sent
				cancel()          // context cancellation should make the last batch incomplete
				yield(42)
			},
		), Batch[int](2),
	)
	require.NoError(t, err)
	require.Equal(t, [][]int{{1, 2}, {3, 4}, {5}}, output)

	output, err = Stream2(
		ctx, ReadSeq(
			func(yield func(int) bool) {
				for i := range 5 {
					if !yield(i + 1) {
						return
					}
				}
				time.Sleep(10 * time.Millisecond) // this should exceed the configured timeout and send the "5" alone
				yield(42)
				time.Sleep(10 * time.Millisecond) // this should exceed the configured timeout and send the "42" alone
				time.Sleep(10 * time.Millisecond) // this should exceed the configured timeout but do nothing as there is no more values
			},
		), Batch[int](2, BatchOptionMaxWait(5*time.Millisecond)),
	)
	require.NoError(t, err)
	require.Equal(t, [][]int{{1, 2}, {3, 4}, {5}, {42}}, output)
}
