package config

import (
	"crypto"
	"net/http"

	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
)

// SecurityConfig ...
type SecurityConfig struct {
	UserPassMinLength int

	MustBasicAuth string
	BasicAuth     struct {
		Username string
		Password string
	}

	MustJWT string
	JWT     middleware.JWTParameters

	InvalidateJWT string // when user logs off, invalidate the tokens

	AuthCookieActivate bool
	AuthCookiePath     string
	AuthCookieDomain   string
	AuthCookieSecure   bool
	AuthCookieHTTPOnly bool
	AuthCookieSameSite http.SameSite
	ServeJwtAsResBody  bool

	MustHash string
	HashPass lib.HashPassConfig
	HashSec  string // optional secret for argon2id hashing

	// data encryption at rest
	MustCipher bool
	CipherKey  []byte // for 256-bit ChaCha20-Poly1305
	Blake2bSec []byte // optional secret for blake2b hashing

	VerifyEmail bool
	RecoverPass bool

	MustFW   string
	Firewall struct {
		ListType string
		IP       string
	}

	MustCORS string
	CORS     []middleware.CORSPolicy

	CheckOrigin     string
	RateLimit       string
	TrustedPlatform string

	Must2FA string
	TwoFA   struct {
		Issuer string
		Crypto crypto.Hash
		Digits int

		Status Status2FA
		PathQR string

		DoubleHash bool
	}
}

// Status2FA - user's 2FA statuses
type Status2FA struct {
	Verified string
	On       string
	Off      string
	Invalid  string
}
