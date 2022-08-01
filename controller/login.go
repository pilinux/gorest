package controller

import (
	"net/http"

	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/lib/renderer"
	"github.com/pilinux/gorest/service"

	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// LoginPayload ...
type LoginPayload struct {
	Email    string `json:"Email"`
	Password string `json:"Password"`
}

// Login ...
func Login(c *gin.Context) {
	var payload LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	if !lib.ValidateEmail(payload.Email) {
		renderer.Render(c, gin.H{"msg": "wrong email address"}, http.StatusBadRequest)
		return
	}

	v, err := service.GetUserByEmail(payload.Email)
	if err != nil {
		renderer.Render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	verifyPass, err := argon2id.ComparePasswordAndHash(payload.Password, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1011")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if !verifyPass {
		renderer.Render(c, gin.H{"msg": "wrong credentials"}, http.StatusUnauthorized)
		return
	}

	// issue new tokens
	claims := middleware.MyCustomClaims{
		AuthID:  v.AuthID,
		Email:   v.Email,
		Role:    "",
		Scope:   "",
		TwoFA:   "",
		SiteLan: "en",
		Custom1: "",
		Custom2: "",
	}
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1012")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1013")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT
	renderer.Render(c, jwtPayload, http.StatusOK)
}

// Refresh ...
func Refresh(c *gin.Context) {
	// get claims
	claims := middleware.MyCustomClaims{
		AuthID:  c.GetUint64("authID"),
		Email:   c.GetString("email"),
		Role:    c.GetString("role"),
		Scope:   c.GetString("scope"),
		TwoFA:   c.GetString("tfa"),
		SiteLan: c.GetString("siteLan"),
		Custom1: c.GetString("custom1"),
		Custom2: c.GetString("custom2"),
	}
	// check validity
	if claims.AuthID == 0 {
		renderer.Render(c, gin.H{"msg": "access denied"}, http.StatusUnauthorized)
		return
	}
	if claims.Email == "" {
		renderer.Render(c, gin.H{"msg": "access denied"}, http.StatusUnauthorized)
		return
	}

	// issue new tokens
	accessJWT, _, err := middleware.GetJWT(claims, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1021")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	refreshJWT, _, err := middleware.GetJWT(claims, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1022")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT
	renderer.Render(c, jwtPayload, http.StatusOK)
}
