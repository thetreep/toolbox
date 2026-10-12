package redact

import "encoding/json"

type jsonState uint8

const (
	stValue jsonState = iota
	stValueOrEnd
	stKeyOrEnd
	stKey
	stColon
	stAfter
)

const (
	redactedString = `"` + Placeholder + `"`
	volatileString = `"` + VolatilePlaceholder + `"`
)

// json compacts src into dst, redacts matching keys and string values in one pass, and reports whether src was valid.
func (r *Redactor) json(dst, src []byte) ([]byte, bool) {
	var (
		stack   []byte
		state   = stValue
		valid   = true
		pending class
		// suppressBase is the stack depth of the redacted container whose content is being dropped.
		suppressBase = -1
		seenValue    bool
		i            int
	)

	emit := func(b ...byte) {
		if suppressBase < 0 {
			dst = append(dst, b...)
		}
	}

	// beginValue consumes a pending redaction; it returns true when the value must not be copied.
	beginValue := func() bool {
		if suppressBase >= 0 {
			return true
		}

		if pending == classNone {
			return false
		}

		if pending == classVolatile {
			dst = append(dst, volatileString...)
		} else {
			dst = append(dst, redactedString...)
		}

		pending = classNone

		return true
	}

	for i < len(src) {
		c := src[i]

		switch c {
		case ' ', '\t', '\r', '\n':
			i++
		case '"':
			end, wellFormed := scanJSONString(src, i)
			if !wellFormed {
				valid = false
			}

			content := src[i+1 : min(end, len(src)+1)-1]
			i = min(end, len(src))

			if state == stKey || state == stKeyOrEnd {
				pending = r.jsonKeyClass(content)
				state = stColon

				emit('"')
				if suppressBase < 0 {
					dst = append(dst, content...)
				}
				emit('"')

				continue
			}

			if state != stValue && state != stValueOrEnd {
				valid = false
			}

			seenValue = true
			state = stAfter

			if beginValue() {
				continue
			}

			dst = append(dst, '"')
			dst, _ = scan(dst, content, true, true)
			dst = append(dst, '"')
		case ':':
			if state != stColon {
				valid = false
			}

			state = stValue
			emit(':')
			i++
		case ',':
			if state != stAfter || len(stack) == 0 {
				valid = false
			}

			if len(stack) > 0 && stack[len(stack)-1] == '{' {
				state = stKey
			} else {
				state = stValue
			}

			pending = classNone
			emit(',')
			i++
		case '{', '[':
			if state != stValue && state != stValueOrEnd {
				valid = false
			}

			seenValue = true

			if beginValue() && suppressBase < 0 {
				suppressBase = len(stack)
			}

			emit(c)

			stack = append(stack, c)
			if c == '{' {
				state = stKeyOrEnd
			} else {
				state = stValueOrEnd
			}

			i++
		case '}', ']':
			open := byte('{')
			if c == ']' {
				open = '['
			}

			okState := state == stAfter || (c == '}' && state == stKeyOrEnd) || (c == ']' && state == stValueOrEnd)
			if !okState || len(stack) == 0 || stack[len(stack)-1] != open {
				valid = false
			}

			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

			if suppressBase >= 0 && len(stack) == suppressBase {
				suppressBase = -1
			} else {
				emit(c)
			}

			pending = classNone
			state = stAfter
			i++
		default:
			end := i
			for end < len(src) && !isJSONDelimiter(src[end]) {
				end++
			}

			token := src[i:end]
			i = end

			if state != stValue && state != stValueOrEnd {
				valid = false
			}

			if !validJSONLiteral(token) {
				valid = false
			}

			seenValue = true
			state = stAfter

			if beginValue() {
				continue
			}

			if validJSONLiteral(token) {
				dst = append(dst, token...)
			} else {
				dst, _ = scan(dst, token, false, true)
			}
		}
	}

	if len(stack) != 0 || state != stAfter || !seenValue {
		valid = false
	}

	return dst, valid
}

func (r *Redactor) jsonKeyClass(raw []byte) class {
	for _, c := range raw {
		if c == '\\' {
			var key string
			if json.Unmarshal(append(append([]byte{'"'}, raw...), '"'), &key) != nil {
				return classNone
			}

			return r.keyClass([]byte(key))
		}
	}

	return r.keyClass(raw)
}

func isJSONDelimiter(c byte) bool {
	switch c {
	case ' ', '\t', '\r', '\n', ',', ':', '[', ']', '{', '}', '"':
		return true
	}

	return false
}

// scanJSONString returns the index past the closing quote (len(src)+1 when unterminated) and whether the grammar holds.
func scanJSONString(src []byte, start int) (end int, wellFormed bool) {
	wellFormed = true

	for j := start + 1; j < len(src); j++ {
		switch c := src[j]; {
		case c == '"':
			return j + 1, wellFormed
		case c == '\\':
			j++
			if j >= len(src) {
				return len(src) + 1, false
			}

			switch src[j] {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			case 'u':
				if j+4 >= len(src) || !isHex4(src[j+1:j+5]) {
					wellFormed = false
				} else {
					j += 4
				}
			default:
				wellFormed = false
			}
		case c < 0x20:
			wellFormed = false
		}
	}

	return len(src) + 1, false
}

func isHex4(b []byte) bool {
	for _, c := range b {
		if !isDigit(c) && !(c >= 'a' && c <= 'f') && !(c >= 'A' && c <= 'F') {
			return false
		}
	}

	return len(b) == 4
}

func validJSONLiteral(t []byte) bool {
	switch string(t) {
	case "true", "false", "null":
		return true
	}

	i := 0
	if i < len(t) && t[i] == '-' {
		i++
	}

	switch {
	case i < len(t) && t[i] == '0':
		i++
	case i < len(t) && t[i] >= '1' && t[i] <= '9':
		for i < len(t) && isDigit(t[i]) {
			i++
		}
	default:
		return false
	}

	if i < len(t) && t[i] == '.' {
		i++

		start := i
		for i < len(t) && isDigit(t[i]) {
			i++
		}

		if i == start {
			return false
		}
	}

	if i < len(t) && (t[i] == 'e' || t[i] == 'E') {
		i++
		if i < len(t) && (t[i] == '+' || t[i] == '-') {
			i++
		}

		start := i
		for i < len(t) && isDigit(t[i]) {
			i++
		}

		if i == start {
			return false
		}
	}

	return i == len(t)
}
