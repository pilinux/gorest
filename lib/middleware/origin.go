package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CheckOrigin - check whether the request generated from the allowed origin
func CheckOrigin(originAllowed []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get origin from request header
		origin := c.GetHeader("Origin")
		origin = strings.TrimSpace(origin)
		origin = strings.ToLower(origin)

		ok := false
		for _, allowed := range originAllowed {
			if allowed == "*" {
				ok = true
				break
			}

			if origin == allowed {
				ok = true
				break
			}
		}

		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, "origin not allowed")
			return
		}

		c.Next()
	}
}
