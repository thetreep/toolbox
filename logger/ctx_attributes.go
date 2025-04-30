package logger

import (
	"context"
	"slices"
)

func CtxWithLogAttributes(ctx context.Context, attrs ...any) context.Context {
	var attributes []any

	if existingAttributes, _ := ctx.Value(
		attrsCtxKey{},
	).([]any); existingAttributes != nil {
		attributes = slices.Clone(existingAttributes)
	} else {
		attributes = make([]any, 0)
	}

	for _, attr := range attrs {
		attributes = append(attributes, attr)
	}

	return context.WithValue(ctx, attrsCtxKey{}, attributes)
}

type attrsCtxKey struct{}

func getAttributesFromContext(ctx context.Context) []any {
	attributes, _ := ctx.Value(attrsCtxKey{}).([]any)
	if attributes == nil {
		return nil
	}

	return attributes
}
