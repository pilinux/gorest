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
		name          string
		originAllowed string
		origin        string
		expectedCode  int
	}{
		{
			name:          "Allowed Origin (*), should pass",
			originAllowed: "*",
			origin:        "http://example.com",
			expectedCode:  http.StatusOK,
		},
		{
			name:          "Allowed Origin (http), should pass",
			originAllowed: "http://example.com",
			origin:        "http://eXample.com",
			expectedCode:  http.StatusOK,
		},
		{
			name:          "Allowed Origin (https), should pass",
			originAllowed: "https://example.com",
			origin:        "https://example.com",
			expectedCode:  http.StatusOK,
		},
		{
			name:          "Allowed Origin (empty), should fail",
			originAllowed: "",
			origin:        "http://example.com",
			expectedCode:  http.StatusForbidden,
		},
		{
			name:          "Forbidden Origin, should fail",
			originAllowed: "http://example.com",
			origin:        "http://other-domain.com",
			expectedCode:  http.StatusForbidden,
		},
		{
			name:          "No Origin, should fail",
			originAllowed: "http://example.com",
			origin:        "",
			expectedCode:  http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cp := []middleware.CORSPolicy{}
			cp = append(cp, middleware.CORSPolicy{"Access-Control-Allow-Origin", test.originAllowed})

			// set up a gin router and handler
			gin.SetMode(gin.TestMode)
			router := gin.New()
			err := router.SetTrustedProxies(nil)
			if err != nil {
				t.Errorf("failed to set trusted proxies to nil")
			}
			router.TrustedPlatform = "X-Real-Ip"
			router.Use(middleware.CORS(cp))
			router.Use(middleware.CheckOrigin(middleware.GetCORS().AllowedOrigins))
			router.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, "success")
			})

			// create a new HTTP request
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Errorf("failed to create an HTTP request")
				return
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Origin", test.origin)

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
