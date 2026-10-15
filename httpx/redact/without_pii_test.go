package redact_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thetreep/toolbox/httpx/redact"
)

func newWithoutPII(opts ...redact.Option) redact.Redactor {
	policy := redact.Policy{
		PII: redact.Rules{
			Keys:        []string{"firstName", "customer.lastName"},
			Headers:     []string{"X-User"},
			QueryParams: []string{"lat"},
		},
		Volatile: redact.Rules{Keys: []string{"nonce"}, Headers: []string{"X-Trace"}},
	}

	return redact.New(policy, append([]redact.Option{redact.WithoutPII()}, opts...)...)
}

func TestWithoutPIIKeepsPIIRules(t *testing.T) {
	r := newWithoutPII(redact.WithVolatile())

	res := r.Body("application/json", []byte(`{"firstName":"Jean","customer":{"lastName":"Dupont"},"nonce":"n","password":"p"}`))
	assert.JSONEq(t, `{"firstName":"Jean","customer":{"lastName":"Dupont"},"nonce":"VOLATILE","password":"REDACTED"}`, string(res.Body))

	res = r.Body("text/xml", []byte(`<a><firstName>Jean</firstName><nonce>n</nonce><token>t</token></a>`))
	assert.Equal(t, `<a><firstName>Jean</firstName><nonce>VOLATILE</nonce><token>REDACTED</token></a>`, string(res.Body))

	res = r.Body("application/x-www-form-urlencoded", []byte(`firstName=Jean&lat=48.85&token=t`))
	assert.Equal(t, `firstName=Jean&lat=48.85&token=REDACTED`, string(res.Body))

	h := r.Header(http.Header{"X-User": {"jean"}, "Authorization": {"Bearer abcdef1234567890xyz"}, "X-Trace": {"t"}})
	assert.Equal(t, http.Header{"X-User": {"jean"}, "Authorization": {"REDACTED"}, "X-Trace": {"VOLATILE"}}, h)

	assert.Equal(t, url.Values{"lat": {"48.85"}, "key": {"REDACTED"}}, r.Query(url.Values{"lat": {"48.85"}, "key": {"k"}}))
	assert.Equal(t, url.Values{"lat": {"48.85"}, "key": {"REDACTED"}}, r.Form(url.Values{"lat": {"48.85"}, "key": {"k"}}))
	assert.Equal(t, "key=REDACTED&lat=48.85", r.RawQuery("lat=48.85&key=k"))
}

func TestWithoutPIIKeepsPIIDetectorsOff(t *testing.T) {
	in := "a@b.fr, 0612345678, +33612345678, 4111111111111111, FR1420041010050500013M02606"
	assert.Equal(t, in, newWithoutPII().Text(in))

	body := `{"note":"` + in + `"}`
	assert.JSONEq(t, body, string(newWithoutPII().Body("application/json", []byte(body)).Body))
	assert.Equal(t, in, string(newWithoutPII().Body("text/plain", []byte(in)).Body))
	assert.Equal(t, in, string(newWithoutPII().Body("application/xml", []byte(in)).Body))
	assert.Equal(t, in, string(newWithoutPII().Body("", []byte(in)).Body))
}

func TestWithoutPIIKeepsSecretDetectors(t *testing.T) {
	jwt := "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	in := "a@b.fr " + jwt + " Bearer abcdef1234567890xyz Basic dXNlcjpwYXNzd29yZA== 0612345678"
	want := "a@b.fr [REDACTED:jwt] [REDACTED:bearer] [REDACTED:basic] 0612345678"

	assert.Equal(t, want, newWithoutPII().Text(in))
	assert.Equal(t, want, string(newWithoutPII().Body("text/plain", []byte(in)).Body))

	res := newWithoutPII().Body("application/json", []byte(`{"note":"`+in+`"}`))
	assert.JSONEq(t, `{"note":"`+want+`"}`, string(res.Body))
}

func TestWithoutPIIKeepsSecretRuleOverPIIDuplicate(t *testing.T) {
	r := redact.New(redact.Policy{
		Secrets: redact.Rules{Keys: []string{"pin"}},
		PII:     redact.Rules{Keys: []string{"pin", "name"}},
	}, redact.WithoutPII())

	assert.JSONEq(t, `{"pin":"REDACTED","name":"Jean"}`, string(r.Body("application/json", []byte(`{"pin":"1","name":"Jean"}`)).Body))
}

func TestWithoutPIIStillValidatesPIIKeys(t *testing.T) {
	assert.Panics(t, func() {
		redact.New(redact.Policy{PII: redact.Rules{Keys: []string{"a..b"}}}, redact.WithoutPII())
	})
}

func TestWithPIIDefaultUnchanged(t *testing.T) {
	in := "a@b.fr 0612345678"
	assert.Equal(t, "[REDACTED:email] [REDACTED:phone]", newRedactor().Text(in))

	res := newRedactor().Body("application/json", []byte(`{"firstName":"Jean"}`))
	assert.JSONEq(t, `{"firstName":"REDACTED"}`, string(res.Body))
}
