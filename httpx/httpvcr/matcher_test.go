package httpvcr

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"

	"github.com/thetreep/toolbox/httpx/redact"
)

func savedRequest() cassette.Request {
	return cassette.Request{
		Proto: "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1,
		Host:          "api.example.test",
		Method:        http.MethodPost,
		URL:           "https://api.example.test/x?key=OBFUSCATED&q=paris",
		Headers:       http.Header{"Content-Type": {"application/json"}, "X-Request-Id": {"OBFUSCATED"}, "Authorization": {"OBFUSCATED"}},
		Body:          "{\n\t\"password\": \"OBFUSCATED\",\n\t\"n\": 1\n}",
		ContentLength: 38,
		Form:          url.Values{"key": {"OBFUSCATED"}, "q": {"paris"}},
	}
}

func liveRequest(t *testing.T, body string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, "https://api.example.test/x?q=paris&key=REAL-KEY", io.NopCloser(strings.NewReader(body)))
	require.NoError(t, err)

	req.GetBody = nil
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Id", "real-id")
	req.Header.Set("Authorization", "Bearer real-token-1234567890")

	return req
}

func TestMatcherComparesNormalizedCopiesAndLeavesBothSidesUntouched(t *testing.T) {
	red := cassetteRedactor(redact.Policy{})
	match := matcher(red)

	live := liveRequest(t, `{"password":"hunter2","n":1}`)
	saved := savedRequest()
	savedBefore := savedRequest()

	assert.True(t, match(live, saved), "formatting, secrets and volatile headers must not matter")

	assert.Equal(t, "https://api.example.test/x?q=paris&key=REAL-KEY", live.URL.String())
	assert.Equal(t, "Bearer real-token-1234567890", live.Header.Get("Authorization"))
	assert.Equal(t, "real-id", live.Header.Get("X-Request-Id"))

	body, err := io.ReadAll(live.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"password":"hunter2","n":1}`, string(body), "the body stays readable for the real call")

	assert.Equal(t, savedBefore, saved)
}

func TestMatcherDetectsRealDifferences(t *testing.T) {
	match := matcher(cassetteRedactor(redact.Policy{}))

	assert.False(t, match(liveRequest(t, `{"password":"hunter2","n":2}`), savedRequest()), "a different body")

	other := liveRequest(t, `{"password":"hunter2","n":1}`)
	other.URL.RawQuery = "q=lyon&key=REAL-KEY"
	assert.False(t, match(other, savedRequest()), "a different query")

	other = liveRequest(t, `{"password":"hunter2","n":1}`)
	other.Header.Set("Content-Type", "text/plain")
	assert.False(t, match(other, savedRequest()), "a different header")
}

func TestMatcherDistinguishesRequestsDifferingByPIIOnly(t *testing.T) {
	match := matcher(cassetteRedactor(redact.Policy{PII: redact.Rules{Keys: []string{"name"}, QueryParams: []string{"lat"}}}))

	saved := savedRequest()
	saved.Body = `{"name":"Jean","n":1}`
	saved.URL = "https://api.example.test/x?lat=48.85&q=paris"
	saved.Form = url.Values{"lat": {"48.85"}, "q": {"paris"}}

	live := liveRequest(t, `{"name":"Jean","n":1}`)
	live.URL.RawQuery = "q=paris&lat=48.85"
	assert.True(t, match(live, saved))

	live = liveRequest(t, `{"name":"Marie","n":1}`)
	live.URL.RawQuery = "q=paris&lat=48.85"
	assert.False(t, match(live, saved), "a different PII body value")

	live = liveRequest(t, `{"name":"Jean","n":1}`)
	live.URL.RawQuery = "q=paris&lat=2.35"
	assert.False(t, match(live, saved), "a different PII query value")
}
