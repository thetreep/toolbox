package convert

import (
  "strconv"
  "strings"
  "time"
  "database/sql"
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

// PointerToOrNil returns a pointer to v if condition is true, otherwise returns nil.
// This is useful for conditionally setting optional fields in API requests.
//
// Example:
//
//	latitude := convert.PointerToOrNil(geo.Latitude, geo != nil && geo.Latitude != 0)
func PointerToOrNil[T any](v T, condition bool) *T {
  if condition {
    return &v
  }
  return nil
}

// PointerToIf returns a pointer to v if the predicate function returns true, otherwise returns nil.
// This is useful for conditional pointer creation with complex validation logic.
//
// Example:
//
//	isValid := func(lat float64) bool { return lat >= -90 && lat <= 90 }
//	latitude := convert.PointerToIf(geo.Latitude, isValid)
func PointerToIf[T any](v T, predicate func(T) bool) *T {
  if predicate(v) {
    return &v
  }
  return nil
}

// PointerToIfNonZero returns a pointer to v if v is not the zero value for its type, otherwise returns nil.
// This is a convenience wrapper for the common case of checking non-zero values.
//
// Example:
//
//	name := convert.PointerToIfNonZero("John") // returns &"John"
//	empty := convert.PointerToIfNonZero("")    // returns nil
func PointerToIfNonZero[T comparable](v T) *T {
  var zero T
  if v != zero {
    return &v
  }
  return nil
}

// Deprecated: use PointerTo instead.
func BoolP(v bool) *bool {
  return &v
}

// Deprecated: use PointerTo instead.
func Float64P(v float64) *float64 {
  return &v
}

// Deprecated: use PointerTo instead.
func FloatP(v float32) *float32 {
  return &v
}

// Deprecated: use PointerTo instead.
func Int64P(v int64) *int64 {
  return &v
}

// Deprecated: use PointerTo instead.
func IntP(v int) *int {
  return &v
}

// Deprecated: use PointerTo instead.
func StrP(s string) *string {
  return &s
}

// Deprecated: use PointerTo instead.
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

// Float32ToString converts a float32 to a string.
func Float32ToString(f float32) string {
  return strconv.FormatFloat(float64(f), 'f', -1, 32)
}

// Float64ToString converts a float64 to a string.
func Float64ToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func ToPtrString(source sql.NullString) *string {
	var result *string
	if source.Valid {
		result = &source.String
	}

	return result
}

func ToPrtInt16(source sql.NullInt16) *int16 {
	var result *int16
	if source.Valid {
		result = &source.Int16
	}

	return result
}

func ToPtrFloat64(source sql.NullFloat64) *float64 {
	var result *float64
	if source.Valid {
		result = &source.Float64
	}

	return result
}

func ToSQLNullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{Valid: false}
	}

	return sql.NullString{
		String: *value,
		Valid:  true,
	}
}

func ToSQLNullInt16(value *int16) sql.NullInt16 {
	if value == nil {
		return sql.NullInt16{Valid: false}
	}

	return sql.NullInt16{
		Int16: *value,
		Valid: true,
	}
}
