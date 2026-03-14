package database

import (
	"crypto/tls"
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
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

// SetSQLOpen replaces the package-level sqlOpen function so tests can
// inject errors. Call ResetSQLOpen to restore the original.
func SetSQLOpen(fn func(string, string) (*sql.DB, error)) {
	sqlOpen = fn
}

// ResetSQLOpen restores sqlOpen to the default sql.Open implementation.
func ResetSQLOpen() {
	sqlOpen = sql.Open
}

// SetDBClient replaces the package-level dbClient with the given
// *gorm.DB so tests can exercise CloseSQL error paths.
func SetDBClient(db *gorm.DB) {
	dbClient = db
}

// ResetDBClient sets dbClient to nil.
func ResetDBClient() {
	dbClient = nil
}
