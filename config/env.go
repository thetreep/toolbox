package config

import (
	"context"
	"os"
	"strings"

	"braces.dev/errtrace"
)

func GetEnvOrError(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists {
		return "", errtrace.Errorf("missing env var %q", key)
	}
	return value, nil
}

func GetEnvOrDefault(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func ResolveEnv(ctx context.Context, key string) (string, error) {
	switch {
	case strings.HasPrefix(key, GCPSecretPrefix):
		return ResolveSecretFromGCP(ctx, os.Getenv(key))
	case strings.HasPrefix(key, OPSecretPrefix):
		return ResolveSecretFromOP(ctx, os.Getenv(key))
	default:
		return GetEnvOrError(key)
	}
}

func ResolveEnvOrDefault(ctx context.Context, key string, defaultValue string) string {
	val, err := ResolveEnv(ctx, key)
	if err != nil {
		return defaultValue
	}
	return val
}
