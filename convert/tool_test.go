package convert

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	type args[T any] struct {
		parser func(string) (value T, err error)
		value  string
	}

	type testCase[T any] struct {
		name      string
		args      args[T]
		want      T
		wantPanic bool
	}

	tests := []testCase[time.Duration]{
		{
			name: "parse duration",
			args: args[time.Duration]{
				parser: time.ParseDuration,
				value:  "1h30m",
			},
			want:      time.Minute * 90, // 1 hour 30 minutes
			wantPanic: false,
		},
		{
			name: "panic",
			args: args[time.Duration]{
				parser: time.ParseDuration,
				value:  "kldqfjkldsf",
			},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() { Must(tt.args.parser(tt.args.value)) }, "Must() should panic")
			} else {
				if got := Must(tt.args.parser(tt.args.value)); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Must() = %v, want %v", got, tt.want)
				}
			}

		})
	}
}
