package httpvcr_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thetreep/toolbox/httpx"
	"github.com/thetreep/toolbox/httpx/httpvcr"
	"github.com/thetreep/toolbox/httpx/redact"
	"github.com/thetreep/toolbox/logger"
)

func chdirTemp(t *testing.T) string {
	t.Helper()

	previous, err := os.Getwd()
	require.NoError(t, err)

	dir := t.TempDir()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { require.NoError(t, os.Chdir(previous)) })

	return dir
}

// A cassette recorded by the former pelican client (OBFUSCATED markers, pretty-printed bodies) replays unchanged.
func TestReplaysV1Cassette(t *testing.T) {
	t.Setenv("CI", "1")

	policy := redact.Policy{Volatile: redact.Rules{Headers: []string{"x-goog-api-client"}}}
	client, recording := httpvcr.NewClient(t, "gmaps", policy, nil,
		httpx.WithUserAgent("imaas"), httpx.WithParentRequestID(func(context.Context) string { return "" }))
	assert.False(t, recording)

	body := `{"origin":{"location":{"latLng":{"latitude":48.8566,"longitude":2.3522}}},` +
		`"destination":{"location":{"latLng":{"latitude":48.8673,"longitude":2.3522}}},"travelMode":3,` +
		`"departureTime":"2026-04-21T20:42:42.000000042+07:00","languageCode":"FR","regionCode":"FR","units":1}`

	ctx := logger.CtxWithLogger(context.Background(), slog.New(slog.NewJSONHandler(io.Discard, nil)))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://routes.googleapis.com/directions/v2:computeRoutes?%24alt=json%3Benum-encoding%3Dint&key=AIzaLIVE-KEY",
		strings.NewReader(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header["x-goog-api-client"] = []string{"gl-go/1.27 gax/2.0"}
	req.Header["x-goog-fieldmask"] = []string{"routes.duration,routes.distanceMeters"}

	resp, err := client.Do(req)
	require.NoError(t, err)

	got, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(got), `"distanceMeters": 1444`)

	unknown, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://routes.googleapis.com/another", nil)
	require.NoError(t, err)

	_, err = client.Do(unknown)
	assert.Error(t, err, "a request that was never recorded must not match")
}

func TestRecordThenReplayWithoutLeakingSecrets(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("K7_RECORD", "")

	dir := chdirTemp(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Set-Cookie", "sid=very-secret-cookie")
		_, _ = io.WriteString(w, `{"ok":true,"access_token":"resp-secret-token","traveler":{"mail":"jean.dupont@example.com"}}`)
	}))
	defer srv.Close()

	call := func(client *http.Client, apiKey, nonce string) {
		t.Helper()

		body := `{"query":"paris","password":"hunter2","nonce":"` + nonce + `","contact":"06 12 34 56 78"}`
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/search?q=paris&key="+apiKey, strings.NewReader(body))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer live-secret-"+apiKey)

		resp, err := client.Do(req)
		require.NoError(t, err)

		got, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		assert.Contains(t, string(got), `"ok":true`)
	}

	policy := redact.Policy{Volatile: redact.Rules{Keys: []string{"nonce"}}}

	recorder := &phaseTB{TB: t, name: "Phase"}
	client, recording := httpvcr.NewClient(recorder, "api", policy, nil)
	assert.True(t, recording)
	call(client, "KEY-1", "n-1")
	recorder.finish()

	cassettes, err := filepath.Glob(filepath.Join(dir, "testdata", "*.k7.yaml"))
	require.NoError(t, err)
	require.Len(t, cassettes, 1)

	raw, err := os.ReadFile(cassettes[0])
	require.NoError(t, err)

	for _, secret := range []string{
		"hunter2", "live-secret", "KEY-1", "very-secret-cookie", "resp-secret-token", "jean.dupont", "06 12 34 56 78", "n-1",
	} {
		assert.NotContains(t, string(raw), secret, "cassette must not hold %q", secret)
	}

	assert.Contains(t, string(raw), "VOLATILE")
	assert.NotContains(t, string(raw), `"92"`, "Content-Length header follows the redacted body")
	httpvcr.AssertNoPII(t, filepath.Join(dir, "testdata"))

	t.Setenv("CI", "1")

	replayer := &phaseTB{TB: t, name: "Phase"}
	client, recording = httpvcr.NewClient(replayer, "api", policy, nil)
	assert.False(t, recording)
	// Secrets and volatile values changed since the recording; the interaction still matches.
	call(client, "KEY-2", "n-2")
	replayer.finish()
}

// phaseTB gives two NewClient calls in one test the same cassette and lets the test stop the recorder between them.
type phaseTB struct {
	testing.TB
	name     string
	cleanups []func()
}

func (p *phaseTB) Name() string      { return p.name }
func (p *phaseTB) Cleanup(fn func()) { p.cleanups = append(p.cleanups, fn) }
func (p *phaseTB) finish() {
	for _, fn := range p.cleanups {
		fn()
	}
}

type captureTB struct {
	testing.TB
	errors []string
}

func (c *captureTB) Errorf(format string, args ...any) {
	c.errors = append(c.errors, fmt.Sprintf(format, args...))
}

func TestAssertNoPIIReportsClassesNeverValues(t *testing.T) {
	dir := t.TempDir()

	cassette := `---
version: 2
interactions:
    - id: 0
      request:
        proto: HTTP/1.1
        proto_major: 1
        proto_minor: 1
        content_length: 0
        host: api.example.test
        body: '{"contact":"jean.dupont@example.com"}'
        headers:
            Authorization:
                - Bearer abcdef1234567890xyz
        url: https://api.example.test/x
        method: POST
      response:
        proto: HTTP/1.1
        proto_major: 1
        proto_minor: 1
        content_length: 2
        body: '{"card":"4111 1111 1111 1111"}'
        headers: {}
        status: 200 OK
        code: 200
        duration: 1ms
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Leaky_api.k7.yaml"), []byte(cassette), 0o644))

	tb := &captureTB{TB: t}
	httpvcr.AssertNoPII(tb, dir)

	require.Len(t, tb.errors, 3)
	joined := strings.Join(tb.errors, "\n")
	assert.Contains(t, joined, "Leaky_api.k7.yaml")
	assert.Contains(t, joined, "email")
	assert.Contains(t, joined, "bearer")
	assert.Contains(t, joined, "pan")

	for _, leaked := range []string{"jean.dupont", "abcdef1234567890xyz", "4111"} {
		assert.NotContains(t, joined, leaked)
	}
}
