package tempo

import (
	"testing"
	"time"
)

func TestParseISO8601Duration(t *testing.T) {

	tcases := []struct {
		duration    string
		expDuration time.Duration
		expErr      error
	}{
		{
			duration:    "P3Y6M4DT12H30M5S",
			expDuration: time.Duration(110550605000000000),
		},
		{
			duration:    "P0Y0M0DT0H30M5S",
			expDuration: time.Duration(30*60+5) * time.Second,
		},
		{
			duration:    "P0Y0M0DT0H0M0S",
			expDuration: 0,
		},
		{
			duration:    "wrong.format",
			expDuration: 0,
			expErr:      errISO8601DurationFormat,
		},
	}

	for i, tc := range tcases {
		value, err := ParseISO8601Duration(tc.duration)
		if got, want := value, tc.expDuration; got != want {
			t.Fatalf("case #%d : unexpected value, got %d, want %d", i+1, got, want)
		}
		if got, want := err, tc.expErr; got != want {
			t.Fatalf("case #%d : unexpected error, got %v, want %v", i+1, got, want)
		}
	}

}
