package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
	router.Use(middleware.SentryCapture(sentryDSN))

	// define test route
	router.GET("/", func(c *gin.Context) {
		// send log to sentry for testing
		log.Info("testing sentry integration")
		c.Status(http.StatusOK)
	})

	// perform request and get response
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("failed to create an HTTP request")
		return
	}
	req.Header.Set("User-Agent", "Test-User-Agent")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	// check response
	if res.Code != http.StatusOK {
		t.Errorf("expected response code %v, got '%v'", http.StatusOK, res.Code)
	}
}
