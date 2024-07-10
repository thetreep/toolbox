package logger

import (
	"context"
	"log/slog"

	"braces.dev/errtrace"
)

func Error(ctx context.Context, err error, attrs ...slog.Attr) {
	logger := GetLoggerFromContext(ctx)

	attrs = append(attrs, slog.Any("stacktrace", errtrace.FormatString(err)))

	logger.LogAttrs(ctx, getErrorLevel(err), err.Error(), attrs...)
}

func getErrorLevel(err error) slog.Level {
	level := slog.LevelError
	// var errWithStatusCode tools.ErrorWithLevel
	// if errors.As(err, &errWithStatusCode) {
	// 	level = errWithStatusCode.GetErrorLevel()
	// }
	return level
}

func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	logger := GetLoggerFromContext(ctx)
	logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

func Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	logger := GetLoggerFromContext(ctx)
	logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

func Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	logger := GetLoggerFromContext(ctx)
	logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}
