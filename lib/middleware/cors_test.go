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
				{"Access-Control-Allow-Origin", "http://example2.com"},
				{"Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS"},
			},
			"GET",
			"Access-Control-Allow-Origin",
			"",
			http.StatusOK,
		},
		{
			[]middleware.CORSPolicy{
				{"Access-Control-Allow-Origin", "http://example2.com"},
				{"Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS"},
			},
			"OPTIONS",
			"Access-Control-Allow-Origin",
			"",
			http.StatusNoContent,
		},
		{
			[]middleware.CORSPolicy{
				{"Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS"},
				{"Access-Control-Allow-Origin", "*"},
			},
			"OPTIONS",
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, PATCH, DELETE, OPTIONS",
			http.StatusNoContent,
		},
		{
			[]middleware.CORSPolicy{
				{"Access-Control-Allow-Headers", "X-Custom-Header, Content-Type"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"OPTIONS",
			"Access-Control-Allow-Headers",
			"X-Custom-Header, Content-Type",
			http.StatusNoContent,
		},
		{
			[]middleware.CORSPolicy{
				{"Access-Control-Expose-Headers", "X-Expose-Header"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"OPTIONS",
			"Access-Control-Expose-Headers",
			"X-Expose-Header",
			http.StatusNoContent,
		},
		{
			[]middleware.CORSPolicy{
				{"Access-Control-Max-Age", "1234"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"OPTIONS",
			"Access-Control-Max-Age",
			"1234",
			http.StatusNoContent,
		},
		{
			[]middleware.CORSPolicy{
				{"Access-Control-Allow-Credentials", "true"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"OPTIONS",
			"Access-Control-Allow-Credentials",
			"true",
			http.StatusNoContent,
		},
		{
			[]middleware.CORSPolicy{
				{"Access-Control-Allow-Credentials", "false"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"OPTIONS",
			"Access-Control-Allow-Credentials",
			"",
			http.StatusNoContent,
		},
		{
			[]middleware.CORSPolicy{
				{"X-Content-Type-Options", "nosniff"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"GET",
			"X-Content-Type-Options",
			"nosniff",
			http.StatusOK,
		},
		{
			[]middleware.CORSPolicy{
				{"X-Frame-Options", "DENY"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"GET",
			"X-Frame-Options",
			"DENY",
			http.StatusOK,
		},
		{
			[]middleware.CORSPolicy{
				{"Referrer-Policy", "no-referrer"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"GET",
			"Referrer-Policy",
			"no-referrer",
			http.StatusOK,
		},
		{
			[]middleware.CORSPolicy{
				{"Content-Security-Policy", "default-src 'self'"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"GET",
			"Content-Security-Policy",
			"default-src 'self'",
			http.StatusOK,
		},
		{
			[]middleware.CORSPolicy{
				{"Timing-Allow-Origin", "http://example.com"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"OPTIONS",
			"Timing-Allow-Origin",
			"http://example.com",
			http.StatusNoContent,
		},
		{
			[]middleware.CORSPolicy{
				{"Timing-Allow-Origin", "http://example2.com"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"OPTIONS",
			"Timing-Allow-Origin",
			"",
			http.StatusNoContent,
		},
		{
			[]middleware.CORSPolicy{
				{"Timing-Allow-Origin", "*"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"OPTIONS",
			"Timing-Allow-Origin",
			"http://example.com",
			http.StatusNoContent,
		},
		{
			[]middleware.CORSPolicy{
				{"Strict-Transport-Security", "max-age=31536000; includeSubDomains"},
				{"Access-Control-Allow-Origin", "http://example.com"},
			},
			"GET",
			"Strict-Transport-Security",
			"max-age=31536000; includeSubDomains",
			http.StatusOK,
		},
		{
			[]middleware.CORSPolicy{
				{"Access-Control-Allow-Origin", "*"},
				{"Access-Control-Allow-Credentials", "true"},
			},
			"GET",
			"error",
			"\"CORS misconfiguration: CORS_CREDENTIALS=true with CORS_ORIGIN=* is forbidden by the CORS spec\"",
			http.StatusInternalServerError,
		},
		{
			[]middleware.CORSPolicy{
				{"Access-Control-Allow-Origin", ""},
				{"Access-Control-Allow-Credentials", "true"},
			},
			"GET",
			"error",
			"\"CORS misconfiguration: CORS_CREDENTIALS=true with CORS_ORIGIN=* is forbidden by the CORS spec\"",
			http.StatusInternalServerError,
		},
		{
			[]middleware.CORSPolicy{},
			"OPTIONS",
			"Access-Control-Allow-Origin",
			"*",
			http.StatusNoContent,
		},
		{
			[]middleware.CORSPolicy{},
			"OPTIONS",
			"Access-Control-Allow-Methods",
			"OPTIONS",
			http.StatusNoContent,
		},
		{
			[]middleware.CORSPolicy{},
			"OPTIONS",
			"Access-Control-Allow-Headers",
			"Content-Type",
			http.StatusNoContent,
		},
	}

	// test each case
	for _, c := range cases {
		if c.headerKey == "error" {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			handler := middleware.CORS(c.cp)
			router.Use(handler)
			router.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req, err := http.NewRequest(c.method, "/", nil)
			if err != nil {
				t.Errorf("failed to create an HTTP request")
				continue
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Origin", "http://example.com")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != c.statusCode {
				t.Errorf("expected status code %d, got %d", c.statusCode, w.Code)
			}
			if w.Body.String() != c.headerVal {
				t.Errorf("expected error message '%s', got '%s'", c.headerVal, w.Body.String())
			}
			continue
		}

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
		// add Origin header for CORS testing
		req.Header.Add("Origin", "http://example.com")

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
