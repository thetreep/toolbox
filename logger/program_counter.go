package logger

import (
	"context"
	"log/slog"
	"runtime"
)

// "program counters" are used to get stack traces
// we store them to skip all the logging helper methods in the trace

type programCounterCtxKey struct{}

func storeProgramCounter(ctx context.Context, skipCount int) context.Context {
	if _, exists := getProgramCounterFromContext(ctx); exists {
		// we do not override previously stored value
		return ctx
	}
	pc, _, _, ok := runtime.Caller(1 + skipCount) // +1 to skip this function too
	if !ok {
		return ctx
	}
	return context.WithValue(ctx, programCounterCtxKey{}, pc)
}

func getProgramCounterFromContext(ctx context.Context) (uintptr, bool) {
	pc, ok := ctx.Value(programCounterCtxKey{}).(uintptr)
	return pc, ok
}

type fixProgramCounterHandler struct {
	wrapped slog.Handler
}

func (h *fixProgramCounterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.wrapped.Enabled(ctx, level)
}

func (h *fixProgramCounterHandler) Handle(ctx context.Context, record slog.Record) error {
	pc, ok := getProgramCounterFromContext(ctx)
	if ok {
		record.PC = pc
	}

	return h.wrapped.Handle(ctx, record)
}

func (h *fixProgramCounterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &fixProgramCounterHandler{h.wrapped.WithAttrs(attrs)}
}

func (h *fixProgramCounterHandler) WithGroup(name string) slog.Handler {
	return &fixProgramCounterHandler{h.wrapped.WithGroup(name)}
}

func WrapHandlerToFixStackTrace(handler slog.Handler) slog.Handler {
	return &fixProgramCounterHandler{
		wrapped: handler,
	}
}
