package config

import (
	"crypto"

	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
)

// SecurityConfigAll - exported variables
var SecurityConfigAll SecurityConfig

// SecurityConfig ...
type SecurityConfig struct {
	MustBasicAuth string
	BasicAuth     struct {
		Username string
		Password string
	}

	MustJWT string
	JWT     middleware.JWTParameters

	MustHash string
	HashPass lib.HashPassConfig

	MustFW   string
	Firewall struct {
		ListType string
		IP       string
	}

	MustCORS string
	CORS     struct {
		Origin      string
		Credentials string
		Headers     string
		Methods     string
		MaxAge      string
	}

	TrustedPlatform string

	Must2FA string
	TwoFA   struct {
		Issuer string
		Crypto crypto.Hash
		Digits int

		Status Status2FA
		PathQR string
	}
	InMemorySecret2FA map[uint64]Secret2FA
}

// Status2FA - user's 2FA statuses
type Status2FA struct {
	Verified string
	On       string
	Off      string
}

// Secret2FA - save encoded secrets in RAM temporarily
type Secret2FA struct {
	Image   string
	Secret  []byte
	PassSHA []byte
}
