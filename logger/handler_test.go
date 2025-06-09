package logger_test

import (
	"context"
	"log/slog"
	"testing"

	"braces.dev/errtrace"
	"github.com/stretchr/testify/require"
	. "github.com/thetreep/toolbox/logger" //nolint:depguard // logger_test package
)

func TestCtxWithLogLevel(t *testing.T) {
	baseHandler := InMemoryHandler(slog.LevelInfo)
	handler := NewCtxLogHandler(baseHandler)
	logger := slog.New(handler)

	Debug(context.Background(), "test") // should be a no-op

	ctx := context.WithValue(CtxWithLogger(context.Background(), logger), "toto", "tutu")

	require.Empty(t, baseHandler.GetLogs())
	Debug(ctx, "test")
	require.Len(t, baseHandler.GetLogs(), 0)
	Info(ctx, "test")
	require.Len(t, baseHandler.GetLogs(), 1)

	ctx = CtxWithLogLevel(ctx, slog.LevelDebug)
	Debug(ctx, "test")
	require.Len(t, baseHandler.GetLogs(), 2)

	ctx = CtxWithLogLevel(ctx, slog.LevelError)
	Warn(ctx, "test")
	require.Len(t, baseHandler.GetLogs(), 2)
	Error(ctx, errtrace.New("test"))
	require.Len(t, baseHandler.GetLogs(), 3)
}
