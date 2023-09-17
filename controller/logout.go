package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"
)

// Logout -
//
// - if 'AUTH_COOKIE_ACTIVATE=yes', delete tokens from client browser.
//
// - if Redis is enabled, save invalid tokens in Redis up until the expiry time.
//
// dependency: JWT
func Logout(c *gin.Context) {
	// verify that JWT service is enabled in .env
	if !config.IsJWT() {
		renderer.Render(c, gin.H{"message": "JWT service not enabled"}, http.StatusNotImplemented)
		return
	}

	// app security settings
	configSecurity := config.GetConfig().Security

	jtiAccess := c.GetString("jtiAccess")
	expAccess := c.GetInt64("expAccess")

	jtiRefresh := c.GetString("jtiRefresh")
	expRefresh := c.GetInt64("expRefresh")

	// if cookie is enabled, delete the cookie from client browser
	if configSecurity.AuthCookieActivate {
		c.SetSameSite(configSecurity.AuthCookieSameSite)
		c.SetCookie(
			"accessJWT",
			"",
			-1,
			configSecurity.AuthCookiePath,
			configSecurity.AuthCookieDomain,
			configSecurity.AuthCookieSecure,
			configSecurity.AuthCookieHTTPOnly,
		)
		c.SetCookie(
			"refreshJWT",
			"",
			-1,
			configSecurity.AuthCookiePath,
			configSecurity.AuthCookieDomain,
			configSecurity.AuthCookieSecure,
			configSecurity.AuthCookieHTTPOnly,
		)
	}

	resp, statusCode := handler.Logout(jtiAccess, jtiRefresh, expAccess, expRefresh)

	renderer.Render(c, resp, statusCode)
}
