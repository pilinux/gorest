// Package controller contains all the controllers
// of the application
package controller

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"
)

// CreateUserAuth - register a new user account
//
// [POST]: /register
//
// dependency: relational database
//
// Accepted JSON payload:
//
// `{"email":"...", "password":"..."}`
func CreateUserAuth(c *gin.Context) {
	// delete existing auth cookie if present
	_, errAccessJWT := c.Cookie("accessJWT")
	_, errRefreshJWT := c.Cookie("refreshJWT")
	if errAccessJWT == nil || errRefreshJWT == nil {
		configSecurity := config.GetConfig().Security
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

	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	auth := model.Auth{}

	// bind JSON
	if err := c.ShouldBindJSON(&auth); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.CreateUserAuth(auth)

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// UpdateEmail - update existing user email
//
// dependency: relational database, JWT
//
// Accepted JSON payload:
//
// `{"emailNew":"...", "password":"..."}`
func UpdateEmail(c *gin.Context) {
	// verify that RDBMS is enabled in .env
	if !config.IsRDBMS() {
		renderer.Render(c, gin.H{"message": "relational database not enabled"}, http.StatusNotImplemented)
		return
	}

	// verify that JWT service is enabled in .env
	if !config.IsJWT() {
		renderer.Render(c, gin.H{"message": "JWT service not enabled"}, http.StatusNotImplemented)
		return
	}

	// get claims
	claims := service.GetClaims(c)

	req := model.TempEmail{}

	// bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.UpdateEmail(claims, req)

	renderer.Render(c, resp, statusCode)
}
