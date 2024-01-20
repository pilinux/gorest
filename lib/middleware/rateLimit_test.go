package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
)

// TestRateLimit - test the RateLimit middleware
func TestRateLimit(t *testing.T) {
	tests := []struct {
		name           string
		rateLimit      string
		trustedProxy   string
		expectedStatus int
	}{
		{
			name:           "Nil limiter instance, should pass",
			rateLimit:      "",
			trustedProxy:   "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-nil limiter instance, pass 2 requests in a second, then fail",
			rateLimit:      "2-S",
			trustedProxy:   "",
			expectedStatus: http.StatusTooManyRequests,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			limiterInstance, err := lib.InitRateLimiter(test.rateLimit, test.trustedProxy)
			if err != nil {
				t.Errorf("expected error: %v, got: %v", nil, err)
			}

			// set up a gin router and handler
			gin.SetMode(gin.TestMode)
			router := gin.New()
			err = router.SetTrustedProxies(nil)
			if err != nil {
				t.Errorf("failed to set trusted proxies to nil")
			}
			router.TrustedPlatform = "X-Real-Ip"

			// define the handler function
			handler := middleware.RateLimit(limiterInstance)

			// add the handler to the router
			router.Use(handler)

			router.GET("/", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "ok",
				})
			})

			// create 3 test requests
			for i := 0; i < 3; i++ {
				req, err := http.NewRequest("GET", "/", nil)
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}

				// create a test recorder
				w := httptest.NewRecorder()

				// serve the request
				router.ServeHTTP(w, req)

				// check the response status code for the first 2 requests
				if i < 2 && w.Code != http.StatusOK {
					t.Errorf("expected status code: %v, got: %v", http.StatusOK, w.Code)
				}

				// check the response status code for the 3rd request
				if i == 3 && w.Code != test.expectedStatus {
					t.Errorf("expected status code: %v, got: %v", test.expectedStatus, w.Code)
				}
			}
		})
	}
}
