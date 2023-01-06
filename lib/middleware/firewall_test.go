package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/gorest/lib/middleware"
)

type testCase struct {
	listType  string
	ipList    string
	remoteIP  string
	statusExp int
}

func TestFirewall(t *testing.T) {
	testCases := []testCase{
		{"whitelist", "192.168.0.1, 192.168.0.2, 192.168.0.3, 192.168.0.4", "192.168.0.1", http.StatusOK},
		{"whitelist", "192.168.0.1, 192.168.0.2, 192.168.0.3, 192.168.0.4", "192.168.0.5", http.StatusUnauthorized},
		{"blacklist", "192.168.0.1, 192.168.0.2, 192.168.0.3, 192.168.0.4", "192.168.0.1", http.StatusUnauthorized},
		{"blacklist", "192.168.0.1, 192.168.0.2, 192.168.0.3, 192.168.0.4", "192.168.0.5", http.StatusOK},

		{"whitelist", "*", "192.168.1.1", http.StatusOK},
		{"blacklist", "*", "192.168.1.1", http.StatusUnauthorized},
	}

	for _, tc := range testCases {
		// set up a gin router and handler
		gin.SetMode(gin.TestMode)
		router := gin.New()
		err := router.SetTrustedProxies(nil)
		if err != nil {
			t.Errorf("failed to set trusted proxies to nil")
		}
		router.TrustedPlatform = "X-Real-Ip"
		router.Use(middleware.Firewall(tc.listType, tc.ipList))
		router.GET("/", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// create a request and response recorder
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Errorf("failed to create an HTTP request")
			return
		}
		req.Header.Set("X-Real-Ip", tc.remoteIP)

		// pass the request to the router and check the response
		router.ServeHTTP(w, req)
		if w.Code != tc.statusExp {
			t.Errorf("expected status code %d, got %d", tc.statusExp, w.Code)
		}
	}
}
