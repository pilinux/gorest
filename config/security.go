package config

import (
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
}
