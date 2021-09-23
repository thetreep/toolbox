package phone_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thetreep/toolbox/phone"
	"github.com/thetreep/toolbox/tests"
)

func TestParse(t *testing.T) {
	tests.Setup(t, func(c context.Context) {
		tcases := []struct {
			num      string
			country  string
			format   phone.Format
			sanitize bool
			expNum   string
			expErr   error
		}{
			{
				num:     "0123456789",
				country: "FR",
				format:  phone.E164,
				expNum:  "+33123456789",
			},
			{
				num:     "0123456789",
				country: "FR",
				format:  phone.NATIONAL,
				expNum:  "01 23 45 67 89",
			},
			{
				num:      "0123456789",
				country:  "FR",
				format:   phone.NATIONAL,
				sanitize: true,
				expNum:   "0123456789",
			},
			{
				num:      "+33123456789",
				country:  "FR",
				format:   phone.NATIONAL,
				sanitize: true,
				expNum:   "0123456789",
			},
			{
				num:     "0033123456789",
				country: "FR",
				format:  phone.E164,
				expNum:  "+33123456789",
			},
			{
				num:      "0033123456789",
				country:  "FR",
				format:   phone.NATIONAL,
				sanitize: true,
				expNum:   "0123456789",
			},
			{
				num:     "0123456789",
				country: "FR",
				format:  phone.INTERNATIONAL,
				expNum:  "+33 1 23 45 67 89",
			},
		}

		for i, tc := range tcases {
			name := fmt.Sprintf("case #%d : ", i+1)
			num, err := phone.Parse(tc.num, tc.country, tc.format, tc.sanitize)
			assert.ErrorIs(t, tc.expErr, err, name+"unexpected error")
			assert.Equal(t, tc.expNum, num, name+"unexpected number")
		}
	})
}
