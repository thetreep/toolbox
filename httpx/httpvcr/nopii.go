package httpvcr

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"

	"github.com/thetreep/toolbox/httpx/redact"
)

// AssertNoPII fails for each cassette under dir holding detectable data, reporting file, field and class only.
func AssertNoPII(t testing.TB, dir string) {
	t.Helper()

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".k7.yaml") {
			return err
		}

		c, loadErr := cassette.Load(strings.TrimSuffix(path, ".yaml"))
		if loadErr != nil {
			t.Errorf("%s: cannot load cassette: %v", path, loadErr)

			return nil
		}

		for n, i := range c.Interactions {
			scanFields(t, path, n, "request", i.Request.URL, i.Request.Headers, i.Request.Body)
			scanFields(t, path, n, "response", "", i.Response.Headers, i.Response.Body)
		}

		return nil
	})
	if err != nil {
		t.Errorf("walking %s: %v", dir, err)
	}
}

func scanFields(t testing.TB, file string, n int, side, url string, headers http.Header, body string) {
	t.Helper()

	check := func(field, text string) {
		if classes := redact.Scan([]byte(text)); len(classes) > 0 {
			t.Errorf("%s: interaction %d %s %s holds unredacted data (%s)", file, n, side, field, strings.Join(classes, ", "))
		}
	}

	check("url", url)
	check("body", body)

	for name, values := range headers {
		for _, v := range values {
			check("header "+name, v)
		}
	}
}
