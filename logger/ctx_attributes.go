package logger

import (
	"context"
	"log/slog"
)

func CtxWithLogAttributes(ctx context.Context, attrs ...any) context.Context {
	existingAttributes, _ := ctx.Value(
		attrsCtxKey{},
	).([]any)

	logAttrs := make(map[string]slog.Attr, 0)
	attributes := make([]any, 0, len(existingAttributes))

	for _, attr := range existingAttributes {
		if logAttr, ok := attr.(slog.Attr); ok {
			logAttrs[logAttr.Key] = logAttr

			continue
		}

		// save attributes that are not slog.Attr
		attributes = append(attributes, attr)
	}

	for _, attr := range attrs {
		if logAttr, ok := attr.(slog.Attr); ok {
			logAttrs[logAttr.Key] = logAttr

			continue
		}

		attributes = append(attributes, attr)
	}

	for _, v := range logAttrs {
		attributes = append(attributes, v)
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
