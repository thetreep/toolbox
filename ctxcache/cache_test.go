package ctxcache

import (
	"context"
	"errors"
	"testing"

	"braces.dev/errtrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFromContextCache(t *testing.T) {
	type testData struct {
		Value int
	}

	type test struct {
		name       string
		setupFunc  func(ctx context.Context) context.Context
		cacheKey   string
		builder    func() (testData, error)
		expectErr  bool
		expectData testData
	}

	tests := []test{
		{
			name:      "cache_not_in_context",
			setupFunc: func(ctx context.Context) context.Context { return ctx },
			cacheKey:  "key1",
			builder: func() (testData, error) {
				return testData{Value: 42}, nil
			},
			expectErr:  false,
			expectData: testData{Value: 42},
		},
		{
			name: "key_not_in_cache",
			setupFunc: func(ctx context.Context) context.Context {
				return ContextWithCache[testData](ctx)
			},
			cacheKey: "key1",
			builder: func() (testData, error) {
				return testData{Value: 42}, nil
			},
			expectErr:  false,
			expectData: testData{Value: 42},
		},
		{
			name: "key_in_cache",
			setupFunc: func(ctx context.Context) context.Context {
				ctx = ContextWithCache[testData](ctx)
				_, _ = GetFromContextCache(ctx, "key1", func() (testData, error) { return testData{Value: 99}, nil })
				return ctx
			},
			cacheKey: "key1",
			builder: func() (testData, error) {
				return testData{Value: 42}, nil
			},
			expectErr:  false,
			expectData: testData{Value: 99},
		},
		{
			name:      "builder_returns_error",
			setupFunc: func(ctx context.Context) context.Context { return ctx },
			cacheKey:  "key1",
			builder: func() (testData, error) {
				return testData{}, errtrace.Wrap(errors.New("builder error"))
			},
			expectErr:  true,
			expectData: testData{},
		},
		{
			name: "invalid_cache_type",
			setupFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, contextCacheCtxKey[string]{}, "invalid_cache")
			},
			cacheKey: "key1",
			builder: func() (testData, error) {
				return testData{Value: 42}, nil
			},
			expectErr:  false,
			expectData: testData{Value: 42},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ctx := tt.setupFunc(context.Background())
				result, err := GetFromContextCache(ctx, tt.cacheKey, tt.builder)

				if tt.expectErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tt.expectData, result)
			},
		)
	}
}

func TestPutInContextCache(t *testing.T) {
	ctx := ContextWithCache[int](context.Background())
	PutInContextCache(ctx, "key1", 42)
	result, err := GetFromContextCache(ctx, "key1", func() (int, error) { return -1, nil })
	require.NoError(t, err)
	assert.Equal(t, 42, result)
	result, err = GetFromContextCache(ctx, "key2", func() (int, error) { return -1, nil })
	require.NoError(t, err)
	assert.Equal(t, -1, result)
	PutInContextCache(ctx, "key2", 42)
	result, err = GetFromContextCache(ctx, "key2", func() (int, error) { return -2, nil })
	require.NoError(t, err)
	assert.Equal(t, 42, result)
}
