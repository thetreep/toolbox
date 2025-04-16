package streamwork

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParallelize(t *testing.T) {
	inputData := []int{1, 2, 3, 4, 5}
	sleepyTime := time.Millisecond * 10

	incrementWorker := WorkerFunc(
		func(ctx context.Context, v int) int {
			time.Sleep(sleepyTime)
			return v + 1
		},
	)

	ctx := context.Background()

	// sequential
	start := time.Now()
	output, err := Stream2(ctx, ReadSlice(inputData), incrementWorker)
	require.NoError(t, err)
	duration := time.Since(start)
	require.GreaterOrEqual(t, duration, sleepyTime*time.Duration(len(inputData)))
	slices.Sort(output)
	require.Equal(t, []int{2, 3, 4, 5, 6}, output)

	// parallelized
	start = time.Now()
	output, err = Stream2(ctx, ReadSlice(inputData), Parallelize(len(inputData), incrementWorker))
	require.NoError(t, err)
	duration = time.Since(start)
	require.Less(t, duration, sleepyTime*2)
	slices.Sort(output)
	require.Equal(t, []int{2, 3, 4, 5, 6}, output)
}
