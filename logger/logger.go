package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/jussi-kalliokoski/slogdriver"
	"github.com/lmittmann/tint"
)

// attribute blacklist is a "constant" that is put here for better readability
// return the attributes that will not be displayed in "text" mode
//
//nolint:gochecknoglobals // see above
var textAttributeBlacklist = []string{
	"body",
}

type LogFormat string

const (
	Text LogFormat = "text"
	Gcp  LogFormat = "gcp"
	Json LogFormat = "json"
)

func Live(ctx context.Context, format LogFormat, level string) *slog.Logger {
	handler := newCtxLogHandler(getBaseHandler(ctx, format, level))

	return slog.New(handler)
}

func getBaseHandler(ctx context.Context, format LogFormat, level string) slog.Handler {
	var handler slog.Handler

	writer := os.Stderr
	options := &slog.HandlerOptions{
		AddSource:   false,
		ReplaceAttr: nil,
		Level:       GetLogLevel(ctx, level),
	}

	const defaultLogFormat = "json"

	switch format {
	case Text:
		handler = tint.NewHandler(writer, &tint.Options{
			AddSource: options.AddSource,
			Level:     options.Level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if slices.Contains(textAttributeBlacklist, a.Key) {
					return slog.String(a.Key, "#hidden#")
				}

				return a
			},
			TimeFormat: time.TimeOnly,
		})
	case Gcp:
		projectID, err := getProjectID()
		if err != nil {
			Warn(
				ctx,
				fmt.Errorf("could not get projectID: %w", err).Error(),
			)

			break
		}

		handler = slogdriver.NewHandler(
			writer, slogdriver.Config{
				ProjectID: projectID,
				Level:     options.Level,
			},
		)
	case "json":
		// default
	default:
		Warn(
			ctx,
			"unknown log format, using default value",
			slog.String("unknownValue", string(format)),
			slog.String("defaultValue", defaultLogFormat),
		)
	}

	if handler == nil {
		handler = slog.NewJSONHandler(writer, options)
	}

	return handler
}

func GetLogLevel(ctx context.Context, level string) slog.Level {
	const defaultLogLevel = slog.LevelInfo

	var logLevel slog.Level

	err := logLevel.UnmarshalText([]byte(level))
	if err != nil {
		Warn(
			ctx,
			"unknown log level, using default value",
			slog.String("unknownValue", level),
			slog.String("defaultValue", defaultLogLevel.String()),
		)

		logLevel = defaultLogLevel
	}

	return logLevel
}

func getProjectID() (string, error) {
	metadataURL := "http://metadata.google.internal/computeMetadata/v1/project/project-id"

	req, err := http.NewRequest(http.MethodGet, metadataURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Metadata-Flavor", "Google")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

type logLevelCtxKey struct{}

func LogLevelFromCtx(ctx context.Context) *slog.Level {
	level, ok := ctx.Value(logLevelCtxKey{}).(slog.Level)
	if !ok {
		return nil
	}

	return &level
}

func CtxWithLogLevel(ctx context.Context, level slog.Level) context.Context {
	return context.WithValue(ctx, logLevelCtxKey{}, level)
}
