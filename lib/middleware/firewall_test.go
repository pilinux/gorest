package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/gorest/lib/middleware"
)

type testCase struct {
	testNo    string
	listType  string
	ipList    string
	remoteIP  string
	statusExp int
}

func TestFirewall(t *testing.T) {
	testCases := []testCase{
		// list of IPs
		{
			"1.1",
			"whitelist",
			"192.168.0.1, 192.168.0.2, 192.168.0.3, 192.168.0.4",
			"192.168.0.1",
			http.StatusOK,
		},
		{
			"1.2",
			"whitelist",
			"192.168.0.1, 192.168.0.2, 192.168.0.3, 192.168.0.4",
			"192.168.0.5",
			http.StatusUnauthorized,
		},
		{
			"1.3",
			"blacklist",
			"192.168.0.1, 192.168.0.2, 192.168.0.3, 192.168.0.4",
			"192.168.0.1",
			http.StatusUnauthorized,
		},
		{
			"1.4",
			"blacklist",
			"192.168.0.1, 192.168.0.2, 192.168.0.3, 192.168.0.4",
			"192.168.0.5",
			http.StatusOK,
		},

		// missing client IP
		{
			"2.1",
			"whitelist",
			"192.168.0.1, 192.168.0.2, 192.168.0.3, 192.168.0.4",
			"",
			http.StatusUnauthorized,
		},
		{
			"2.2",
			"blacklist",
			"192.168.0.1, 192.168.0.2, 192.168.0.3, 192.168.0.4",
			"",
			http.StatusUnauthorized,
		},

		// wildcard
		{
			"3.1",
			"whitelist",
			"*",
			"192.168.1.1",
			http.StatusOK,
		},
		{
			"3.2",
			"blacklist",
			"*",
			"192.168.1.1",
			http.StatusUnauthorized,
		},

		// CIDR
		{
			"4.1",
			"whitelist",
			"192.168.0.0/16",
			"192.168.1.1",
			http.StatusOK,
		},
		{
			"4.2",
			"whitelist",
			"192.168.0.0/16",
			"192.169.1.1",
			http.StatusUnauthorized,
		},
		{
			"4.3",
			"whitelist",
			"192.168.10.0/24",
			"192.168.10.255",
			http.StatusOK,
		},
		{
			"4.4",
			"whitelist",
			"192.168.10.0/24",
			"192.168.11.1",
			http.StatusUnauthorized,
		},
		{
			"4.5",
			"whitelist",
			"192.168.10.10/32",
			"192.168.10.10",
			http.StatusOK,
		},
		{
			"4.6",
			"whitelist",
			"192.168.10.10/32",
			"192.168.10.11",
			http.StatusUnauthorized,
		},
		{
			"4.7",
			"whitelist",
			"172.16.0.0/12",
			"172.22.0.1",
			http.StatusOK,
		},
		{
			"4.8",
			"whitelist",
			"172.16.0.0/12",
			"172.32.0.1",
			http.StatusUnauthorized,
		},
		{
			"4.9",
			"blacklist",
			"192.168.0.0/16",
			"192.168.1.1",
			http.StatusUnauthorized,
		},
		{
			"4.10",
			"blacklist",
			"192.168.0.0/16",
			"192.169.1.1",
			http.StatusOK,
		},

		// CIDR and IPs
		{
			"5.1",
			"whitelist",
			"192.168.0.0/16, 192.169.0.1, 192.169.0.2, 192.169.0.3, 192.169.0.4, 192.170.10.0/24",
			"192.168.0.1",
			http.StatusOK,
		},
		{
			"5.2",
			"whitelist",
			"192.168.0.0/16, 192.169.0.1, 192.169.0.2, 192.169.0.3, 192.169.0.4, 192.170.10.0/24",
			"192.169.0.1",
			http.StatusOK,
		},
		{
			"5.3",
			"whitelist",
			"192.168.0.0/16, 192.169.0.1, 192.169.0.2, 192.169.0.3, 192.169.0.4, 192.170.10.0/24",
			"192.170.10.240",
			http.StatusOK,
		},
		{
			"5.4",
			"blacklist",
			"192.168.0.0/16, 192.169.0.1, 192.169.0.2, 192.169.0.3, 192.169.0.4, 192.170.10.0/24",
			"192.170.0.1",
			http.StatusOK,
		},
		{
			"5.5",
			"blacklist",
			"192.168.0.0/16, 192.169.0.1, 192.169.0.2, 192.169.0.3, 192.169.0.4, 192.170.10.0/24",
			"192.168.10.1",
			http.StatusUnauthorized,
		},

		// *, CIDR and IPs
		{
			"6.1",
			"whitelist",
			"*, 192.168.0.0/16, 192.169.0.1, 192.169.0.2, 192.169.0.3, 192.169.0.4, 192.170.10.0/24",
			"192.171.0.1",
			http.StatusOK,
		},
		{
			"6.2",
			"blacklist",
			"*, 192.168.0.0/16, 192.169.0.1, 192.169.0.2, 192.169.0.3, 192.169.0.4, 192.170.10.0/24",
			"192.171.0.1",
			http.StatusUnauthorized,
		},

		// IPv6
		{
			"7.1",
			"whitelist",
			"2001:db8::",
			"2001:db8::",
			http.StatusOK,
		},
		{
			"7.2",
			"whitelist",
			"2001:db8::/32",
			"2001:db8::1",
			http.StatusOK,
		},
		{
			"7.3",
			"whitelist",
			"2001:db8::/128",
			"2001:db8::1",
			http.StatusUnauthorized,
		},
		{
			"7.4",
			"blacklist",
			"2001:db8::/32",
			"2001:db8::1",
			http.StatusUnauthorized,
		},
		{
			"7.5",
			"blacklist",
			"2001:db8::/128",
			"2001:db8::ffff",
			http.StatusOK,
		},

		// IPv4 and IPv6
		{
			"8.1",
			"whitelist",
			"2001:db8::/32,, 192.168.10.10/32,",
			"2001:db8::1",
			http.StatusOK,
		},
		{
			"8.2",
			"blacklist",
			"2001:db8::/32,, 192.168.10.10/32,",
			"192.168.10.10",
			http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run("TestCase"+tc.testNo, func(t *testing.T) {
			// reset firewall state between test cases
			middleware.ResetFirewallState()

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
				t.Errorf("testCase no %s, expected status code %d, got %d", tc.testNo, tc.statusExp, w.Code)
			}
		})
	}
}
