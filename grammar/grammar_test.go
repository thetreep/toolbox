package grammar_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thetreep/toolbox/grammar"
	"github.com/thetreep/toolbox/tests"
)

func TestNormalize(t *testing.T) {
	tests.Setup(t, func(ctx context.Context) {
		tcases := []struct {
			in, expectOut string
		}{
			{"test", "test"},
			{"test1234", "test1234"},
			{"pierre-francois", "pierre francois"},
			{"Pierre-Francois", "pierre francois"},
		}

		for _, tcase := range tcases {
			got := grammar.Normalize(tcase.in)
			assert.Equal(t, tcase.expectOut, got)
		}
	})

}
func TestCapitalize(t *testing.T) {
	tcases := []struct {
		in, expectOut string
	}{
		{"", ""},
		{"test", "Test"},
		{"test1234", "Test1234"},
		{"pierre-francois", "Pierre-francois"},
		{"Pierre-Francois", "Pierre-francois"},
	}

	for _, tcase := range tcases {
		got := grammar.Capitalize(tcase.in)
		assert.Equal(t, tcase.expectOut, got)
	}
}
func TestJustCapitalize(t *testing.T) {
	tcases := []struct {
		in, expectOut string
	}{
		{"", ""},
		{"test", "Test"},
		{"tEST1234", "TEST1234"},
		{"pierre-francois", "Pierre-francois"},
		{"Pierre-Francois", "Pierre-Francois"},
	}

	for _, tcase := range tcases {
		got := grammar.JustCapitalize(tcase.in)
		assert.Equal(t, tcase.expectOut, got)
	}
}

func TestSanitizePhone(t *testing.T) {
	tcases := []struct {
		in        string
		expectOut string
	}{
		{in: "0123456789", expectOut: "0123456789"},
		{in: "0123456789   ", expectOut: "0123456789"},
		{in: "+33123456789", expectOut: "+33123456789"},
		{in: "01.23.45.67.89", expectOut: "0123456789"},
		{in: "123-456-789", expectOut: "123456789"},
		{in: "01 23 45 67 89", expectOut: "0123456789"},
		{in: "(+33) 123-456-789", expectOut: "+33123456789"},
		{in: "(+33) 123 456 789", expectOut: "+33123456789"},
		{in: "01-23-45-67-89", expectOut: "0123456789"},
	}
	for _, tcase := range tcases {
		got := grammar.SanitizePhone(tcase.in)
		assert.Equal(t, tcase.expectOut, got)
	}
}
