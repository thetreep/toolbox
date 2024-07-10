package logger

import (
	"context"
	"log/slog"
	"runtime"

	"braces.dev/errtrace"
)

// ctxLogHandler does two things:
// - it extracts values from the context to add them as attributes
// - it allows overriding the log level via the context
type ctxLogHandler struct {
	slog.Handler
}

func newCtxLogHandler(wrappedHandler slog.Handler) slog.Handler {
	return ctxLogHandler{Handler: wrappedHandler}
}

func (handler ctxLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	ctxLevel := loglevelFromCtx(ctx)
	if ctxLevel == nil {
		return handler.Handler.Enabled(ctx, level)
	}

	return level >= *ctxLevel
}

func (handler ctxLogHandler) Handle(ctx context.Context, record slog.Record) error {
	fixProgramCounter(&record)

	// requestID := middleware.GetReqID(ctx)
	// if requestID != "" {
	// 	record.AddAttrs(slog.String("requestID", requestID))
	// }
	// userSession := session.GetSessionFromContext(ctx)
	// if userSession.IsAuthenticated() {
	// 	record.AddAttrs(slog.String("userID", userSession.GetUserID()))
	// 	record.AddAttrs(slog.String("userEmail", userSession.GetUserEmail()))
	// }

	// you can extract more values from context here as needed

	return errtrace.Wrap(handler.Handler.Handle(ctx, record))
}

// fixProgramCounter changes a record PC (program counter) so it doesn't point to our helper functions
// but to the actual place where the log is written
func fixProgramCounter(record *slog.Record) {
	var pcs [1]uintptr
	const callersToSkip = 6 // magic number found by trial-and-error, may need to be updated if the logger/handlers stack changes
	runtime.Callers(callersToSkip, pcs[:])
	record.PC = pcs[0]
}

func loglevelFromCtx(ctx context.Context) *slog.Level {
	level, ok := ctx.Value(logLevelCtxKey).(slog.Level)
	if !ok {
		return nil
	}
	return &level
}

func ctxWithLogLevel(ctx context.Context, level slog.Level) context.Context {
	return context.WithValue(ctx, logLevelCtxKey, level)
}

type logLevelCtxKeyType string

const logLevelCtxKey logLevelCtxKeyType = "logLevelCtxKey"
