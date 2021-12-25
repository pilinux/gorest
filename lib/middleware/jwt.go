package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// AccessKey ...
var AccessKey []byte

// AccessKeyTTL ...
var AccessKeyTTL int

// RefreshKey ...
var RefreshKey []byte

// RefreshKeyTTL ...
var RefreshKeyTTL int

// MyCustomClaims ...
type MyCustomClaims struct {
	ID    uint64 `json:"Id"`
	Email string `json:"Email"`
	jwt.StandardClaims
}

// AuthID - Access details
var AuthID uint64

// Email - Access details
var Email string

// JWTPayload ...
type JWTPayload struct {
	AccessJWT  string `json:"AccessJWT"`
	RefreshJWT string `json:"RefreshJWT"`
}

// JWT ...
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		val := c.Request.Header.Get("Authorization")
		if len(val) == 0 || !strings.Contains(val, "Bearer ") {
			// log.Println("no vals or no Bearer found")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		vals := strings.Split(val, " ")
		if len(vals) != 2 {
			// log.Println("result split not valid")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(vals[1], &MyCustomClaims{}, validateAccessJWT)

		if err != nil {
			// log.Println("error parsing JWT", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
			// fmt.Println(claims.ID, claims.Email)
			AuthID = claims.ID
		}
	}
}

// RefreshJWT ...
func RefreshJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var jwtPayload JWTPayload
		if err := c.ShouldBindJSON(&jwtPayload); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		token, err := jwt.ParseWithClaims(jwtPayload.RefreshJWT, &MyCustomClaims{}, validateRefreshJWT)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
			AuthID = claims.ID
			Email = claims.Email
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

// GetJWT ...
func GetJWT(id uint64, email string, tokenType string) (string, error) {
	var key []byte
	var ttl int
	if tokenType == "access" {
		key = AccessKey
		ttl = AccessKeyTTL
	}
	if tokenType == "refresh" {
		key = RefreshKey
		ttl = RefreshKeyTTL
	}
	// Create the Claims
	claims := MyCustomClaims{
		id,
		email,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * time.Duration(ttl)).Unix(),
			Issuer:    "GoRest API",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtValue, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return jwtValue, nil
}
