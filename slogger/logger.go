package logger

import (
	"context"
	"log/slog"
	"os"
	"slices"

	"github.com/jussi-kalliokoski/slogdriver"
	"github.com/thetreep/express/backend/internal/tools/config"
)

// return the attributes that will not be displayed in "text" mode
//
//nolint:gochecknoglobals // attribute blacklist is a "constant" that is put here for better readability
var textAttributeBlacklist = []string{
	"body",
}

func NewProductionLogger(ctx context.Context) *slog.Logger {
	handler := getBaseHandler(ctx)

	handler = newCtxLogHandler(handler)

	return slog.New(handler)
}

func getBaseHandler(ctx context.Context) slog.Handler {
	var handler slog.Handler

	writer := os.Stderr
	options := &slog.HandlerOptions{
		AddSource:   false,
		ReplaceAttr: nil,
		Level:       getLogLevel(ctx),
	}

	const defaultLogFormat = "json"
	logFormat := config.GetEnvOrDefault("LOG_FORMAT", defaultLogFormat)
	switch logFormat {
	case "text":
		options.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
			if slices.Contains(textAttributeBlacklist, a.Key) {
				return slog.String(a.Key, "#hidden#")
			}
			return a
		}
		handler = slog.NewTextHandler(writer, options)
	case "gcp":
		handler = slogdriver.NewHandler(
			writer, slogdriver.Config{
				// TODO: make it configurable or fetch from [metadata server](https://cloud.google.com/run/docs/container-contract#metadata-server)
				ProjectID: "thetreep-express",
				Level:     options.Level,
			},
		)
	case "json":
		// default
	default:
		Warn(
			ctx,
			"unknown log format, using default value",
			slog.String("unknownValue", logFormat),
			slog.String("defaultValue", defaultLogFormat),
		)
	}
	if handler == nil {
		handler = slog.NewJSONHandler(writer, options)
	}
	return handler
}

func getLogLevel(ctx context.Context) slog.Level {
	const defaultLogLevel = slog.LevelInfo

	logLevelString := config.GetEnvOrDefault("LOG_LEVEL", defaultLogLevel.String())
	var logLevel slog.Level
	err := logLevel.UnmarshalText([]byte(logLevelString))
	if err != nil {
		Warn(
			ctx,
			"unknown log level, using default value",
			slog.String("unknownValue", logLevelString),
			slog.String("defaultValue", defaultLogLevel.String()),
		)
		logLevel = defaultLogLevel
	}
	return logLevel
}
