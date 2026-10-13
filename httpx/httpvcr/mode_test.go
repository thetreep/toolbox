package httpvcr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

func TestRecordMode(t *testing.T) {
	tests := []struct {
		name string
		ci   string
		k7   string
		want recorder.Mode
	}{
		{"default records once", "", "", recorder.ModeRecordOnce},
		{"CI replays only", "true", "", recorder.ModeReplayOnly},
		{"CI wins over K7_RECORD", "1", "1", recorder.ModeReplayOnly},
		{"K7_RECORD=1 records only", "", "1", recorder.ModeRecordOnly},
		{"K7_RECORD=true records only", "", "true", recorder.ModeRecordOnly},
		{"other K7_RECORD values are ignored", "", "yes", recorder.ModeRecordOnce},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CI", tc.ci)
			t.Setenv("K7_RECORD", tc.k7)

			assert.Equal(t, tc.want, recordMode())
		})
	}
}
