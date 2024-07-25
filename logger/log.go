// Package logger provides some helpers over the standard `log/slog` package for structured logging.
package logger

import (
	"context"
	"log/slog"

	"braces.dev/errtrace"
)

func Error(ctx context.Context, err error, attrs ...slog.Attr) {
	attrs = append(attrs, slog.Any("stacktrace", errtrace.FormatString(err)))
	attrs = append(attrs, getAttributesFromErr(err)...)

	logAttrs(ctx, getErrorLevel(err), err.Error(), attrs...)
}

func ErrorAsWarning(ctx context.Context, err error, attrs ...slog.Attr) {
	Error(ctx, ErrWithLogLevel(err, slog.LevelWarn), attrs...)
}

func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	logAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

func Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	logAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

func Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	logAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

func logAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	logger := GetLoggerFromContext(ctx)
	attrs = append(attrs, getAttributesFromContext(ctx)...)

	logger.LogAttrs(ctx, level, msg, attrs...)
}
