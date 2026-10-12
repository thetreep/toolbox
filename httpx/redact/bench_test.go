package redact_test

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
)

func largeJSON() []byte {
	var b bytes.Buffer

	b.WriteString(`{"offers":[`)

	for i := 0; i < 2400; i++ {
		if i > 0 {
			b.WriteByte(',')
		}

		fmt.Fprintf(&b, `{"id":"offer-%d","price":{"amount":%d.50,"currency":"EUR"},"token":"tok-%d",`+
			`"segments":[{"from":"Paris Gare de Lyon","to":"Lyon Part-Dieu","departure":"2026-04-21T08:%02d:00+02:00",`+
			`"carriage":{"seats":[{"n":1,"free":true},{"n":2,"free":false},{"n":3,"free":true}]}}],`+
			`"conditions":"Billet echangeable et remboursable jusqu'a la veille du depart, frais de dossier inclus.",`+
			`"contact":{"email":"traveler%d@example.com","phone":"06 12 34 56 %02d"}}`, i, 40+i%50, i, i%60, i, i%100)
	}

	b.WriteString(`]}`)

	return b.Bytes()
}

var smallJSON = []byte(`{"username":"jean","password":"hunter2","origin":{"latLng":{"latitude":48.8566,"longitude":2.3522}},` +
	`"departureTime":"2026-04-21T20:42:42+02:00","contact":"jean.dupont@example.com"}`)

// v1Baseline mimics the former pipeline: one regex pass per field and per detector over the whole body.
func v1Baseline() func([]byte) []byte {
	var passes []*regexp.Regexp

	for _, key := range []string{"password", "token", "access_token", "refresh_token", "id_token", "client_secret", "secret", "apiKey", "api_key"} {
		passes = append(passes, regexp.MustCompile(`("`+key+`":\s*)("[^"]*"|[0-9\.]+|null)`))
	}

	for _, expr := range []string{
		`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`,
		`(?:\+33|0033|0)\s?[1-9](?:[ .-]?\d{2}){4}`,
		`\b(?:\d[ -]?){13,19}\b`,
		`\b[A-Z]{2}\d{2}(?: ?[A-Z0-9]{4}){2,7} ?[A-Z0-9]{1,4}\b`,
		`eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]*`,
		`(?i)\b(?:bearer|basic) +[A-Za-z0-9._~+/=-]{8,}`,
	} {
		passes = append(passes, regexp.MustCompile(expr))
	}

	return func(b []byte) []byte {
		for _, re := range passes {
			b = re.ReplaceAll(b, []byte("${1}REDACTED"))
		}

		return b
	}
}

func benchBody(b *testing.B, body []byte, fn func([]byte)) {
	b.SetBytes(int64(len(body)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fn(body)
	}
}

func BenchmarkLargeJSON(b *testing.B) {
	body := largeJSON()
	r := newRedactor()

	b.Run("v2", func(b *testing.B) {
		benchBody(b, body, func(in []byte) { r.Body("application/json", in) })
	})
	b.Run("v1-style-regex", func(b *testing.B) {
		fn := v1Baseline()
		benchBody(b, body, func(in []byte) { fn(in) })
	})
}

func BenchmarkSmallJSON(b *testing.B) {
	r := newRedactor()

	b.Run("v2", func(b *testing.B) {
		benchBody(b, smallJSON, func(in []byte) { r.Body("application/json", in) })
	})
	b.Run("v1-style-regex", func(b *testing.B) {
		fn := v1Baseline()
		benchBody(b, smallJSON, func(in []byte) { fn(in) })
	})
}

func BenchmarkLargeJSONCleanStrings(b *testing.B) {
	body := bytes.ReplaceAll(largeJSON(), []byte("example.com"), []byte("example"))
	r := newRedactor()

	benchBody(b, body, func(in []byte) { r.Body("application/json", in) })
}

func TestLargeJSONIsAtLeastOneMB(t *testing.T) {
	if n := len(largeJSON()); n < 1<<20 {
		t.Fatalf("benchmark payload is only %d bytes", n)
	}
}
