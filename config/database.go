package config

import (
	"time"
)

// DatabaseConfig - database variables
type DatabaseConfig struct {
	// relational database
	RDBMS struct {
		Activate string
		Env      struct {
			Driver   string
			Host     string
			Port     string
			TimeZone string
		}
		Access struct {
			DbName string
			User   string
			Pass   string
		}
		Ssl struct {
			Sslmode string
		}
		Conn struct {
			MaxIdleConns    int
			MaxOpenConns    int
			ConnMaxLifetime time.Duration
		}
		Log struct {
			LogLevel int
		}
	}

	// redis database
	REDIS struct {
		Activate string
		Env      struct {
			Host string
			Port string
		}
		Conn struct {
			PoolSize int
			ConnTTL  int
		}
	}

	// mongo database
	MongoDB struct {
		Activate string
		Env      struct {
			AppName  string
			URI      string
			PoolSize uint64
			PoolMon  string
			ConnTTL  int
		}
	}
}
