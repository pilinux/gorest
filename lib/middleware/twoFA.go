package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TwoFA validates 2-FA status from JWT before
// forwarding the request to the controller
func TwoFA(keywordOn, keywordOff, keywordVerified string) gin.HandlerFunc {
	return func(c *gin.Context) {
		statusFromJWT := c.GetString("tfa")
		var requestAllowed bool

		// user account is protected by 2-FA, requires validation
		if statusFromJWT == keywordOn {
			requestAllowed = false
		}

		// 2-FA verified
		if statusFromJWT == keywordVerified {
			requestAllowed = true
		}

		// user account is not protected by 2-FA
		if statusFromJWT == keywordOff || statusFromJWT == "" {
			requestAllowed = true
		}

		if !requestAllowed {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}
