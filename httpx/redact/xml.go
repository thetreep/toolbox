package redact

import (
	"bytes"
	"net/url"
)

// xml redacts matching elements, attributes and text nodes in one pass, copying everything else as-is.
func (r *redactor) xml(dst, src []byte) []byte {
	var (
		depth int
		names nameStack
		// suppressBase is the depth of the redacted element whose content is being dropped.
		suppressBase = -1
		i            int
	)

	text := func(b []byte) {
		if suppressBase < 0 {
			dst, _ = r.scan(dst, b, false)
		}
	}

	raw := func(b []byte) {
		if suppressBase < 0 {
			dst = append(dst, b...)
		}
	}

	for i < len(src) {
		lt := bytes.IndexByte(src[i:], '<')
		if lt < 0 {
			text(src[i:])

			break
		}

		text(src[i : i+lt])
		i += lt

		rest := src[i:]

		switch {
		case bytes.HasPrefix(rest, []byte("<!--")):
			i = r.xmlSpan(&dst, src, i, "<!--", "-->", suppressBase >= 0)
		case bytes.HasPrefix(rest, []byte("<![CDATA[")):
			i = r.xmlSpan(&dst, src, i, "<![CDATA[", "]]>", suppressBase >= 0)
		case bytes.HasPrefix(rest, []byte("<?")):
			i = r.xmlSpan(&dst, src, i, "<?", "?>", suppressBase >= 0)
		case bytes.HasPrefix(rest, []byte("<!")):
			end := bytes.IndexByte(rest, '>')
			if end < 0 {
				end = len(rest) - 1
			}

			raw(rest[:end+1])

			i += end + 1
		case bytes.HasPrefix(rest, []byte("</")):
			end := bytes.IndexByte(rest, '>')
			if end < 0 {
				end = len(rest) - 1
			}

			if depth > 0 {
				depth--

				if r.hasPaths {
					names.pop()
				}
			}

			if suppressBase >= 0 && depth == suppressBase {
				suppressBase = -1
			}

			raw(rest[:end+1])

			i += end + 1
		default:
			var next int

			next, depth, suppressBase = r.xmlStartTag(&dst, &names, src, i, depth, suppressBase)
			i = next
		}
	}

	return dst
}

// xmlSpan copies a comment, CDATA section or processing instruction, running the detectors on its content.
func (r *redactor) xmlSpan(dst *[]byte, src []byte, i int, open, closing string, suppressed bool) int {
	body := i + len(open)

	end := bytes.Index(src[body:], []byte(closing))
	if end < 0 {
		end = len(src) - body
		closing = ""
	}

	if !suppressed {
		*dst = append(*dst, open...)
		*dst, _ = r.scan(*dst, src[body:body+end], false)
		*dst = append(*dst, closing...)
	}

	return body + end + len(closing)
}

// xmlStartTag handles one start tag at src[i] and returns the next index and updated depth and suppressBase.
func (r *redactor) xmlStartTag(dst *[]byte, names *nameStack, src []byte, i, depth, suppressBase int) (int, int, int) {
	suppressed := suppressBase >= 0
	j := i + 1

	for j < len(src) && !isXMLSpace(src[j]) && src[j] != '>' && src[j] != '/' {
		j++
	}

	name := src[i+1 : j]
	elementClass := r.xmlNameClass(name, names)

	if r.hasPaths {
		names.push()
		names.setTop(localName(name))
	}

	if !suppressed {
		*dst = append(*dst, src[i:j]...)
	}

	selfClosing, closed := false, false

	for j < len(src) {
		c := src[j]

		switch {
		case c == '>':
			closed = true
			j++

			goto tagEnd
		case c == '/' && j+1 < len(src) && src[j+1] == '>':
			selfClosing = true
			j += 2

			goto tagEnd
		case isXMLSpace(c) || c == '/':
			if !suppressed {
				*dst = append(*dst, c)
			}

			j++
		default:
			j = r.xmlAttribute(dst, names, src, j, suppressed)
		}
	}

tagEnd:
	if !suppressed {
		switch {
		case selfClosing:
			*dst = append(*dst, '/', '>')
		case closed:
			*dst = append(*dst, '>')
		}
	}

	if selfClosing {
		if r.hasPaths {
			names.pop()
		}

		return j, depth, suppressBase
	}

	if !suppressed && elementClass != classNone {
		*dst = append(*dst, elementClass.placeholder()...)
		suppressBase = depth
	}

	return j, depth + 1, suppressBase
}

// xmlAttribute copies one name="value" attribute starting at src[j] and returns the index after it.
func (r *redactor) xmlAttribute(dst *[]byte, names *nameStack, src []byte, j int, suppressed bool) int {
	nameStart := j
	for j < len(src) && src[j] != '=' && !isXMLSpace(src[j]) && src[j] != '>' && src[j] != '/' {
		j++
	}

	name := src[nameStart:j]

	if !suppressed {
		*dst = append(*dst, name...)
	}

	for j < len(src) && isXMLSpace(src[j]) {
		if !suppressed {
			*dst = append(*dst, src[j])
		}

		j++
	}

	if j >= len(src) || src[j] != '=' {
		return j
	}

	if !suppressed {
		*dst = append(*dst, '=')
	}

	j++

	for j < len(src) && isXMLSpace(src[j]) {
		if !suppressed {
			*dst = append(*dst, src[j])
		}

		j++
	}

	if j >= len(src) {
		return j
	}

	quote := src[j]

	var valueStart, valueEnd, next int

	if quote == '"' || quote == '\'' {
		valueStart = j + 1

		end := bytes.IndexByte(src[valueStart:], quote)
		if end < 0 {
			valueEnd, next = len(src), len(src)
		} else {
			valueEnd, next = valueStart+end, valueStart+end+1
		}
	} else {
		valueStart, valueEnd = j, j
		for valueEnd < len(src) && !isXMLSpace(src[valueEnd]) && src[valueEnd] != '>' {
			valueEnd++
		}

		next = valueEnd
	}

	if !suppressed {
		matched := r.xmlNameClass(name, names)

		if quote == '"' || quote == '\'' {
			*dst = append(*dst, quote)
		}

		if matched != classNone {
			*dst = append(*dst, matched.placeholder()...)
		} else {
			*dst, _ = r.scan(*dst, src[valueStart:valueEnd], false)
		}

		if (quote == '"' || quote == '\'') && next > valueEnd {
			*dst = append(*dst, quote)
		}
	}

	return next
}

func isXMLSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\r' || c == '\n' }

// form rewrites an application/x-www-form-urlencoded body.
func (r *redactor) form(dst, src []byte) []byte {
	for len(src) > 0 {
		var pair []byte

		if amp := bytes.IndexByte(src, '&'); amp >= 0 {
			pair, src = src[:amp], src[amp+1:]
		} else {
			pair, src = src, nil
		}

		dst = r.formPair(dst, pair)

		if len(src) > 0 {
			dst = append(dst, '&')
		}
	}

	return dst
}

func (r *redactor) formPair(dst, pair []byte) []byte {
	eq := bytes.IndexByte(pair, '=')
	if eq < 0 {
		out, _ := r.scan(dst, pair, false)

		return out
	}

	key, value := pair[:eq], pair[eq+1:]

	decodedKey, err := url.QueryUnescape(string(key))
	if err != nil {
		decodedKey = string(key)
	}

	dst = append(dst, key...)
	dst = append(dst, '=')

	if c := r.formKeys[lowerASCII(decodedKey)]; c != classNone {
		return append(dst, c.placeholder()...)
	}

	decodedValue, err := url.QueryUnescape(string(value))
	if err != nil {
		decodedValue = string(value)
	}

	if out, found := r.scan(nil, []byte(decodedValue), false); found != 0 {
		return append(dst, url.QueryEscape(string(out))...)
	}

	return append(dst, value...)
}

func lowerASCII(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			b := []byte(s)
			for j := i; j < len(b); j++ {
				if b[j] >= 'A' && b[j] <= 'Z' {
					b[j] += 'a' - 'A'
				}
			}

			return string(b)
		}
	}

	return s
}
