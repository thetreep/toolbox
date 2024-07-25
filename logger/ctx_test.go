package logger

import (
	"context"
	"log/slog"
	"testing"

	"braces.dev/errtrace"
	"github.com/stretchr/testify/require"
)

func TestCtxWithLogger(t *testing.T) {
	handler := InMemoryHandler(slog.LevelDebug)
	logger := slog.New(handler)

	Debug(context.Background(), "test") // should be a no-op

	ctx := context.WithValue(CtxWithLogger(context.Background(), logger), "toto", "tutu")

	require.Empty(t, handler.GetLogs())
	Debug(ctx, "test")
	require.Len(t, handler.GetLogs(), 1)
	Info(ctx, "test")
	require.Len(t, handler.GetLogs(), 2)
	Warn(ctx, "test")
	require.Len(t, handler.GetLogs(), 3)
	Error(ctx, errtrace.New("test"))
	require.Len(t, handler.GetLogs(), 4)
	ErrorAsWarning(ctx, errtrace.New("test"))
	require.Len(t, handler.GetLogs(), 5)
}
