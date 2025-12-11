// Package ctxcache provides a way to cache values inside a context.
// It is useful to avoid calling an expensive function several times in the same request/process,
// without having to pass its results deep into many unrelated functions.
// It is concurrency safe.
package ctxcache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"braces.dev/errtrace"
	"github.com/thetreep/toolbox/logger"
)

type contextCache[T any] struct {
	data map[string]func() (T, error)
	lock sync.Mutex
}

type contextCacheCtxKey[T any] struct{}

func (c *contextCache[T]) get(ctx context.Context, cacheKey string, builder func() (T, error)) (T, error) {
	ctx = c.ctxWithAttributes(ctx, cacheKey)
	promise := func() func() (T, error) {
		c.lock.Lock()
		defer c.lock.Unlock()
		if cached, exists := c.data[cacheKey]; exists {
			if verboseLogsEnabled(ctx) {
				logger.Debug(ctx, "context cache hit")
			}
			return cached
		}
		if verboseLogsEnabled(ctx) {
			logger.Debug(ctx, "context cache miss")
		}
		promise := sync.OnceValues(
			func() (T, error) {
				return errtrace.Wrap2(builder())
			},
		)
		c.data[cacheKey] = promise
		return promise
	}()

	return errtrace.Wrap2(promise())
}

func (c *contextCache[T]) ctxWithAttributes(ctx context.Context, cacheKey string) context.Context {
	attributes := []any{slog.String("cacheType", fmt.Sprintf("%T", *new(T)))}
	if cacheKey != "" {
		attributes = append(attributes, slog.String("cacheKey", cacheKey))
	}
	return logger.CtxWithLogAttributes(ctx, attributes...)
}

func ContextWithCache[T any](ctx context.Context) context.Context {
	cache, hasCache := ctx.Value(contextCacheCtxKey[T]{}).(*contextCache[T])
	if hasCache {
		if verboseLogsEnabled(ctx) {
			logger.Debug(cache.ctxWithAttributes(ctx, ""), "context cache already exists")
		}
		return ctx
	}
	cache = &contextCache[T]{
		data: make(map[string]func() (T, error)),
	}
	logger.Debug(cache.ctxWithAttributes(ctx, ""), "context cache created")
	return context.WithValue(ctx, contextCacheCtxKey[T]{}, cache)
}

func GetFromContextCache[T any](ctx context.Context, cacheKey string, builder func() (T, error)) (T, error) {
	cache, hasCache := ctx.Value(contextCacheCtxKey[T]{}).(*contextCache[T])
	if !hasCache {
		if verboseLogsEnabled(ctx) {
			logger.Debug(cache.ctxWithAttributes(ctx, ""), "context cache does not exists")
		}
		return errtrace.Wrap2(builder())
	}

	return errtrace.Wrap2(cache.get(ctx, cacheKey, builder))
}

// PutInContextCache manually stores a value in the context cache.
// THIS IS NOT THE RECOMMENDED WAY OF USING THIS CACHE.
// Usually the cache is populated by the builder of GetFromContextCache
func PutInContextCache[T any](ctx context.Context, cacheKey string, value T) {
	cache, hasCache := ctx.Value(contextCacheCtxKey[T]{}).(*contextCache[T])
	if !hasCache {
		if verboseLogsEnabled(ctx) {
			logger.Debug(cache.ctxWithAttributes(ctx, ""), "context cache does not exists")
		}
		return
	}
	ctx = cache.ctxWithAttributes(ctx, cacheKey)
	cache.lock.Lock()
	defer cache.lock.Unlock()
	if verboseLogsEnabled(ctx) {
		logger.Debug(ctx, "context cache put")
	}
	cache.data[cacheKey] = func() (T, error) { //nolint:unparam // matches the cache signature
		return value, nil
	}
}
