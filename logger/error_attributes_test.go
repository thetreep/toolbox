package logger

import (
	"context"
	"log/slog"
	"testing"

	"braces.dev/errtrace"
	"github.com/stretchr/testify/require"
)

func TestErrWithAttributes(t *testing.T) {
	handler := InMemoryHandler(slog.LevelDebug)
	logger := slog.New(handler)
	ctx := context.WithValue(CtxWithLogger(context.Background(), logger), "toto", "tutu")

	require.Empty(t, handler.GetLogs())
	Error(ctx, ErrWithAttributes(errtrace.New("test"), slog.Any("toto", "tutu")))
	require.Len(t, handler.GetLogs(), 1)
	require.Equal(t, 2, handler.GetLogs()[0].NumAttrs())
	ErrorAsWarning(ctx, ErrWithAttributes(errtrace.New("test"), slog.Any("toto", "tutu")), slog.Any("tata", "titi"))
	require.Len(t, handler.GetLogs(), 2)
	require.Equal(t, 3, handler.GetLogs()[1].NumAttrs())
}
