package logger

import "context"

// CtxWithErrorAttributesFunc sets a function whose result is added to the attributes of every error logged with this context.
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
