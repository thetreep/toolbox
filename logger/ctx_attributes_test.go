package logger

import (
	"context"
	"log/slog"
	"testing"

	"braces.dev/errtrace"
	"github.com/stretchr/testify/require"
)

func TestCtxWithLogAttributes(t *testing.T) {
	handler := InMemoryHandler(slog.LevelDebug)
	logger := slog.New(handler)
	ctx := CtxWithLogger(context.Background(), logger)
	ctx = CtxWithLogAttributes(ctx, slog.String("toto", "tutu"), slog.Int("?", 42))
	ctx2 := CtxWithLogAttributes(ctx, slog.String("tata", "tonton"), slog.Int("?", 404))
	ctx3 := CtxWithLogger(context.Background(), logger)

	require.Empty(t, handler.GetLogs())
	Debug(ctx, "test")
	require.Len(t, handler.GetLogs(), 1)
	require.Equal(t, 2, handler.GetLogs()[0].NumAttrs())
	Info(ctx, "test", slog.Any("tata", "titi"))
	require.Len(t, handler.GetLogs(), 2)
	require.Equal(t, 3, handler.GetLogs()[1].NumAttrs())
	Warn(ctx, "test")
	require.Len(t, handler.GetLogs(), 3)
	require.Equal(t, 2, handler.GetLogs()[2].NumAttrs())
	Error(ctx, errtrace.New("test"))
	require.Len(t, handler.GetLogs(), 4)
	require.Equal(t, 3, handler.GetLogs()[3].NumAttrs())
	ErrorAsWarning(ctx, errtrace.New("test"))
	require.Len(t, handler.GetLogs(), 5)
	require.Equal(t, 3, handler.GetLogs()[4].NumAttrs())

	Debug(ctx2, "test")
	require.Len(t, handler.GetLogs(), 6)
	require.Equal(t, 3, handler.GetLogs()[5].NumAttrs())

	Debug(ctx3, "test", "toto", "tonton", "tata", "ok", slog.String("key", "value"))
	require.Len(t, handler.GetLogs(), 7)
	require.Equal(t, 3, handler.GetLogs()[6].NumAttrs())
}
