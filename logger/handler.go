package logger

import (
	"context"
	"log/slog"
)

// ctxLogHandler implement slog.Handler,
// it override Enabled method to get the log level from the context.
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
