// Package httpvcr records and replays HTTP calls in tests, with the redaction policy applied to cassettes.
package httpvcr

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"

	"github.com/thetreep/toolbox/httpx"
	"github.com/thetreep/toolbox/httpx/redact"
)

// IsRecording is true when the client talks to the real API.
type IsRecording = bool

// NewClient returns a logging client backed by testdata/<t.Name()>_<apiName>.k7.yaml (CI replays, K7_RECORD=1 re-records, else records when missing).
func NewClient(
	t testing.TB, apiName string, policy redact.Policy, base http.RoundTripper, opts ...httpx.LoggingOption,
) (*http.Client, IsRecording) {
	t.Helper()

	if base == nil {
		base = http.DefaultTransport
	}

	redactor := redact.New(policy, redact.WithVolatile())

	rec, err := recorder.New(
		"testdata/"+t.Name()+"_"+apiName+".k7",
		recorder.WithSkipRequestLatency(true),
		recorder.WithMode(recordMode()),
		recorder.WithMatcher(matcher(redactor)),
		recorder.WithRealTransport(base),
		recorder.WithHook(saveHook(redactor), recorder.AfterCaptureHook),
	)
	if err != nil {
		t.Fatalf("httpvcr: opening the cassette: %v", err)
	}

	t.Cleanup(func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("httpvcr: saving the cassette: %v", err)
		}
	})

	transport := httpx.NewLoggingTransport(apiName, rec, policy, opts...)

	return &http.Client{Transport: transport}, rec.IsRecording()
}

func recordMode() recorder.Mode {
	if os.Getenv("CI") != "" {
		return recorder.ModeReplayOnly
	}

	if v := os.Getenv("K7_RECORD"); v == "1" || v == "true" {
		return recorder.ModeRecordOnly
	}

	return recorder.ModeRecordOnce
}

// saveHook redacts a freshly recorded interaction before it reaches the cassette.
func saveHook(red redact.Redactor) recorder.HookFunc {
	return func(i *cassette.Interaction) error {
		u, err := url.Parse(i.Request.URL)
		if err != nil {
			return fmt.Errorf("httpvcr: parsing recorded URL: %w", err)
		}

		i.Request.URL = redactURL(red, u).String()

		i.Request.Headers = red.Header(i.Request.Headers)
		i.Response.Headers = red.Header(i.Response.Headers)

		i.Request.Body = redactBody(red, i.Request.Headers, i.Request.Body)
		i.Response.Body = redactBody(red, i.Response.Headers, i.Response.Body)

		syncLength(&i.Request.ContentLength, i.Request.Headers, i.Request.Body)
		syncLength(&i.Response.ContentLength, i.Response.Headers, i.Response.Body)

		// The form recorded with the capture still holds the raw values.
		req, err := i.GetHTTPRequest()
		if err != nil {
			return err
		}

		req.Form, req.PostForm = nil, nil
		if err := req.ParseForm(); err != nil {
			return err
		}

		i.Request.Form = req.Form

		return nil
	}
}

// syncLength keeps the declared lengths in line with the rewritten body, or replay would read a short body.
func syncLength(contentLength *int64, headers http.Header, body string) {
	if *contentLength >= 0 {
		*contentLength = int64(len(body))
	}

	if headers.Get("Content-Length") != "" {
		headers.Set("Content-Length", strconv.Itoa(len(body)))
	}
}

func redactBody(red redact.Redactor, headers http.Header, body string) string {
	return string(red.Body(headers.Get("Content-Type"), []byte(body)).Body)
}

func redactURL(red redact.Redactor, u *url.URL) *url.URL {
	out := *u
	out.RawQuery = red.RawQuery(u.RawQuery)
	out.Path, out.RawPath = red.Text(u.Path), ""

	if u.User != nil {
		out.User = url.User(redact.Placeholder)
	}

	return &out
}

// matcher compares normalized copies: neither the live request nor the cassette entry is modified.
func matcher(red redact.Redactor) cassette.MatcherFunc {
	base := cassette.NewDefaultMatcher(cassette.WithIgnoreAuthorization())

	return func(live *http.Request, saved cassette.Request) bool {
		liveCopy, ok := normalizeLive(red, live)
		if !ok {
			return false
		}

		return base(liveCopy, normalizeSaved(red, saved))
	}
}

func normalizeLive(red redact.Redactor, live *http.Request) (*http.Request, bool) {
	body, ok := readBody(live)
	if !ok {
		return nil, false
	}

	clone := live.Clone(live.Context())
	clone.URL = redactURL(red, live.URL)
	clone.Header = red.Header(live.Header)
	clone.Form, clone.PostForm = nil, nil

	normalized := redactBody(red, live.Header, string(body))
	clone.ContentLength = int64(len(normalized))

	if live.Body != nil {
		clone.Body = io.NopCloser(bytes.NewReader([]byte(normalized)))
	}

	return clone, true
}

func normalizeSaved(red redact.Redactor, saved cassette.Request) cassette.Request {
	if u, err := url.Parse(saved.URL); err == nil {
		saved.URL = redactURL(red, u).String()
	}

	saved.Headers = red.Header(saved.Headers)
	saved.Body = redactBody(red, saved.Headers, saved.Body)
	saved.ContentLength = int64(len(saved.Body))
	saved.Form = red.Form(saved.Form)

	return saved
}

// readBody reads the live body without consuming it for the caller.
func readBody(r *http.Request) ([]byte, bool) {
	if r.Body == nil || r.Body == http.NoBody {
		return nil, true
	}

	if r.GetBody != nil {
		fresh, err := r.GetBody()
		if err != nil {
			return nil, false
		}

		defer fresh.Close()

		b, err := io.ReadAll(fresh)

		return b, err == nil
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, false
	}

	// Restoring the stream is the one unavoidable touch on the live request.
	r.Body = io.NopCloser(bytes.NewReader(b))

	return b, true
}
