package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/gorest/lib/middleware"
)

func TestTwoFA(t *testing.T) {
	keywordOn := "on"
	keywordOff := "off"
	keywordVerified := "verified"

	// set up test cases
	testCases := []struct {
		name           string
		tfa            string
		expectedStatus int
	}{
		{
			name:           "2-FA is off or not configured",
			tfa:            "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "2-FA off for this account",
			tfa:            "off",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "2-FA on, status from JWT not verified",
			tfa:            "on",
			expectedStatus: http.StatusUnauthorized,
		},

		{
			name:           "2-FA on, status from JWT verified",
			tfa:            "verified",
			expectedStatus: http.StatusOK,
		},

		{
			name:           "2-FA on, status from JWT not defined",
			tfa:            "not.defined",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	// run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// create a new gin router and apply the TwoFA middleware to it
			gin.SetMode(gin.TestMode)
			router := gin.New()
			err := router.SetTrustedProxies(nil)
			if err != nil {
				t.Errorf("failed to set trusted proxies to nil")
			}
			router.TrustedPlatform = "X-Real-Ip"

			router.Use(func(c *gin.Context) {
				c.Set("tfa", tc.tfa)
				c.Next()
			})
			router.Use(middleware.TwoFA(keywordOn, keywordOff, keywordVerified))

			// create a handler function that always returns 200
			router.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// create an HTTP request
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Errorf("failed to create an HTTP request")
				return
			}

			// set the key and value in the request header
			req.Header.Set("tfa", tc.tfa)

			// create a response recorder
			w := httptest.NewRecorder()

			// send the request to the router
			router.ServeHTTP(w, req)

			// check that the response status is as expected
			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}
