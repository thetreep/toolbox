package logger

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
)

// Response transforms a http.Response into a slog.Attr by extracting various information and grouping them.
// The `withBody` parameter controls if the response's body should be logged or completely obfuscated.
func Response(r *http.Response, withBody bool) slog.Attr {
	if r == nil {
		return slog.Any("response", nil)
	}
	headers := maps.Clone(r.Header)
	// obfuscate Set-Cookie header, because it often contains sensitive data
	if _, ok := headers["Set-Cookie"]; ok {
		headers.Set("Set-Cookie", "OBFUSCATED")
	}

	return slog.Group(
		"response",
		slog.String("status", r.Status),
		slog.Int("statusCode", r.StatusCode),
		slog.Any("headers", headers),
		responseBodyToAttr(r, withBody),
	)
}

func responseBodyToAttr(r *http.Response, withBody bool) slog.Attr {
	if !withBody {
		return slog.String("no_body", "body logging disabled")
	}

	if r.Body == nil {
		return slog.String("no_body", "Body == nil")
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return slog.String("no_body", fmt.Sprintf("ReadAll err: %s", err.Error()))
	}
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	return slog.String("body", string(bodyBytes))
}
