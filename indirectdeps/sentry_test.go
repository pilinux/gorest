package indirectdeps_test

// Sentry path: middleware.NewSentryHook (logrus hook -> sentry-go) with no
// release passed; sentry-go runs git describe through x/sys/execabs.
//
// Indirect modules:
//   - golang.org/x/sys
//
//	Identifier                     Purpose
//	-----------------------------  -------
//	sentryReleaseEnvs              env vars sentry-go reads for the release before git describe;
//	                               the test empties each one
//	TestSentryReleaseAndRateLimit  a fake envelope endpoint records each event's release; it
//	                               must equal `git describe --long --always --dirty`. After the
//	                               first event the endpoint returns X-Sentry-Rate-Limits (error
//	                               plus an unknown category) and the second event must be
//	                               dropped. Skips without git or outside a checkout

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	gmiddleware "github.com/pilinux/gorest/lib/middleware"
)

// env vars sentry-go checks before falling back to `git describe`
var sentryReleaseEnvs = []string{
	"SENTRY_RELEASE", "HEROKU_BUILD_COMMIT", "HEROKU_SLUG_COMMIT", "SOURCE_VERSION",
	"CODEBUILD_RESOLVED_SOURCE_VERSION", "CIRCLE_SHA1", "GAE_DEPLOYMENT_ID", "GITHUB_SHA",
	"COMMIT_REF", "VERCEL_GIT_COMMIT_SHA", "ZEIT_GITHUB_COMMIT_SHA", "ZEIT_GITLAB_COMMIT_SHA",
	"ZEIT_BITBUCKET_COMMIT_SHA",
}

// TestSentryReleaseAndRateLimit covers golang.org/x/sys/execabs: our hook
// passes no release, so sentry-go runs `git describe` through execabs. It
// also checks the transport honors X-Sentry-Rate-Limits.
func TestSentryReleaseAndRateLimit(t *testing.T) {
	logIndirect(t)
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	out, err := exec.Command("git", "describe", "--long", "--always", "--dirty").Output()
	if err != nil {
		t.Skipf("not inside a git checkout: %v", err)
	}
	wantRelease := strings.TrimSpace(string(out))
	for _, env := range sentryReleaseEnvs {
		t.Setenv(env, "")
	}

	var (
		mu       sync.Mutex
		releases []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		sc := bufio.NewScanner(bytes.NewReader(body))
		sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		sc.Scan() // envelope header
		sawEvent := false
		for sc.Scan() {
			var item struct {
				Type string `json:"type"`
			}
			_ = json.Unmarshal(sc.Bytes(), &item)
			if !sc.Scan() {
				break
			}
			if item.Type != "event" {
				continue
			}
			var ev struct {
				Release string `json:"release"`
			}
			_ = json.Unmarshal(sc.Bytes(), &ev)
			sawEvent = true
			mu.Lock()
			releases = append(releases, ev.Release)
			mu.Unlock()
		}
		// the hook also ships a separate log envelope; only throttle after the event
		if sawEvent {
			w.Header().Set("X-Sentry-Rate-Limits", "60:error:organization, 60:mystery_category:organization")
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	dsn := strings.Replace(srv.URL, "http://", "http://public@", 1) + "/1" // ggignore
	hook, err := gmiddleware.NewSentryHook(dsn, "production")
	if err != nil {
		t.Fatalf("NewSentryHook: %v", err)
	}
	logger := log.New()
	logger.Out = io.Discard

	fire := func(msg string) {
		entry := log.NewEntry(logger).WithField("code", "1001.1")
		entry.Level = log.ErrorLevel
		entry.Message = msg
		entry.Time = time.Now()
		if err := hook.Fire(entry); err != nil {
			t.Fatalf("Fire: %v", err)
		}
		hook.Flush(5 * time.Second)
	}

	fire("first")
	mu.Lock()
	got := append([]string(nil), releases...)
	mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("events received = %d, want 1", len(got))
	}
	if got[0] != wantRelease {
		t.Errorf("release = %q, want git describe %q", got[0], wantRelease)
	}

	fire("second, rate limited")
	mu.Lock()
	defer mu.Unlock()
	if len(releases) != 1 {
		t.Errorf("events received = %d after rate limit, want 1", len(releases))
	}
}
