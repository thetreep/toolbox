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
			sanitize phone.SanitizeMode
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
				sanitize: phone.SanitizeDefault,
				expNum:   "0123456789",
			},
			{
				num:      "+33123456789",
				country:  "FR",
				format:   phone.NATIONAL,
				sanitize: phone.SanitizeDefault,
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
				sanitize: phone.SanitizeDefault,
				expNum:   "0123456789",
			},
			{
				num:     "0123456789",
				country: "FR",
				format:  phone.INTERNATIONAL,
				expNum:  "+33 1 23 45 67 89",
			},
			{
				num:      "(248) 123-7654",
				country:  "US",
				format:   phone.INTERNATIONAL,
				sanitize: phone.SanitizeStrict,
				expNum:   "12481237654",
			},
			{
				num:      "+1-248-123-7654",
				country:  "FR",
				format:   phone.INTERNATIONAL,
				sanitize: phone.SanitizeStrict,
				expNum:   "12481237654",
			},
		}

		for i, tc := range tcases {
			name := fmt.Sprintf("case #%d : ", i+1)
			num, err := phone.ParseWithFallback(tc.num, tc.country, tc.format, tc.sanitize)
			assert.ErrorIs(t, tc.expErr, err, name+"unexpected error")
			assert.Equal(t, tc.expNum, num, name+"unexpected number")
		}
	})
}

func TestGetRegionFromNumber(t *testing.T) {
	tests := []struct {
		number   string
		expected string
		expErr   error
	}{
		{"+33123456789", "FR", nil},
		{"+1-212-456-7890", "US", nil},
		{"+442079460958", "GB", nil},
		{"+34912345678", "ES", nil},
		{"0123456789", "", phone.ErrInvalidInternationalFormat},
		{"12481237654", "", phone.ErrInvalidInternationalFormat},
	}

	for _, test := range tests {
		t.Run(test.number, func(t *testing.T) {
			region, err := phone.GetRegionFromInternationalNumber(test.number)
			assert.ErrorIs(t, err, test.expErr)
			assert.Equal(t, test.expected, region)
		})
	}
}
