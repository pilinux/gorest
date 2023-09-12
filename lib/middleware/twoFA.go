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

		statusChecked := false
		requestAllowed := false

		// JWT is not activated for this app
		// or, JWT is activated for this app, but 2-FA is off or not configured
		if !statusChecked && statusFromJWT == "" {
			statusChecked = true
			requestAllowed = true
		}

		// JWT is activated for this app
		// user account is not protected by 2-FA
		// 2-FA status from JWT = off
		if !statusChecked && statusFromJWT == keywordOff {
			statusChecked = true
			requestAllowed = true
		}

		// JWT is activated for this app
		// user account is protected by 2-FA, requires validation
		// 2-FA on, 2-FA status from JWT = not verified
		if !statusChecked && statusFromJWT == keywordOn {
			statusChecked = true
			requestAllowed = false
		}

		// JWT is activated for this app
		// user account is protected by 2-FA, requires validation
		// 2-FA on, 2-FA status from JWT = verified
		if !statusChecked && statusFromJWT == keywordVerified {
			statusChecked = true
			requestAllowed = true
		}

		if !statusChecked {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "2-fa: status unknown")
			return
		}

		if !requestAllowed {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "2-fa: required valid OTP")
			return
		}

		c.Next()
	}
}
