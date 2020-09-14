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

func TestEqualNorm(t *testing.T) {
	tcases := []struct {
		s1, s2       string
		shoudBeEqual bool
	}{
		{"test", "test", true},
		{"test1234", "TEST1234", true},
		{"test1234", "TEST", false},
		{"PieRre-francois", "pierre francois", true},
		{"André-Niño", "ANDRE-NINO", true},
		{"André-Niño", "Andre ninho", false},
	}
	for _, tcase := range tcases {
		got := grammar.EqualNorm(tcase.s1, tcase.s2)
		if tcase.shoudBeEqual {
			assert.True(t, got)
		} else {
			assert.False(t, got)
		}
	}
}

func TestRandStringRunes(t *testing.T) {
	tcases := []struct {
		n, expLen int
	}{
		{0, 0},
		{1, 1},
		{10, 10},
		{33, 33},
		{200, 200},
		{5555, 5555},
		{123456, 123456},
	}
	for _, tcase := range tcases {
		assert.Len(t, grammar.RandStringRunes(tcase.n), tcase.expLen)
	}
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
