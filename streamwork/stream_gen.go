package streamwork

/* GENERATED FILE, DO NOT EDIT BY HAND */

import "context"

func Stream2[T1 any, T2 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	options ...StreamOption,
) ([]T2, error) {
	return stream(
		func(cfg streamConfig) <-chan T2 { return worker1(ctx, source(ctx, cfg), cfg) }, options...,
	)
}
func Stream3[T1 any, T2 any, T3 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	options ...StreamOption,
) ([]T3, error) {
	return stream(
		func(cfg streamConfig) <-chan T3 { return worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg) }, options...,
	)
}
func Stream4[T1 any, T2 any, T3 any, T4 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	options ...StreamOption,
) ([]T4, error) {
	return stream(
		func(cfg streamConfig) <-chan T4 {
			return worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream5[T1 any, T2 any, T3 any, T4 any, T5 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	options ...StreamOption,
) ([]T5, error) {
	return stream(
		func(cfg streamConfig) <-chan T5 {
			return worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream6[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	options ...StreamOption,
) ([]T6, error) {
	return stream(
		func(cfg streamConfig) <-chan T6 {
			return worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream7[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	options ...StreamOption,
) ([]T7, error) {
	return stream(
		func(cfg streamConfig) <-chan T7 {
			return worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream8[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	options ...StreamOption,
) ([]T8, error) {
	return stream(
		func(cfg streamConfig) <-chan T8 {
			return worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream9[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	options ...StreamOption,
) ([]T9, error) {
	return stream(
		func(cfg streamConfig) <-chan T9 {
			return worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream10[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	options ...StreamOption,
) ([]T10, error) {
	return stream(
		func(cfg streamConfig) <-chan T10 {
			return worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream11[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	options ...StreamOption,
) ([]T11, error) {
	return stream(
		func(cfg streamConfig) <-chan T11 {
			return worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream12[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	options ...StreamOption,
) ([]T12, error) {
	return stream(
		func(cfg streamConfig) <-chan T12 {
			return worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream13[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	options ...StreamOption,
) ([]T13, error) {
	return stream(
		func(cfg streamConfig) <-chan T13 {
			return worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream14[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	options ...StreamOption,
) ([]T14, error) {
	return stream(
		func(cfg streamConfig) <-chan T14 {
			return worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream15[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	options ...StreamOption,
) ([]T15, error) {
	return stream(
		func(cfg streamConfig) <-chan T15 {
			return worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream16[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	options ...StreamOption,
) ([]T16, error) {
	return stream(
		func(cfg streamConfig) <-chan T16 {
			return worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream17[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	options ...StreamOption,
) ([]T17, error) {
	return stream(
		func(cfg streamConfig) <-chan T17 {
			return worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream18[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	options ...StreamOption,
) ([]T18, error) {
	return stream(
		func(cfg streamConfig) <-chan T18 {
			return worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream19[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	options ...StreamOption,
) ([]T19, error) {
	return stream(
		func(cfg streamConfig) <-chan T19 {
			return worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream20[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	options ...StreamOption,
) ([]T20, error) {
	return stream(
		func(cfg streamConfig) <-chan T20 {
			return worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream21[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	options ...StreamOption,
) ([]T21, error) {
	return stream(
		func(cfg streamConfig) <-chan T21 {
			return worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream22[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	options ...StreamOption,
) ([]T22, error) {
	return stream(
		func(cfg streamConfig) <-chan T22 {
			return worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream23[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	options ...StreamOption,
) ([]T23, error) {
	return stream(
		func(cfg streamConfig) <-chan T23 {
			return worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream24[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	options ...StreamOption,
) ([]T24, error) {
	return stream(
		func(cfg streamConfig) <-chan T24 {
			return worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream25[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	options ...StreamOption,
) ([]T25, error) {
	return stream(
		func(cfg streamConfig) <-chan T25 {
			return worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream26[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	options ...StreamOption,
) ([]T26, error) {
	return stream(
		func(cfg streamConfig) <-chan T26 {
			return worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream27[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	options ...StreamOption,
) ([]T27, error) {
	return stream(
		func(cfg streamConfig) <-chan T27 {
			return worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream28[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	options ...StreamOption,
) ([]T28, error) {
	return stream(
		func(cfg streamConfig) <-chan T28 {
			return worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream29[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any, T29 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	worker28 Worker[T28, T29],
	options ...StreamOption,
) ([]T29, error) {
	return stream(
		func(cfg streamConfig) <-chan T29 {
			return worker28(ctx, worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream30[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any, T29 any, T30 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	worker28 Worker[T28, T29],
	worker29 Worker[T29, T30],
	options ...StreamOption,
) ([]T30, error) {
	return stream(
		func(cfg streamConfig) <-chan T30 {
			return worker29(ctx, worker28(ctx, worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream31[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any, T29 any, T30 any, T31 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	worker28 Worker[T28, T29],
	worker29 Worker[T29, T30],
	worker30 Worker[T30, T31],
	options ...StreamOption,
) ([]T31, error) {
	return stream(
		func(cfg streamConfig) <-chan T31 {
			return worker30(ctx, worker29(ctx, worker28(ctx, worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream32[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any, T29 any, T30 any, T31 any, T32 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	worker28 Worker[T28, T29],
	worker29 Worker[T29, T30],
	worker30 Worker[T30, T31],
	worker31 Worker[T31, T32],
	options ...StreamOption,
) ([]T32, error) {
	return stream(
		func(cfg streamConfig) <-chan T32 {
			return worker31(ctx, worker30(ctx, worker29(ctx, worker28(ctx, worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream33[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any, T29 any, T30 any, T31 any, T32 any, T33 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	worker28 Worker[T28, T29],
	worker29 Worker[T29, T30],
	worker30 Worker[T30, T31],
	worker31 Worker[T31, T32],
	worker32 Worker[T32, T33],
	options ...StreamOption,
) ([]T33, error) {
	return stream(
		func(cfg streamConfig) <-chan T33 {
			return worker32(ctx, worker31(ctx, worker30(ctx, worker29(ctx, worker28(ctx, worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream34[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any, T29 any, T30 any, T31 any, T32 any, T33 any, T34 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	worker28 Worker[T28, T29],
	worker29 Worker[T29, T30],
	worker30 Worker[T30, T31],
	worker31 Worker[T31, T32],
	worker32 Worker[T32, T33],
	worker33 Worker[T33, T34],
	options ...StreamOption,
) ([]T34, error) {
	return stream(
		func(cfg streamConfig) <-chan T34 {
			return worker33(ctx, worker32(ctx, worker31(ctx, worker30(ctx, worker29(ctx, worker28(ctx, worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream35[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any, T29 any, T30 any, T31 any, T32 any, T33 any, T34 any, T35 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	worker28 Worker[T28, T29],
	worker29 Worker[T29, T30],
	worker30 Worker[T30, T31],
	worker31 Worker[T31, T32],
	worker32 Worker[T32, T33],
	worker33 Worker[T33, T34],
	worker34 Worker[T34, T35],
	options ...StreamOption,
) ([]T35, error) {
	return stream(
		func(cfg streamConfig) <-chan T35 {
			return worker34(ctx, worker33(ctx, worker32(ctx, worker31(ctx, worker30(ctx, worker29(ctx, worker28(ctx, worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream36[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any, T29 any, T30 any, T31 any, T32 any, T33 any, T34 any, T35 any, T36 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	worker28 Worker[T28, T29],
	worker29 Worker[T29, T30],
	worker30 Worker[T30, T31],
	worker31 Worker[T31, T32],
	worker32 Worker[T32, T33],
	worker33 Worker[T33, T34],
	worker34 Worker[T34, T35],
	worker35 Worker[T35, T36],
	options ...StreamOption,
) ([]T36, error) {
	return stream(
		func(cfg streamConfig) <-chan T36 {
			return worker35(ctx, worker34(ctx, worker33(ctx, worker32(ctx, worker31(ctx, worker30(ctx, worker29(ctx, worker28(ctx, worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream37[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any, T29 any, T30 any, T31 any, T32 any, T33 any, T34 any, T35 any, T36 any, T37 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	worker28 Worker[T28, T29],
	worker29 Worker[T29, T30],
	worker30 Worker[T30, T31],
	worker31 Worker[T31, T32],
	worker32 Worker[T32, T33],
	worker33 Worker[T33, T34],
	worker34 Worker[T34, T35],
	worker35 Worker[T35, T36],
	worker36 Worker[T36, T37],
	options ...StreamOption,
) ([]T37, error) {
	return stream(
		func(cfg streamConfig) <-chan T37 {
			return worker36(ctx, worker35(ctx, worker34(ctx, worker33(ctx, worker32(ctx, worker31(ctx, worker30(ctx, worker29(ctx, worker28(ctx, worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream38[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any, T29 any, T30 any, T31 any, T32 any, T33 any, T34 any, T35 any, T36 any, T37 any, T38 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	worker28 Worker[T28, T29],
	worker29 Worker[T29, T30],
	worker30 Worker[T30, T31],
	worker31 Worker[T31, T32],
	worker32 Worker[T32, T33],
	worker33 Worker[T33, T34],
	worker34 Worker[T34, T35],
	worker35 Worker[T35, T36],
	worker36 Worker[T36, T37],
	worker37 Worker[T37, T38],
	options ...StreamOption,
) ([]T38, error) {
	return stream(
		func(cfg streamConfig) <-chan T38 {
			return worker37(ctx, worker36(ctx, worker35(ctx, worker34(ctx, worker33(ctx, worker32(ctx, worker31(ctx, worker30(ctx, worker29(ctx, worker28(ctx, worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream39[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any, T29 any, T30 any, T31 any, T32 any, T33 any, T34 any, T35 any, T36 any, T37 any, T38 any, T39 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	worker28 Worker[T28, T29],
	worker29 Worker[T29, T30],
	worker30 Worker[T30, T31],
	worker31 Worker[T31, T32],
	worker32 Worker[T32, T33],
	worker33 Worker[T33, T34],
	worker34 Worker[T34, T35],
	worker35 Worker[T35, T36],
	worker36 Worker[T36, T37],
	worker37 Worker[T37, T38],
	worker38 Worker[T38, T39],
	options ...StreamOption,
) ([]T39, error) {
	return stream(
		func(cfg streamConfig) <-chan T39 {
			return worker38(ctx, worker37(ctx, worker36(ctx, worker35(ctx, worker34(ctx, worker33(ctx, worker32(ctx, worker31(ctx, worker30(ctx, worker29(ctx, worker28(ctx, worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream40[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any, T20 any, T21 any, T22 any, T23 any, T24 any, T25 any, T26 any, T27 any, T28 any, T29 any, T30 any, T31 any, T32 any, T33 any, T34 any, T35 any, T36 any, T37 any, T38 any, T39 any, T40 any](
	ctx context.Context,
	source Source[T1],
	worker1 Worker[T1, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	worker7 Worker[T7, T8],
	worker8 Worker[T8, T9],
	worker9 Worker[T9, T10],
	worker10 Worker[T10, T11],
	worker11 Worker[T11, T12],
	worker12 Worker[T12, T13],
	worker13 Worker[T13, T14],
	worker14 Worker[T14, T15],
	worker15 Worker[T15, T16],
	worker16 Worker[T16, T17],
	worker17 Worker[T17, T18],
	worker18 Worker[T18, T19],
	worker19 Worker[T19, T20],
	worker20 Worker[T20, T21],
	worker21 Worker[T21, T22],
	worker22 Worker[T22, T23],
	worker23 Worker[T23, T24],
	worker24 Worker[T24, T25],
	worker25 Worker[T25, T26],
	worker26 Worker[T26, T27],
	worker27 Worker[T27, T28],
	worker28 Worker[T28, T29],
	worker29 Worker[T29, T30],
	worker30 Worker[T30, T31],
	worker31 Worker[T31, T32],
	worker32 Worker[T32, T33],
	worker33 Worker[T33, T34],
	worker34 Worker[T34, T35],
	worker35 Worker[T35, T36],
	worker36 Worker[T36, T37],
	worker37 Worker[T37, T38],
	worker38 Worker[T38, T39],
	worker39 Worker[T39, T40],
	options ...StreamOption,
) ([]T40, error) {
	return stream(
		func(cfg streamConfig) <-chan T40 {
			return worker39(ctx, worker38(ctx, worker37(ctx, worker36(ctx, worker35(ctx, worker34(ctx, worker33(ctx, worker32(ctx, worker31(ctx, worker30(ctx, worker29(ctx, worker28(ctx, worker27(ctx, worker26(ctx, worker25(ctx, worker24(ctx, worker23(ctx, worker22(ctx, worker21(ctx, worker20(ctx, worker19(ctx, worker18(ctx, worker17(ctx, worker16(ctx, worker15(ctx, worker14(ctx, worker13(ctx, worker12(ctx, worker11(ctx, worker10(ctx, worker9(ctx, worker8(ctx, worker7(ctx, worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker1(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
