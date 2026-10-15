package redact

import (
	"bytes"
	"encoding/base64"
)

// Detector classes used in the typed placeholders.
const (
	ClassEmail  = "email"
	ClassPhone  = "phone"
	ClassPAN    = "pan"
	ClassIBAN   = "iban"
	ClassJWT    = "jwt"
	ClassBasic  = "basic"
	ClassBearer = "bearer"
)

type detector uint8

const (
	detNone detector = iota
	detEmail
	detPhone
	detPAN
	detIBAN
	detJWT
	detBasic
	detBearer
	detCount
)

type detectorSet uint16

func setOf(ds ...detector) detectorSet {
	var s detectorSet

	for _, d := range ds {
		s |= 1 << d
	}

	return s
}

func (s detectorSet) has(d detector) bool { return s&(1<<d) != 0 }

var (
	piiDetectors    = setOf(detEmail, detPhone, detPAN, detIBAN)
	secretDetectors = setOf(detJWT, detBasic, detBearer)
	allDetectors    = piiDetectors | secretDetectors
)

var detectorNames = [detCount]string{
	detEmail:  ClassEmail,
	detPhone:  ClassPhone,
	detPAN:    ClassPAN,
	detIBAN:   ClassIBAN,
	detJWT:    ClassJWT,
	detBasic:  ClassBasic,
	detBearer: ClassBearer,
}

var detectorPlaceholders = func() (out [detCount][]byte) {
	for d, name := range detectorNames {
		out[d] = []byte("[REDACTED:" + name + "]")
	}
	return out
}()

// triggers marks the bytes at which a detector may start; everything else is skipped without a call.
var triggers = func() (t [256]bool) {
	for c := '0'; c <= '9'; c++ {
		t[c] = true
	}
	for c := 'A'; c <= 'Z'; c++ {
		t[c] = true
	}
	t['+'], t['@'], t['b'], t['e'] = true, true, true, true
	return t
}()

func (r *redactor) scan(dst, s []byte, jsonStr bool) ([]byte, uint16) {
	return scan(dst, s, jsonStr, r.detectors)
}

// scan runs the enabled detectors in a single left-to-right pass; it appends s to dst with matches replaced.
func scan(dst, s []byte, jsonStr bool, enabled detectorSet) ([]byte, uint16) {
	// FIXME: escaped JSON strings (\u0040, \") and XML entities (&#64;) aren't decoded before detectors run.
	var (
		found uint16
		last  int
	)

	for i := 0; i < len(s); {
		if !triggers[s[i]] {
			i++

			continue
		}

		start, end, det := matchAt(s, i, last, jsonStr, enabled)
		if det == detNone {
			i++

			continue
		}

		found |= 1 << det

		dst = append(dst, s[last:start]...)
		dst = append(dst, detectorPlaceholders[det]...)

		last, i = end, end
	}

	dst = append(dst, s[last:]...)

	return dst, found
}

func matchAt(s []byte, i, last int, jsonStr bool, enabled detectorSet) (start, end int, det detector) {
	c := s[i]

	switch {
	case c == '@':
		if !enabled.has(detEmail) {
			break
		}

		if start, end, ok := matchEmail(s, i, last, jsonStr); ok {
			return start, end, detEmail
		}
	case !startBoundary(s, i, jsonStr):
	case c == '+':
		if !enabled.has(detPhone) {
			break
		}

		if end, ok := matchInternational(s, i); ok {
			return i, end, detPhone
		}
	case c >= '0' && c <= '9':
		if enabled.has(detPAN) {
			if end, ok := matchPAN(s, i); ok {
				return i, end, detPAN
			}
		}

		if enabled.has(detPhone) {
			if end, ok := matchFRPhone(s, i); ok {
				return i, end, detPhone
			}
		}
	case c >= 'A' && c <= 'Z':
		if enabled.has(detIBAN) {
			if end, ok := matchIBAN(s, i); ok {
				return i, end, detIBAN
			}
		}

		if end, d, ok := matchScheme(s, i); ok {
			return i, end, d
		}
	case c == 'b':
		if end, d, ok := matchScheme(s, i); ok {
			return i, end, d
		}
	case c == 'e':
		if end, ok := matchJWT(s, i); ok {
			return i, end, detJWT
		}
	}

	return 0, 0, detNone
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func isAlnum(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isLocalChar(c byte) bool {
	return isAlnum(c) || c == '.' || c == '_' || c == '%' || c == '+' || c == '-'
}

func isBase64URL(c byte) bool { return isAlnum(c) || c == '-' || c == '_' }

func endBoundary(s []byte, j int) bool { return j >= len(s) || !isAlnum(s[j]) }

// startBoundary also accepts the byte right after a JSON short escape such as \n, which would otherwise glue to the value.
func startBoundary(s []byte, i int, jsonStr bool) bool {
	if i == 0 || !isAlnum(s[i-1]) {
		return true
	}

	return jsonStr && i >= 2 && s[i-2] == '\\' && bytes.IndexByte([]byte("bfnrt"), s[i-1]) >= 0
}

func matchEmail(s []byte, at, last int, jsonStr bool) (start, end int, ok bool) {
	const maxLocal = 64

	start = at
	for start > last && at-start < maxLocal && isLocalChar(s[start-1]) {
		start--
	}

	// An escape glued to the local part must stay whole, or the rewritten JSON string becomes invalid.
	if jsonStr && start > 0 && s[start-1] == '\\' && start < at {
		if s[start] == 'u' {
			start += 5
		} else {
			start++
		}
	}

	if start >= at {
		return 0, 0, false
	}

	end = at + 1
	for end < len(s) && (isAlnum(s[end]) || s[end] == '.' || s[end] == '-') {
		end++
	}

	for end > at+1 && (s[end-1] == '.' || s[end-1] == '-') {
		end--
	}

	domain := s[at+1 : end]

	dot := bytes.LastIndexByte(domain, '.')
	if dot <= 0 || len(domain)-dot-1 < 2 || !isAlnum(domain[0]) {
		return 0, 0, false
	}

	for _, c := range domain[dot+1:] {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			return 0, 0, false
		}
	}

	return start, end, true
}

// matchInternational matches '+' followed by 8 to 15 digits, optionally grouped by single separators.
func matchInternational(s []byte, i int) (end int, ok bool) {
	j := i + 1
	if j >= len(s) || s[j] < '1' || s[j] > '9' {
		return 0, false
	}

	digits := 0

	for j < len(s) {
		switch {
		case isDigit(s[j]):
			digits++
			j++
		case isSeparator(s[j]) && j+1 < len(s) && isDigit(s[j+1]):
			j++
		default:
			goto done
		}

		if digits > 15 {
			return 0, false
		}
	}

done:
	if digits < 8 || !endBoundary(s, j) {
		return 0, false
	}

	return j, true
}

func isSeparator(c byte) bool { return c == ' ' || c == '.' || c == '-' }

// matchFRPhone matches 0X XX XX XX XX and 0033 X XX XX XX XX with optional single separators.
func matchFRPhone(s []byte, i int) (end int, ok bool) {
	j := i

	switch {
	case bytes.HasPrefix(s[i:], []byte("0033")):
		j += 4
		if j < len(s) && s[j] == ' ' {
			j++
		}
	case s[i] == '0':
		j++
	default:
		return 0, false
	}

	if j >= len(s) || s[j] < '1' || s[j] > '9' {
		return 0, false
	}

	j++

	for group := 0; group < 4; group++ {
		if j < len(s) && isSeparator(s[j]) {
			j++
		}

		if j+1 >= len(s) || !isDigit(s[j]) || !isDigit(s[j+1]) {
			return 0, false
		}

		j += 2
	}

	if !endBoundary(s, j) {
		return 0, false
	}

	return j, true
}

// matchPAN matches 13 to 19 digits, optionally grouped by single spaces or dashes, that pass the Luhn check.
func matchPAN(s []byte, i int) (end int, ok bool) {
	// TODO: PAN detector false positives on Luhn-valid 13-19 digit strings (e.g. ms timestamps as strings).
	const minDigits, maxDigits = 13, 19

	var (
		digits [maxDigits]byte
		count  int
	)

	j := i

	for j < len(s) {
		switch {
		case isDigit(s[j]):
			if count == maxDigits {
				return 0, false
			}

			digits[count] = s[j] - '0'
			count++
			j++
		case (s[j] == ' ' || s[j] == '-') && count > 0 && j+1 < len(s) && isDigit(s[j+1]):
			j++
		default:
			goto done
		}
	}

done:
	if count < minDigits || !endBoundary(s, j) || !luhn(digits[:count]) {
		return 0, false
	}

	return j, true
}

func luhn(digits []byte) bool {
	sum := 0

	for k := len(digits) - 1; k >= 0; k-- {
		d := int(digits[k])
		if (len(digits)-1-k)%2 == 1 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}

		sum += d
	}

	return sum%10 == 0
}

// matchIBAN matches an upper-case IBAN (15 to 34 chars, optional single spaces) that passes the mod-97 check.
func matchIBAN(s []byte, i int) (end int, ok bool) {
	const minLen, maxLen = 15, 34

	if i+3 >= len(s) || s[i+1] < 'A' || s[i+1] > 'Z' || !isDigit(s[i+2]) || !isDigit(s[i+3]) {
		return 0, false
	}

	var (
		compact [maxLen]byte
		count   int
	)

	j := i

	for j < len(s) {
		c := s[j]

		switch {
		case isDigit(c) || (c >= 'A' && c <= 'Z'):
			if count == maxLen {
				return 0, false
			}

			compact[count] = c
			count++
			j++
		case c == ' ' && j+1 < len(s) && (isDigit(s[j+1]) || (s[j+1] >= 'A' && s[j+1] <= 'Z')):
			j++
		default:
			goto done
		}
	}

done:
	if count < minLen || !endBoundary(s, j) || !ibanChecksum(compact[:count]) {
		return 0, false
	}

	return j, true
}

func ibanChecksum(iban []byte) bool {
	rem := 0
	feed := func(c byte) {
		if isDigit(c) {
			rem = (rem*10 + int(c-'0')) % 97

			return
		}

		rem = (rem*100 + int(c-'A') + 10) % 97
	}

	for _, c := range iban[4:] {
		feed(c)
	}

	for _, c := range iban[:4] {
		feed(c)
	}

	return rem == 1
}

func matchJWT(s []byte, i int) (end int, ok bool) {
	if !bytes.HasPrefix(s[i:], []byte("eyJ")) || (i > 0 && (isBase64URL(s[i-1]) || s[i-1] == '.')) {
		return 0, false
	}

	j := i

	for seg := 0; seg < 3; seg++ {
		segStart := j
		for j < len(s) && isBase64URL(s[j]) {
			j++
		}

		// The signature segment may be empty (unsecured JWT).
		if j == segStart && seg < 2 {
			return 0, false
		}

		if seg < 2 {
			if j >= len(s) || s[j] != '.' {
				return 0, false
			}

			j++
		}
	}

	if j < len(s) && (s[j] == '.' || isBase64URL(s[j])) {
		return 0, false
	}

	return j, true
}

// matchScheme matches "Bearer <token>" and "Basic <base64 user:pass>", case-insensitively.
func matchScheme(s []byte, i int) (end int, det detector, ok bool) {
	rest := s[i:]

	var scheme int

	switch {
	case hasPrefixFold(rest, "bearer "):
		scheme, det = len("bearer "), detBearer
	case hasPrefixFold(rest, "basic "):
		scheme, det = len("basic "), detBasic
	default:
		return 0, detNone, false
	}

	j := i + scheme
	for j < len(s) && s[j] == ' ' {
		j++
	}

	tokenStart := j

	for j < len(s) && isTokenChar(s[j], det) {
		j++
	}

	token := s[tokenStart:j]
	for det == detBearer && len(token) > 0 && (token[len(token)-1] == '.' || token[len(token)-1] == '-') {
		token, j = token[:len(token)-1], j-1
	}

	if det == detBasic {
		if len(token) < 8 || !decodesToCredentials(token) {
			return 0, detNone, false
		}

		return j, det, true
	}

	if len(token) < 16 || bytes.IndexFunc(token, func(r rune) bool { return r >= '0' && r <= '9' }) < 0 {
		return 0, detNone, false
	}

	return j, det, true
}

func isTokenChar(c byte, det detector) bool {
	if isAlnum(c) || c == '+' || c == '/' || c == '=' {
		return true
	}

	return det == detBearer && (c == '.' || c == '_' || c == '~' || c == '-')
}

func hasPrefixFold(b []byte, prefix string) bool {
	if len(b) < len(prefix) {
		return false
	}

	for k := 0; k < len(prefix); k++ {
		c := b[k]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}

		if c != prefix[k] {
			return false
		}
	}

	return true
}

func decodesToCredentials(token []byte) bool {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(token)))

	n, err := base64.StdEncoding.Decode(decoded, token)
	if err != nil {
		return false
	}

	return bytes.IndexByte(decoded[:n], ':') >= 0
}
