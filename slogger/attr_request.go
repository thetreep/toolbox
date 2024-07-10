package logger

// REVIEW: https://github.com/thetreep/express-api/pull/1#discussion_r1560911950
// REVIEW: move this whole package to toolbox package?

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
)

// Request transforms a http.Request into a slog.Attr by extracting various information and grouping them.
// The `withBody` parameter controls if the request's body should be logged or completely obfuscated.
func Request(r *http.Request, withBody bool) slog.Attr {
	headers := maps.Clone(r.Header)
	// obfuscate Authorization header so we don't log secrets
	if _, ok := headers["Authorization"]; ok {
		headers.Set("Authorization", "OBFUSCATED")
	}
	// obfuscate Cookie header, because it often contains sensitive data
	if _, ok := headers["Cookie"]; ok {
		headers.Set("Cookie", "OBFUSCATED")
	}

	return slog.Group(
		"request",
		slog.String("hostname", r.URL.Hostname()),
		slog.String("path", r.URL.EscapedPath()),
		queryToLogAttrs(r.URL.Query()),
		slog.String("method", r.Method),
		slog.Any("headers", headers),
		requestBodyToAttr(r, withBody),
	)
}

func requestBodyToAttr(r *http.Request, withBody bool) slog.Attr {
	if !withBody {
		return slog.String("no_body", "body logging disabled")
	}

	if r.GetBody == nil {
		if r.Body == nil {
			return slog.String("no_body", "Body == nil")
		}
		defer r.Body.Close()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return slog.String("no_body", fmt.Sprintf("ReadAll err: %s", err.Error()))
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		r.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	bodyReader, err := r.GetBody()
	if err != nil {
		return slog.String("no_body", fmt.Sprintf("GetBody err: %s", err.Error()))
	}
	bodyBytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return slog.String("no_body", fmt.Sprintf("ReadAll err: %s", err.Error()))
	}
	return slog.String("body", string(bodyBytes))
}
