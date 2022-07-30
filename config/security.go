package config

import (
	"github.com/pilinux/gorest/lib"
)

// SecurityConfig ...
type SecurityConfig struct {
	BasicAuth struct {
		Username string
		Password string
	}
	JWT struct {
		AccessKey     string
		RefreshKey    string
		AccessKeyTTL  int
		RefreshKeyTTL int

		Audience string
		Issuer   string
		AccNbf   int
		RefNbf   int
		Subject  string
	}
	HashPass lib.HashPassConfig
	Firewall struct {
		ListType string
		IP       string
	}
	CORS struct {
		Origin      string
		Credentials string
		Headers     string
		Methods     string
		MaxAge      string
	}
	TrustedPlatform string
}
