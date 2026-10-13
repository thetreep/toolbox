package redact

import (
	"fmt"
	"slices"
	"strings"
)

// pathRule is a dotted key such as "customer.name": the last segment is looked up by name, the parents are checked against the enclosing names.
type pathRule struct {
	parents  []string // lowercase, outermost first
	anchored bool
	class    class
}

// parsePath splits a dotted key; it panics on an empty segment since policies are static code.
func parsePath(key string) (parents []string, last string, anchored, isPath bool) {
	rest, anchored := strings.CutPrefix(key, "$.")
	if !anchored && !strings.Contains(rest, ".") {
		return nil, "", false, false
	}

	segments := strings.Split(lowerASCII(rest), ".")
	if slices.Contains(segments, "") {
		panic(fmt.Sprintf("redact: key %q has an empty path segment", key))
	}

	return segments[:len(segments)-1], segments[len(segments)-1], anchored, true
}

func (r *redactor) addPath(parents []string, last string, anchored bool, c class) {
	for i, existing := range r.paths[last] {
		if existing.anchored == anchored && strings.Join(existing.parents, ".") == strings.Join(parents, ".") {
			if existing.class != classSecret {
				r.paths[last][i].class = c
			}

			return
		}
	}

	r.paths[last] = append(r.paths[last], pathRule{parents: parents, anchored: anchored, class: c})
	r.maxKeyLen = max(r.maxKeyLen, len(last))
}

// nameStack holds the lowercase names of the enclosing JSON objects (by their key) or XML elements, outermost first.
type nameStack struct {
	buf  []byte
	ends []int
}

func (s *nameStack) push() { s.ends = append(s.ends, len(s.buf)) }

func (s *nameStack) pop() {
	if n := len(s.ends); n > 0 {
		s.ends = s.ends[:n-1]
		if n > 1 {
			s.buf = s.buf[:s.ends[n-2]]
		} else {
			s.buf = s.buf[:0]
		}
	}
}

// setTop replaces the name of the innermost level and returns it lowercased; the slice is only valid until the next call.
func (s *nameStack) setTop(name []byte) []byte {
	n := len(s.ends)
	if n == 0 {
		return nil
	}

	base := 0
	if n > 1 {
		base = s.ends[n-2]
	}

	s.buf = s.buf[:base]
	for _, c := range name {
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}

		s.buf = append(s.buf, c)
	}

	s.ends[n-1] = len(s.buf)

	return s.buf[base:]
}

func (s *nameStack) level(i int) []byte {
	start := 0
	if i > 0 {
		start = s.ends[i-1]
	}

	return s.buf[start:s.ends[i]]
}

// matches reports whether the first n levels end with the rule's parents, or equal them when the rule is anchored.
func (s *nameStack) matches(n int, rule *pathRule) bool {
	if n > len(s.ends) || len(rule.parents) > n || (rule.anchored && len(rule.parents) != n) {
		return false
	}

	for i, p := range rule.parents {
		if string(s.level(n-len(rule.parents)+i)) != p {
			return false
		}
	}

	return true
}

// nameClass classifies an already lowercased name, checking plain keys then path rules against the first n levels of names.
func (r *redactor) nameClass(low []byte, names *nameStack, n int) class {
	if len(low) == 0 || len(low) > r.maxKeyLen {
		return classNone
	}

	c := r.keys[string(low)]

	rules := r.paths[string(low)]
	for i := range rules {
		if c == classSecret {
			break
		}

		if rules[i].class != classNone && names.matches(n, &rules[i]) {
			c = rules[i].class
		}
	}

	return c
}

// xmlNameClass classifies an element or attribute name; the names stack holds the open elements, the owner element included for attributes.
func (r *redactor) xmlNameClass(name []byte, names *nameStack) class {
	if !r.hasPaths {
		if c := r.keyClass(name); c != classNone {
			return c
		}

		return r.keyClass(localName(name))
	}

	if c := r.lowerNameClass(name, names); c != classNone {
		return c
	}

	return r.lowerNameClass(localName(name), names)
}

func (r *redactor) lowerNameClass(name []byte, names *nameStack) class {
	if len(name) == 0 || len(name) > r.maxKeyLen {
		return classNone
	}

	var buf [64]byte

	low := append(buf[:0], name...)
	for i, c := range low {
		if c >= 'A' && c <= 'Z' {
			low[i] = c + 'a' - 'A'
		}
	}

	return r.nameClass(low, names, len(names.ends))
}
