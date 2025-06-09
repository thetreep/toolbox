package logger

import (
	"context"
	"log/slog"
)

// ctxLogHandler does two things:
// - it allows overriding the log level via the context.
type ctxLogHandler struct {
	slog.Handler
}

func newCtxLogHandler(wrappedHandler slog.Handler) slog.Handler {
	return ctxLogHandler{Handler: wrappedHandler}
}

func (handler ctxLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	ctxLevel := LogLevelFromCtx(ctx)
	if ctxLevel == nil {
		return handler.Handler.Enabled(ctx, level)
	}

	return level >= *ctxLevel
}
