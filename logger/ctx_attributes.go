package logger

import (
	"context"
	"log/slog"
)

func CtxWithLogAttributes(ctx context.Context, attrs ...slog.Attr) context.Context {
	attributes, _ := ctx.Value(attributesCtxKey{}).(map[string]slog.Attr)
	if attributes == nil {
		attributes = make(map[string]slog.Attr)
	}

	for _, attr := range attrs {
		attributes[attr.Key] = attr
	}

	return context.WithValue(ctx, attributesCtxKey{}, attributes)
}

type attributesCtxKey struct{}

func getAttributesFromContext(ctx context.Context) []slog.Attr {
	attributes, _ := ctx.Value(attributesCtxKey{}).(map[string]slog.Attr)
	if attributes == nil {
		return nil
	}

	attributesSlice := make([]slog.Attr, 0, len(attributes))
	for _, attr := range attributes {
		attributesSlice = append(attributesSlice, attr)
	}

	return attributesSlice
}
