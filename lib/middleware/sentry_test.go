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
	router.Use(middleware.SentryCapture(sentryDSN, "production", "v0.0.1", "yes", "1.0"))

	// check sentry in a separate goroutine
	var wg sync.WaitGroup
	var GoroutineLogger *log.Logger
	sentryHook, err := middleware.InitSentry(
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
		sentryHook.SetFlushTimeout(5 * time.Second)
		defer sentryHook.Flush()
		GoroutineLogger = log.New()
		GoroutineLogger.AddHook(sentryHook)
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
