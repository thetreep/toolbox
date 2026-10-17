package logger

import "context"

// CtxWithErrorAttributesFunc sets fn to run on each logged error, its attributes coming after the error's own.
func CtxWithErrorAttributesFunc(ctx context.Context, fn func(err error) []any) context.Context {
	return context.WithValue(ctx, errorAttributesFuncCtxKey{}, fn)
}

type errorAttributesFuncCtxKey struct{}

func getAttributesFromErrorFunc(ctx context.Context, err error) []any {
	fn, _ := ctx.Value(errorAttributesFuncCtxKey{}).(func(err error) []any)
	if fn == nil {
		return nil
	}

	return fn(err)
}
