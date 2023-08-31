package convert

import (
	"strconv"
	"strings"
	"time"
)

// ToFloat converts string to float64. It returns 0 if conversion is impossible.
func ToFloat(s string) float32 {
	f, _ := strconv.ParseFloat(strings.Replace(s, ",", ".", -1), 32)

	return float32(f)
}

// ToFloat64 converts string to float64. It returns 0 if conversion is impossible.
func ToFloat64(s string) float64 {
	f, _ := strconv.ParseFloat(strings.Replace(s, ",", ".", -1), 64)

	return f
}

// ToInt converts string to int. It returns 0 if conversion is impossible.
func ToInt(s string) int {
	i, _ := strconv.Atoi(s)

	return i
}

// ToInt64 converts string to int64. It returns 0 if conversion is impossible.
func ToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 0, 64)

	return i
}

func PointerTo[T any](v T) *T {
	return &v
}

// Deprecated: use PointerTo instead
func BoolP(v bool) *bool {
	return &v
}

// Deprecated: use PointerTo instead
func Float64P(v float64) *float64 {
	return &v
}

// Deprecated: use PointerTo instead
func FloatP(v float32) *float32 {
	return &v
}

// Deprecated: use PointerTo instead
func Int64P(v int64) *int64 {
	return &v
}

// Deprecated: use PointerTo instead
func IntP(v int) *int {
	return &v
}

// Deprecated: use PointerTo instead
func StrP(s string) *string {
	return &s
}

// Deprecated: use PointerTo instead
func TimeP(t time.Time) *time.Time {
	return &t
}

func EqualInt64P(a, b *int64) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
