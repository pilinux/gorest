package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/gorest/lib/middleware"
)

func TestCORS(t *testing.T) {
	// define the test cases
	cases := []struct {
		cp         []middleware.CORSPolicy
		method     string
		headerKey  string
		headerVal  string
		statusCode int
	}{
		{
			[]middleware.CORSPolicy{
				{"Access-Control-Allow-Origin", "*"},
				{"Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS"},
			},
			"GET",
			"Access-Control-Allow-Origin",
			"*",
			http.StatusOK,
		},
		{
			[]middleware.CORSPolicy{
				{"Access-Control-Allow-Origin", "*"},
				{"Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS"},
			},
			"OPTIONS",
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, PATCH, DELETE, OPTIONS",
			http.StatusNoContent,
		},
	}

	// test each case
	for _, c := range cases {
		// set up a gin router and handler
		gin.SetMode(gin.TestMode)
		router := gin.New()
		err := router.SetTrustedProxies(nil)
		if err != nil {
			t.Errorf("failed to set trusted proxies to nil")
		}
		router.TrustedPlatform = "X-Real-Ip"

		// define the handler function
		handler := middleware.CORS(c.cp)

		// add the handler to the router
		router.Use(handler)

		router.GET("/", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// create a new HTTP request
		req, err := http.NewRequest(c.method, "/", nil)
		if err != nil {
			t.Errorf("failed to create an HTTP request")
			return
		}
		req.Header.Add("Content-Type", "application/json")

		// create a new recorder to capture the response
		w := httptest.NewRecorder()

		// call the handler function and pass in the recorder and request
		router.ServeHTTP(w, req)

		// check the status code
		if w.Code != c.statusCode {
			t.Errorf("expected status code %d, got %d", c.statusCode, w.Code)
		}

		// check the header value
		if c.headerKey != "" && w.Header().Get(c.headerKey) != c.headerVal {
			t.Errorf("expected header '%s' to be '%s', got '%s'", c.headerKey, c.headerVal, w.Header().Get(c.headerKey))
		}
	}
}
