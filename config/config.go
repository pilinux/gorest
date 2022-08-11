// Package config is responsible for reading all environment
// variables and set up the base configuration for a
// functional application
package config

import (
	"crypto"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
	log "github.com/sirupsen/logrus"
)

// Activated - "yes" keyword to activate a service
const Activated string = "yes"

// Configuration - server and db configuration variables
type Configuration struct {
	Database   DatabaseConfig
	Logger     LoggerConfig
	Server     ServerConfig
	Security   SecurityConfig
	ViewConfig ViewConfig
}

// env - load the configurations from .env
func env() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.WithError(err).Panic("panic code: 101")
	}
}

// Config - load all the configurations
func Config() Configuration {
	var configuration Configuration

	configuration.Database = Database()
	configuration.Logger = Logger()
	configuration.Security = Security()
	configuration.Server = Server()
	configuration.ViewConfig = View()

	return configuration
}

// Database - all DB variables
func Database() DatabaseConfig {
	var databaseConfig DatabaseConfig

	// Load environment variables
	env()

	// RDBMS
	activateRDBMS := os.Getenv("ACTIVATE_RDBMS")
	if activateRDBMS == Activated {
		databaseConfig.RDBMS = DatabaseRDBMS().RDBMS

		// set params globally
		setParamsDatabaseRDBMS(databaseConfig.RDBMS)
	}
	DBConfigAll.RDBMS.Activate = activateRDBMS
	databaseConfig.RDBMS.Activate = activateRDBMS

	// REDIS
	activateRedis := os.Getenv("ACTIVATE_REDIS")
	if activateRedis == Activated {
		databaseConfig.REDIS = DatabaseRedis().REDIS

		// set params globally
		setParamsDatabaseRedis(databaseConfig.REDIS)
	}
	DBConfigAll.REDIS.Activate = activateRedis
	databaseConfig.REDIS.Activate = activateRedis

	// MongoDB
	activateMongo := os.Getenv("ACTIVATE_MONGO")
	if activateMongo == Activated {
		databaseConfig.MongoDB = DatabaseMongo().MongoDB

		// set params globally
		setParamsDatabaseMongo(databaseConfig.MongoDB)
	}
	DBConfigAll.MongoDB.Activate = activateMongo
	databaseConfig.MongoDB.Activate = activateMongo

	return databaseConfig
}

// DatabaseRDBMS - all RDBMS variables
func DatabaseRDBMS() DatabaseConfig {
	var databaseConfig DatabaseConfig
	var err error

	// Load environment variables
	env()

	// Env
	databaseConfig.RDBMS.Env.Driver = os.Getenv("DBDRIVER")
	databaseConfig.RDBMS.Env.Host = os.Getenv("DBHOST")
	databaseConfig.RDBMS.Env.Port = os.Getenv("DBPORT")
	databaseConfig.RDBMS.Env.TimeZone = os.Getenv("DBTIMEZONE")
	// Access
	databaseConfig.RDBMS.Access.DbName = os.Getenv("DBNAME")
	databaseConfig.RDBMS.Access.User = os.Getenv("DBUSER")
	databaseConfig.RDBMS.Access.Pass = os.Getenv("DBPASS")
	// SSL
	databaseConfig.RDBMS.Ssl.Sslmode = os.Getenv("DBSSLMODE")
	// Conn
	dbMaxIdleConns := os.Getenv("DBMAXIDLECONNS")
	dbMaxOpenConns := os.Getenv("DBMAXOPENCONNS")
	dbConnMaxLifetime := os.Getenv("DBCONNMAXLIFETIME")
	databaseConfig.RDBMS.Conn.MaxIdleConns, err = strconv.Atoi(dbMaxIdleConns)
	if err != nil {
		log.WithError(err).Panic("panic code: 131")
	}
	databaseConfig.RDBMS.Conn.MaxOpenConns, err = strconv.Atoi(dbMaxOpenConns)
	if err != nil {
		log.WithError(err).Panic("panic code: 132")
	}
	databaseConfig.RDBMS.Conn.ConnMaxLifetime, err = time.ParseDuration(dbConnMaxLifetime)
	if err != nil {
		log.WithError(err).Panic("panic code: 133")
	}

	// Logger
	dbLogLevel := os.Getenv("DBLOGLEVEL")
	databaseConfig.RDBMS.Log.LogLevel, err = strconv.Atoi(dbLogLevel)
	if err != nil {
		log.WithError(err).Panic("panic code: 134")
	}

	return databaseConfig
}

// DatabaseRedis - all REDIS DB variables
func DatabaseRedis() DatabaseConfig {
	var databaseConfig DatabaseConfig

	// Load environment variables
	env()

	// REDIS
	poolSize, err := strconv.Atoi(os.Getenv("POOLSIZE"))
	if err != nil {
		log.WithError(err).Panic("panic code: 135")
	}
	connTTL, err := strconv.Atoi(os.Getenv("CONNTTL"))
	if err != nil {
		log.WithError(err).Panic("panic code: 136")
	}

	databaseConfig.REDIS.Env.Host = os.Getenv("REDISHOST")
	databaseConfig.REDIS.Env.Port = os.Getenv("REDISPORT")
	databaseConfig.REDIS.Conn.PoolSize = poolSize
	databaseConfig.REDIS.Conn.ConnTTL = connTTL

	return databaseConfig
}

// DatabaseMongo - all MongoDB variables
func DatabaseMongo() DatabaseConfig {
	var databaseConfig DatabaseConfig

	// Load environment variables
	env()

	// MongoDB
	poolSize, err := strconv.ParseUint(os.Getenv("MONGO_POOLSIZE"), 10, 64)
	if err != nil {
		log.WithError(err).Panic("panic code: 137")
	}
	connTTL, err := strconv.Atoi(os.Getenv("MONGO_CONNTTL"))
	if err != nil {
		log.WithError(err).Panic("panic code: 138")
	}

	databaseConfig.MongoDB.Env.URI = os.Getenv("MONGO_URI")
	databaseConfig.MongoDB.Env.AppName = os.Getenv("MONGO_APP")
	databaseConfig.MongoDB.Env.PoolSize = poolSize
	databaseConfig.MongoDB.Env.PoolMon = os.Getenv("MONGO_MONITOR_POOL")
	databaseConfig.MongoDB.Env.ConnTTL = connTTL

	return databaseConfig
}

// Logger ...
func Logger() LoggerConfig {
	var loggerConfig LoggerConfig

	// Load environment variables
	env()

	loggerConfig.Activate = os.Getenv("ACTIVATE_SENTRY")
	if loggerConfig.Activate == Activated {
		loggerConfig.SentryDsn = os.Getenv("SentryDSN")
	}

	return loggerConfig
}

// Security - configs for generating tokens and hashes
func Security() SecurityConfig {
	var securityConfig SecurityConfig

	// Load environment variables
	env()

	// Basic auth
	SecurityConfigAll.MustBasicAuth = os.Getenv("ACTIVATE_BASIC_AUTH")
	if SecurityConfigAll.MustBasicAuth == Activated {
		username := os.Getenv("USERNAME")
		password := os.Getenv("PASSWORD")

		SecurityConfigAll.BasicAuth.Username = username
		SecurityConfigAll.BasicAuth.Password = password

		securityConfig.BasicAuth.Username = username
		securityConfig.BasicAuth.Password = password
	}
	securityConfig.MustBasicAuth = SecurityConfigAll.MustBasicAuth

	// JWT
	SecurityConfigAll.MustJWT = os.Getenv("ACTIVATE_JWT")
	if SecurityConfigAll.MustJWT == Activated {
		securityConfig.JWT = getParamsJWT()

		// set params globally
		setParamsJWT(securityConfig.JWT)
	}
	securityConfig.MustJWT = SecurityConfigAll.MustJWT

	// Hashing passwords
	SecurityConfigAll.MustHash = os.Getenv("ACTIVATE_HASHING")
	if SecurityConfigAll.MustHash == Activated {
		securityConfig.HashPass = getParamsHash()

		// set params globally
		setParamsHash(securityConfig.HashPass)
	}
	securityConfig.MustHash = SecurityConfigAll.MustHash

	// Two-factor authentication
	SecurityConfigAll.Must2FA = os.Getenv("ACTIVATE_2FA")
	if SecurityConfigAll.Must2FA == Activated {
		SecurityConfigAll.TwoFA.Issuer = os.Getenv("TWO_FA_ISSUER")

		cryptoAlg := os.Getenv("TWO_FA_CRYPTO")
		if cryptoAlg == "1" {
			SecurityConfigAll.TwoFA.Crypto = crypto.SHA1
		}
		if cryptoAlg == "256" {
			SecurityConfigAll.TwoFA.Crypto = crypto.SHA256
		}
		if cryptoAlg == "512" {
			SecurityConfigAll.TwoFA.Crypto = crypto.SHA512
		}

		digits, err := strconv.Atoi(os.Getenv("TWO_FA_DIGITS"))
		if err != nil {
			log.WithError(err).Panic("panic code: 130")
		}
		SecurityConfigAll.TwoFA.Digits = digits

		// define different statuses of individual user
		SecurityConfigAll.TwoFA.Status.Verified = os.Getenv("TWO_FA_VERIFIED")
		SecurityConfigAll.TwoFA.Status.On = os.Getenv("TWO_FA_ON")
		SecurityConfigAll.TwoFA.Status.Off = os.Getenv("TWO_FA_OFF")
		SecurityConfigAll.TwoFA.Status.Invalid = os.Getenv("TWO_FA_INVALID")

		// for saving QR temporarily
		SecurityConfigAll.TwoFA.PathQR = os.Getenv("TWO_FA_QR_PATH")

		securityConfig.TwoFA.Issuer = SecurityConfigAll.TwoFA.Issuer
		securityConfig.TwoFA.Crypto = SecurityConfigAll.TwoFA.Crypto
		securityConfig.TwoFA.Digits = SecurityConfigAll.TwoFA.Digits
		securityConfig.TwoFA.Status.Verified = SecurityConfigAll.TwoFA.Status.Verified
		securityConfig.TwoFA.Status.On = SecurityConfigAll.TwoFA.Status.On
		securityConfig.TwoFA.Status.Off = SecurityConfigAll.TwoFA.Status.Off
		securityConfig.TwoFA.Status.Invalid = SecurityConfigAll.TwoFA.Status.Invalid
		securityConfig.TwoFA.PathQR = SecurityConfigAll.TwoFA.PathQR
	}
	securityConfig.Must2FA = SecurityConfigAll.Must2FA

	// App firewall
	SecurityConfigAll.MustFW = os.Getenv("ACTIVATE_FIREWALL")
	if SecurityConfigAll.MustFW == Activated {
		listType := os.Getenv("LISTTYPE")
		ip := os.Getenv("IP")

		SecurityConfigAll.Firewall.ListType = listType
		SecurityConfigAll.Firewall.IP = ip

		securityConfig.Firewall.ListType = listType
		securityConfig.Firewall.IP = ip
	}
	securityConfig.MustFW = SecurityConfigAll.MustFW

	// CORS
	SecurityConfigAll.MustCORS = os.Getenv("ACTIVATE_CORS")
	if SecurityConfigAll.MustCORS == Activated {
		securityConfig.CORS.Origin = os.Getenv("CORS_ORIGIN")
		securityConfig.CORS.Credentials = os.Getenv("CORS_CREDENTIALS")
		securityConfig.CORS.Headers = os.Getenv("CORS_HEADERS")
		securityConfig.CORS.Methods = os.Getenv("CORS_METHODS")
		securityConfig.CORS.MaxAge = os.Getenv("CORS_MAXAGE")
	}
	securityConfig.MustCORS = SecurityConfigAll.MustCORS

	// Important for getting real client IP
	securityConfig.TrustedPlatform = os.Getenv("TRUSTED_PLATFORM")

	return securityConfig
}

// Server - port and env
func Server() ServerConfig {
	var serverConfig ServerConfig

	// Load environment variables
	env()

	serverConfig.ServerPort = os.Getenv("APP_PORT")
	serverConfig.ServerEnv = os.Getenv("APP_ENV")

	return serverConfig
}

// View - HTML renderer
func View() ViewConfig {
	var viewConfig ViewConfig

	// Load environment variables
	env()

	viewConfig.Activate = os.Getenv("ACTIVATE_VIEW")
	if viewConfig.Activate == Activated {
		viewConfig.Directory = os.Getenv("TEMPLATE_DIR")
	}

	return viewConfig
}

// setParamsDatabaseRDBMS - set parameters for RDBMS
func setParamsDatabaseRDBMS(c RDBMS) {
	DBConfigAll.RDBMS.Env.Driver = c.Env.Driver
	DBConfigAll.RDBMS.Env.Host = c.Env.Host
	DBConfigAll.RDBMS.Env.Port = c.Env.Port
	DBConfigAll.RDBMS.Env.TimeZone = c.Env.TimeZone

	DBConfigAll.RDBMS.Access.DbName = c.Access.DbName
	DBConfigAll.RDBMS.Access.User = c.Access.User
	DBConfigAll.RDBMS.Access.Pass = c.Access.Pass

	DBConfigAll.RDBMS.Ssl.Sslmode = c.Ssl.Sslmode

	DBConfigAll.RDBMS.Conn.MaxIdleConns = c.Conn.MaxIdleConns
	DBConfigAll.RDBMS.Conn.MaxOpenConns = c.Conn.MaxOpenConns
	DBConfigAll.RDBMS.Conn.ConnMaxLifetime = c.Conn.ConnMaxLifetime

	DBConfigAll.RDBMS.Log.LogLevel = c.Log.LogLevel
}

// setParamsDatabaseRedis - set parameters for Redis
func setParamsDatabaseRedis(c REDIS) {
	DBConfigAll.REDIS.Env.Host = c.Env.Host
	DBConfigAll.REDIS.Env.Port = c.Env.Port

	DBConfigAll.REDIS.Conn.PoolSize = c.Conn.PoolSize
	DBConfigAll.REDIS.Conn.ConnTTL = c.Conn.ConnTTL
}

// setParamsDatabaseMongo - set parameters for MongoDB
func setParamsDatabaseMongo(c MongoDB) {
	DBConfigAll.MongoDB.Env.AppName = c.Env.AppName
	DBConfigAll.MongoDB.Env.URI = c.Env.URI
	DBConfigAll.MongoDB.Env.PoolSize = c.Env.PoolSize
	DBConfigAll.MongoDB.Env.PoolMon = c.Env.PoolMon
	DBConfigAll.MongoDB.Env.ConnTTL = c.Env.ConnTTL
}

// getParamsJWT - read parameters from env
func getParamsJWT() middleware.JWTParameters {
	env()

	params := middleware.JWTParameters{}
	var err error

	params.AccessKey = []byte(os.Getenv("ACCESS_KEY"))
	params.AccessKeyTTL, err = strconv.Atoi(os.Getenv("ACCESS_KEY_TTL"))
	if err != nil {
		log.WithError(err).Panic("panic code: 111")
	}
	params.RefreshKey = []byte(os.Getenv("REFRESH_KEY"))
	params.RefreshKeyTTL, err = strconv.Atoi(os.Getenv("REFRESH_KEY_TTL"))
	if err != nil {
		log.WithError(err).Panic("panic code: 112")
	}
	params.Audience = os.Getenv("AUDIENCE")
	params.Issuer = os.Getenv("ISSUER")
	params.AccNbf, err = strconv.Atoi(os.Getenv("NOT_BEFORE_ACC"))
	if err != nil {
		log.WithError(err).Panic("panic code: 113")
	}
	params.RefNbf, err = strconv.Atoi(os.Getenv("NOT_BEFORE_REF"))
	if err != nil {
		log.WithError(err).Panic("panic code: 114")
	}
	params.Subject = os.Getenv("SUBJECT")

	return params
}

// setParamsJWT - set parameters for JWT
func setParamsJWT(c middleware.JWTParameters) {
	middleware.JWTParams.AccessKey = c.AccessKey
	middleware.JWTParams.AccessKeyTTL = c.AccessKeyTTL
	middleware.JWTParams.RefreshKey = c.RefreshKey
	middleware.JWTParams.RefreshKeyTTL = c.RefreshKeyTTL
	middleware.JWTParams.Audience = c.Audience
	middleware.JWTParams.Issuer = c.Issuer
	middleware.JWTParams.AccNbf = c.AccNbf
	middleware.JWTParams.RefNbf = c.RefNbf
	middleware.JWTParams.Subject = c.Subject
}

// getParamsHash - read parameters from env
func getParamsHash() lib.HashPassConfig {
	env()

	params := lib.HashPassConfig{}
	var err error

	hashPassMemory64, err := strconv.ParseUint((os.Getenv("HASHPASSMEMORY")), 10, 32)
	if err != nil {
		log.WithError(err).Panic("panic code: 121")
	}
	hashPassIterations64, err := strconv.ParseUint((os.Getenv("HASHPASSITERATIONS")), 10, 32)
	if err != nil {
		log.WithError(err).Panic("panic code: 122")
	}
	hashPassParallelism64, err := strconv.ParseUint((os.Getenv("HASHPASSPARALLELISM")), 10, 8)
	if err != nil {
		log.WithError(err).Panic("panic code: 123")
	}
	hashPassSaltLength64, err := strconv.ParseUint((os.Getenv("HASHPASSSALTLENGTH")), 10, 32)
	if err != nil {
		log.WithError(err).Panic("panic code: 124")
	}
	hashPassKeyLength64, err := strconv.ParseUint((os.Getenv("HASHPASSKEYLENGTH")), 10, 32)
	if err != nil {
		log.WithError(err).Panic("panic code: 125")
	}

	params.Memory = uint32(hashPassMemory64)
	params.Iterations = uint32(hashPassIterations64)
	params.Parallelism = uint8(hashPassParallelism64)
	params.SaltLength = uint32(hashPassSaltLength64)
	params.KeyLength = uint32(hashPassKeyLength64)

	return params
}

// setParamsHash - set parameters for hashing
func setParamsHash(c lib.HashPassConfig) {
	SecurityConfigAll.HashPass.Memory = c.Memory
	SecurityConfigAll.HashPass.Iterations = c.Iterations
	SecurityConfigAll.HashPass.Parallelism = c.Parallelism
	SecurityConfigAll.HashPass.SaltLength = c.SaltLength
	SecurityConfigAll.HashPass.KeyLength = c.KeyLength
}
