package convert

import (
	"time"
)

func BoolP(v bool) *bool {
	return &v
}

func Float64P(v float64) *float64 {
	return &v
}

func FloatP(v float32) *float32 {
	return &v
}

func Int64P(v int64) *int64 {
	return &v
}

func IntP(v int) *int {
	return &v
}

func StrP(s string) *string {
	return &s
}

func TimeP(t time.Time) *time.Time {
	return &t
}
