package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CheckOrigin - check whether the request generated from the allowed origin
func CheckOrigin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get Access-Control-Allow-Origin from CORS header
		origin := c.Writer.Header().Get("Access-Control-Allow-Origin")
		origin = strings.TrimSpace(origin)
		origin = strings.ToLower(origin)

		// if Access-Control-Allow-Origin is *, null or empty, continue
		if origin == "*" || origin == "null" || origin == "" {
			c.Next()
			return
		}

		// Access-Control-Allow-Origin is not *, null or empty
		if strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://") {
			origin = strings.Split(origin, "//")[1]
		}

		// get host from the request
		host := c.Request.Host
		host = strings.TrimSpace(host)
		host = strings.ToLower(host)

		// if origin and host are different, abort
		if origin != host {
			c.AbortWithStatusJSON(http.StatusForbidden, "origin not allowed")
			return
		}

		c.Next()
	}
}
