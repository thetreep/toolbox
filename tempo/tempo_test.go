package tempo_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thetreep/toolbox/tempo"
	"github.com/thetreep/toolbox/tests"
)

func TestParseISO8601Duration(t *testing.T) {
	tests.Setup(t, func(ctx context.Context) {
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
				expErr:      tempo.ErrISO8601DurationFormat,
			},
		}

		for i, tc := range tcases {
			gotDuration, gotErr := tempo.ParseISO8601Duration(tc.duration)
			assert.Equal(t, gotDuration, tc.expDuration, fmt.Sprintf("case #%d", i+1))
			assert.Equal(t, gotErr, tc.expErr, fmt.Sprintf("case #%d", i+1))
		}
	})
}
