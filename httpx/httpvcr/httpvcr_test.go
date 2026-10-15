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

// A cassette recorded by the v1 client (OBFUSCATED markers, pretty-printed bodies) replays unchanged.
func TestReplaysV1Cassette(t *testing.T) {
	t.Setenv("CI", "1")

	policy := redact.Policy{Volatile: redact.Rules{Headers: []string{"x-goog-api-client"}}}
	client, recording := httpvcr.NewClient(t, "gmaps", policy, nil,
		httpx.WithUserAgent("my-service"), httpx.WithParentRequestID(func(context.Context) string { return "" }))
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
		"hunter2", "live-secret", "KEY-1", "very-secret-cookie", "resp-secret-token", "n-1",
	} {
		assert.NotContains(t, string(raw), secret, "cassette must not hold %q", secret)
	}

	for _, kept := range []string{"jean.dupont@example.com", "06 12 34 56 78"} {
		assert.Contains(t, string(raw), kept, "PII is left as is in cassettes")
	}

	assert.Contains(t, string(raw), "VOLATILE")
	assert.NotContains(t, string(raw), `"92"`, "Content-Length header follows the redacted body")

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

// Requests differing only by a PII value must each replay their own response, whatever the call order.
func TestReplayKeepsRequestsDifferingByPIIApart(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("K7_RECORD", "")

	dir := chdirTemp(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"greeting":"hello %s","lat":%q}`, strings.Trim(string(body), `{}"name:`), r.URL.Query().Get("lat"))
	}))
	defer srv.Close()

	call := func(client *http.Client, name, lat string) string {
		t.Helper()

		req, err := http.NewRequest(http.MethodPost, srv.URL+"/hello?lat="+lat, strings.NewReader(`{"name":"`+name+`"}`))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)

		got, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		return string(got)
	}

	policy := redact.Policy{PII: redact.Rules{Keys: []string{"name"}, QueryParams: []string{"lat"}}}

	recorder := &phaseTB{TB: t, name: "Phase"}
	client, recording := httpvcr.NewClient(recorder, "api", policy, nil)
	assert.True(t, recording)

	jean := call(client, "Jean", "48.85")
	marie := call(client, "Marie", "2.35")
	recorder.finish()

	require.NotEqual(t, jean, marie)

	cassettes, err := filepath.Glob(filepath.Join(dir, "testdata", "*.k7.yaml"))
	require.NoError(t, err)
	require.Len(t, cassettes, 1)

	raw, err := os.ReadFile(cassettes[0])
	require.NoError(t, err)

	for _, kept := range []string{"Jean", "Marie", "48.85", "2.35"} {
		assert.Contains(t, string(raw), kept)
	}

	t.Setenv("CI", "1")

	replayer := &phaseTB{TB: t, name: "Phase"}
	client, recording = httpvcr.NewClient(replayer, "api", policy, nil)
	assert.False(t, recording)

	assert.Equal(t, marie, call(client, "Marie", "2.35"))
	assert.Equal(t, jean, call(client, "Jean", "48.85"))
	replayer.finish()
}
