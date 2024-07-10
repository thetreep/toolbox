package logger

import (
	"log/slog"
	"net/http"
	"net/url"

	fingerscrossed "github.com/hectorj/slog-fingerscrossed"
	"github.com/thetreep/toolbox/convert"
)

// func ChiRequestLogger() func(next http.Handler) http.Handler {
// 	return middleware.RequestLogger(newChiLogFormatter())
// }

// LogLevelModifierHTTPMiddleware returns a middleware which allows client to set the log level via a cookie.
// It makes it possible to debug some requests in production without changing the log level globally.
func LogLevelModifierHTTPMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(responseWriter http.ResponseWriter, request *http.Request) {
				logLevelCookie, _ := request.Cookie("LOG_LEVEL")
				if logLevelCookie != nil {
					var logLevel slog.Level
					err := logLevel.UnmarshalText([]byte(logLevelCookie.Value))
					if err == nil {
						request = request.WithContext(ctxWithLogLevel(request.Context(), logLevel))
						Info(
							request.Context(),
							"log level changed via cookie",
							slog.String("logLevel", logLevel.String()),
						)
					}
				}

				next.ServeHTTP(responseWriter, request)
			},
		)
	}
}

// FingersCrossedHTTPMiddleware returns a middleware implementing the "fingers crossed" logging strategy.
// It gives us all logs for a request if an error is logged, even those which should be filtered by the configured log level.
// See https://github.com/hectorj/slog-fingerscrossed for more details.
func FingersCrossedHTTPMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(responseWriter http.ResponseWriter, request *http.Request) {
				ctx := request.Context()
				oldLogger := GetLoggerFromContext(ctx)
				oldHandler := oldLogger.Handler()
				// this is only useful if we aren't already logging everything
				if !oldHandler.Enabled(ctx, slog.LevelDebug) {
					newHandler := fingerscrossed.NewHandler(oldHandler)
					newLogger := slog.New(newHandler)
					ctx = CtxWithLogger(ctx, newLogger)
					request = request.WithContext(ctx)

					baseLogLevel := loglevelFromCtx(ctx)
					if baseLogLevel == nil {
						baseLogLevel = convert.PointerTo(getLogLevel(ctx))
					}
					defer func() { _ = newHandler.FlushLogs(*baseLogLevel) }()
				}

				next.ServeHTTP(responseWriter, request)
			},
		)
	}
}

// func newChiLogFormatter() middleware.LogFormatter {
// 	return &chiLogFormatter{}
// }
//
// type chiLogFormatter struct{}
//
// func (f *chiLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
// 	return &chiLogEntry{
// 		request: r,
// 	}
// }
//
// type chiLogEntry struct {
// 	request *http.Request
// }
//
// func (e *chiLogEntry) Write(status, bytes int, responseHeaders http.Header, elapsed time.Duration, extra interface{}) {
// 	level := slog.LevelInfo
// 	if status >= http.StatusInternalServerError {
// 		level = slog.LevelError
// 	} else if status >= http.StatusBadRequest {
// 		level = slog.LevelWarn
// 	}
//
// 	ctx := e.request.Context()
// 	l := GetLoggerFromContext(ctx)
//
// 	loggedHeaders := maps.Clone(responseHeaders)
// 	loggedHeaders.Del("Set-Cookie")
//
// 	l.LogAttrs(
// 		ctx, level, "HTTP request completed", slog.Group(
// 			"http",
// 			Request(e.request, false),
// 			slog.Int("status", status),
// 			slog.Int("responseBytes", bytes),
// 			slog.Any("responseHeaders", loggedHeaders),
// 			slog.Int64("elapsedNanoSeconds", elapsed.Nanoseconds()),
// 			slog.Any("extra", extra),
// 		),
// 	)
// }

func queryToLogAttrs(query url.Values) slog.Attr {
	valuesAttrs := make([]any, 0, len(query))

	for key, values := range query {
		valuesAttrs = append(valuesAttrs, slog.Any(key, values))
	}

	return slog.Group("query", valuesAttrs...)
}

//
// func (e *chiLogEntry) Panic(v interface{}, stack []byte) {
// 	ctx := e.request.Context()
// 	l := GetLoggerFromContext(ctx)
//
// 	l.LogAttrs(
// 		ctx, slog.LevelError, "HTTP request panicked", slog.Group(
// 			"http",
// 			Request(e.request, false),
// 		),
// 		slog.Any("error", v),
// 		slog.String("stack", string(stack)),
// 	)
// }
