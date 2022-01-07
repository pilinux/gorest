package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Firewall ...
func Firewall(listType string, ipList string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get IP address by reading off the forwarded-for
		// header (for proxies) and falls back to use the remote address
		ipAddress := c.Request.RemoteAddr
		forwardedAddress := c.Request.Header.Get("X-Forwarded-For")
		if forwardedAddress != "" {
			//X-Forwarded-For
			ipAddress = forwardedAddress //Single IP

			//Array of IPs
			ip := strings.Split(forwardedAddress, ", ")
			if len(ip) > 1 {
				ipAddress = ip[0] //First IP
			}
		}

		if !strings.Contains(ipList, "*") {
			if listType == "whitelist" {
				if !strings.Contains(ipList, ipAddress) {
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
			}

			if listType == "blacklist" {
				if strings.Contains(ipList, ipAddress) {
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
			}
		}
	}
}
