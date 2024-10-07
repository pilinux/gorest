package controller

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"
)

// Login - issue new JWTs after user:pass verification
//
// [POST]: /login
//
// dependency: relational database, JWT
//
// Accepted JSON payload:
//
// `{"email":"...", "password":"..."}`
func Login(c *gin.Context) {
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

	var payload model.AuthPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	resp, statusCode := handler.Login(payload)

	// auth verification failed
	if statusCode != http.StatusOK {
		renderer.Render(c, resp, statusCode)
		return
	}

	// auth verification OK
	// set cookie if the feature is enabled in app settings
	configSecurity := config.GetConfig().Security
	if configSecurity.AuthCookieActivate {
		tokens, ok := resp.Message.(middleware.JWTPayload)
		if ok {
			c.SetSameSite(configSecurity.AuthCookieSameSite)
			c.SetCookie(
				"accessJWT",
				tokens.AccessJWT,
				middleware.JWTParams.AccessKeyTTL*60,
				configSecurity.AuthCookiePath,
				configSecurity.AuthCookieDomain,
				configSecurity.AuthCookieSecure,
				configSecurity.AuthCookieHTTPOnly,
			)
			c.SetCookie(
				"refreshJWT",
				tokens.RefreshJWT,
				middleware.JWTParams.RefreshKeyTTL*60,
				configSecurity.AuthCookiePath,
				configSecurity.AuthCookieDomain,
				configSecurity.AuthCookieSecure,
				configSecurity.AuthCookieHTTPOnly,
			)

			if !configSecurity.ServeJwtAsResBody {
				resp.Message = "login successful"
				if configSecurity.Must2FA == config.Activated {
					tokens.AccessJWT = ""
					tokens.RefreshJWT = ""
					resp.Message = tokens
				}
			}
		}

		if !ok {
			log.Error("error code: 1011.1")
			resp.Message = "failed to prepare auth cookie"
			statusCode = http.StatusInternalServerError
		}
	}

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}

// Refresh - issue new JWTs after validation
//
// dependency: JWT
func Refresh(c *gin.Context) {
	// verify that JWT service is enabled in .env
	if !config.IsJWT() {
		renderer.Render(c, gin.H{"message": "JWT service not enabled"}, http.StatusNotImplemented)
		return
	}

	// get claims
	claims := service.GetClaims(c)

	resp, statusCode := handler.Refresh(claims)

	configSecurity := config.GetConfig().Security

	// JWT verification failed
	if statusCode != http.StatusOK {
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

		renderer.Render(c, resp, statusCode)
		return
	}

	// JWT verification OK
	// set cookie if the feature is enabled in app settings
	if configSecurity.AuthCookieActivate {
		tokens, ok := resp.Message.(middleware.JWTPayload)
		if ok {
			c.SetSameSite(configSecurity.AuthCookieSameSite)
			c.SetCookie(
				"accessJWT",
				tokens.AccessJWT,
				middleware.JWTParams.AccessKeyTTL*60,
				configSecurity.AuthCookiePath,
				configSecurity.AuthCookieDomain,
				configSecurity.AuthCookieSecure,
				configSecurity.AuthCookieHTTPOnly,
			)
			c.SetCookie(
				"refreshJWT",
				tokens.RefreshJWT,
				middleware.JWTParams.RefreshKeyTTL*60,
				configSecurity.AuthCookiePath,
				configSecurity.AuthCookieDomain,
				configSecurity.AuthCookieSecure,
				configSecurity.AuthCookieHTTPOnly,
			)

			if !configSecurity.ServeJwtAsResBody {
				resp.Message = "new tokens issued"
				if configSecurity.Must2FA == config.Activated {
					tokens.AccessJWT = ""
					tokens.RefreshJWT = ""
					resp.Message = tokens
				}
			}
		}

		if !ok {
			log.Error("error code: 1012.1")
			resp.Message = "failed to prepare auth cookie"
			statusCode = http.StatusInternalServerError
		}
	}

	if reflect.TypeOf(resp.Message).Kind() == reflect.String {
		renderer.Render(c, resp, statusCode)
		return
	}

	renderer.Render(c, resp.Message, statusCode)
}
