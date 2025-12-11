package automapper

import "context"

type MapFuncWithErr[T any, U any] func(T) (U, error)

func (a MapFuncWithErr[T, U]) Map(v T) (U, error) {
	return a(v)
}

func (a MapFuncWithErr[T, U]) MapEach(v []T) ([]U, error) {
	var values []U //nolint:prealloc // need to keep nil value

	if a == nil || v == nil {
		return values, nil
	}

	for _, value := range v {
		to, err := a.Map(value)
		if err != nil {
			return nil, err
		}
		values = append(values, to)
	}
	return values, nil
}

type MapFuncWithCtxErr[T any, U any] func(context.Context, T) (U, error)

func (a MapFuncWithCtxErr[T, U]) Map(ctx context.Context, v T) (U, error) {
	return a(ctx, v)
}

func (a MapFuncWithCtxErr[T, U]) MapEach(ctx context.Context, v []T) ([]U, error) {
	var values []U //nolint:prealloc // need to keep nil value

	if a == nil || v == nil {
		return values, nil
	}

	for _, value := range v {
		to, err := a.Map(ctx, value)
		if err != nil {
			return nil, err
		}
		values = append(values, to)
	}
	return values, nil
}

type MapFuncWithCtx[T any, U any] func(context.Context, T) U

func (a MapFuncWithCtx[T, U]) Map(ctx context.Context, v T) U {
	return a(ctx, v)
}

func (a MapFuncWithCtx[T, U]) MapEach(ctx context.Context, v []T) []U {
	var values []U //nolint:prealloc // need to keep nil value

	if a == nil || v == nil {
		return values
	}

	for _, value := range v {
		values = append(values, a.Map(ctx, value))
	}
	return values
}
