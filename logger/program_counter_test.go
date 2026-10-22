package logger

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapHandlerToFixStackTrace(t *testing.T) {
	handler := InMemoryHandler(slog.LevelDebug)
	logger := slog.New(WrapHandlerToFixStackTrace(handler))
	ctx := CtxWithLogger(context.Background(), logger)

	buf := bytes.NewBuffer(nil)
	printer := slog.NewTextHandler(
		buf, &slog.HandlerOptions{
			AddSource:   true,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
		},
	)

	// line numbers must match the line at which the log was done

	Debug(ctx, "test")
	require.Len(t, handler.GetLogs(), 1)
	require.NoError(t, printer.Handle(ctx, handler.GetLogs()[0]))
	require.Contains(t, buf.String(), "toolbox/logger/program_counter_test.go:29")

	buf.Reset()
	Info(ctx, "test")
	require.Len(t, handler.GetLogs(), 2)
	require.NoError(t, printer.Handle(ctx, handler.GetLogs()[1]))
	require.Contains(t, buf.String(), "toolbox/logger/program_counter_test.go:35")

	buf.Reset()
	Warn(ctx, "test")
	require.Len(t, handler.GetLogs(), 3)
	require.NoError(t, printer.Handle(ctx, handler.GetLogs()[2]))
	require.Contains(t, buf.String(), "toolbox/logger/program_counter_test.go:41")

	buf.Reset()
	Error(ctx, errors.New("test"))
	require.Len(t, handler.GetLogs(), 4)
	require.NoError(t, printer.Handle(ctx, handler.GetLogs()[3]))
	require.Contains(t, buf.String(), "toolbox/logger/program_counter_test.go:47")

	buf.Reset()
	ErrorAsWarning(ctx, errors.New("test"))
	require.Len(t, handler.GetLogs(), 5)
	require.NoError(t, printer.Handle(ctx, handler.GetLogs()[4]))
	require.Contains(t, buf.String(), "toolbox/logger/program_counter_test.go:53")
}

func TestWrapHandlerToFixStackTrace_throughCtxProxyHandler(t *testing.T) {
	handler := InMemoryHandler(slog.LevelDebug)
	ctx := CtxWithLogger(context.Background(), slog.New(WrapHandlerToFixStackTrace(handler)))
	proxyLogger := slog.New(CtxProxyHandler{})

	buf := bytes.NewBuffer(nil)
	printer := slog.NewTextHandler(buf, &slog.HandlerOptions{AddSource: true})

	proxyLogger.ErrorContext(ctx, "test")
	require.Len(t, handler.GetLogs(), 1)
	require.NoError(t, printer.Handle(ctx, handler.GetLogs()[0]))
	require.Contains(t, buf.String(), "toolbox/logger/program_counter_test.go:67")
}

func TestWrapHandlerToFixStackTrace_directSlogCall(t *testing.T) {
	handler := InMemoryHandler(slog.LevelDebug)
	logger := slog.New(WrapHandlerToFixStackTrace(handler))

	buf := bytes.NewBuffer(nil)
	printer := slog.NewTextHandler(buf, &slog.HandlerOptions{AddSource: true})

	logger.ErrorContext(context.Background(), "test")
	require.Len(t, handler.GetLogs(), 1)
	require.NoError(t, printer.Handle(context.Background(), handler.GetLogs()[0]))
	require.Contains(t, buf.String(), "toolbox/logger/program_counter_test.go:80")
}
