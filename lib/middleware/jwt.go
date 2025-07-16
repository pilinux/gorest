package middleware

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTParameters - params to configure JWT
type JWTParameters struct {
	Algorithm     string
	AccessKey     []byte
	AccessKeyTTL  int
	RefreshKey    []byte
	RefreshKeyTTL int
	PrivKeyECDSA  *ecdsa.PrivateKey
	PubKeyECDSA   *ecdsa.PublicKey
	PrivKeyRSA    *rsa.PrivateKey
	PubKeyRSA     *rsa.PublicKey

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
	Azp     string `json:"azp,omitempty"` // authorized party
	Fva     []int  `json:"fva,omitempty"` // factor verification age
	Sid     string `json:"sid,omitempty"` // session ID
	V       int    `json:"v,omitempty"`   // version
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
func JWT(namedCookie ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var jwtPayload JWTPayload
		var token *jwt.Token
		var val string
		var vals []string
		var accessJWT string
		var err error

		// if namedCookie is provided, use it to read the JWT from the cookie
		if len(namedCookie) > 0 {
			accessJWT, err = c.Cookie(namedCookie[0])
			// accessJWT is available in the cookie
			if err == nil {
				jwtPayload.AccessJWT = accessJWT
				goto VerifyClaims
			}
		}

		// first try to read the cookie
		accessJWT, err = c.Cookie("accessJWT")
		// accessJWT is available in the cookie
		if err == nil {
			jwtPayload.AccessJWT = accessJWT
			goto VerifyClaims
		}

		// accessJWT is not available in the cookie
		// try to read the Authorization header
		val = c.Request.Header.Get("Authorization")
		if len(val) == 0 || !strings.Contains(val, "Bearer") {
			// no vals or no bearer found
			c.AbortWithStatusJSON(http.StatusUnauthorized, "token missing")
			return
		}
		vals = strings.Split(val, " ")
		// Authorization: Bearer {access} => length is 2
		// Authorization: Bearer {access} {refresh} => length is 3
		if len(vals) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "token missing")
			return
		}
		jwtPayload.AccessJWT = vals[1]

	VerifyClaims:
		token, err = jwt.ParseWithClaims(jwtPayload.AccessJWT, &JWTClaims{}, ValidateAccessJWT)
		if err != nil {
			// error parsing JWT
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
			return
		}
		if token == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "token missing")
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid token claims")
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid token")
			return
		}

		c.Set("authID", claims.AuthID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("scope", claims.Scope)
		c.Set("tfa", claims.TwoFA)
		c.Set("siteLan", claims.SiteLan)
		c.Set("custom1", claims.Custom1)
		c.Set("custom2", claims.Custom2)
		if claims.ExpiresAt != nil {
			c.Set("expAccess", claims.ExpiresAt.Unix()) // in Unix epoch time
		}
		if claims.IssuedAt != nil {
			c.Set("iatAccess", claims.IssuedAt.Unix()) // in Unix epoch time
		}
		c.Set("jtiAccess", claims.ID)

		// set values for external auth providers if available
		c.Set("azp", claims.Azp) // authorized party
		c.Set("fva", claims.Fva) // factor verification age
		c.Set("sid", claims.Sid) // session ID
		c.Set("v", claims.V)     // version

		// set values from RegisteredClaims
		//
		// token issuer
		c.Set("iss", claims.Issuer)
		//
		// token subject
		c.Set("sub", claims.Subject)
		//
		// token audience
		c.Set("aud", claims.Audience)
		//
		// token issued at
		if claims.IssuedAt != nil {
			c.Set("iat", claims.IssuedAt.Unix())
		}
		//
		// token expiration time
		if claims.ExpiresAt != nil {
			c.Set("exp", claims.ExpiresAt.Unix())
		}
		//
		// token not before time
		if claims.NotBefore != nil {
			c.Set("nbf", claims.NotBefore.Unix())
		}
		//
		// token ID
		c.Set("jti", claims.ID)

		c.Next()
	}
}

// RefreshJWT - validate refresh token
func RefreshJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var jwtPayload JWTPayload
		var val string
		var vals []string

		// first try to read the cookie
		refreshJWT, err := c.Cookie("refreshJWT")
		// refreshJWT is available in the cookie
		if err == nil {
			jwtPayload.RefreshJWT = refreshJWT
			goto VerifyClaims
		}

		// refreshJWT is not available in the cookie
		// try to read the Authorization header
		val = c.Request.Header.Get("Authorization")
		if len(val) == 0 || !strings.Contains(val, "Bearer") {
			// no vals or no bearer found
			goto CheckReqBody
		}
		vals = strings.Split(val, " ")
		// Authorization: Bearer {refresh} => length is 2
		// Authorization: Bearer {access} {refresh} => length is 3
		if len(vals) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "token missing")
			return
		}
		jwtPayload.RefreshJWT = vals[1]
		if len(vals) == 3 {
			jwtPayload.RefreshJWT = vals[2]
		}
		goto VerifyClaims

	CheckReqBody:
		// refreshJWT is not available in the cookie or Authorization header
		// try to read the request body
		if err := c.ShouldBindJSON(&jwtPayload); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
			return
		}
		jwtPayload.RefreshJWT = strings.TrimSpace(jwtPayload.RefreshJWT)

	VerifyClaims:
		token, err := jwt.ParseWithClaims(jwtPayload.RefreshJWT, &JWTClaims{}, ValidateRefreshJWT)
		if err != nil {
			// error parsing JWT
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
			return
		}
		if token == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "token missing")
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid token claims")
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "invalid token")
			return
		}

		c.Set("authID", claims.AuthID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("scope", claims.Scope)
		c.Set("tfa", claims.TwoFA)
		c.Set("siteLan", claims.SiteLan)
		c.Set("custom1", claims.Custom1)
		c.Set("custom2", claims.Custom2)
		if claims.ExpiresAt != nil {
			c.Set("expRefresh", claims.ExpiresAt.Unix()) // in Unix epoch time
		}
		if claims.IssuedAt != nil {
			c.Set("iatRefresh", claims.IssuedAt.Unix()) // in Unix epoch time
		}
		c.Set("jtiRefresh", claims.ID)

		// set values for external auth providers if available
		c.Set("azp", claims.Azp) // authorized party
		c.Set("fva", claims.Fva) // factor verification age
		c.Set("sid", claims.Sid) // session ID
		c.Set("v", claims.V)     // version

		// set values from RegisteredClaims
		//
		// token issuer
		c.Set("iss", claims.Issuer)
		//
		// token subject
		c.Set("sub", claims.Subject)
		//
		// token audience
		c.Set("aud", claims.Audience)
		//
		// token issued at
		if claims.IssuedAt != nil {
			c.Set("iat", claims.IssuedAt.Unix())
		}
		//
		// token expiration time
		if claims.ExpiresAt != nil {
			c.Set("exp", claims.ExpiresAt.Unix())
		}
		//
		// token not before time
		if claims.NotBefore != nil {
			c.Set("nbf", claims.NotBefore.Unix())
		}
		//
		// token ID
		c.Set("jti", claims.ID)

		c.Next()
	}
}

// ValidateHMACAccess - validate hash based access token
func ValidateHMACAccess(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return JWTParams.AccessKey, nil
}

// ValidateHMACRefresh - validate hash based refresh token
func ValidateHMACRefresh(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return JWTParams.RefreshKey, nil
}

// ValidateECDSA - validate elliptic curve digital signature algorithm based token
func ValidateECDSA(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return JWTParams.PubKeyECDSA, nil
}

// ValidateRSA - validate Rivest–Shamir–Adleman cryptosystem based token
func ValidateRSA(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return JWTParams.PubKeyRSA, nil
}

// ValidateAccessJWT - verify the access JWT's signature, and validate its claims
func ValidateAccessJWT(token *jwt.Token) (any, error) {
	alg := JWTParams.Algorithm

	switch alg {
	case "HS256", "HS384", "HS512":
		return ValidateHMACAccess(token)
	case "ES256", "ES384", "ES512":
		return ValidateECDSA(token)
	case "RS256", "RS384", "RS512":
		return ValidateRSA(token)
	default:
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
}

// ValidateRefreshJWT - verify the refresh JWT's signature, and validate its claims
func ValidateRefreshJWT(token *jwt.Token) (any, error) {
	alg := JWTParams.Algorithm

	switch alg {
	case "HS256", "HS384", "HS512":
		return ValidateHMACRefresh(token)
	case "ES256", "ES384", "ES512":
		return ValidateECDSA(token)
	case "RS256", "RS384", "RS512":
		return ValidateRSA(token)
	default:
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
}

// GetJWT - issue new tokens
func GetJWT(customClaims MyCustomClaims, tokenType string) (string, string, error) {
	var (
		key     []byte
		privKey any
		ttl     int
		nbf     int
	)

	switch tokenType {
	case "access":
		key = JWTParams.AccessKey
		ttl = JWTParams.AccessKeyTTL
		nbf = JWTParams.AccNbf
	case "refresh":
		key = JWTParams.RefreshKey
		ttl = JWTParams.RefreshKeyTTL
		nbf = JWTParams.RefNbf
	default:
		return "", "", errors.New("invalid token type")
	}

	// Create the Claims
	claims := JWTClaims{
		MyCustomClaims: customClaims,
		RegisteredClaims: jwt.RegisteredClaims{
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

	var token *jwt.Token
	alg := jwt.GetSigningMethod(JWTParams.Algorithm)

	switch JWTParams.Algorithm {
	case "HS256", "HS384", "HS512":
		token = jwt.NewWithClaims(alg, claims)
		privKey = key
	case "ES256", "ES384", "ES512":
		token = jwt.NewWithClaims(alg, claims)
		privKey = JWTParams.PrivKeyECDSA
	case "RS256", "RS384", "RS512":
		token = jwt.NewWithClaims(alg, claims)
		privKey = JWTParams.PrivKeyRSA
	default:
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	}

	// HMAC
	//
	// HS256
	// openssl rand -base64 32
	//
	// HS384
	// openssl rand -base64 48
	//
	// HS512
	// openssl rand -base64 64

	// ECDSA
	//
	// ES256
	// prime256v1: X9.62/SECG curve over a 256 bit prime field, also known as P-256 or NIST P-256
	// widely used, recommended for general-purpose cryptographic operations
	// openssl ecparam -name prime256v1 -genkey -noout -out private-key.pem
	// openssl ec -in private-key.pem -pubout -out public-key.pem
	//
	// ES384
	// secp384r1: NIST/SECG curve over a 384 bit prime field
	// openssl ecparam -name secp384r1 -genkey -noout -out private-key.pem
	// openssl ec -in private-key.pem -pubout -out public-key.pem
	//
	// ES512
	// secp521r1: NIST/SECG curve over a 521 bit prime field
	// openssl ecparam -name secp521r1 -genkey -noout -out private-key.pem
	// openssl ec -in private-key.pem -pubout -out public-key.pem

	// RSA
	//
	// RS256
	// openssl genpkey -algorithm RSA -out private-key.pem -pkeyopt rsa_keygen_bits:2048
	// openssl rsa -in private-key.pem -pubout -out public-key.pem
	//
	// RS384
	// openssl genpkey -algorithm RSA -out private-key.pem -pkeyopt rsa_keygen_bits:3072
	// openssl rsa -in private-key.pem -pubout -out public-key.pem
	//
	// RS512
	// openssl genpkey -algorithm RSA -out private-key.pem -pkeyopt rsa_keygen_bits:4096
	// openssl rsa -in private-key.pem -pubout -out public-key.pem

	jwtValue, err := token.SignedString(privKey)
	if err != nil {
		return "", "", err
	}

	return jwtValue, claims.ID, nil
}
