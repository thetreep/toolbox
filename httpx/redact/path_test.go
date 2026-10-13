package redact_test

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thetreep/toolbox/httpx/redact"
)

func pathRedactor(keys ...string) redact.Redactor {
	return redact.New(redact.Policy{PII: redact.Rules{Keys: keys}})
}

func TestJSONPathKeys(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		in   string
		want string
	}{
		{"direct parent", []string{"customer.name"}, `{"customer":{"name":"n","age":1}}`, `{"customer":{"name":"REDACTED","age":1}}`},
		{"parent at depth", []string{"customer.name"}, `{"a":{"customer":{"name":"n"}}}`, `{"a":{"customer":{"name":"REDACTED"}}}`},
		{"parent is not direct", []string{"customer.name"}, `{"customer":{"x":{"name":"n"}}}`, `{"customer":{"x":{"name":"n"}}}`},
		{"no parent", []string{"customer.name"}, `{"name":"n"}`, `{"name":"n"}`},
		{"other parent", []string{"customer.name"}, `{"agency":{"name":"n"}}`, `{"agency":{"name":"n"}}`},
		{"parent name is only a suffix", []string{"customer.name"}, `{"mycustomer":{"name":"n"}}`, `{"mycustomer":{"name":"n"}}`},
		{"array is transparent", []string{"customers.name"}, `{"customers":[{"name":"n"},{"name":"m"}]}`, `{"customers":[{"name":"REDACTED"},{"name":"REDACTED"}]}`},
		{"nested arrays are transparent", []string{"customer.name"}, `{"customer":[[{"name":"n"}]]}`, `{"customer":[[{"name":"REDACTED"}]]}`},
		{"array of scalars between", []string{"customer.name"}, `{"customer":[1,{"name":"n"},"x"]}`, `{"customer":[1,{"name":"REDACTED"},"x"]}`},
		{"three segments", []string{"a.b.c"}, `{"x":{"a":{"b":{"c":1,"d":2}}},"b":{"c":3},"a":{"x":{"c":4}}}`, `{"x":{"a":{"b":{"c":"REDACTED","d":2}}},"b":{"c":3},"a":{"x":{"c":4}}}`},
		{"suffix longer than chain", []string{"a.b.c"}, `{"b":{"c":1}}`, `{"b":{"c":1}}`},
		{"case insensitive", []string{"Customer.NAME"}, `{"CUSTOMER":{"Name":"n"}}`, `{"CUSTOMER":{"Name":"REDACTED"}}`},
		{"escaped key in chain", []string{"customer.name"}, `{"customer":{"name":"n"}}`, `{"customer":{"name":"REDACTED"}}`},
		{"escaped last key", []string{"customer.name"}, `{"customer":{"name":"n"}}`, `{"customer":{"name":"REDACTED"}}`},
		{"value is an object", []string{"driver.name"}, `{"driver":{"name":{"first":"a"},"age":3}}`, `{"driver":{"name":"REDACTED","age":3}}`},
		{"siblings reuse the chain slot", []string{"customer.name"}, `{"customer":{"id":1},"name":"n","other":{"name":"m"}}`, `{"customer":{"id":1},"name":"n","other":{"name":"m"}}`},
		{"same path twice", []string{"customer.name"}, `{"a":{"customer":{"name":"1"}},"b":{"customer":{"name":"2"}}}`, `{"a":{"customer":{"name":"REDACTED"}},"b":{"customer":{"name":"REDACTED"}}}`},
		{"after a redacted subtree", []string{"customer.name"}, `{"token":{"customer":{"name":"x"}},"customer":{"name":"y"}}`, `{"token":"REDACTED","customer":{"name":"REDACTED"}}`},
		{"dotted key also matches the literal name", []string{"customer.name", "$.pin"}, `{"customer.name":"n","Customer.Name":"m","a":{"customer.name":"o"},"$.pin":1}`, `{"customer.name":"REDACTED","Customer.Name":"REDACTED","a":{"customer.name":"REDACTED"},"$.pin":"REDACTED"}`},
		{"key whose parent is a value in an array root", []string{"customer.name"}, `[{"customer":{"name":"n"}}]`, `[{"customer":{"name":"REDACTED"}}]`},

		{"anchor top-level", []string{"$.name"}, `{"name":"n","a":{"name":"m"}}`, `{"name":"REDACTED","a":{"name":"m"}}`},
		{"anchor array root is transparent", []string{"$.name"}, `[{"name":"n","a":{"name":"m"}}]`, `[{"name":"REDACTED","a":{"name":"m"}}]`},
		{"anchor nested", []string{"$.customer.name"}, `{"customer":{"name":"n"},"x":{"customer":{"name":"m"}}}`, `{"customer":{"name":"REDACTED"},"x":{"customer":{"name":"m"}}}`},
		{"anchor needs the whole chain", []string{"$.customer.name"}, `{"customer":{"x":{"name":"n"}}}`, `{"customer":{"x":{"name":"n"}}}`},
		{"anchor is case insensitive", []string{"$.Name"}, `{"NAME":"n"}`, `{"NAME":"REDACTED"}`},

		{"plain key keeps matching everywhere", []string{"name", "customer.id"}, `{"name":"a","customer":{"name":"b","id":1}}`, `{"name":"REDACTED","customer":{"name":"REDACTED","id":"REDACTED"}}`},
		{"plain and path on the same last segment", []string{"id", "customer.id"}, `{"id":1,"customer":{"id":2},"x":{"id":3}}`, `{"id":"REDACTED","customer":{"id":"REDACTED"},"x":{"id":"REDACTED"}}`},

		{"dollar without dot is a plain name", []string{"$name"}, `{"$name":1,"name":2}`, `{"$name":"REDACTED","name":2}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := pathRedactor(tt.keys...).Body("application/json", []byte(tt.in))
			assert.Equal(t, tt.want, string(res.Body))
			assert.True(t, res.ValidJSON)
			assert.True(t, json.Valid(res.Body))
		})
	}
}

func TestJSONPathVolatileAndSecret(t *testing.T) {
	policy := redact.Policy{
		Secrets:  redact.Rules{Keys: []string{"phone", "a.ts"}},
		Volatile: redact.Rules{Keys: []string{"driver.phone", "ts", "$.nonce", "driver.seat"}},
	}
	in := []byte(`{"nonce":"n","x":{"nonce":"m"},"ts":1,"a":{"ts":2},"phone":"p","driver":{"phone":"q","seat":4}}`)

	off := redact.New(policy).Body("application/json", in)
	assert.JSONEq(t, `{"nonce":"n","x":{"nonce":"m"},"ts":1,"a":{"ts":"REDACTED"},"phone":"REDACTED","driver":{"phone":"REDACTED","seat":4}}`, string(off.Body))

	on := redact.New(policy, redact.WithVolatile()).Body("application/json", in)
	assert.JSONEq(t, `{"nonce":"VOLATILE","x":{"nonce":"m"},"ts":"VOLATILE","a":{"ts":"REDACTED"},"phone":"REDACTED","driver":{"phone":"REDACTED","seat":"VOLATILE"}}`, string(on.Body))
}

func TestXMLPathKeys(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		in   string
		want string
	}{
		{"element", []string{"driver.last_name"}, `<driver><last_name>x</last_name><age>3</age></driver>`, `<driver><last_name>REDACTED</last_name><age>3</age></driver>`},
		{"attribute", []string{"driver.last_name"}, `<driver last_name="x" age="3"/>`, `<driver last_name="REDACTED" age="3"/>`},
		{"attribute on an open element", []string{"driver.last_name"}, `<driver last_name='x'>t</driver>`, `<driver last_name='REDACTED'>t</driver>`},
		{"element and attribute together", []string{"driver.last_name"}, `<a><driver last_name="x"><last_name>y</last_name></driver></a>`, `<a><driver last_name="REDACTED"><last_name>REDACTED</last_name></driver></a>`},
		{"parent at depth", []string{"driver.last_name"}, `<a><b><driver><last_name>x</last_name></driver></b></a>`, `<a><b><driver><last_name>REDACTED</last_name></driver></b></a>`},
		{"parent is not direct", []string{"driver.last_name"}, `<driver><x><last_name>x</last_name></x></driver>`, `<driver><x><last_name>x</last_name></x></driver>`},
		{"attribute of a non-parent", []string{"driver.last_name"}, `<driver><x last_name="x"/></driver>`, `<driver><x last_name="x"/></driver>`},
		{"no parent", []string{"driver.last_name"}, `<last_name>x</last_name>`, `<last_name>x</last_name>`},
		{"other element", []string{"driver.last_name"}, `<a><driver last_name="x"/><passenger last_name="y"/></a>`, `<a><driver last_name="REDACTED"/><passenger last_name="y"/></a>`},
		{"case insensitive", []string{"Driver.LAST_NAME"}, `<DRIVER Last_Name="x"><Last_Name>y</Last_Name></DRIVER>`, `<DRIVER Last_Name="REDACTED"><Last_Name>REDACTED</Last_Name></DRIVER>`},
		{"namespace prefixes are ignored in the chain", []string{"driver.last_name"}, `<s:a xmlns:s="u"><s:driver><s:last_name>x</s:last_name></s:driver></s:a>`, `<s:a xmlns:s="u"><s:driver><s:last_name>REDACTED</s:last_name></s:driver></s:a>`},
		{"sibling after a self-closing element", []string{"driver.last_name"}, `<a><driver/><last_name>x</last_name></a>`, `<a><driver/><last_name>x</last_name></a>`},
		{"after a redacted subtree", []string{"driver.last_name"}, `<a><token><driver><last_name>x</last_name></driver></token><driver><last_name>y</last_name></driver></a>`, `<a><token>REDACTED</token><driver><last_name>REDACTED</last_name></driver></a>`},
		{"element with children", []string{"driver.address"}, `<driver><address><city>x</city></address><age>3</age></driver>`, `<driver><address>REDACTED</address><age>3</age></driver>`},
		{"comments and declaration do not change the chain", []string{"driver.last_name"}, `<?xml version="1.0"?><driver><!-- c --><![CDATA[z]]><last_name>x</last_name></driver>`, `<?xml version="1.0"?><driver><!-- c --><![CDATA[z]]><last_name>REDACTED</last_name></driver>`},

		{"anchor root element", []string{"$.driver"}, `<driver>x</driver>`, `<driver>REDACTED</driver>`},
		{"anchor root element only", []string{"$.driver"}, `<a><driver>x</driver></a>`, `<a><driver>x</driver></a>`},
		{"anchor child of root", []string{"$.a.driver"}, `<a><driver>x</driver><b><driver>y</driver></b></a>`, `<a><driver>REDACTED</driver><b><driver>y</driver></b></a>`},
		{"anchor attribute of root", []string{"$.a.id"}, `<a id="1"><b id="2"/></a>`, `<a id="REDACTED"><b id="2"/></a>`},
		{"anchor attribute of child", []string{"$.a.b.id"}, `<a id="1"><b id="2"/><c><b id="3"/></c></a>`, `<a id="1"><b id="REDACTED"/><c><b id="3"/></c></a>`},
		{"anchor survives stray closing tags", []string{"$.a.id"}, `</x></x><a id="1"><b id="2"/></a>`, `</x></x><a id="REDACTED"><b id="2"/></a>`},

		{"dotted key also matches the literal name", []string{"driver.last_name"}, `<a driver.last_name="x"><driver.last_name>y</driver.last_name></a>`, `<a driver.last_name="REDACTED"><driver.last_name>REDACTED</driver.last_name></a>`},

		{"plain and path on the same last segment", []string{"id", "driver.last_name"}, `<a id="1"><driver id="2" last_name="x"/></a>`, `<a id="REDACTED"><driver id="REDACTED" last_name="REDACTED"/></a>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := pathRedactor(tt.keys...).Body("text/xml", []byte(tt.in))
			assert.Equal(t, tt.want, string(res.Body))
			assert.Equal(t, redact.KindXML, res.Kind)
		})
	}
}

func TestXMLPathVolatile(t *testing.T) {
	policy := redact.Policy{Volatile: redact.Rules{Keys: []string{"a.ts", "$.a.nonce"}}}
	in := []byte(`<a nonce="n" ts="1"><b nonce="m" ts="2"/></a>`)

	assert.Equal(t, string(in), string(redact.New(policy).Body("text/xml", in).Body))
	assert.Equal(t, `<a nonce="VOLATILE" ts="VOLATILE"><b nonce="m" ts="2"/></a>`,
		string(redact.New(policy, redact.WithVolatile()).Body("text/xml", in).Body))
}

func TestPathKeysAreLiteralInFlatContexts(t *testing.T) {
	r := pathRedactor("customer.name", "$.pin", "plain")

	form := r.Body("application/x-www-form-urlencoded", []byte(`customer.name=x&name=y&$.pin=z&pin=w&plain=v`))
	assert.Equal(t, `customer.name=REDACTED&name=y&$.pin=REDACTED&pin=w&plain=REDACTED`, string(form.Body))

	got := r.Form(url.Values{"customer.name": {"x"}, "name": {"y"}})
	assert.Equal(t, url.Values{"customer.name": {"REDACTED"}, "name": {"y"}}, got)

	assert.Equal(t, "customer.name=1&name=2", r.RawQuery("customer.name=1&name=2"))
}

func TestMalformedPathKeyPanics(t *testing.T) {
	for _, key := range []string{"a..b", "a.", ".a", "$.", "$..a", "$.a."} {
		t.Run(key, func(t *testing.T) {
			want := `redact: key "` + key + `" has an empty path segment`
			assert.PanicsWithValue(t, want, func() { pathRedactor(key) })
			assert.PanicsWithValue(t, want, func() { redact.New(redact.Policy{Secrets: redact.Rules{Keys: []string{key}}}) })
			assert.PanicsWithValue(t, want, func() { redact.New(redact.Policy{Volatile: redact.Rules{Keys: []string{key}}}) })
		})
	}

	assert.NotPanics(t, func() { pathRedactor("$name", "$", "a.b", "$.a") })
}

func TestPathRulesWithMalformedInputFailClosed(t *testing.T) {
	r := pathRedactor("customer.name", "$.pin", "driver.last_name")

	tests := []struct {
		name, contentType, in string
	}{
		{"truncated value", "application/json", `{"customer":{"name":"hunter2`},
		{"truncated after colon", "application/json", `{"customer":{"name":`},
		{"unbalanced", "application/json", `{"customer":{"name":"hunter2"`},
		{"trailing comma", "application/json", `{"customer":{"name":"hunter2",}}`},
		{"extra closers", "application/json", `{}}]]{"customer":{"name":"hunter2"}}`},
		{"mismatched closer then key", "application/json", `{"customer":[}{"name":"hunter2"}`},
		{"anchored invalid", "application/json", `{"pin":"hunter2" "x":1`},
		{"sniffed truncated", "text/plain", `{"a":[1,{"customer":{"name":"hunter2`},
		{"xml truncated element", "text/xml", `<driver><last_name>hunter2`},
		{"xml truncated attribute", "text/xml", `<driver last_name="hunter2`},
		{"xml unquoted attribute", "text/xml", `<driver last_name=hunter2>`},
		{"xml unbalanced", "text/xml", `<a><driver last_name="hunter2"><b></a></a></a><driver last_name="hunter2">`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res redact.Result

			require.NotPanics(t, func() { res = r.Body(tt.contentType, []byte(tt.in)) })
			assert.NotContains(t, string(res.Body), "hunter2")
		})
	}

	res := r.Body("application/json", []byte(`{"customer":{"name":"hunter2"`))
	assert.False(t, res.ValidJSON)
}

func TestPathRulesNeverPanic(t *testing.T) {
	r := redact.New(redact.Policy{
		Secrets:  redact.Rules{Keys: []string{"a.b", "$.a", "a.b.c.d.e"}},
		Volatile: redact.Rules{Keys: []string{"b.c"}},
	}, redact.WithVolatile())

	inputs := []string{
		"", "{", "}", "[", "]", `{"a":`, `{"a":{"b":`, `]}{"a":{"b":1}}`, `{"a":{"b":{"c":{"d":{"e":1}}}}}`, "<a><b", "<a><b>", "</a></a><a><b>x",
		"<a b=", `<a b="`, "<a/><a/>", `{"a\`, `{"a\u00`, `{"` + string([]byte{0xff}) + `":{"b":1}}`,
	}

	for _, in := range inputs {
		for _, ct := range []string{"application/json", "text/xml", ""} {
			require.NotPanics(t, func() { r.Body(ct, []byte(in)) }, "%q", in)
		}
	}
}
