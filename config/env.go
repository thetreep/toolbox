package config

import (
	"os"

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
