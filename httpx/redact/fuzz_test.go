package redact_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/thetreep/toolbox/httpx/redact"
)

func fuzzPathRedactor() redact.Redactor {
	return redact.New(redact.Policy{
		Secrets:  redact.Rules{Keys: []string{"customer.name", "$.pin", "a.b.c", "s:driver.id"}},
		Volatile: redact.Rules{Keys: []string{"driver.seat", "$.nonce"}},
	}, redact.WithVolatile())
}

func fuzzEmbeddedRedactor() redact.Redactor {
	return redact.New(redact.Policy{
		PII:      redact.Rules{Keys: []string{"lastName", "customer.name"}},
		Volatile: redact.Rules{Keys: []string{"nonce"}},
		Embedded: []string{"payload", "$.wrapper.doc", "customer.doc"},
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
		`{"payload":"{` + b + `"lastName` + b + `":` + b + `"x` + b + `"}"}`, `{"payload":"<a><lastName>x</lastName></a>"}`, `{"payload":"{` + b + `"payload` + b + `":` + b + `"{}` + b + `"}"}`,
		`{"payload":"{` + b + `"a` + b + `":"`, `{"payload":"` + b + `u0000"}`, `{"payload":12}`, `{"payload":{"lastName":1}}`, `{"payload":null}`,
		`{"wrapper":{"doc":"[1,{` + b + `"lastName` + b + `":2}]"}}`, `{"customer":[{"doc":"<a b=` + b + `"c` + b + `"/>"}]}`, strings.Repeat(`{"payload":`, 2000),
	}
}

func FuzzJSON(f *testing.F) {
	for _, s := range jsonSeeds() {
		f.Add(s)
	}

	redactors := []redact.Redactor{newRedactor(redact.WithVolatile()), fuzzPathRedactor(), newWithoutPII(redact.WithVolatile()), fuzzEmbeddedRedactor()}

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
		`<payload>&lt;a&gt;&lt;lastName&gt;x&lt;/lastName&gt;&lt;/a&gt;</payload>`, `<payload><![CDATA[{"lastName":"x"}]]></payload>`, `<payload>&lt;payload&gt;&amp;lt;a/&amp;gt;&lt;/payload&gt;</payload>`,
		`<payload>&#60;a/&#62;</payload>`, `<payload>&#0;</payload>`, `<payload>&`, `<payload><![CDATA[`, `<payload><b/></payload>`, `<wrapper><doc>{}</doc></wrapper>`, `<customer><doc>&lt;a&gt;</doc></customer>`,
		strings.Repeat("<payload>&lt;", 2000), strings.Repeat("<a>", 5000), `<a mail="jean@example.com">06 12 34 56 78</a>`, "<a>\xff</a>", `<a b=c d=e>`,
		`<customer name="x"><name>y</name></customer>`, `<a><customer><name>x</name></customer></a>`, `</a></a><customer name='x'/>`,
		`<driver seat="1"><seat>2</seat></driver>`, `<nonce>n</nonce>`, `<s:driver id="1"><s:id>2</s:id></s:driver>`, `<customer name=`, strings.Repeat("<customer>", 2000),
	} {
		f.Add(s)
	}

	redactors := []redact.Redactor{newRedactor(redact.WithVolatile()), fuzzPathRedactor(), newWithoutPII(redact.WithVolatile()), fuzzEmbeddedRedactor()}

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

	redactors := []redact.Redactor{newRedactor(redact.WithVolatile()), fuzzPathRedactor(), newWithoutPII(redact.WithVolatile()), fuzzEmbeddedRedactor()}

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

func FuzzEmbeddedJSONValue(f *testing.F) {
	f.Add(`secret`)
	f.Add(`a "quoted" \ value` + "\n\t")
	f.Add(`</payload> jean@example.com <![CDATA[`)

	r := fuzzEmbeddedRedactor()

	f.Fuzz(func(t *testing.T, value string) {
		inner, _ := json.Marshal(map[string]any{"lastName": value, "n": 1})
		outer, _ := json.Marshal(map[string]any{"payload": string(inner)})

		res := r.Body("application/json", outer)
		if !res.ValidJSON || !json.Valid(res.Body) {
			t.Fatalf("invalid output %q for %q", res.Body, outer)
		}

		var got struct{ Payload string }
		if err := json.Unmarshal(res.Body, &got); err != nil {
			t.Fatalf("payload is no string in %q: %v", res.Body, err)
		}

		if got.Payload != `{"lastName":"REDACTED","n":1}` {
			t.Fatalf("value %q leaked: %q", value, got.Payload)
		}
	})
}

func FuzzEmbeddedXMLContent(f *testing.F) {
	f.Add(`secret`)
	f.Add(`a & b <c> jean@example.com ]]>`)
	f.Add("\x00\xff")

	r := fuzzEmbeddedRedactor()

	f.Fuzz(func(t *testing.T, content string) {
		var inner, outer bytes.Buffer

		inner.WriteString(`<a><lastName>`)
		_ = xml.EscapeText(&inner, []byte(content))
		inner.WriteString(`</lastName><n>1</n></a>`)

		outer.WriteString(`<e><payload>`)
		_ = xml.EscapeText(&outer, inner.Bytes())
		outer.WriteString(`</payload></e>`)

		res := r.Body("text/xml", outer.Bytes())

		var got struct {
			Payload string `xml:"payload"`
		}

		if err := xml.Unmarshal(res.Body, &got); err != nil {
			t.Fatalf("invalid output %q for %q: %v", res.Body, outer.Bytes(), err)
		}

		if got.Payload != `<a><lastName>REDACTED</lastName><n>1</n></a>` {
			t.Fatalf("content %q leaked: %q", content, got.Payload)
		}
	})
}

// An embedded document is either masked or redacted exactly as the same document sent on its own.
func FuzzEmbeddedDifferential(f *testing.F) {
	f.Add(`{"lastName":"Dupont","city":"Lyon"}`)
	f.Add(` [{"token":"x"},{"lastName":"a"}] `)
	f.Add(`<o><lastName>Dupont</lastName><t>jean@example.com</t></o>`)
	f.Add(`<?xml version="1.0"?><o a="1"><![CDATA[x]]><lastName/></o>`)
	f.Add(`<o>&amp;nbsp;</o>`)

	r := fuzzEmbeddedRedactor()
	escape := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

	f.Fuzz(func(t *testing.T, doc string) {
		if !utf8.ValidString(doc) || strings.TrimSpace(doc) == "" || strings.Contains(strings.ToLower(doc), "payload") {
			return
		}

		var ct string

		switch strings.TrimSpace(doc)[0] {
		case '{', '[':
			ct = "application/json"
		case '<':
			ct = "application/xml"
		default:
			return
		}

		alone := r.Body(ct, []byte(doc))

		outer, _ := json.Marshal(map[string]string{"payload": doc})

		var inJSON struct{ Payload string }
		if err := json.Unmarshal(r.Body("application/json", outer).Body, &inJSON); err != nil {
			t.Fatalf("invalid JSON envelope for %q: %v", doc, err)
		}

		if ct == "application/json" && !alone.ValidJSON {
			if inJSON.Payload != redact.Placeholder {
				t.Fatalf("invalid JSON document %q not masked: %q", doc, inJSON.Payload)
			}

			return
		}

		if inJSON.Payload != redact.Placeholder && !sameDocument(ct, inJSON.Payload, alone.Body) {
			t.Fatalf("JSON envelope: got %q, alone %q", inJSON.Payload, alone.Body)
		}

		var inXML struct {
			Payload string `xml:"payload"`
		}
		envelope := []byte("<r><payload>" + escape.Replace(doc) + "</payload></r>")
		// XML parsers turn CR into LF, so the decoded payload would differ for a reason unrelated to redaction.
		if strings.Contains(doc, "\r") || xml.Unmarshal(envelope, &inXML) != nil {
			return
		}

		if err := xml.Unmarshal(r.Body("application/xml", envelope).Body, &inXML); err != nil {
			t.Fatalf("invalid XML envelope for %q: %v", doc, err)
		}

		if inXML.Payload != redact.Placeholder && !sameDocument(ct, inXML.Payload, alone.Body) {
			t.Fatalf("XML envelope: got %q, alone %q", inXML.Payload, alone.Body)
		}
	})
}

func sameDocument(ct, nested string, alone []byte) bool {
	if ct != "application/json" {
		return strings.TrimSpace(nested) == strings.TrimSpace(string(alone))
	}

	var a, b bytes.Buffer

	return json.Compact(&a, []byte(nested)) == nil && json.Compact(&b, alone) == nil && a.String() == b.String()
}
