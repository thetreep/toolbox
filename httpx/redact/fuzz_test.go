package redact_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/thetreep/toolbox/httpx/redact"
)

func fuzzPathRedactor() redact.Redactor {
	return redact.New(redact.Policy{
		Secrets:  redact.Rules{Keys: []string{"customer.name", "$.pin", "a.b.c", "s:driver.id"}},
		Volatile: redact.Rules{Keys: []string{"driver.seat", "$.nonce"}},
	}, redact.WithVolatile())
}

func jsonSeeds() []string {
	b := string(rune(92))

	return []string{
		``, `{}`, `[]`, `null`, `{"password":"x"}`, `{"a":{"token":[1,2,{"secret":null}]}}`,
		`{"password":"a` + b + `"b","z":1}`, `{"x":"a` + b + `"password` + b + `":` + b + `"y","password":"p"}`,
		`{"pass` + b + `u0077ord":"p"}`, `{"a":"` + b + `u00e9 ` + b + `n jean@example.com"}`,
		`{"name":"Zoé 日本 😀","token":"é"}`, `{"password":"hunter2`, `{"token":{"a":[1,2`, `{"a":1,}`, `[1 2]`,
		`{"mail":"jean@example.com","n":4111111111111111,"p":"06 12 34 56 78"}`, "{\"a\":\"\x01\"}", "\xff\xfe{",
		strings.Repeat("[", 5000), strings.Repeat(`{"a":`, 2000), `{"a":"` + b, `{"a":"` + b + `u12`,
		`-1.5e+10`, `{"a":-}`, ` { "secret" : true } `,
		`{"customer":{"name":"x"},"pin":1}`, `[{"customer":[{"name":"x"}],"pin":1}]`, `{"a":{"b":{"c":[1,{"c":2}]}}}`,
		`{"cust` + b + `u006fmer":{"name":"x"}}`, `{"customer":{"name":"x"`, `{"customer":[}{"name":"x"}`, `}]{"customer":{"name":1}}`,
		`{"driver":{"seat":{"a":[1]}},"nonce":"n"}`, strings.Repeat(`{"customer":`, 2000),
	}
}

func FuzzJSON(f *testing.F) {
	for _, s := range jsonSeeds() {
		f.Add(s)
	}

	redactors := []redact.Redactor{newRedactor(redact.WithVolatile()), fuzzPathRedactor(), newWithoutPII(redact.WithVolatile())}

	f.Fuzz(func(t *testing.T, in string) {
		for _, r := range redactors {
			res := r.Body("application/json", []byte(in))
			if len(in) == 0 {
				return
			}

			if res.ValidJSON != json.Valid([]byte(in)) {
				t.Fatalf("ValidJSON=%v but json.Valid=%v for %q", res.ValidJSON, json.Valid([]byte(in)), in)
			}

			if res.ValidJSON && !json.Valid(res.Body) {
				t.Fatalf("valid input produced invalid output %q -> %q", in, res.Body)
			}
		}
	})
}

func FuzzJSONKeyValue(f *testing.F) {
	for _, s := range jsonSeeds() {
		f.Add(s)
	}

	r := newRedactor()

	f.Fuzz(func(t *testing.T, value string) {
		if !json.Valid([]byte(value)) {
			return
		}

		res := r.Body("application/json", []byte(`{"a":1,"password":`+value+`,"b":[2]}`))

		const want = `{"a":1,"password":"REDACTED","b":[2]}`
		if string(res.Body) != want {
			t.Fatalf("value %q leaked: %q", value, res.Body)
		}
	})
}

func FuzzJSONPathValue(f *testing.F) {
	for _, s := range jsonSeeds() {
		f.Add(s)
	}

	r := fuzzPathRedactor()

	f.Fuzz(func(t *testing.T, value string) {
		if !json.Valid([]byte(value)) {
			return
		}

		res := r.Body("application/json", []byte(`{"a":1,"x":{"customer":[{"name":`+value+`}]},"pin":`+value+`,"b":[2]}`))

		const want = `{"a":1,"x":{"customer":[{"name":"REDACTED"}]},"pin":"REDACTED","b":[2]}`
		if string(res.Body) != want {
			t.Fatalf("value %q leaked: %q", value, res.Body)
		}
	})
}

func FuzzXML(f *testing.F) {
	for _, s := range []string{
		``, `<a/>`, `<a><token>x</token></a>`, `<a password="x" b='y'>t</a>`, `<![CDATA[x]]>`, `<!-- c -->`, `<?xml version="1.0"?>`,
		`<s:E xmlns:s="u"><ns:secret>p</ns:secret></s:E>`, `<a b=`, `<a b="`, `<token>x`, `</`, `<`, `<!DOCTYPE a [<!ENTITY x "y">]>`,
		strings.Repeat("<a>", 5000), `<a mail="jean@example.com">06 12 34 56 78</a>`, "<a>\xff</a>", `<a b=c d=e>`,
		`<customer name="x"><name>y</name></customer>`, `<a><customer><name>x</name></customer></a>`, `</a></a><customer name='x'/>`,
		`<driver seat="1"><seat>2</seat></driver>`, `<nonce>n</nonce>`, `<s:driver id="1"><s:id>2</s:id></s:driver>`, `<customer name=`, strings.Repeat("<customer>", 2000),
	} {
		f.Add(s)
	}

	redactors := []redact.Redactor{newRedactor(redact.WithVolatile()), fuzzPathRedactor(), newWithoutPII(redact.WithVolatile())}

	f.Fuzz(func(t *testing.T, in string) {
		for _, r := range redactors {
			res := r.Body("text/xml", []byte(in))
			if len(res.Body) > 4*len(in)+64 {
				t.Fatalf("output %d bytes for %d input bytes", len(res.Body), len(in))
			}
		}
	})
}

func FuzzXMLKeyContent(f *testing.F) {
	f.Add(`secret`)
	f.Add(`a & b jean@example.com`)

	r := newRedactor()

	f.Fuzz(func(t *testing.T, content string) {
		if strings.Contains(content, "<") {
			return
		}

		res := r.Body("text/xml", []byte(`<a><token>`+content+`</token><b>1</b></a>`))
		if string(res.Body) != `<a><token>REDACTED</token><b>1</b></a>` {
			t.Fatalf("content %q leaked: %q", content, res.Body)
		}
	})
}

func FuzzXMLPathContent(f *testing.F) {
	f.Add(`secret`)
	f.Add(`a & b jean@example.com`)

	r := fuzzPathRedactor()

	f.Fuzz(func(t *testing.T, content string) {
		if strings.Contains(content, "<") || strings.Contains(content, `"`) {
			return
		}

		res := r.Body("text/xml", []byte(`<a><customer name="`+content+`"><name>`+content+`</name></customer><name>1</name></a>`))
		if string(res.Body) != `<a><customer name="REDACTED"><name>REDACTED</name></customer><name>1</name></a>` {
			t.Fatalf("content %q leaked: %q", content, res.Body)
		}
	})
}

func FuzzBody(f *testing.F) {
	for _, s := range jsonSeeds() {
		f.Add(s, "application/json")
		f.Add(s, "")
	}

	f.Add(`a=b&password=c&%zz=1&=&&`, "application/x-www-form-urlencoded")
	f.Add(`jean@example.com 0612345678 4111111111111111 FR1420041010050500013M02606`, "text/plain")
	f.Add(`<a/>`, "text/html; charset=")

	redactors := []redact.Redactor{newRedactor(redact.WithVolatile()), fuzzPathRedactor(), newWithoutPII(redact.WithVolatile())}

	f.Fuzz(func(t *testing.T, in, contentType string) {
		for _, r := range redactors {
			r.Body(contentType, []byte(in))
		}
	})
}

func FuzzDetectors(f *testing.F) {
	for _, s := range []string{
		`a@b.fr`, `0612345678`, `+33 6 12 34 56 78`, `4111 1111 1111 1111`, `FR14 2004 1010 0505 0001 3M02 606`,
		`eyJhbGciOiJub25lIn0.eyJzdWIiOiIxIn0.`, `Basic dXNlcjpwYXNzd29yZA==`, `Bearer abcdef1234567890xyz`, `@@@@`, `++++`,
		`0033`, `00330`, strings.Repeat("1-", 100), strings.Repeat("a.", 100) + "@",
	} {
		f.Add(s)
	}

	r, noPII := newRedactor(), newWithoutPII()

	f.Fuzz(func(t *testing.T, in string) {
		out := r.Text(in)
		if len(redact.Scan([]byte(in))) == 0 && out != in {
			t.Fatalf("Text changed input without a finding: %q -> %q", in, out)
		}

		if strings.Contains(in, "[REDACTED:") {
			return
		}

		got := noPII.Text(in)
		for _, class := range []string{redact.ClassEmail, redact.ClassPhone, redact.ClassPAN, redact.ClassIBAN} {
			if strings.Contains(got, "[REDACTED:"+class+"]") {
				t.Fatalf("WithoutPII applied the %s detector: %q -> %q", class, in, got)
			}
		}
	})
}
