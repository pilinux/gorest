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
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// variables for issuing or validating tokens
var (
	AccessKey     []byte
	AccessKeyTTL  int
	RefreshKey    []byte
	RefreshKeyTTL int

	Audience string
	Issuer   string
	AccNbf   int
	RefNbf   int
	Subject  string
)

// myCustomClaims ...
type myCustomClaims struct {
	AuthID  uint64 `json:"authID,omitempty"`
	Email   string `json:"email,omitempty"`
	Role    string `json:"role,omitempty"`
	Scope   string `json:"scope,omitempty"`
	SiteLan string `json:"siteLan,omitempty"`
	Custom1 string `json:"custom1,omitempty"`
	Custom2 string `json:"custom2,omitempty"`
}

// JWTClaims ...
type JWTClaims struct {
	myCustomClaims
	jwt.StandardClaims
}

// user-related info for JWT
var (
	AuthID  uint64
	Email   string
	Role    string
	Scope   string
	SiteLan string
	Custom1 string
	Custom2 string
)

// JWTPayload ...
type JWTPayload struct {
	AccessJWT  string `json:"accessJWT,omitempty"`
	RefreshJWT string `json:"refreshJWT,omitempty"`
}

// JWT - validate access token
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		val := c.Request.Header.Get("Authorization")
		if len(val) == 0 || !strings.Contains(val, "Bearer ") {
			// no vals or no bearer found
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		vals := strings.Split(val, " ")
		if len(vals) != 2 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(vals[1], &JWTClaims{}, validateAccessJWT)

		if err != nil {
			// error parsing JWT
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			AuthID = claims.AuthID
			Email = claims.Email
			Role = claims.Role
			Scope = claims.Scope
			SiteLan = claims.SiteLan
			Custom1 = claims.Custom1
			Custom2 = claims.Custom2
		}
	}
}

// RefreshJWT - validate refresh token
func RefreshJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var jwtPayload JWTPayload
		if err := c.ShouldBindJSON(&jwtPayload); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		token, err := jwt.ParseWithClaims(jwtPayload.RefreshJWT, &JWTClaims{}, validateRefreshJWT)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			AuthID = claims.AuthID
			Email = claims.Email
			Role = claims.Role
			Scope = claims.Scope
			SiteLan = claims.SiteLan
			Custom1 = claims.Custom1
			Custom2 = claims.Custom2
		}
	}
}

// validateAccessJWT ...
func validateAccessJWT(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return AccessKey, nil
}

// validateRefreshJWT ...
func validateRefreshJWT(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return RefreshKey, nil
}

// GetJWT - issue new tokens
func GetJWT(id uint64, email, role, scope, siteLan, custom1, custom2, tokenType string) (string, string, error) {
	var (
		key []byte
		ttl int
		nbf int
	)

	if tokenType == "access" {
		key = AccessKey
		ttl = AccessKeyTTL
		nbf = AccNbf
	}
	if tokenType == "refresh" {
		key = RefreshKey
		ttl = RefreshKeyTTL
		nbf = RefNbf
	}
	// Create the Claims
	claims := JWTClaims{
		myCustomClaims{
			AuthID:  id,
			Email:   email,
			Role:    role,
			Scope:   scope,
			SiteLan: siteLan,
			Custom1: custom1,
			Custom2: custom2,
		},
		jwt.StandardClaims{
			Audience:  Audience,
			ExpiresAt: time.Now().Add(time.Minute * time.Duration(ttl)).Unix(),
			Id:        uuid.NewString(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    Issuer,
			Subject:   Subject,
		},
	}

	if nbf > 0 {
		claims.NotBefore = time.Now().Add(time.Second * time.Duration(nbf)).Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtValue, err := token.SignedString(key)
	if err != nil {
		return "", "", err
	}
	return jwtValue, claims.Id, nil
}
