package config

import (
	"crypto"

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

	MustHash string
	HashPass lib.HashPassConfig

	VerifyEmail bool
	RecoverPass bool

	MustFW   string
	Firewall struct {
		ListType string
		IP       string
	}

	MustCORS string
	CORS     []middleware.CORSPolicy

	TrustedPlatform string

	Must2FA string
	TwoFA   struct {
		Issuer string
		Crypto crypto.Hash
		Digits int

		Status Status2FA
		PathQR string
	}
}

// Status2FA - user's 2FA statuses
type Status2FA struct {
	Verified string
	On       string
	Off      string
	Invalid  string
}
