package grammar

import "testing"

func TestNormalize(t *testing.T) {
	tcases := []struct {
		in, expectOut string
	}{
		{"test", "test"},
		{"test1234", "test1234"},
		{"pierre-francois", "pierre francois"},
		{"Pierre-Francois", "pierre francois"},
	}
	for _, tcase := range tcases {
		if got, want := Normalize(tcase.in), tcase.expectOut; got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
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
		if got, want := Capitalize(tcase.in), tcase.expectOut; got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
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
		if got, want := JustCapitalize(tcase.in), tcase.expectOut; got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	}
}

func TestSanitizePhone(t *testing.T) {
	tcases := []struct {
		got  string
		want string
	}{
		{got: "0123456789", want: "0123456789"},
		{got: "0123456789   ", want: "0123456789"},
		{got: "+33123456789", want: "+33123456789"},
		{got: "01.23.45.67.89", want: "0123456789"},
		{got: "123-456-789", want: "123456789"},
		{got: "01 23 45 67 89", want: "0123456789"},
		{got: "(+33) 123-456-789", want: "+33123456789"},
		{got: "(+33) 123 456 789", want: "+33123456789"},
		{got: "01-23-45-67-89", want: "0123456789"},
	}
	for i, tcase := range tcases {
		if got, want := SanitizePhone(tcase.got), tcase.want; got != want {
			t.Fatalf("%d: got %s, want %s", i+1, got, want)
		}
	}
}
