package config

import (
	"time"
)

// DatabaseConfig holds all database variables.
type DatabaseConfig struct {
	// relational database
	RDBMS RDBMS

	// redis database
	REDIS REDIS

	// mongo database
	MongoDB MongoDB
}

// RDBMS holds relational database variables.
type RDBMS struct {
	Activate string
	Env      struct {
		Driver   string `json:"-"`
		URI      string `json:"-"`
		Host     string `json:"-"`
		Port     string `json:"-"`
		TimeZone string `json:"-"`
	}
	Access struct {
		DbName string `json:"-"`
		User   string `json:"-"`
		Pass   string `json:"-"`
	}
	Ssl struct {
		Sslmode    string `json:"-"`
		MinTLS     string `json:"-"`
		RootCA     string `json:"-"`
		ServerCert string `json:"-"`
		ClientCert string `json:"-"`
		ClientKey  string `json:"-"`
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

// REDIS holds redis database variables.
type REDIS struct {
	Activate string
	Env      struct {
		URI  string `json:"-"`
		Host string `json:"-"`
		Port string `json:"-"`
	}
	Conn struct {
		PoolSize int
		ConnTTL  int
	}
}

// MongoDB holds mongo database variables.
type MongoDB struct {
	Activate string
	Env      struct {
		AppName  string
		URI      string `json:"-"`
		PoolSize uint64 `json:"-"`
		PoolMon  string
		ConnTTL  int
	}
}
