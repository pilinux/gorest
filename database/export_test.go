package database

import (
	"crypto/tls"

	"github.com/go-sql-driver/mysql"
)

// SetRegisterTLSConfig replaces the package-level registerTLSConfig
// function so tests can inject errors. Call ResetRegisterTLSConfig
// to restore the original.
func SetRegisterTLSConfig(fn func(string, *tls.Config) error) {
	registerTLSConfig = fn
}

// ResetRegisterTLSConfig restores registerTLSConfig to the default
// mysql.RegisterTLSConfig implementation.
func ResetRegisterTLSConfig() {
	registerTLSConfig = mysql.RegisterTLSConfig
}
