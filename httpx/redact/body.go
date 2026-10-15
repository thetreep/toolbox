package redact

import (
	"mime"
	"strings"
)

// Kind is the body format the redactor handled.
type Kind uint8

const (
	KindText Kind = iota
	KindJSON
	KindXML
	KindForm
)

// Result is a redacted body; ValidJSON is only meaningful for KindJSON.
type Result struct {
	Body      []byte
	Kind      Kind
	ValidJSON bool
}

// Body redacts body according to its content type, sniffing the format when the type is not conclusive.
func (r *redactor) Body(contentType string, body []byte) Result {
	if len(body) == 0 {
		return Result{Body: []byte{}, Kind: KindText}
	}

	kind, declared := kindFromContentType(contentType)
	if !declared {
		kind = sniff(body)
	}

	switch kind {
	case KindJSON:
		out, valid := r.json(make([]byte, 0, len(body)+len(body)/8), body, true)
		if valid {
			return Result{Body: out, Kind: KindJSON, ValidJSON: true}
		}

		// Invalid (truncated, HTML error page…): keep whitespace so detectors also see values split across tokens.
		keyed, _ := r.json(make([]byte, 0, len(body)+len(body)/8), body, false)
		out, _ = r.scan(make([]byte, 0, len(keyed)), keyed, false, true)

		if declared {
			return Result{Body: out, Kind: KindJSON}
		}

		return Result{Body: out, Kind: KindText}
	case KindXML:
		return Result{Body: r.xml(make([]byte, 0, len(body)+len(body)/8), body), Kind: KindXML}
	case KindForm:
		return Result{Body: r.form(make([]byte, 0, len(body)), body), Kind: KindForm}
	}

	out, _ := r.scan(make([]byte, 0, len(body)), body, false, true)

	return Result{Body: out, Kind: KindText}
}

func kindFromContentType(contentType string) (kind Kind, declared bool) {
	if contentType == "" {
		return KindText, false
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))
	}

	switch {
	case mediaType == "application/json", mediaType == "text/json", strings.HasSuffix(mediaType, "+json"):
		return KindJSON, true
	case mediaType == "application/xml", mediaType == "text/xml", strings.HasSuffix(mediaType, "+xml"):
		return KindXML, true
	case mediaType == "application/x-www-form-urlencoded":
		return KindForm, true
	}

	return KindText, false
}

func sniff(body []byte) Kind {
	for _, c := range body {
		switch c {
		case ' ', '\t', '\r', '\n':
			continue
		case '{', '[':
			return KindJSON
		case '<':
			return KindXML
		}

		break
	}

	return KindText
}
