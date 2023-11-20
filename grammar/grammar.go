package grammar

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	reg            = regexp.MustCompile("[^a-zA-Z0-9]+")
	phoneSanitizer = strings.NewReplacer(" ", "", ".", "", "_", "", "(", "", ")", "", "-", "")
)

func Normalizer() transform.Transformer {
	return transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
}

// Capitalize sets the first letter in upper case and all the others in lower case.
func Capitalize(s string) string {
	if s == "" {
		return ""
	}

	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// JustCapitalize lize sets the first letter in upper case without changes the others.
func JustCapitalize(s string) string {
	if s == "" {
		return ""
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

// SanitizePhone removes unexpected character of a phone string.
func SanitizePhone(phone string) string {
	return strings.TrimSpace(phoneSanitizer.Replace(phone))
}

// Normalize normalizes a string by replacing special letter by its normalized version (e.g. : `é` -> `e`).
func Normalize(str string) string {
	out, _, _ := transform.String(Normalizer(), str)

	return strings.ToLower(strings.TrimSpace(reg.ReplaceAllString(out, " ")))
}

// EqualNorm compare two strings with their lower case and normalized version.
func EqualNorm(str1, str2 string) bool {
	return strings.EqualFold(Normalize(str1), Normalize(str2))
}

// ContainsNorm checks is a normalized lower string contains another normalized string.
func ContainsNorm(str1, str2 string) bool {
	return strings.Contains(
		strings.ToLower(Normalize(str1)),
		strings.ToLower(Normalize(str2)))
}
