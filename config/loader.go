package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Loader reads configuration values from environment variables,
// optionally resolving GCP secrets. Errors are accumulated and
// surfaced via Err().
type Loader struct {
	ctx context.Context
	err error
}

// NewLoader returns a new Loader bound to the given context.
func NewLoader(ctx context.Context) *Loader {
	return &Loader{ctx: ctx}
}

// Err returns the first error encountered during loading, if any.
func (l *Loader) Err() error {
	return l.err
}

// Must panics if any error was encountered during loading.
// Useful for fail-fast initialisation in main().
func (l *Loader) Must() {
	if l.err != nil {
		panic(l.err)
	}
}

// Str returns the string value of key, or fallback if unset.
func (l *Loader) Str(key, fallback string) string {
	return load(l, key, fallback, castString)
}

// Int returns the int value of key, or fallback if unset.
func (l *Loader) Int(key string, fallback int) int {
	return load(l, key, fallback, castInt)
}

// Int64 returns the int64 value of key, or fallback if unset.
func (l *Loader) Int64(key string, fallback int64) int64 {
	return load(l, key, fallback, castInt64)
}

// Float64 returns the float64 value of key, or fallback if unset.
func (l *Loader) Float64(key string, fallback float64) float64 {
	return load(l, key, fallback, castFloat64)
}

// Bool returns the boolean value of key, or fallback if unset.
// Accepted truthy values: "1", "t", "true" (case-insensitive).
func (l *Loader) Bool(key string, fallback bool) bool {
	return load(l, key, fallback, castBool)
}

// Duration returns the time.Duration value of key, or fallback if unset.
// Values must be valid Go duration strings, e.g. "5s", "1m30s".
func (l *Loader) Duration(key string, fallback time.Duration) time.Duration {
	return load(l, key, fallback, castDuration)
}

// Required returns the string value of key. If the key is unset, it
// records an error and returns an empty string. Use when there is no
// sensible fallback.
func (l *Loader) Required(key string) string {
	if l.err != nil {
		return ""
	}
	raw := os.Getenv(key)
	if raw == "" {
		l.err = fmt.Errorf("config: required key %q is not set", key)
		return ""
	}
	resolved, err := ResolveSecretFromGCP(l.ctx, raw)
	if err != nil {
		l.err = fmt.Errorf("config: resolve %q: %w", key, err)
		return ""
	}
	return resolved
}

// load is the generic backbone used by all typed accessors.
// It resolves the env var through GCP if needed, then casts it.
func load[T any](l *Loader, key string, fallback T, cast func(string) (T, error)) T {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	resolved, err := ResolveSecretFromGCP(l.ctx, raw)
	if err != nil {
		l.err = errors.Join(l.err, fmt.Errorf("config: resolve %q: %w", key, err))
		return fallback
	}
	val, err := cast(resolved)
	if err != nil {
		l.err = errors.Join(l.err, fmt.Errorf("config: cast %q=%q: %w", key, resolved, err))
		return fallback
	}
	return val
}

func castString(s string) (string, error) {
	return s, nil
}

func castInt(s string) (int, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("expected int, got %q", s)
	}
	return v, nil
}

func castInt64(s string) (int64, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("expected int64, got %q", s)
	}
	return v, nil
}

func castFloat64(s string) (float64, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("expected float64, got %q", s)
	}
	return v, nil
}

func castBool(s string) (bool, error) {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("expected bool, got %q", s)
	}
	return v, nil
}

func castDuration(s string) (time.Duration, error) {
	v, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("expected duration (e.g. \"5s\"), got %q", s)
	}
	return v, nil
}
