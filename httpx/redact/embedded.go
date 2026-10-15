package redact

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"strconv"
	"unicode/utf8"
)

// maxEmbeddedDepth bounds documents nested in documents; deeper ones are masked.
const maxEmbeddedDepth = 3

// nested redacts doc, found in an Embedded value, with the same policy; false means it must be masked.
func (r *redactor) nested(doc []byte) ([]byte, bool) {
	if r.nest >= maxEmbeddedDepth {
		return nil, false
	}

	sub := *r
	sub.nest++

	switch sniff(doc) {
	case KindJSON:
		return sub.json(make([]byte, 0, len(doc)+len(doc)/8), doc, false)
	case KindXML:
		if !wellFormedXML(doc) {
			return nil, false
		}

		return sub.xml(make([]byte, 0, len(doc)+len(doc)/8), doc), true
	}

	return nil, false
}

// jsonEmbedded appends the redacted JSON string whose raw content is given, masked when it is not a well-formed document.
func (r *redactor) jsonEmbedded(dst, content []byte, wellFormed bool) []byte {
	if doc, ok := decodeJSONString(content, wellFormed); ok {
		if len(bytes.TrimSpace(doc)) == 0 {
			dst = append(dst, '"')
			dst = append(dst, content...)

			return append(dst, '"')
		}

		if out, ok := r.nested(doc); ok {
			dst = append(dst, '"')
			dst = appendJSONEscaped(dst, out)

			return append(dst, '"')
		}
	}

	return append(dst, redactedString...)
}

func decodeJSONString(content []byte, wellFormed bool) ([]byte, bool) {
	if !wellFormed {
		return nil, false
	}

	if bytes.IndexByte(content, '\\') < 0 {
		return content, true
	}

	var s string
	if json.Unmarshal(append(append([]byte{'"'}, content...), '"'), &s) != nil {
		return nil, false
	}

	return []byte(s), true
}

func appendJSONEscaped(dst, s []byte) []byte {
	const hex = "0123456789abcdef"

	for _, c := range s {
		switch {
		case c == '"' || c == '\\':
			dst = append(dst, '\\', c)
		case c == '\n':
			dst = append(dst, '\\', 'n')
		case c == '\r':
			dst = append(dst, '\\', 'r')
		case c == '\t':
			dst = append(dst, '\\', 't')
		case c < 0x20:
			dst = append(dst, '\\', 'u', '0', '0', hex[c>>4], hex[c&0xf])
		default:
			dst = append(dst, c)
		}
	}

	return dst
}

// xmlEmbedded appends the escaped redacted document held by the element whose start tag ends at src[i], returning its end tag index; false appends nothing.
func (r *redactor) xmlEmbedded(dst *[]byte, src []byte, i int) (int, bool) {
	var doc []byte

	j := i

	for {
		lt := bytes.IndexByte(src[j:], '<')
		if lt < 0 {
			return 0, false
		}

		var ok bool

		if doc, ok = appendXMLUnescaped(doc, src[j:j+lt]); !ok {
			return 0, false
		}

		j += lt

		const cdata = "<![CDATA["

		switch rest := src[j:]; {
		case bytes.HasPrefix(rest, []byte(cdata)):
			end := bytes.Index(rest, []byte("]]>"))
			if end < 0 {
				return 0, false
			}

			doc = append(doc, rest[len(cdata):end]...)
			j += end + len("]]>")
		case bytes.HasPrefix(rest, []byte("</")):
			if len(bytes.TrimSpace(doc)) == 0 {
				*dst = append(*dst, src[i:j]...)

				return j, true
			}

			out, ok := r.nested(doc)
			if !ok {
				return 0, false
			}

			*dst = appendXMLEscaped(*dst, out)

			return j, true
		default:
			return 0, false
		}
	}
}

func appendXMLEscaped(dst, s []byte) []byte {
	for _, c := range s {
		switch c {
		case '&':
			dst = append(dst, "&amp;"...)
		case '<':
			dst = append(dst, "&lt;"...)
		case '>':
			dst = append(dst, "&gt;"...)
		default:
			dst = append(dst, c)
		}
	}

	return dst
}

// appendXMLUnescaped decodes the predefined and numeric entities; any other reference makes the text undecodable.
func appendXMLUnescaped(dst, s []byte) ([]byte, bool) {
	for len(s) > 0 {
		amp := bytes.IndexByte(s, '&')
		if amp < 0 {
			return append(dst, s...), true
		}

		dst = append(dst, s[:amp]...)
		s = s[amp+1:]

		semi := bytes.IndexByte(s, ';')
		if semi < 0 {
			return nil, false
		}

		ref := string(s[:semi])
		s = s[semi+1:]

		switch ref {
		case "lt":
			dst = append(dst, '<')
		case "gt":
			dst = append(dst, '>')
		case "amp":
			dst = append(dst, '&')
		case "quot":
			dst = append(dst, '"')
		case "apos":
			dst = append(dst, '\'')
		default:
			n, ok := parseCharRef(ref)
			if !ok {
				return nil, false
			}

			dst = utf8.AppendRune(dst, n)
		}
	}

	return dst, true
}

func parseCharRef(ref string) (rune, bool) {
	var (
		digits string
		base   int
	)

	switch {
	case len(ref) > 2 && ref[0] == '#' && (ref[1] == 'x' || ref[1] == 'X'):
		digits, base = ref[2:], 16
	case len(ref) > 1 && ref[0] == '#':
		digits, base = ref[1:], 10
	default:
		return 0, false
	}

	n, err := strconv.ParseUint(digits, base, 32)
	if err != nil || n == 0 || !utf8.ValidRune(rune(n)) {
		return 0, false
	}

	return rune(n), true
}

// wellFormedXML checks the structure only: charsets are not decoded.
func wellFormedXML(doc []byte) bool {
	d := xml.NewDecoder(bytes.NewReader(doc))
	d.CharsetReader = func(_ string, in io.Reader) (io.Reader, error) { return in, nil }

	for {
		if _, err := d.Token(); err != nil {
			return errors.Is(err, io.EOF)
		}
	}
}
