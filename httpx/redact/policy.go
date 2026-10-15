// Package redact removes secrets and personal data from HTTP headers, queries and bodies.
package redact

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	// Placeholder replaces secrets and PII.
	Placeholder = "REDACTED"
	// VolatilePlaceholder replaces values that change on every call, so recorded cassettes still match.
	VolatilePlaceholder = "VOLATILE"
)

// Rules are case-insensitive names to redact; Keys match JSON keys, XML elements and XML attributes, dotted keys are paths (see README.md).
type Rules struct {
	Headers     []string
	QueryParams []string
	Keys        []string
}

// Policy classifies what must be redacted; Secrets and PII are always redacted, Volatile only when asked for.
type Policy struct {
	Secrets  Rules
	PII      Rules
	Volatile Rules
}

// Merge appends policies in order; nothing is ever removed.
func Merge(policies ...Policy) Policy {
	var out Policy

	for _, p := range policies {
		out.Secrets = mergeRules(out.Secrets, p.Secrets)
		out.PII = mergeRules(out.PII, p.PII)
		out.Volatile = mergeRules(out.Volatile, p.Volatile)
	}

	return out
}

func mergeRules(a, b Rules) Rules {
	return Rules{
		Headers:     append(append([]string(nil), a.Headers...), b.Headers...),
		QueryParams: append(append([]string(nil), a.QueryParams...), b.QueryParams...),
		Keys:        append(append([]string(nil), a.Keys...), b.Keys...),
	}
}

// Default is the baseline every Redactor starts from.
func Default() Policy {
	return Policy{
		Secrets: Rules{
			Headers: []string{
				"Authorization", "Proxy-Authorization", "Cookie", "Set-Cookie", "X-Api-Key", "X-Goog-Api-Key",
			},
			QueryParams: []string{"key", "api_key", "access_token"},
			Keys: []string{
				"password", "token", "access_token", "refresh_token", "id_token",
				"client_secret", "secret", "apiKey", "api_key",
			},
		},
		Volatile: Rules{
			Headers: []string{"X-Request-ID", "X-Parent-Request-ID"},
		},
	}
}

type class uint8

const (
	classNone class = iota
	classSecret
	classVolatile
)

func (c class) placeholder() string {
	if c == classVolatile {
		return VolatilePlaceholder
	}

	return Placeholder
}

// Redactor applies a compiled Policy and is safe for concurrent use; the pointer alias leaves no usable zero value.
type Redactor = *redactor

type redactor struct {
	headers     map[string]class
	queryParams map[string]class
	keys        map[string]class
	paths       map[string][]pathRule
	formKeys    map[string]class
	maxKeyLen   int
	hasPaths    bool
	detectors   detectorSet
}

// Option customizes New.
type Option func(*options)

type options struct{ volatile, withoutPII bool }

// WithVolatile also redacts the Volatile rules, for cassette recording and matching.
func WithVolatile() Option { return func(o *options) { o.volatile = true } }

// WithoutPII ignores the PII rules and the email, phone, PAN and IBAN detectors; secrets stay redacted.
func WithoutPII() Option { return func(o *options) { o.withoutPII = true } }

// New compiles p on top of Default(), which cannot be dropped; it panics on a dotted key with an empty segment (a..b, a., .a, $.).
func New(p Policy, opts ...Option) Redactor {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	p = Merge(Default(), p)
	r := &redactor{
		headers:     map[string]class{},
		queryParams: map[string]class{},
		keys:        map[string]class{},
		paths:       map[string][]pathRule{},
		formKeys:    map[string]class{},
		detectors:   allDetectors,
	}

	if o.withoutPII {
		r.detectors = secretDetectors
	}

	add := func(m map[string]class, names []string, c class) {
		for _, name := range names {
			name = strings.ToLower(name)
			if name == "" {
				continue
			}

			if m[name] != classSecret {
				m[name] = c
			}
		}
	}

	rulesets := []struct {
		rules Rules
		class class
	}{{p.Secrets, classSecret}}
	if !o.withoutPII {
		rulesets = append(rulesets, struct {
			rules Rules
			class class
		}{p.PII, classSecret})
	}

	if o.volatile {
		rulesets = append(rulesets, struct {
			rules Rules
			class class
		}{p.Volatile, classVolatile})
	}

	// Keys are checked even when not compiled, so a malformed one surfaces whatever the options.
	for _, key := range append(append([]string(nil), p.Volatile.Keys...), p.PII.Keys...) {
		parsePath(key)
	}

	for _, rs := range rulesets {
		add(r.headers, rs.rules.Headers, rs.class)
		add(r.queryParams, rs.rules.QueryParams, rs.class)

		for _, key := range rs.rules.Keys {
			if parents, last, anchored, isPath := parsePath(key); isPath {
				r.addPath(parents, last, anchored, rs.class)
			}

			// A dotted key also matches a key spelled with the dots, so a path rule never fails open on flattened payloads.
			add(r.keys, []string{key}, rs.class)
		}

		add(r.formKeys, rs.rules.QueryParams, rs.class)
		add(r.formKeys, rs.rules.Keys, rs.class)
	}

	for name := range r.keys {
		r.maxKeyLen = max(r.maxKeyLen, len(name))
	}

	r.hasPaths = len(r.paths) > 0

	return r
}

// Header returns a copy of h with matching headers replaced and detectors run on the other values.
func (r *redactor) Header(h http.Header) http.Header {
	if h == nil {
		return nil
	}

	out := make(http.Header, len(h))

	for name, values := range h {
		if c := r.headers[strings.ToLower(name)]; c != classNone {
			out[name] = []string{c.placeholder()}

			continue
		}

		redacted := make([]string, len(values))
		for i, v := range values {
			redacted[i] = r.Text(v)
		}

		out[name] = redacted
	}

	return out
}

// Query returns a copy of q with matching parameters replaced and detectors run on the other values.
func (r *redactor) Query(q url.Values) url.Values {
	if q == nil {
		return nil
	}

	out := make(url.Values, len(q))

	for name, values := range q {
		if c := r.queryParams[strings.ToLower(name)]; c != classNone {
			out[name] = []string{c.placeholder()}

			continue
		}

		redacted := make([]string, len(values))
		for i, v := range values {
			redacted[i] = r.Text(v)
		}

		out[name] = redacted
	}

	return out
}

// Form redacts parsed form values: query parameter rules and body key rules both apply.
func (r *redactor) Form(f url.Values) url.Values {
	if f == nil {
		return nil
	}

	out := make(url.Values, len(f))

	for name, values := range f {
		if c := r.formKeys[strings.ToLower(name)]; c != classNone {
			out[name] = []string{c.placeholder()}

			continue
		}

		redacted := make([]string, len(values))
		for i, v := range values {
			redacted[i] = r.Text(v)
		}

		out[name] = redacted
	}

	return out
}

// RawQuery redacts an encoded query string and returns it re-encoded with sorted keys.
func (r *redactor) RawQuery(raw string) string {
	// Pairs that do not parse are dropped, which is the safe side for a redactor.
	values, _ := url.ParseQuery(raw)

	return r.Query(values).Encode()
}

// Text runs the detectors on arbitrary text.
func (r *redactor) Text(s string) string {
	out, found := r.scan(nil, []byte(s), false)
	if found == 0 {
		return s
	}

	return string(out)
}

func (r *redactor) keyClass(name []byte) class {
	if len(name) == 0 || len(name) > r.maxKeyLen {
		return classNone
	}

	var buf [64]byte

	if len(name) > len(buf) {
		return r.keys[strings.ToLower(string(name))]
	}

	for i, c := range name {
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}

		buf[i] = c
	}

	return r.keys[string(buf[:len(name)])]
}

// localName strips an XML namespace prefix.
func localName(name []byte) []byte {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == ':' {
			return name[i+1:]
		}
	}

	return name
}
