package tests_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thetreep/toolbox/tests"
)

func TestShowCaseAssert(t *testing.T) {
	tests.Setup(t, func(ctx context.Context) {
		tcases := []struct {
			in, expectOut interface{}
		}{
			{"test", "test"},
			{"test1234", "test1234"},
			{&struct{ r int }{1}, &struct{ r int }{1}},
		}

		for _, tcase := range tcases {
			assert.Equal(t, tcase.expectOut, tcase.in)
		}

		tcases = []struct {
			in, expectOut interface{}
		}{
			{"test", "test1"},
			{"test1234", "test124"},
			{&struct{ r int }{1}, &struct{ r int64 }{2}},
		}
	})
}
