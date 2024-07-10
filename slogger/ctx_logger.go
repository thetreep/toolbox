package logger

import (
	"context"
	"log/slog"
)

func CtxWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}

func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerCtxKey).(*slog.Logger)
	if !ok {
		logger = NewProductionLogger(ctx)
	}
	return logger
}

type loggerCtxKeyType string

const loggerCtxKey loggerCtxKeyType = "express-api_slog_logger"
