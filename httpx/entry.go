package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/thetreep/toolbox/httpx/redact"
	"github.com/thetreep/toolbox/logger"
)

const (
	// entrySlack leaves room for the fields the log handler and the context add around our attributes.
	entrySlack = 4096
	// entryStructure covers keys, braces and the short fixed fields of an entry.
	entryStructure = 1536
	maxHeaderValue = 4096
)

// loggedBody is a redacted body ready to be sized and rendered.
type loggedBody struct {
	size        int64
	body        []byte
	jsonOK      bool
	captureCut  bool
	placeholder string
}

func (t *loggingTransport) prepareBody(h http.Header, s captureSnapshot) loggedBody {
	lb := loggedBody{size: s.total, captureCut: s.truncated}

	if s.binary {
		lb.placeholder = fmt.Sprintf("[binary: %s, %d bytes]", s.binaryLabel, s.total)

		return lb
	}

	res := t.redactor.Body(h.Get("Content-Type"), s.buf)
	lb.body = res.Body
	// encoding/json is the final judge: an invalid RawMessage would make the log handler drop the attribute.
	lb.jsonOK = res.Kind == redact.KindJSON && res.ValidJSON && !s.truncated && json.Valid(res.Body)

	return lb
}

// need is the serialized size of the body when logged in full.
func (b loggedBody) need() int {
	switch {
	case b.placeholder != "":
		return len(b.placeholder) + 2
	case b.jsonOK:
		return rawJSONSize(b.body)
	default:
		return escapedSize(b.body) + 2
	}
}

type renderedBody struct {
	value     any
	truncated bool
	ref       string
}

func (t *loggingTransport) emit(c *call, resp *http.Response, elapsed time.Duration, callErr error) {
	req := c.req
	reqHeaders := t.redactHeaders(req.Header)
	query := t.redactor.Query(req.URL.Query())
	path := t.redactedPath(req.URL)
	hostname := req.URL.Hostname()

	reqSnap := c.reqBody.snapshot()
	reqBody := t.prepareBody(req.Header, reqSnap)

	var (
		respHeaders http.Header
		respBody    loggedBody
		status      = "no response"
	)

	if resp != nil {
		respHeaders = t.redactHeaders(resp.Header)
		respBody = t.prepareBody(resp.Header, c.respBody.snapshot())
		status = strconv.Itoa(resp.StatusCode)
	}

	summary := fmt.Sprintf("%s %s %s%s -> %s (%s)", t.apiName, req.Method, req.URL.Host, path, status, elapsed.Round(time.Millisecond))

	overhead := entrySlack + entryStructure + 2*len(summary) + 2*len(path) + 2*len(hostname) +
		jsonSize(reqHeaders) + jsonSize(query) + jsonSize(respHeaders)
	room := max(t.budget-overhead, 0)

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}

	reqRoom, respRoom := splitBudget(room, reqBody.need(), respBody.need(), resp != nil, statusCode >= http.StatusBadRequest)

	sinkCtx := context.WithoutCancel(c.ctx)
	prefix := c.start.UTC().Format("2006-01-02") + "/" + t.apiName + "/" + c.childID + "/"

	reqRendered := t.renderBody(sinkCtx, prefix+"request", reqBody, reqRoom)
	requestAttr := slog.Group("request",
		slog.String("hostname", hostname),
		slog.String("path", path),
		slog.Any("query", query),
		slog.String("method", req.Method),
		slog.Any("headers", reqHeaders),
		slog.Any("body", reqRendered.value),
		slog.Int64("bodySize", reqBody.size),
	)
	requestAttr = withBodyExtras(requestAttr, reqRendered)

	var responseAttr any = slog.Any("response", nil)

	if resp != nil {
		respRendered := t.renderBody(sinkCtx, prefix+"response", respBody, respRoom)
		attr := slog.Group("response",
			slog.String("status", resp.Status),
			slog.Int("statusCode", resp.StatusCode),
			slog.Any("headers", respHeaders),
			slog.Any("body", respRendered.value),
			slog.Int64("bodySize", respBody.size),
		)
		responseAttr = withBodyExtras(attr, respRendered)
	}

	attrs := []any{
		slog.String("event", "outgoing_http_call"),
		slog.Duration("durationNanoSeconds", elapsed),
		slog.String("duration", elapsed.String()),
		requestAttr,
		responseAttr,
	}

	if callErr != nil {
		logger.Warn(c.ctx, summary, append(attrs, slog.String("error", t.redactor.Text(callErr.Error())))...)

		return
	}

	logger.Info(c.ctx, summary, attrs...)
}

func withBodyExtras(group slog.Attr, r renderedBody) slog.Attr {
	attrs := group.Value.Group()

	if r.ref != "" {
		attrs = append(attrs, slog.String("bodyRef", r.ref))
	}

	if r.truncated {
		attrs = append(attrs, slog.Bool("bodyTruncated", true))
	}

	return slog.Attr{Key: group.Key, Value: slog.GroupValue(attrs...)}
}

func (t *loggingTransport) redactHeaders(h http.Header) http.Header {
	out := t.redactor.Header(h)

	for name, values := range out {
		for i, v := range values {
			if len(v) > maxHeaderValue {
				out[name][i] = v[:cutRuneBoundary(v, maxHeaderValue)] + "…"
			}
		}
	}

	return out
}

func cutRuneBoundary(s string, n int) int {
	for n > 0 && n < len(s) && !utf8.RuneStart(s[n]) {
		n--
	}

	return n
}

// splitBudget shares room between the bodies: evenly, or response first when the call failed.
func splitBudget(room, needReq, needResp int, hasResp, failed bool) (reqRoom, respRoom int) {
	if !hasResp {
		return room, 0
	}

	if needReq+needResp <= room {
		return needReq, needResp
	}

	if failed {
		respRoom = min(needResp, room-min(needReq, room/4))

		return room - respRoom, respRoom
	}

	half := room / 2

	switch {
	case needReq <= half:
		return needReq, room - needReq
	case needResp <= half:
		return room - needResp, needResp
	default:
		return half, room - half
	}
}

func (t *loggingTransport) renderBody(ctx context.Context, key string, b loggedBody, room int) renderedBody {
	if b.placeholder != "" {
		return renderedBody{value: b.placeholder}
	}

	if !b.captureCut && b.need() <= room {
		if b.jsonOK {
			return renderedBody{value: json.RawMessage(b.body)}
		}

		return renderedBody{value: string(b.body)}
	}

	r := renderedBody{truncated: true}

	if t.sink != nil && len(b.body) > 0 {
		ref, err := t.sink.Put(ctx, key, b.body)
		if err != nil {
			logger.Warn(ctx, "could not store the full HTTP body, keeping the preview only",
				slog.String("error", err.Error()), slog.String("bodyKey", key))
		} else {
			r.ref = ref
		}
	}

	total := int64(len(b.body))
	if b.captureCut {
		total = b.size
	}

	r.value = truncatedPreview(b.body, total, room)

	return r
}

// truncatedPreview cuts body on a rune boundary so that its serialized form, marker included, fits in room.
func truncatedPreview(body []byte, total int64, room int) string {
	keep := fitEscaped(body, room-2-len(fmt.Sprintf("…[truncated %d/%d bytes]", len(body), total))-8)

	return fmt.Sprintf("%s…[truncated %d/%d bytes]", body[:keep], keep, total)
}

// fitEscaped returns the longest rune-aligned prefix length whose serialized size is at most room.
func fitEscaped(b []byte, room int) int {
	used := 0

	for i := 0; i < len(b); {
		r, w := utf8.DecodeRune(b[i:])
		cost := runeCost(r, w)

		if used+cost > room {
			return i
		}

		used += cost
		i += w
	}

	return len(b)
}

// runeCost is a worst case over the JSON encoders in play (HTML-escaping or not).
func runeCost(r rune, width int) int {
	switch {
	case r == utf8.RuneError && width == 1, r == 0x2028, r == 0x2029:
		return 6
	case r == '"', r == '\\', r == '\n', r == '\r', r == '\t':
		return 2
	case r < 0x20, r == '<', r == '>', r == '&':
		return 6
	}

	return width
}

func escapedSize(b []byte) int {
	size := 0

	for i := 0; i < len(b); {
		r, w := utf8.DecodeRune(b[i:])
		size += runeCost(r, w)
		i += w
	}

	return size
}

func rawJSONSize(b []byte) int {
	return len(b) + 5*(bytes.Count(b, []byte("<"))+bytes.Count(b, []byte(">"))+bytes.Count(b, []byte("&"))) +
		3*(bytes.Count(b, []byte("\u2028"))+bytes.Count(b, []byte("\u2029")))
}

func jsonSize(v any) int {
	b, err := json.Marshal(v)
	if err != nil {
		return 0
	}

	return len(b)
}
