package logger

import (
	"context"
	"log/slog"
	"testing"

	"braces.dev/errtrace"
	"github.com/stretchr/testify/require"
)

func TestErrWithLogLevel(t *testing.T) {
	handler := InMemoryHandler(slog.LevelDebug)
	logger := slog.New(handler)
	ctx := context.WithValue(CtxWithLogger(context.Background(), logger), "toto", "tutu")

	require.Empty(t, handler.GetLogs())
	Error(ctx, ErrWithLogLevel(errtrace.New("test"), slog.LevelInfo))
	require.Len(t, handler.GetLogs(), 1)
	require.Equal(t, slog.LevelInfo, handler.GetLogs()[0].Level)
	ErrorAsWarning(ctx, ErrWithLogLevel(errtrace.New("test"), slog.LevelInfo))
	require.Len(t, handler.GetLogs(), 2)
	require.Equal(t, slog.LevelWarn, handler.GetLogs()[1].Level)
}
