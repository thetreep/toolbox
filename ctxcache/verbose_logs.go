package ctxcache

import "context"

type verboseLogsCtxKey struct{}

func CtxWithVerboseLogs(ctx context.Context) context.Context {
	return context.WithValue(ctx, verboseLogsCtxKey{}, true)
}

func CtxWithoutVerboseLogs(ctx context.Context) context.Context {
	return context.WithValue(ctx, verboseLogsCtxKey{}, false)
}

func verboseLogsEnabled(ctx context.Context) bool {
	enabled, _ := ctx.Value(verboseLogsCtxKey{}).(bool)
	return enabled
}
