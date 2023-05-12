package middleware

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// JWTParameters - params to configure JWT
type JWTParameters struct {
	AccessKey     []byte
	AccessKeyTTL  int
	RefreshKey    []byte
	RefreshKeyTTL int

	Audience string
	Issuer   string
	AccNbf   int
	RefNbf   int
	Subject  string
}

// JWTParams - exported variables
var JWTParams JWTParameters

// MyCustomClaims ...
type MyCustomClaims struct {
	AuthID  uint64 `json:"authID,omitempty"`
	Email   string `json:"email,omitempty"`
	Role    string `json:"role,omitempty"`
	Scope   string `json:"scope,omitempty"`
	TwoFA   string `json:"twoFA,omitempty"`
	SiteLan string `json:"siteLan,omitempty"`
	Custom1 string `json:"custom1,omitempty"`
	Custom2 string `json:"custom2,omitempty"`
}

// JWTClaims ...
type JWTClaims struct {
	MyCustomClaims
	jwt.RegisteredClaims
}

// JWTPayload ...
type JWTPayload struct {
	AccessJWT   string `json:"accessJWT,omitempty"`
	RefreshJWT  string `json:"refreshJWT,omitempty"`
	TwoAuth     string `json:"twoFA,omitempty"`
	RecoveryKey string `json:"recoveryKey,omitempty"`
}

// JWT - validate access token
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token *jwt.Token
		var val string
		var vals []string

		// first try to read the cookie
		accessJWT, err := c.Cookie("accessJWT")
		// accessJWT is available in the cookie
		if err == nil {
			token, err = jwt.ParseWithClaims(accessJWT, &JWTClaims{}, validateAccessJWT)
			if err != nil {
				// error parsing JWT
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			goto VerifyClaims
		}

		// accessJWT is not available in the cookie
		// try to read the Authorization header
		val = c.Request.Header.Get("Authorization")
		if len(val) == 0 || !strings.Contains(val, "Bearer ") {
			// no vals or no bearer found
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		vals = strings.Split(val, " ")
		if len(vals) != 2 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err = jwt.ParseWithClaims(vals[1], &JWTClaims{}, validateAccessJWT)
		if err != nil {
			// error parsing JWT
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

	VerifyClaims:
		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			c.Set("authID", claims.AuthID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Set("scope", claims.Scope)
			c.Set("tfa", claims.TwoFA)
			c.Set("siteLan", claims.SiteLan)
			c.Set("custom1", claims.Custom1)
			c.Set("custom2", claims.Custom2)
		}

		c.Next()
	}
}

// RefreshJWT - validate refresh token
func RefreshJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var jwtPayload JWTPayload

		// first try to read the cookie
		refreshJWT, err := c.Cookie("refreshJWT")
		// refreshJWT is available in the cookie
		if err == nil {
			jwtPayload.RefreshJWT = refreshJWT

			goto VerifyClaims
		}

		// refreshJWT is not available in the cookie
		// try to read the request body
		if err := c.ShouldBindJSON(&jwtPayload); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

	VerifyClaims:
		token, err := jwt.ParseWithClaims(jwtPayload.RefreshJWT, &JWTClaims{}, validateRefreshJWT)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			c.Set("authID", claims.AuthID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Set("scope", claims.Scope)
			c.Set("tfa", claims.TwoFA)
			c.Set("siteLan", claims.SiteLan)
			c.Set("custom1", claims.Custom1)
			c.Set("custom2", claims.Custom2)
		}

		c.Next()
	}
}

// validateAccessJWT ...
func validateAccessJWT(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return JWTParams.AccessKey, nil
}

// validateRefreshJWT ...
func validateRefreshJWT(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return JWTParams.RefreshKey, nil
}

// GetJWT - issue new tokens
func GetJWT(customClaims MyCustomClaims, tokenType string) (string, string, error) {
	var (
		key []byte
		ttl int
		nbf int
	)

	if tokenType == "access" {
		key = JWTParams.AccessKey
		ttl = JWTParams.AccessKeyTTL
		nbf = JWTParams.AccNbf
	}
	if tokenType == "refresh" {
		key = JWTParams.RefreshKey
		ttl = JWTParams.RefreshKeyTTL
		nbf = JWTParams.RefNbf
	}
	// Create the Claims
	claims := JWTClaims{
		MyCustomClaims{
			AuthID:  customClaims.AuthID,
			Email:   customClaims.Email,
			Role:    customClaims.Role,
			Scope:   customClaims.Scope,
			TwoFA:   customClaims.TwoFA,
			SiteLan: customClaims.SiteLan,
			Custom1: customClaims.Custom1,
			Custom2: customClaims.Custom2,
		},
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(ttl))),
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    JWTParams.Issuer,
			Subject:   JWTParams.Subject,
		},
	}

	if JWTParams.Audience != "" {
		claims.Audience = []string{JWTParams.Audience}
	}
	if nbf > 0 {
		claims.NotBefore = jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(nbf)))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtValue, err := token.SignedString(key)
	if err != nil {
		return "", "", err
	}
	return jwtValue, claims.ID, nil
}
