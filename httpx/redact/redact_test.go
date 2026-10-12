package redact_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thetreep/toolbox/httpx/redact"
)

func newRedactor(opts ...redact.Option) *redact.Redactor {
	return redact.New(redact.Policy{
		PII:      redact.Rules{Keys: []string{"firstName"}, Headers: []string{"X-User"}},
		Volatile: redact.Rules{Keys: []string{"nonce"}, Headers: []string{"X-Trace"}, QueryParams: []string{"ts"}},
	}, opts...)
}

func TestJSONKeys(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"string", `{"password":"hunter2","a":1}`, `{"password":"REDACTED","a":1}`},
		{"case insensitive", `{"PassWord":"x"}`, `{"PassWord":"REDACTED"}`},
		{"number", `{"token": 12345}`, `{"token":"REDACTED"}`},
		{"bool", `{"secret": true, "x": false}`, `{"secret":"REDACTED","x":false}`},
		{"null", `{"secret": null}`, `{"secret":"REDACTED"}`},
		{"object", `{"token": {"a": [1, 2, {"b": "c"}]}, "z": 1}`, `{"token":"REDACTED","z":1}`},
		{"array", `{"token": [1, "x"], "z": 1}`, `{"token":"REDACTED","z":1}`},
		{"nested", `{"a":{"b":[{"api_key":"k"}]}}`, `{"a":{"b":[{"api_key":"REDACTED"}]}}`},
		{"escaped quote in value", `{"password":"a\"b","z":"ok"}`, `{"password":"REDACTED","z":"ok"}`},
		{"escaped quote before key", `{"x":"a\"password\":\"y","password":"p"}`, `{"x":"a\"password\":\"y","password":"REDACTED"}`},
		{"escaped key", `{"pass\u0077ord":"p"}`, `{"pass\u0077ord":"REDACTED"}`},
		{"key name in value is not a key", `{"a":"password"}`, `{"a":"password"}`},
		{"pii key", `{"firstName":"Jean"}`, `{"firstName":"REDACTED"}`},
		{"volatile off", `{"nonce":"abc"}`, `{"nonce":"abc"}`},
		{"top-level array", ` [ {"token":"x"} , 2 ] `, `[{"token":"REDACTED"},2]`},
		{"unicode", `{"name":"Zoé 日本","token":"é"}`, `{"name":"Zoé 日本","token":"REDACTED"}`},
		{"empty containers", `{"a":{},"b":[],"token":{}}`, `{"a":{},"b":[],"token":"REDACTED"}`},
	}

	r := newRedactor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := r.Body("application/json", []byte(tt.in))
			assert.Equal(t, tt.want, string(res.Body))
			assert.True(t, res.ValidJSON)
			assert.True(t, json.Valid(res.Body))
		})
	}
}

func TestJSONVolatile(t *testing.T) {
	res := newRedactor(redact.WithVolatile()).Body("application/json", []byte(`{"nonce":"abc","password":"p"}`))
	assert.JSONEq(t, `{"nonce":"VOLATILE","password":"REDACTED"}`, string(res.Body))
}

func TestJSONValidity(t *testing.T) {
	invalid := []string{
		``, ` `, `{`, `[1,`, `{"a":1,}`, `{"a" 1}`, `{"a":1 "b":2}`, `[1 2]`, `{"a":01}`, `nul`, `{"a":"x`,
		`{'a':1}`, "{\"a\":\"\x01\"}", `{"a":"\x"}`, `{"a":"\u12"}`, `1 2`, `]`, `{"a":1}}`, `{"a":-}`,
		`{"token":"x"`,
	}

	r := newRedactor()

	for _, in := range invalid {
		res := r.Body("application/json", []byte(in))
		assert.False(t, res.ValidJSON, "%q", in)
	}

	for _, in := range []string{`{}`, `[]`, `1`, `"x"`, `null`, `-1.5e+10`, `{"a":[true,false,null,"\u00e9\n"]}`} {
		assert.True(t, r.Body("application/json", []byte(in)).ValidJSON, "%q", in)
	}
}

func TestTruncatedJSONStillRedacts(t *testing.T) {
	res := newRedactor().Body("application/json", []byte(`{"a":[1,2],"password":"hunter2","b":{"token":"abc`))
	assert.NotContains(t, string(res.Body), "hunter2")
	assert.NotContains(t, string(res.Body), "abc")
	assert.False(t, res.ValidJSON)
}

func TestJSONDetectorsOnStringsOnly(t *testing.T) {
	in := `{"mail":"jean.dupont@example.com","n":4111111111111111,"s":"card 4111 1111 1111 1111 ok","phone":"06 12 34 56 78"}`
	res := newRedactor().Body("application/json", []byte(in))

	assert.JSONEq(t,
		`{"mail":"[REDACTED:email]","n":4111111111111111,"s":"card [REDACTED:pan] ok","phone":"[REDACTED:phone]"}`,
		string(res.Body),
	)
	assert.True(t, json.Valid(res.Body))
}

func TestJSONDetectorsKeepEscapesValid(t *testing.T) {
	for _, in := range []string{
		`{"a":"x\nuser@example.com"}`,
		`{"a":"\u00e9mile@example.com"}`,
		`{"a":"line\n0612345678\n"}`,
		`{"a":"\"jean@example.com\""}`,
		`{"a":"\\jean@example.com"}`,
	} {
		res := newRedactor().Body("application/json", []byte(in))
		assert.True(t, json.Valid(res.Body), "%s -> %s", in, res.Body)
		assert.NotContains(t, string(res.Body), "example.com", in)
	}
}

func TestXML(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{
			"element text",
			`<a><Password>hunter2</Password><b>ok</b></a>`,
			`<a><Password>REDACTED</Password><b>ok</b></a>`,
		},
		{
			"element with children",
			`<a><token><x>1</x><y>2</y></token><b>ok</b></a>`,
			`<a><token>REDACTED</token><b>ok</b></a>`,
		},
		{
			"namespace prefix",
			`<s:Envelope xmlns:s="u"><s:Body><ns:password>p</ns:password></s:Body></s:Envelope>`,
			`<s:Envelope xmlns:s="u"><s:Body><ns:password>REDACTED</ns:password></s:Body></s:Envelope>`,
		},
		{
			"attribute double and single quotes",
			`<login user="bob" password="p1" token='t1'/>`,
			`<login user="bob" password="REDACTED" token='REDACTED'/>`,
		},
		{
			"attribute on element with text",
			`<a apiKey="k">text</a>`,
			`<a apiKey="REDACTED">text</a>`,
		},
		{
			"cdata and comments are text",
			`<a><![CDATA[mail jean@example.com]]><!-- jean@example.com --></a>`,
			`<a><![CDATA[mail [REDACTED:email]]]><!-- [REDACTED:email] --></a>`,
		},
		{
			"detectors on text and attributes",
			`<a mail="jean@example.com">06 12 34 56 78</a>`,
			`<a mail="[REDACTED:email]">[REDACTED:phone]</a>`,
		},
		{
			"self closing matched element",
			`<a><token/><b>1</b></a>`,
			`<a><token/><b>1</b></a>`,
		},
		{
			"declaration kept",
			`<?xml version="1.0"?><a><secret>s</secret></a>`,
			`<?xml version="1.0"?><a><secret>REDACTED</secret></a>`,
		},
		{
			"nested same name",
			`<token><token>a</token>b</token><c/>`,
			`<token>REDACTED</token><c/>`,
		},
	}

	r := newRedactor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := r.Body("text/xml; charset=utf-8", []byte(tt.in))
			assert.Equal(t, tt.want, string(res.Body))
			assert.Equal(t, redact.KindXML, res.Kind)
		})
	}
}

func TestForm(t *testing.T) {
	res := newRedactor().Body(
		"application/x-www-form-urlencoded",
		[]byte(`grant_type=password&Client_Secret=abc%20d&mail=jean%40example.com&q=a+b&flag`),
	)
	assert.Equal(t, redact.KindForm, res.Kind)
	assert.Equal(t,
		`grant_type=password&Client_Secret=REDACTED&mail=%5BREDACTED%3Aemail%5D&q=a+b&flag`,
		string(res.Body),
	)
}

func TestSniffing(t *testing.T) {
	r := newRedactor()

	json1 := r.Body("", []byte(` {"password":"x"}`))
	assert.Equal(t, redact.KindJSON, json1.Kind)
	assert.Contains(t, string(json1.Body), "REDACTED")

	// A text/plain body that happens to be JSON still gets its keys redacted.
	json2 := r.Body("text/plain", []byte(`{"password":"x"}`))
	assert.Equal(t, `{"password":"REDACTED"}`, string(json2.Body))

	// Invalid sniffed JSON falls back to plain text handling.
	text := r.Body("text/plain", []byte(`[INFO] mail jean@example.com`))
	assert.Equal(t, redact.KindText, text.Kind)
	assert.Equal(t, `[INFO] mail [REDACTED:email]`, string(text.Body))

	xml1 := r.Body("", []byte(`<a><secret>x</secret></a>`))
	assert.Equal(t, redact.KindXML, xml1.Kind)
}

func TestEmptyBody(t *testing.T) {
	res := newRedactor().Body("application/json", nil)
	assert.Empty(t, res.Body)
	assert.NotNil(t, res.Body)
}

func TestHeaderAndQuery(t *testing.T) {
	r := newRedactor()

	h := http.Header{
		"Authorization":  {"Bearer abcdefghij1234567890"},
		"x-goog-api-key": {"k"},
		"X-User":         {"bob"},
		"X-Trace":        {"abc"},
		"Accept":         {"a", "b"},
		"X-Mail":         {"jean@example.com"},
	}
	got := r.Header(h)

	assert.Equal(t, []string{"REDACTED"}, got["Authorization"])
	assert.Equal(t, []string{"REDACTED"}, got["x-goog-api-key"])
	assert.Equal(t, []string{"REDACTED"}, got["X-User"])
	assert.Equal(t, []string{"abc"}, got["X-Trace"])
	assert.Equal(t, []string{"a", "b"}, got["Accept"])
	assert.Equal(t, []string{"[REDACTED:email]"}, got["X-Mail"])
	assert.Equal(t, []string{"Bearer abcdefghij1234567890"}, h["Authorization"], "input must not be mutated")

	vol := newRedactor(redact.WithVolatile()).Header(h)
	assert.Equal(t, []string{"VOLATILE"}, vol["X-Trace"])

	q := r.Query(url.Values{"key": {"k"}, "TS": {"1"}, "q": {"paris"}})
	assert.Equal(t, url.Values{"key": {"REDACTED"}, "TS": {"1"}, "q": {"paris"}}, q)
	assert.Equal(t, "key=REDACTED&q=paris&ts=VOLATILE",
		newRedactor(redact.WithVolatile()).RawQuery("q=paris&key=abc&ts=17"))
}

func TestDefaultsCannotBeDropped(t *testing.T) {
	r := redact.New(redact.Policy{})
	assert.Equal(t, []string{"REDACTED"}, r.Header(http.Header{"Set-Cookie": {"a=b"}})["Set-Cookie"])
	assert.Contains(t, string(r.Body("application/json", []byte(`{"client_secret":"s"}`)).Body), "REDACTED")
}

func TestMergeIsAppendOnly(t *testing.T) {
	a := redact.Policy{Secrets: redact.Rules{Keys: []string{"a"}}}
	b := redact.Policy{Secrets: redact.Rules{Keys: []string{"b"}}, PII: redact.Rules{Headers: []string{"h"}}}
	got := redact.Merge(a, b)

	assert.Equal(t, []string{"a", "b"}, got.Secrets.Keys)
	assert.Equal(t, []string{"h"}, got.PII.Headers)
	assert.Equal(t, []string{"a"}, a.Secrets.Keys)
}

func TestDetectors(t *testing.T) {
	positives := map[string][]string{
		redact.ClassEmail: {"jean.dupont+tag@sub.example.co.uk", "mail: a@b.fr.", "(x_y@z-w.com)"},
		redact.ClassPhone: {
			"0612345678", "06 12 34 56 78", "06.12.34.56.78", "+33612345678", "+33 6 12 34 56 78", "+14155552671",
			"0033 6 12 34 56 78", "+44 20 7946 0958",
		},
		redact.ClassPAN:  {"4111111111111111", "4111 1111 1111 1111", "4111-1111-1111-1111", "378282246310005"},
		redact.ClassIBAN: {"FR1420041010050500013M02606", "FR14 2004 1010 0505 0001 3M02 606", "GB82WEST12345698765432"},
		redact.ClassJWT: {
			"eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			"eyJhbGciOiJub25lIn0.eyJzdWIiOiIxIn0.",
		},
		redact.ClassBasic:  {"Basic dXNlcjpwYXNzd29yZA==", "basic dXNlcjpwYXNzd29yZA=="},
		redact.ClassBearer: {"Bearer abcdef1234567890xyz", "bearer ya29.a0AfH6SMBx-12345678901234"},
	}

	for class, inputs := range positives {
		for _, in := range inputs {
			assert.Contains(t, redact.Scan([]byte("x "+in+" y")), class, in)
			assert.Contains(t, redact.Scan([]byte(in)), class, in)
		}
	}

	negatives := []string{
		"", "hello world", "2026-04-21T20:42:42.000000042+07:00", "4111111111111112", "1234567890123456",
		"550e8400-e29b-41d4-a716-446655440000", "FR1420041010050500013M02607", "0123", "+33", "+1 2",
		"user@localhost", "a@b.c", "see @mention", "Basic information", "Bearer authentication", "Bearer abcdef",
		"version 1.2.3", "12345678901234567890", "ab0612345678", "0612345678ab", "eyJhbGciOiJIUzI1NiJ9",
		"Basic dXNlcg==", "0000000001", "price 1299.99 EUR",
	}

	for _, in := range negatives {
		assert.Empty(t, redact.Scan([]byte(in)), "%q", in)
	}
}

func TestDetectorsRewrite(t *testing.T) {
	in := "a@b.fr, 0612345678, 4111111111111111 and FR1420041010050500013M02606 / Bearer abcdef1234567890xyz."
	want := "[REDACTED:email], [REDACTED:phone], [REDACTED:pan] and [REDACTED:iban] / [REDACTED:bearer]."
	assert.Equal(t, want, newRedactor().Text(in))
	assert.Equal(t, "nothing here", newRedactor().Text("nothing here"))
}

func TestNeverPanics(t *testing.T) {
	r := newRedactor(redact.WithVolatile())
	deep := strings.Repeat("[", 200000)
	inputs := []string{
		"", "{", "}", `"`, `\`, `{"a":`, `{"token":`, `{"token":"\`, `{"token":"\u12`, `[[[[`, `]]]]`, `<`, `<a`, `<a b`, `<a b=`,
		`<a b="`, `<!--`, `<![CDATA[`, `<?`, `</`, `<token>`, `<token>x`, "\xff\xfe", "{\"a\":\"\xff\"}", `%`, `&&&=`, `=`, deep,
		strings.Repeat("{\"a\":", 50000), strings.Repeat("<a>", 50000), strings.Repeat("a@", 10000),
		strings.Repeat("1 ", 100000), strings.Repeat("Bearer ", 10000),
	}

	for _, in := range inputs {
		for _, ct := range []string{"application/json", "text/xml", "application/x-www-form-urlencoded", "text/plain", ""} {
			require.NotPanics(t, func() { r.Body(ct, []byte(in)) })
		}
	}
}
