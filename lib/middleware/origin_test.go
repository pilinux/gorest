package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/gorest/lib/middleware"
)

// TestCheckOrigin - test the CheckOrigin middleware
func TestCheckOrigin(t *testing.T) {
	tests := []struct {
		name         string
		origin       string
		host         string
		expectedCode int
	}{
		{
			name:         "Allowed Origin (*), should pass",
			origin:       "*",
			host:         "example.com",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Allowed Origin (null), should pass",
			origin:       "null",
			host:         "example.com",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Allowed Origin (empty), should pass",
			origin:       "",
			host:         "example.com",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Allowed Origin (http), should pass",
			origin:       "http://Example.com",
			host:         "eXample.com",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Allowed Origin (https), should pass",
			origin:       "https://example.com",
			host:         "example.com",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Forbidden Origin, should fail",
			origin:       "http://example.com",
			host:         "other-domain.com",
			expectedCode: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cp := []middleware.CORSPolicy{}
			cp = append(cp, middleware.CORSPolicy{"Access-Control-Allow-Origin", test.origin})

			// set up a gin router and handler
			gin.SetMode(gin.TestMode)
			router := gin.New()
			err := router.SetTrustedProxies(nil)
			if err != nil {
				t.Errorf("failed to set trusted proxies to nil")
			}
			router.TrustedPlatform = "X-Real-Ip"
			router.Use(middleware.CORS(cp))
			router.Use(middleware.CheckOrigin())
			router.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, "success")
			})

			// create a new HTTP request
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Errorf("failed to create an HTTP request")
				return
			}
			req.Host = test.host

			// create a new HTTP response recorder
			w := httptest.NewRecorder()

			// pass the request to the router and check the response
			router.ServeHTTP(w, req)
			if w.Code != test.expectedCode {
				t.Errorf("expected status code %d, got %d", test.expectedCode, w.Code)
			}
		})
	}
}
