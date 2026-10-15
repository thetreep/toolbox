package redact_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thetreep/toolbox/httpx/redact"
)

func newEmbedded(opts ...redact.Option) redact.Redactor {
	return redact.New(redact.Policy{
		PII:      redact.Rules{Keys: []string{"lastName", "passenger.phone"}},
		Volatile: redact.Rules{Keys: []string{"nonce"}},
		Embedded: []string{"payload", "$.wrapper.doc", "fileContent"},
	}, opts...)
}

func jsonString(s string) string {
	var b strings.Builder

	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)

	return strings.TrimSpace(b.String())
}

func TestEmbeddedJSONInJSON(t *testing.T) {
	inner := `{"token":"abc","lastName":"Dupont","amount":12,"passenger":{"phone":"0612"},"contact":"jean@example.com"}`
	in := `{"id":7,"payload":` + jsonString(inner) + `,"lastName":"Outer"}`

	res := newEmbedded().Body("application/json", []byte(in))

	assert.True(t, res.ValidJSON)
	assert.JSONEq(t, `{"id":7,"lastName":"REDACTED","payload":`+jsonString(
		`{"token":"REDACTED","lastName":"REDACTED","amount":12,"passenger":{"phone":"REDACTED"},"contact":"[REDACTED:email]"}`)+`}`, string(res.Body))
}

func TestEmbeddedKeepsNestedFormatting(t *testing.T) {
	inner := "{\n  \"lastName\": \"Dupont\",\n  \"city\": \"Lyon\"\n}"

	res := newEmbedded().Body("application/json", []byte(`{"payload":`+jsonString(inner)+`}`))

	assert.JSONEq(t, `{"payload":`+jsonString("{\n  \"lastName\": \"REDACTED\",\n  \"city\": \"Lyon\"\n}")+`}`, string(res.Body))
}

func TestEmbeddedXMLInJSON(t *testing.T) {
	inner := `<?xml version="1.0"?><order><lastName>Dupont</lastName><price>9</price></order>`

	res := newEmbedded().Body("application/json", []byte(`{"payload":`+jsonString(inner)+`}`))

	assert.JSONEq(t, `{"payload":`+jsonString(`<?xml version="1.0"?><order><lastName>REDACTED</lastName><price>9</price></order>`)+`}`, string(res.Body))
}

func TestEmbeddedJSONInXML(t *testing.T) {
	in := `<Envelope><payload>{&quot;lastName&quot;:&quot;Dupont&quot;,&quot;price&quot;:9}</payload></Envelope>`

	res := newEmbedded().Body("text/xml", []byte(in))

	assert.Equal(t, `<Envelope><payload>{"lastName":"REDACTED","price":9}</payload></Envelope>`, string(res.Body))
}

func TestEmbeddedXMLInXML(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{
			"entities",
			`<Envelope><fileContent>&lt;order ref=&quot;R1&quot;&gt;&lt;lastName&gt;Dupont&lt;/lastName&gt;&lt;note&gt;a &amp;amp; b&lt;/note&gt;&lt;/order&gt;</fileContent></Envelope>`,
			`<Envelope><fileContent>&lt;order ref="R1"&gt;&lt;lastName&gt;REDACTED&lt;/lastName&gt;&lt;note&gt;a &amp;amp; b&lt;/note&gt;&lt;/order&gt;</fileContent></Envelope>`,
		},
		{
			"numeric references",
			`<a><payload>&#60;o&#x3E;&#60;lastName&#62;Dupont&#60;/lastName&#62;&#60;/o&#x3E;</payload></a>`,
			`<a><payload>&lt;o&gt;&lt;lastName&gt;REDACTED&lt;/lastName&gt;&lt;/o&gt;</payload></a>`,
		},
		{
			"cdata",
			`<a><payload><![CDATA[<o><lastName>Dupont</lastName><price>9</price></o>]]></payload></a>`,
			`<a><payload>&lt;o&gt;&lt;lastName&gt;REDACTED&lt;/lastName&gt;&lt;price&gt;9&lt;/price&gt;&lt;/o&gt;</payload></a>`,
		},
		{
			"text around cdata",
			`<a><payload> <![CDATA[<o><lastName>D</lastName>]]>&lt;/o&gt;</payload></a>`,
			`<a><payload> &lt;o&gt;&lt;lastName&gt;REDACTED&lt;/lastName&gt;&lt;/o&gt;</payload></a>`,
		},
		{
			"namespaced element and attributes",
			`<s:Envelope xmlns:s="u"><s:Body><ns:payload a="1">&lt;x:o xmlns:x="v" lastName="Dupont"/&gt;</ns:payload></s:Body></s:Envelope>`,
			`<s:Envelope xmlns:s="u"><s:Body><ns:payload a="1">&lt;x:o xmlns:x="v" lastName="REDACTED"/&gt;</ns:payload></s:Body></s:Envelope>`,
		},
		{"empty element", `<a><payload></payload><b/></a>`, `<a><payload></payload><b/></a>`},
		{"self closing", `<a><payload/><b>1</b></a>`, `<a><payload/><b>1</b></a>`},
		{"whitespace only", "<a><payload>\n  </payload></a>", "<a><payload>\n  </payload></a>"},
	}

	r := newEmbedded()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(r.Body("text/xml", []byte(tt.in)).Body))
		})
	}
}

func TestEmbeddedDepthLimit(t *testing.T) {
	wrap := func(levels int) string {
		doc := `{"lastName":"Dupont","ok":1}`
		for range levels {
			doc = `{"payload":` + jsonString(doc) + `}`
		}

		return doc
	}

	unwrap := func(t *testing.T, doc string, levels int) string {
		t.Helper()

		for range levels {
			var m map[string]string

			require.NoError(t, json.Unmarshal([]byte(doc), &m))
			doc = m["payload"]
		}

		return doc
	}

	r := newEmbedded()

	got := string(r.Body("application/json", []byte(wrap(3))).Body)
	assert.JSONEq(t, `{"lastName":"REDACTED","ok":1}`, unwrap(t, got, 3))

	got = string(r.Body("application/json", []byte(wrap(4))).Body)
	assert.JSONEq(t, `{"payload":"REDACTED"}`, unwrap(t, got, 3))
	assert.NotContains(t, got, "Dupont")
}

func TestEmbeddedFailsClosed(t *testing.T) {
	tests := []struct{ name, in, ctype string }{
		{"truncated json", `{"payload":"{\"lastName\":\"Dupont\""}`, "application/json"},
		{"trailing json garbage", `{"payload":"{\"lastName\":\"Dupont\"} x"}`, "application/json"},
		{"json scalar", `{"payload":"\"Dupont\""}`, "application/json"},
		{"plain text", `{"payload":"Dupont Jean 12 rue X"}`, "application/json"},
		{"base64", `{"payload":"PG8+RHVwb250PC9vPg=="}`, "application/json"},
		{"invalid escape", `{"payload":"{\"a\":\"\q\"}"}`, "application/json"},
		{"mismatched xml", `{"payload":"<o><lastName>Dupont</o>"}`, "application/json"},
		{"unclosed xml", `{"payload":"<o><b>Dupont</b>"}`, "application/json"},
		{"number", `{"payload":12345}`, "application/json"},
		{"bool", `{"payload":true}`, "application/json"},
		{"xml unknown entity", `<a><payload>&lt;o&gt;&nbsp;Dupont&lt;/o&gt;</payload></a>`, "text/xml"},
		{"xml bare ampersand", `<a><payload>&lt;o&gt;Dupont &lt;/o&gt; &</payload></a>`, "text/xml"},
		{"xml bad numeric", `<a><payload>&#0;&lt;o/&gt;Dupont</payload></a>`, "text/xml"},
		{"xml plain text", `<a><payload>Dupont</payload></a>`, "text/xml"},
		{"xml malformed nested", `<a><payload>&lt;o&gt;&lt;b&gt;Dupont&lt;/o&gt;</payload></a>`, "text/xml"},
		{"xml child element", `<a><payload><o>Dupont</o></payload></a>`, "text/xml"},
		{"xml comment inside", `<a><payload>&lt;o/&gt;<!-- Dupont --></payload></a>`, "text/xml"},
		{"xml unterminated cdata", `<a><payload><![CDATA[<o>Dupont</o></payload></a>`, "text/xml"},
		{"xml attribute", `<a payload="&lt;o&gt;Dupont&lt;/o&gt;"/>`, "text/xml"},
	}

	r := newEmbedded()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := r.Body(tt.ctype, []byte(tt.in))

			assert.NotContains(t, string(res.Body), "Dupont")
			assert.NotContains(t, string(res.Body), "12345")
			assert.Contains(t, string(res.Body), "REDACTED")

			if tt.ctype == "application/json" {
				assert.True(t, json.Valid(res.Body), string(res.Body))
			}
		})
	}
}

func TestEmbeddedMaskExactOutput(t *testing.T) {
	r := newEmbedded()

	assert.JSONEq(t, `{"payload":"REDACTED","b":1}`, string(r.Body("application/json", []byte(`{"payload":"nope","b":1}`)).Body))
	assert.Equal(t, `<a><payload>REDACTED</payload><b>1</b></a>`, string(r.Body("text/xml", []byte(`<a><payload>nope</payload><b>1</b></a>`)).Body))
}

func TestEmbeddedNonStringValues(t *testing.T) {
	r := newEmbedded()

	tests := []struct{ name, in, want string }{
		{"object is already structured", `{"payload":{"lastName":"Dupont","a":1}}`, `{"payload":{"lastName":"REDACTED","a":1}}`},
		{"array", `{"payload":[{"lastName":"Dupont"}],"b":1}`, `{"payload":[{"lastName":"REDACTED"}],"b":1}`},
		{"null", `{"payload":null}`, `{"payload":null}`},
		{"empty string", `{"payload":""}`, `{"payload":""}`},
		{"blank string", `{"payload":"  "}`, `{"payload":"  "}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(r.Body("application/json", []byte(tt.in)).Body))
		})
	}
}

func TestEmbeddedPaths(t *testing.T) {
	r := newEmbedded()
	inner := jsonString(`{"lastName":"Dupont"}`)
	want := jsonString(`{"lastName":"REDACTED"}`)

	res := r.Body("application/json", []byte(`{"wrapper":{"doc":`+inner+`},"x":{"wrapper":{"doc":"{\"lastName\":\"Dupont\"} nope"}}}`))
	assert.JSONEq(t, `{"wrapper":{"doc":`+want+`},"x":{"wrapper":{"doc":"{\"lastName\":\"Dupont\"} nope"}}}`, string(res.Body))

	res = newEmbedded().Body("text/xml", []byte(`<wrapper><doc>&lt;o lastName="D"/&gt;</doc></wrapper>`))
	assert.Equal(t, `<wrapper><doc>&lt;o lastName="REDACTED"/&gt;</doc></wrapper>`, string(res.Body))

	assert.Panics(t, func() { redact.New(redact.Policy{Embedded: []string{"a..b"}}) })
}

func TestEmbeddedPrecedence(t *testing.T) {
	in := `{"payload":"{\"lastName\":\"Dupont\"}","nonce":"{\"lastName\":\"Dupont\"}"}`

	t.Run("secret wins", func(t *testing.T) {
		r := redact.New(redact.Policy{Secrets: redact.Rules{Keys: []string{"payload"}}, Embedded: []string{"payload"}})
		assert.JSONEq(t, `{"payload":"REDACTED","nonce":"{\"lastName\":\"Dupont\"}"}`, string(r.Body("application/json", []byte(in)).Body))
	})

	t.Run("pii wins", func(t *testing.T) {
		r := redact.New(redact.Policy{PII: redact.Rules{Keys: []string{"payload"}}, Embedded: []string{"payload"}})
		assert.Contains(t, string(r.Body("application/json", []byte(in)).Body), `"payload":"REDACTED"`)
	})

	t.Run("secret path wins over embedded key", func(t *testing.T) {
		r := redact.New(redact.Policy{Secrets: redact.Rules{Keys: []string{"a.payload"}}, Embedded: []string{"payload"}})
		res := r.Body("application/json", []byte(`{"a":{"payload":"{}"},"payload":"{\"x\":1}"}`))
		assert.JSONEq(t, `{"a":{"payload":"REDACTED"},"payload":"{\"x\":1}"}`, string(res.Body))
	})

	t.Run("embedded path does not override a plain secret", func(t *testing.T) {
		r := redact.New(redact.Policy{Secrets: redact.Rules{Keys: []string{"payload"}}, Embedded: []string{"a.payload"}})
		assert.JSONEq(t, `{"a":{"payload":"REDACTED"}}`, string(r.Body("application/json", []byte(`{"a":{"payload":"{}"}}`)).Body))
	})

	t.Run("pii ignored without pii, embedded applies", func(t *testing.T) {
		r := redact.New(redact.Policy{PII: redact.Rules{Keys: []string{"payload"}}, Embedded: []string{"payload"}}, redact.WithoutPII())
		assert.JSONEq(t, `{"payload":`+jsonString(`{"x":1}`)+`}`, string(r.Body("application/json", []byte(`{"payload":"{\"x\":1}"}`)).Body))
	})

	t.Run("volatile wins when enabled", func(t *testing.T) {
		r := redact.New(redact.Policy{Volatile: redact.Rules{Keys: []string{"payload"}}, Embedded: []string{"payload"}}, redact.WithVolatile())
		assert.JSONEq(t, `{"payload":"VOLATILE"}`, string(r.Body("application/json", []byte(`{"payload":"{}"}`)).Body))
	})

	t.Run("volatile ignored when disabled", func(t *testing.T) {
		r := redact.New(redact.Policy{Volatile: redact.Rules{Keys: []string{"payload"}}, Embedded: []string{"payload"}})
		assert.JSONEq(t, `{"payload":"{}"}`, string(r.Body("application/json", []byte(`{"payload":"{}"}`)).Body))
	})

	t.Run("nested document gets volatile too", func(t *testing.T) {
		r := newEmbedded(redact.WithVolatile())
		assert.JSONEq(t, `{"payload":`+jsonString(`{"nonce":"VOLATILE"}`)+`}`, string(r.Body("application/json", []byte(`{"payload":"{\"nonce\":\"n\"}"}`)).Body))
	})
}

func TestEmbeddedWithoutPII(t *testing.T) {
	r := newEmbedded(redact.WithoutPII())

	inner := `{"lastName":"Dupont","token":"abc","contact":"jean@example.com","auth":"Bearer abcdef1234567890xyz"}`
	res := r.Body("application/json", []byte(`{"payload":`+jsonString(inner)+`}`))

	assert.JSONEq(t, `{"payload":`+jsonString(`{"lastName":"Dupont","token":"REDACTED","contact":"jean@example.com","auth":"[REDACTED:bearer]"}`)+`}`, string(res.Body))

	res = r.Body("text/xml", []byte(`<a><payload>&lt;o&gt;&lt;lastName&gt;Dupont&lt;/lastName&gt;&lt;token&gt;t&lt;/token&gt;&lt;/o&gt;</payload></a>`))
	assert.Equal(t, `<a><payload>&lt;o&gt;&lt;lastName&gt;Dupont&lt;/lastName&gt;&lt;token&gt;REDACTED&lt;/token&gt;&lt;/o&gt;</payload></a>`, string(res.Body))
}

func TestEmbeddedMerge(t *testing.T) {
	merged := redact.Merge(redact.Policy{Embedded: []string{"a"}}, redact.Policy{Embedded: []string{"b"}})
	assert.Equal(t, []string{"a", "b"}, merged.Embedded)

	r := redact.New(merged)
	assert.JSONEq(t, `{"a":"REDACTED","b":"REDACTED","c":"nope"}`, string(r.Body("application/json", []byte(`{"a":"nope","b":"nope","c":"nope"}`)).Body))
}

func TestEmbeddedPlainPoliciesUntouched(t *testing.T) {
	res := newRedactor().Body("application/json", []byte(`{"payload":"{\"password\":\"p\"}"}`))
	assert.JSONEq(t, `{"payload":"{\"password\":\"p\"}"}`, string(res.Body))
}
