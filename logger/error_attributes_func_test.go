package logger

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func recordAttrs(record slog.Record) map[string]any {
	attrs := make(map[string]any, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value.Any()

		return true
	})

	return attrs
}

func TestCtxWithErrorAttributesFunc(t *testing.T) {
	kindFunc := func(error) []any { return []any{slog.String("kind", "test")} }

	tests := map[string]struct {
		fn        func(err error) []any
		err       error
		log       func(ctx context.Context, err error)
		wantAttrs map[string]any
		wantLevel slog.Level
	}{
		"no function": {
			fn:        nil,
			err:       errors.New("boom"),
			log:       func(ctx context.Context, err error) { Error(ctx, err) },
			wantAttrs: map[string]any{},
			wantLevel: slog.LevelError,
		},
		"function returns attributes": {
			fn:        kindFunc,
			err:       errors.New("boom"),
			log:       func(ctx context.Context, err error) { Error(ctx, err) },
			wantAttrs: map[string]any{"kind": "test"},
			wantLevel: slog.LevelError,
		},
		"function returns nil": {
			fn:        func(error) []any { return nil },
			err:       errors.New("boom"),
			log:       func(ctx context.Context, err error) { Error(ctx, err) },
			wantAttrs: map[string]any{},
			wantLevel: slog.LevelError,
		},
		"function receives the logged error": {
			fn:        func(err error) []any { return []any{slog.String("message", err.Error())} },
			err:       errors.New("boom"),
			log:       func(ctx context.Context, err error) { Error(ctx, err) },
			wantAttrs: map[string]any{"message": "boom"},
			wantLevel: slog.LevelError,
		},
		"combined with error attributes and call attributes": {
			fn:  kindFunc,
			err: ErrWithAttributes(errors.New("boom"), slog.String("from_err", "yes")),
			log: func(ctx context.Context, err error) { Error(ctx, err, slog.String("from_call", "yes")) },
			wantAttrs: map[string]any{
				"kind":      "test",
				"from_err":  "yes",
				"from_call": "yes",
			},
			wantLevel: slog.LevelError,
		},
		"error as warning": {
			fn:        kindFunc,
			err:       errors.New("boom"),
			log:       func(ctx context.Context, err error) { ErrorAsWarning(ctx, err) },
			wantAttrs: map[string]any{"kind": "test"},
			wantLevel: slog.LevelWarn,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := InMemoryHandler(slog.LevelDebug)
			ctx := CtxWithLogger(context.Background(), slog.New(handler))
			ctx = CtxWithErrorAttributesFunc(ctx, tt.fn)

			tt.log(ctx, tt.err)

			require.Len(t, handler.GetLogs(), 1)
			record := handler.GetLogs()[0]
			require.Equal(t, tt.wantLevel, record.Level)

			attrs := recordAttrs(record)
			require.Contains(t, attrs, "stacktrace")
			delete(attrs, "stacktrace")
			require.Equal(t, tt.wantAttrs, attrs)
		})
	}
}

func TestCtxWithErrorAttributesFuncOrder(t *testing.T) {
	handler := InMemoryHandler(slog.LevelDebug)
	ctx := CtxWithLogger(context.Background(), slog.New(handler))
	ctx = CtxWithErrorAttributesFunc(ctx, func(error) []any { return []any{slog.String("from_func", "yes")} })

	Error(ctx, ErrWithAttributes(errors.New("boom"), slog.String("from_err", "yes")), slog.String("from_call", "yes"))

	require.Len(t, handler.GetLogs(), 1)
	var keys []string
	handler.GetLogs()[0].Attrs(func(attr slog.Attr) bool {
		keys = append(keys, attr.Key)

		return true
	})
	require.Equal(t, []string{"from_call", "stacktrace", "from_err", "from_func"}, keys)
}

func TestCtxWithErrorAttributesFuncKeepsSourceLocation(t *testing.T) {
	handler := InMemoryHandler(slog.LevelDebug)
	ctx := CtxWithLogger(context.Background(), slog.New(WrapHandlerToFixStackTrace(handler)))
	ctx = CtxWithErrorAttributesFunc(ctx, func(error) []any { return []any{slog.String("kind", "test")} })

	buf := bytes.NewBuffer(nil)
	sourcePrinter := slog.NewTextHandler(buf, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug})

	Error(ctx, errors.New("test"))
	require.Len(t, handler.GetLogs(), 1)
	require.NoError(t, sourcePrinter.Handle(ctx, handler.GetLogs()[0]))
	require.Contains(t, buf.String(), "logger/error_attributes_func_test.go:127")

	buf.Reset()
	ErrorAsWarning(ctx, errors.New("test"))
	require.Len(t, handler.GetLogs(), 2)
	require.NoError(t, sourcePrinter.Handle(ctx, handler.GetLogs()[1]))
	require.Contains(t, buf.String(), "logger/error_attributes_func_test.go:133")
}
