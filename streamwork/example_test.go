package streamwork

import (
	"context"
	"fmt"
	"iter"
	"slices"
	"strconv"
	"strings"
)

func Example() {
	ctx := context.Background()
	// let's define a stream which goes string -> []string -> string -> int -> string
	result, err := Stream5(
		// StreamX: exists from 2 to 10, depending on how many different jobs you want to chain
		ctx,
		// first a data source: see ReadSlice, ReadSeq, and ReadSeqErr…
		ReadSlice([]string{"1", "2", "3", "NaN", "5"}),
		// …and then you list your workers
		// you can batch values at any point…
		Batch[string](2),
		// …and flatten them back later
		Flatten[string](),
		// by default, workers "fail fast": they stop on the first error, but you can override this behavior with AlwaysContinue
		AlwaysContinue(
			// parallelize any job with a fixed-size worker pool
			Parallelize(
				2, WorkerFuncErr(
					// define a worker from a func which transforms its input into some output and an error
					func(ctx context.Context, v string) (int, error) {
						return strconv.Atoi(v)
					},
				),
			),
		),
		WorkerSeq(
			// define a worker from a func which transforms an iter.Seq into another one
			func(ctx context.Context, vs iter.Seq[int]) iter.Seq[string] {
				return func(yield func(string) bool) {
					for v := range vs {
						if !yield(strconv.Itoa(v * 2)) {
							return
						}
					}
				}
			},
		),
	)
	if err != nil {
		fmt.Println(err.Error())
	}
	slices.Sort(result) // sorting so the output is stable
	fmt.Println(strings.Join(result, "\n"))

	// Output:
	// strconv.Atoi: parsing "NaN": invalid syntax
	// 10
	// 2
	// 4
	// 6
}
