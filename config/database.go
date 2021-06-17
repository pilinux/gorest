package config

import "time"

// DatabaseConfig - database variables
type DatabaseConfig struct {
	DbDriver string
	DbUser   string
	DbPass   string
	DbName   string
	DbHost   string
	DbPort   string

	DbSslmode  string
	DbTimeZone string

	DbMaxIdleConns    int
	DbMaxOpenConns    int
	DbConnMaxLifetime time.Duration
	DbLogLevel        int
}
