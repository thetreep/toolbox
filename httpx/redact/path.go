package redact

import (
	"fmt"
	"slices"
	"strings"
)

// pathRule is a dotted key such as "customer.name": the last segment is looked up by name, the parents are checked against the enclosing names.
type pathRule struct {
	parents    []string // lowercase, outermost first
	xmlParents []string // parents without namespace prefixes, as the XML name stack holds local names
	anchored   bool
	class      class
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
			if c > existing.class {
				r.paths[last][i].class = c
			}

			return
		}
	}

	xmlParents := make([]string, len(parents))
	for i, p := range parents {
		xmlParents[i] = string(localName([]byte(p)))
	}

	r.paths[last] = append(r.paths[last], pathRule{parents: parents, xmlParents: xmlParents, anchored: anchored, class: c})
	r.maxKeyLen = max(r.maxKeyLen, len(last))
}

// nameStack holds the lowercase names of the open JSON keys or XML elements, outermost first; the innermost is the one being classified.
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

// setTop replaces the name of the innermost level and returns it lowercased; the slice is only valid until the stack next changes.
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
func (s *nameStack) matches(n int, rule *pathRule, xml bool) bool {
	parents := rule.parents
	if xml {
		parents = rule.xmlParents
	}

	if n > len(s.ends) || len(parents) > n || (rule.anchored && len(parents) != n) {
		return false
	}

	for i, p := range parents {
		if string(s.level(n-len(parents)+i)) != p {
			return false
		}
	}

	return true
}

// nameClass classifies an already lowercased name, checking plain keys then path rules against the first n levels of names.
func (r *redactor) nameClass(low []byte, names *nameStack, n int, xml bool) class {
	if len(low) == 0 || len(low) > r.maxKeyLen {
		return classNone
	}

	c := r.keys[string(low)]

	rules := r.paths[string(low)]
	for i := range rules {
		if c == classSecret {
			break
		}

		if rules[i].class > c && names.matches(n, &rules[i], xml) {
			c = rules[i].class
		}
	}

	return c
}

// xmlNameClass classifies an element or attribute name; the names stack holds the open elements, the owner element included for attributes.
func (r *redactor) xmlNameClass(name []byte, names *nameStack) class {
	local := localName(name)

	if !r.hasPaths {
		c := r.keyClass(name)
		if len(local) < len(name) {
			c = max(c, r.keyClass(local))
		}

		return c
	}

	c := r.lowerNameClass(name, names)
	if len(local) < len(name) {
		c = max(c, r.lowerNameClass(local, names))
	}

	return c
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

	return r.nameClass(low, names, len(names.ends), true)
}
