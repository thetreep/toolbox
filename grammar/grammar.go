package grammar

import (
	"math/rand"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	empty = ""
	tab   = "\t"
)

var (
	normalizer     = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	reg, _         = regexp.Compile("[^a-zA-Z0-9]+")
	phoneSanitizer = strings.NewReplacer(" ", "", ".", "", "_", "", "(", "", ")", "", "-", "")
	letterRunes    = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

//RandStringRunes generates a random n-lenght string
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

//Capitalize sets the first letter in upper case and all the others in lower case
func Capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

//JustCapitalize lize sets the first letter in upper case without changes the others
func JustCapitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

//SanitizePhone removes unexpected character of a phone string
func SanitizePhone(phone string) string {
	return strings.TrimSpace(phoneSanitizer.Replace(phone))
}

//Normalize normalizes a string by replacing special letter by its normalized version (e.g. : `é` -> `e`)
func Normalize(str string) string {
	out, _, _ := transform.String(normalizer, strings.ToLower(str))
	return strings.TrimSpace(reg.ReplaceAllString(out, " "))
}

//EqualNorm compare two strings with their lower case and normalized version
func EqualNorm(str1, str2 string) bool {
	return strings.EqualFold(Normalize(str1), Normalize(str2))
}
