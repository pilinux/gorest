package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	sentry "github.com/getsentry/sentry-go"
	sentrylogrus "github.com/getsentry/sentry-go/logrus"
	"github.com/gin-gonic/gin"
	"github.com/pilinux/gorest/lib/middleware"
	log "github.com/sirupsen/logrus"
)

// testDSN is a well-formed DSN. It never reaches sentry.io in tests, but it lets
// sentry.NewClient succeed so the hook owns a non-nil client and events flow
// through BeforeSend.
const testDSN = "https://username@sentry.io/project-id"

// newGoroutineHook returns a hook that is NOT attached to the global logrus
// logger, so tests can call Fire directly without mutating global state.
func newGoroutineHook(t *testing.T) sentrylogrus.Hook {
	t.Helper()
	sentryDSN := os.Getenv("TEST_SENTRY_DSN")
	hook, err := middleware.NewSentryHook(sentryDSN, "production", "v0.0.1")
	if err != nil {
		t.Fatalf("failed to create sentry hook: %v", err)
	}
	if hook == nil {
		t.Fatal("expected a non-nil sentry hook")
	}
	t.Cleanup(func() { hook.Flush(2 * time.Second) })
	return hook
}

func TestSentryCapture(t *testing.T) {
	// set up a gin router and handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	err := router.SetTrustedProxies(nil)
	if err != nil {
		t.Errorf("failed to set trusted proxies to nil")
	}
	router.TrustedPlatform = "X-Real-Ip"

	// register middleware with valid sentry dsn
	sentryDSN := os.Getenv("TEST_SENTRY_DSN")
	_, err = middleware.InitSentry(sentryDSN, "production", "v0.0.1", "yes", "1.0")
	if err != nil {
		t.Errorf("failed to initialize sentry: %v", err)
	}
	// avoid leaking the global hook/log level into subsequent tests
	t.Cleanup(middleware.DestroySentry)
	router.Use(middleware.SentryCapture())

	// check sentry in a separate goroutine
	var wg sync.WaitGroup
	var GoroutineLogger *log.Logger
	sentryHook, err := middleware.NewSentryHook(
		sentryDSN,
		"production",
		"v0.0.1",
		"yes",
		"1.0",
	)
	if err != nil {
		t.Errorf("failed to initialize sentry for separate goroutines")
	}
	if err == nil {
		if sentryHook != nil {
			t.Cleanup(func() {
				// ensure sentry flushes all events before exiting
				sentryHook.Flush(5 * time.Second)
			})

			GoroutineLogger = log.New()
			GoroutineLogger.SetLevel(log.DebugLevel)
			GoroutineLogger.SetFormatter(&log.JSONFormatter{})
			GoroutineLogger.AddHook(sentryHook)
		}
	}
	if GoroutineLogger == nil {
		t.Errorf("failed to create a logger for separate goroutines")
	}

	if GoroutineLogger != nil {
		if sentryDSN != "" {
			wg.Go(func() {
				GoroutineLogger.
					WithFields(log.Fields{
						"time": time.Now().Format(time.RFC3339),
						"ref":  "goroutine - 1",
					}).
					Info("testing sentry integration in a separate goroutine")
			})
			wg.Wait()
		}
	}

	// define test route
	router.GET("/", func(c *gin.Context) {
		// send log to sentry for testing
		log.
			WithContext(c.Request.Context()).
			WithFields(log.Fields{
				"time": time.Now().Format(time.RFC3339),
				"ref":  "middleware",
			}).
			Info("testing sentry integration in the middleware")
		c.Status(http.StatusOK)
	})

	// perform request and get response
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("failed to create an HTTP request")
		return
	}
	req.Host = "localhost"
	req.RemoteAddr = "192.168.0.1"
	req.Header.Set("User-Agent", "Test-User-Agent")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	// check response
	if res.Code != http.StatusOK {
		t.Errorf("expected response code %v, got '%v'", http.StatusOK, res.Code)
	}

	// check sentry in another goroutine
	if GoroutineLogger != nil {
		if sentryDSN != "" {
			wg.Go(func() {
				GoroutineLogger.
					WithFields(log.Fields{
						"time": time.Now().Format(time.RFC3339),
						"ref":  "goroutine - 2",
					}).
					Info("testing sentry integration in a separate goroutine")
			})
			wg.Wait()
		}
	}
}

func TestInitSentryLogLevel(t *testing.T) {
	// test case 1: when sentry dsn is wrong
	t.Run("Sentry DSN wrong", func(t *testing.T) {
		testLogLevel("mockDSN", []string{"development"}, log.TraceLevel, t)
	})

	// test case 2: when v[0] is "production", log level should be InfoLevel
	t.Run("Log level production", func(t *testing.T) {
		testLogLevel("https://username@sentry.io/project-id", []string{"production"}, log.InfoLevel, t)
	})

	// test case 3: when v[0] is not "production", log level should be TraceLevel
	t.Run("Log level non-production", func(t *testing.T) {
		testLogLevel("https://username@sentry.io/project-id", []string{"development"}, log.TraceLevel, t)
	})

	// test case 4: when v is empty, log level should default to TraceLevel
	t.Run("Log level default", func(t *testing.T) {
		testLogLevel("https://username@sentry.io/project-id", []string{}, log.TraceLevel, t)
	})
}

// testLogLevel tests log level setting.
func testLogLevel(sentryDsn string, v []string, expectedLevel log.Level, t *testing.T) {
	// destroy any existing global hook
	middleware.DestroySentry()

	// set the log level using the InitSentry function
	_, err := middleware.InitSentry(sentryDsn, v...)
	if err == nil && sentryDsn == "mockDSN" {
		// if DSN is incorrect, expect an error
		t.Fatalf("Expected error, but got nil")
	}
	if err != nil && sentryDsn != "mockDSN" {
		// if DSN is correct, but there was an error
		t.Fatalf("Error initializing sentry: %v", err)
	}
	if err != nil && sentryDsn == "mockDSN" {
		// if DSN is incorrect and there was an error, stop the test here
		return
	}

	// check if the log level was set correctly
	if log.GetLevel() != expectedLevel {
		t.Errorf("Expected log level %v, but got %v", expectedLevel, log.GetLevel())
	}
}

// TestSentryCombinedHookFire exercises the Fire/captureIssue paths that are not
// reached by the higher-level middleware test.
func TestSentryCombinedHookFire(t *testing.T) {
	// Entries at or below FatalLevel trigger the in-Fire flush branch before
	// the entry is delegated to the underlying log hook. A FatalLevel entry would
	// os.Exit(1) inside that hook, so PanicLevel (which is also <= FatalLevel and
	// likewise exercises the flush) is used: it panics instead of exiting, and the
	// panic is recovered so the test process survives.
	//
	//	if entry.Level <= log.FatalLevel {
	//		h.Flush(2 * time.Second)
	//	}
	t.Run("panic level flushes before delegating", func(t *testing.T) {
		hook := newGoroutineHook(t)
		entry := &log.Entry{
			Level:   log.PanicLevel,
			Message: "testing sentry hook panic-level flush",
			Data:    log.Fields{},
		}
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected the underlying log hook to panic on a panic-level entry")
			}
		}()
		_ = hook.Fire(entry)
		t.Errorf("Fire should not return for a panic-level entry")
	})

	// An entry whose context carries a sentry hub makes scopeFromEntry clone
	// the request-scoped hub's scope.
	//
	//	if entry.Context != nil {
	//		if hub := sentry.GetHubFromContext(entry.Context); hub != nil && hub.Scope() != nil {
	//			return hub.Scope().Clone()
	//		}
	//	}
	t.Run("context hub clones scope", func(t *testing.T) {
		hook := newGoroutineHook(t)
		hub := sentry.NewHub(nil, sentry.NewScope())
		hub.Scope().SetTag("origin", "context-hub")
		ctx := sentry.SetHubOnContext(context.Background(), hub)
		entry := &log.Entry{
			Context: ctx,
			Level:   log.ErrorLevel,
			Message: "testing sentry hook scopeFromEntry context branch",
			Data:    log.Fields{},
		}
		if err := hook.Fire(entry); err != nil {
			t.Errorf("Fire returned an unexpected error: %v", err)
		}
	})

	// A non-string field value is converted with fmt.Sprint in the default
	// branch of the tag loop.
	//
	//	default:
	//		hub.Scope().SetTag(k, fmt.Sprint(v))
	//	}
	t.Run("non-string field tag", func(t *testing.T) {
		hook := newGoroutineHook(t)
		entry := &log.Entry{
			Level:   log.WarnLevel,
			Message: "testing sentry hook non-string field tag",
			Data:    log.Fields{"attempts": 3, "ok": true},
		}
		if err := hook.Fire(entry); err != nil {
			t.Errorf("Fire returned an unexpected error: %v", err)
		}
	})

	// An error stored under log.ErrorKey is captured as an exception.
	//
	//	if errVal, ok := entry.Data[log.ErrorKey].(error); ok {
	//		hub.Scope().SetContext("logrus", sentry.Context{"message": entry.Message})
	//		hub.CaptureException(errVal)
	//		return
	//	}
	t.Run("error field captured as exception", func(t *testing.T) {
		hook := newGoroutineHook(t)
		entry := &log.Entry{
			Level:   log.ErrorLevel,
			Message: "testing sentry hook error field captured as exception",
			Data:    log.Fields{log.ErrorKey: errors.New("boom")},
		}
		if err := hook.Fire(entry); err != nil {
			t.Errorf("Fire returned an unexpected error: %v", err)
		}
	})

	// An out-of-range level is dropped by the log hook, which then invokes
	// the fallback closure set in createSentryHook.
	//
	//	hook.SetFallback(func(entry *log.Entry) error {
	//		log.
	//			WithContext(entry.Context).
	//			WithFields(entry.Data).
	//			Error("Sentry fallback executed")
	//
	//		return nil
	//	})
	t.Run("invalid level triggers fallback", func(t *testing.T) {
		hook := newGoroutineHook(t)
		entry := &log.Entry{
			Level:   log.Level(99),
			Message: "testing sentry hook invalid level triggers fallback",
			Data:    log.Fields{},
		}
		if err := hook.Fire(entry); err != nil {
			t.Errorf("Fire returned an unexpected error: %v", err)
		}
	})
}

// TestInitSentryDoubleHook covers the guard that returns the existing global hook
// instead of attaching a second one.
//
//	if globalSentryHook != nil {
//		return globalSentryHook, nil
//	}
func TestInitSentryDoubleHook(t *testing.T) {
	middleware.DestroySentry()
	t.Cleanup(middleware.DestroySentry)

	first, err := middleware.InitSentry(testDSN, "production")
	if err != nil {
		t.Fatalf("failed to initialize sentry: %v", err)
	}
	if first == nil {
		t.Fatal("expected a non-nil hook on first init")
	}

	second, err := middleware.InitSentry(testDSN, "development")
	if err != nil {
		t.Fatalf("second InitSentry returned an error: %v", err)
	}
	if second != first {
		t.Errorf("expected the second InitSentry to return the existing global hook")
	}

	middleware.DestroySentry()
	third, err := middleware.InitSentry(testDSN, "production")
	if err != nil {
		t.Fatalf("failed to initialize sentry after destroy: %v", err)
	}
	if third == nil {
		t.Fatal("expected a non-nil hook on third init")
	}
	if third == first {
		t.Errorf("expected a new hook after destroy, but got the previous hook")
	}
}
