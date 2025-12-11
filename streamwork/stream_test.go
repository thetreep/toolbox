package streamwork

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStreamX(t *testing.T) {
	inputData := []int{1, 2, 3, 4, 5}
	incrementWorker := WorkerFunc(
		func(ctx context.Context, v int) int {
			return v + 1
		},
	)

	ctx := context.Background()

	// Testing Stream2
	output, err := Stream2(ctx, ReadSlice(inputData), incrementWorker)
	require.NoError(t, err)
	slices.Sort(output)
	require.Equal(t, []int{2, 3, 4, 5, 6}, output)

	// Testing Stream3
	output, err = Stream3(ctx, ReadSlice(inputData), incrementWorker, incrementWorker)
	require.NoError(t, err)
	slices.Sort(output)
	require.Equal(t, []int{3, 4, 5, 6, 7}, output)

	// Testing Stream4
	output, err = Stream4(ctx, ReadSlice(inputData), incrementWorker, incrementWorker, incrementWorker)
	require.NoError(t, err)
	slices.Sort(output)
	require.Equal(t, []int{4, 5, 6, 7, 8}, output)

	// Testing Stream5
	output, err = Stream5(ctx, ReadSlice(inputData), incrementWorker, incrementWorker, incrementWorker, incrementWorker)
	require.NoError(t, err)
	slices.Sort(output)
	require.Equal(t, []int{5, 6, 7, 8, 9}, output)

	// Testing Stream6
	output, err = Stream6(
		ctx,
		ReadSlice(inputData),
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
	)
	require.NoError(t, err)
	slices.Sort(output)
	require.Equal(t, []int{6, 7, 8, 9, 10}, output)

	// Testing Stream7
	output, err = Stream7(
		ctx,
		ReadSlice(inputData),
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
	)
	require.NoError(t, err)
	slices.Sort(output)
	require.Equal(t, []int{7, 8, 9, 10, 11}, output)

	// Testing Stream8
	output, err = Stream8(
		ctx,
		ReadSlice(inputData),
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
	)
	require.NoError(t, err)
	slices.Sort(output)
	require.Equal(t, []int{8, 9, 10, 11, 12}, output)

	// Testing Stream9
	output, err = Stream9(
		ctx,
		ReadSlice(inputData),
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
	)
	require.NoError(t, err)
	slices.Sort(output)
	require.Equal(t, []int{9, 10, 11, 12, 13}, output)

	// Testing Stream10
	output, err = Stream10(
		ctx,
		ReadSlice(inputData),
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
		incrementWorker,
	)
	require.NoError(t, err)
	slices.Sort(output)
	require.Equal(t, []int{10, 11, 12, 13, 14}, output)
}
