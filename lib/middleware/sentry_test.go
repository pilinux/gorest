package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/gorest/lib/middleware"
	log "github.com/sirupsen/logrus"
)

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
			wg.Add(1)
			go func() {
				defer wg.Done()
				GoroutineLogger.
					WithFields(log.Fields{
						"time": time.Now().Format(time.RFC3339),
						"ref":  "goroutine - 1",
					}).
					Info("testing sentry integration in a separate goroutine")
			}()
			wg.Wait()
		}
	}

	// define test route
	router.GET("/", func(c *gin.Context) {
		// send log to sentry for testing
		log.
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
			wg.Add(1)
			go func() {
				defer wg.Done()
				GoroutineLogger.
					WithFields(log.Fields{
						"time": time.Now().Format(time.RFC3339),
						"ref":  "goroutine - 2",
					}).
					Info("testing sentry integration in a separate goroutine")
			}()
			wg.Wait()
		}
	}
}

func TestInitSentryLogLevel(t *testing.T) {
	// test case 1: when sentry dsn is wrong
	t.Run("Sentry DSN wrong", func(t *testing.T) {
		testLogLevel("mockDSN", []string{"development"}, log.DebugLevel, t)
	})

	// test case 2: when v[0] is "production", log level should be InfoLevel
	t.Run("Log level production", func(t *testing.T) {
		testLogLevel("https://username@sentry.io/project-id", []string{"production"}, log.InfoLevel, t)
	})

	// test case 3: when v[0] is not "production", log level should be DebugLevel
	t.Run("Log level non-production", func(t *testing.T) {
		testLogLevel("https://username@sentry.io/project-id", []string{"development"}, log.DebugLevel, t)
	})

	// test case 4: when v is empty, log level should default to DebugLevel
	t.Run("Log level default", func(t *testing.T) {
		testLogLevel("https://username@sentry.io/project-id", []string{}, log.DebugLevel, t)
	})
}

// Helper function to test log level setting
func testLogLevel(sentryDsn string, v []string, expectedLevel log.Level, t *testing.T) {
	// destroy any existing global hook
	middleware.DestroySentry()

	// set the log level using the InitSentry function
	_, err := middleware.InitSentry(sentryDsn, v...)
	if err == nil && sentryDsn == "mockDSN" {
		// if DSN is incorrect, expect an error
		t.Fatalf("Expected error, but got nil")
		return
	}
	if err != nil && sentryDsn != "mockDSN" {
		// if DSN is correct, but there was an error
		t.Fatalf("Error initializing sentry: %v", err)
		return
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
