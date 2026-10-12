package httpx

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/thetreep/toolbox/httpx/redact"
	"github.com/thetreep/toolbox/logger"
)

const (
	// DefaultEntryBudget keeps a log entry well under Cloud Logging's 256 KB limit.
	DefaultEntryBudget = 180_000
	// DefaultMaxCapture bounds the bytes buffered per body when a BodySink can store the whole body.
	DefaultMaxCapture = 8 << 20

	tracerName = "github.com/thetreep/toolbox/httpx"
)

// BodySink stores a redacted body that does not fit in the log entry and returns a reference to it.
type BodySink interface {
	Put(ctx context.Context, key string, body []byte) (ref string, err error)
}

// LoggingOption customizes NewLoggingTransport.
type LoggingOption func(*LoggingTransport)

// WithParentRequestID sets X-Parent-Request-ID from the request context ("unknown" when empty).
func WithParentRequestID(fn func(ctx context.Context) string) LoggingOption {
	return func(t *LoggingTransport) { t.parentRequestID = fn }
}

// WithUserAgent overrides the User-Agent header of outgoing requests.
func WithUserAgent(userAgent string) LoggingOption {
	return func(t *LoggingTransport) { t.userAgent = userAgent }
}

// WithBodySink stores bodies that exceed the entry budget and logs a reference instead.
func WithBodySink(sink BodySink) LoggingOption {
	return func(t *LoggingTransport) { t.sink = sink }
}

// WithEntryBudget caps the serialized size of a log entry, in bytes.
func WithEntryBudget(bytes int) LoggingOption {
	return func(t *LoggingTransport) { t.budget = bytes }
}

// WithMaxCapture caps the bytes buffered per body when a BodySink is configured.
func WithMaxCapture(bytes int) LoggingOption {
	return func(t *LoggingTransport) { t.maxCapture = bytes }
}

// LoggingTransport logs every outgoing call as a redacted structured entry and traces it.
// Compose retries above it so that each try is logged on its own.
type LoggingTransport struct {
	apiName         string
	base            http.RoundTripper
	redactor        *redact.Redactor
	tracer          trace.Tracer
	parentRequestID func(ctx context.Context) string
	userAgent       string
	sink            BodySink
	budget          int
	maxCapture      int
	now             func() time.Time
	newID           func() string
}

// NewLoggingTransport wraps base (a default transport when nil); policy is merged on top of redact.Default().
func NewLoggingTransport(
	apiName string, base http.RoundTripper, policy redact.Policy, opts ...LoggingOption,
) *LoggingTransport {
	if base == nil {
		base = NewTransport()
	}

	t := &LoggingTransport{
		apiName:    apiName,
		base:       base,
		redactor:   redact.New(policy),
		tracer:     otel.GetTracerProvider().Tracer(tracerName),
		budget:     DefaultEntryBudget,
		maxCapture: DefaultMaxCapture,
		now:        time.Now,
		newID:      newRequestID,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// RoundTrip implements http.RoundTripper without mutating req.
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := t.now()
	childID := t.newID()

	ctx, span := t.tracer.Start(req.Context(), t.apiName+" outgoing HTTP call", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("target-api", t.apiName),
		attribute.String("http.request.method", req.Method),
		attribute.String("server.address", req.URL.Hostname()),
		attribute.String("url.path", req.URL.Path),
		attribute.String("request-id", childID),
	)

	attrs := []any{
		slog.String("target-api", t.apiName),
		slog.String("child-request-id", childID),
	}
	ctx = logger.CtxWithLogAttributes(ctx, attrs...)

	out := req.Clone(ctx)
	if out.Header == nil {
		out.Header = http.Header{}
	}

	out.Header.Set("X-Request-ID", childID)

	if t.parentRequestID != nil {
		parent := t.parentRequestID(ctx)
		if parent == "" {
			parent = "unknown"
		}

		out.Header.Set("X-Parent-Request-ID", parent)
	}

	if t.userAgent != "" {
		out.Header.Set("User-Agent", t.userAgent)
	}

	c := &call{
		t:        t,
		ctx:      ctx,
		span:     span,
		req:      out,
		childID:  childID,
		start:    start,
		reqBody:  t.newCapture(out.Header, out.ContentLength),
		respBody: nil,
	}

	if out.Body != nil && out.Body != http.NoBody {
		out.Body = &captureReadCloser{ReadCloser: out.Body, capture: c.reqBody}
	}

	resp, err := t.base.RoundTrip(out)
	if err != nil {
		c.finish(nil, err)

		return nil, logger.ErrWithAttributes(err, attrs...)
	}

	c.resp = resp
	span.SetAttributes(attribute.Int("http.response.status_code", resp.StatusCode))

	// A 101 body is a bidirectional stream that must reach the caller untouched.
	if resp.Body == nil || resp.Body == http.NoBody || resp.StatusCode == http.StatusSwitchingProtocols {
		c.finish(resp, nil)

		return resp, nil
	}

	c.respBody = t.newCapture(resp.Header, resp.ContentLength)
	resp.Body = &logOnCloseBody{ReadCloser: resp.Body, capture: c.respBody, done: func() { c.finish(resp, nil) }}

	return resp, nil
}

type call struct {
	t        *LoggingTransport
	ctx      context.Context //nolint:containedctx // carries the span and the log attributes until the body is closed
	span     trace.Span
	req      *http.Request
	resp     *http.Response
	childID  string
	start    time.Time
	reqBody  *capture
	respBody *capture
	once     sync.Once
}

func (c *call) finish(resp *http.Response, err error) {
	c.once.Do(func() {
		elapsed := c.t.now().Sub(c.start)

		c.t.emit(c, resp, elapsed)

		switch {
		case err != nil:
			c.span.SetStatus(codes.Error, "HTTP call failed")
			c.span.RecordError(err)
		case resp != nil && resp.StatusCode >= http.StatusBadRequest:
			c.span.SetStatus(codes.Error, http.StatusText(resp.StatusCode))
		}

		c.span.End()
	})
}

// logOnCloseBody tees the body into a bounded capture and logs the call once it is fully read or closed.
type logOnCloseBody struct {
	io.ReadCloser
	capture *capture
	done    func()
}

func (b *logOnCloseBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.capture.write(p[:n])

	if err != nil {
		b.done()
	}

	return n, err
}

func (b *logOnCloseBody) Close() error {
	err := b.ReadCloser.Close()
	b.done()

	return err
}

type captureReadCloser struct {
	io.ReadCloser
	capture *capture
}

func (b *captureReadCloser) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.capture.write(p[:n])

	return n, err
}

func newRequestID() string {
	var b [16]byte

	_, _ = rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func isBinaryContent(h http.Header) (label string, binary bool) {
	if enc := strings.ToLower(strings.TrimSpace(h.Get("Content-Encoding"))); enc != "" && enc != "identity" {
		return enc + "-encoded " + mediaType(h), true
	}

	mt := mediaType(h)

	switch {
	case mt == "application/pdf", mt == "application/octet-stream", mt == "application/zip",
		mt == "application/gzip", mt == "application/x-gzip",
		strings.HasPrefix(mt, "image/"), strings.HasPrefix(mt, "audio/"), strings.HasPrefix(mt, "video/"),
		strings.HasPrefix(mt, "multipart/"):
		return mt, true
	}

	return "", false
}

func mediaType(h http.Header) string {
	mt, _, _ := strings.Cut(h.Get("Content-Type"), ";")

	return strings.ToLower(strings.TrimSpace(mt))
}
