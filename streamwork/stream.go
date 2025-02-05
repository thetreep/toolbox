package streamwork

import (
	"context"
)

type streamConfig struct {
	continueOnError bool
	bufferSize      int
	errHandler      ErrHandler
}

type StreamOption func(*streamConfig)

func StreamOptionContinueOnError() StreamOption {
	return func(c *streamConfig) {
		c.continueOnError = true
	}
}

func StreamOptionBufferSize(size int) StreamOption {
	return func(c *streamConfig) {
		c.bufferSize = size
	}
}

type streamSink[T any] struct {
	c    <-chan T
	errG ErrGetter
}

func (s streamSink[T]) Collect() ([]T, error) {
	var result []T
	for v := range s.c {
		result = append(result, v)
	}
	return result, s.errG()
}

func stream[T any](fn func(cfg streamConfig) <-chan T, options ...StreamOption) ([]T, error) {
	cfg := streamConfig{}
	for _, opt := range options {
		opt(&cfg)
	}
	errC := &errCollector{}
	cfg.errHandler = errC.Handle
	return streamSink[T]{
		c:    fn(cfg),
		errG: errC.Err,
	}.Collect()
}

func Stream2[T any, T2 any](ctx context.Context, source Source[T], worker Worker[T, T2], options ...StreamOption) ([]T2, error) {
	return stream(
		func(cfg streamConfig) <-chan T2 {
			return worker(ctx, source(ctx, cfg), cfg)
		}, options...,
	)
}

func Stream3[T any, T2 any, T3 any](
	ctx context.Context,
	source Source[T],
	worker Worker[T, T2],
	worker2 Worker[T2, T3],
	options ...StreamOption,
) ([]T3, error) {
	return stream(
		func(cfg streamConfig) <-chan T3 {
			return worker2(ctx, worker(ctx, source(ctx, cfg), cfg), cfg)
		}, options...,
	)
}

func Stream4[T any, T2 any, T3 any, T4 any](
	ctx context.Context,
	source Source[T],
	worker Worker[T, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	options ...StreamOption,
) ([]T4, error) {
	return stream(
		func(cfg streamConfig) <-chan T4 {
			return worker3(ctx, worker2(ctx, worker(ctx, source(ctx, cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream5[T any, T2 any, T3 any, T4 any, T5 any](
	ctx context.Context,
	source Source[T],
	worker Worker[T, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	options ...StreamOption,
) ([]T5, error) {
	return stream(
		func(cfg streamConfig) <-chan T5 {
			return worker4(ctx, worker3(ctx, worker2(ctx, worker(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream6[T any, T2 any, T3 any, T4 any, T5 any, T6 any](
	ctx context.Context,
	source Source[T],
	worker Worker[T, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	options ...StreamOption,
) ([]T6, error) {
	return stream(
		func(cfg streamConfig) <-chan T6 {
			return worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg)
		}, options...,
	)
}
func Stream7[T any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any](
	ctx context.Context,
	source Source[T],
	worker Worker[T, T2],
	worker2 Worker[T2, T3],
	worker3 Worker[T3, T4],
	worker4 Worker[T4, T5],
	worker5 Worker[T5, T6],
	worker6 Worker[T6, T7],
	options ...StreamOption,
) ([]T7, error) {
	return stream(
		func(cfg streamConfig) <-chan T7 {
			return worker6(
				ctx,
				worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg),
				cfg,
			)
		}, options...,
	)
}
func Stream8[T any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any](
	ctx context.Context,
	source Source[T],
	worker Worker[T, T2],
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
			return worker7(
				ctx,
				worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg),
				cfg,
			)
		}, options...,
	)
}
func Stream9[T any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any](
	ctx context.Context,
	source Source[T],
	worker Worker[T, T2],
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
			return worker8(
				ctx,
				worker7(
					ctx,
					worker6(ctx, worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg), cfg),
					cfg,
				),
				cfg,
			)
		}, options...,
	)
}
func Stream10[T any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any](
	ctx context.Context,
	source Source[T],
	worker Worker[T, T2],
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
			return worker9(
				ctx,
				worker8(
					ctx,
					worker7(
						ctx,
						worker6(
							ctx,
							worker5(ctx, worker4(ctx, worker3(ctx, worker2(ctx, worker(ctx, source(ctx, cfg), cfg), cfg), cfg), cfg), cfg),
							cfg,
						),
						cfg,
					),
					cfg,
				),
				cfg,
			)
		}, options...,
	)
}
