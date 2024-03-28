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
		originAllowed := c.Writer.Header().Get("Access-Control-Allow-Origin")
		originAllowed = strings.TrimSpace(originAllowed)
		originAllowed = strings.ToLower(originAllowed)

		// if Access-Control-Allow-Origin is *, null or empty, continue
		if originAllowed == "*" || originAllowed == "null" || originAllowed == "" {
			c.Next()
			return
		}

		// get origin from request header
		origin := c.GetHeader("Origin")
		origin = strings.TrimSpace(origin)
		origin = strings.ToLower(origin)

		// if origin and host are different, abort
		if origin != originAllowed {
			c.AbortWithStatusJSON(http.StatusForbidden, "origin not allowed")
			return
		}

		c.Next()
	}
}
