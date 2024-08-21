package logger

import (
	"context"
	"log/slog"

	"braces.dev/errtrace"
)

func CtxWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, logger)
}

func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerCtxKey{}).(*slog.Logger)
	if !ok {
		logger = NoopLogger()
	}

	return logger
}

type loggerCtxKey struct{}

// CtxProxyHandler is a slog.Handler which proxies all calls to
// the handler stored in the context.
//
// It is an alternative to the global Error, Warn, Info and Debug functions
type CtxProxyHandler struct {
	withAttrs  []slog.Attr
	withGroups []string
}

func (c CtxProxyHandler) getHandler(ctx context.Context) slog.Handler {
	handler := GetLoggerFromContext(ctx).Handler().WithAttrs(c.withAttrs)
	for _, group := range c.withGroups {
		handler = handler.WithGroup(group)
	}
	return handler
}

func (c CtxProxyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return c.getHandler(ctx).Enabled(ctx, level)
}

func (c CtxProxyHandler) Handle(ctx context.Context, record slog.Record) error {
	ctx = storeProgramCounter(ctx, 1)
	record.AddAttrs(getAttributesFromContext(ctx)...)
	return errtrace.Wrap(c.getHandler(ctx).Handle(ctx, record))
}

func (c CtxProxyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return CtxProxyHandler{
		withAttrs:  append(c.withAttrs, attrs...),
		withGroups: c.withGroups,
	}
}

func (c CtxProxyHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return c
	}
	return CtxProxyHandler{
		withAttrs:  c.withAttrs,
		withGroups: append(c.withGroups, name),
	}
}

var _ slog.Handler = (*CtxProxyHandler)(nil)
