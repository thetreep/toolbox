package convert

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	type args struct {
		parser func(string) (value time.Duration, err error)
		value  string
	}

	type testCase struct {
		name      string
		args      args
		want      time.Duration
		wantPanic bool
	}

	tests := []testCase{
		{
			name: "parse duration",
			args: args{
				parser: time.ParseDuration,
				value:  "1h30m",
			},
			want:      time.Minute * 90, // 1 hour 30 minutes
			wantPanic: false,
		},
		{
			name: "panic",
			args: args{
				parser: time.ParseDuration,
				value:  "kldqfjkldsf",
			},
			wantPanic: true,
		},
	}
	// TODO add other parsers

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

func TestMustForTest(t *testing.T) {
	type args struct {
		parser func(string) (value time.Duration, err error)
		value  string
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "parse duration",
			args: args{
				parser: time.ParseDuration,
				value:  "1h30m",
			},
			want: time.Minute * 90, // 1 hour 30 minutes
		},
		// this make test fail :point_down:
		// {
		// 	name: "panic",
		// 	args: args{
		// 		parser: time.ParseDuration,
		// 		value:  "kldqfjkldsf",
		// 	},
		// },
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustForTest(tt.args.parser(tt.args.value))(t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustForTest() = %v, want %v", got, tt.want)
			}
		})
	}
}
