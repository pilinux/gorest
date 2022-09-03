// Package config is responsible for reading all environment
// variables and set up the base configuration for a
// functional application
package config

import (
	"crypto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	EmailConf  EmailConfig
	Logger     LoggerConfig
	Server     ServerConfig
	Security   SecurityConfig
	ViewConfig ViewConfig
}

var configAll *Configuration

// env - load the configurations from .env
func env() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.WithError(err).Panic("panic code: 101")
	}
}

// Config - load all the configurations
func Config() *Configuration {
	var configuration Configuration

	configuration.Database = database()
	configuration.EmailConf = email()
	configuration.Logger = logger()
	configuration.Security = security()
	configuration.Server = server()
	configuration.ViewConfig = view()

	configAll = &configuration

	return configAll
}

// GetConfig - return all the config variables
func GetConfig() *Configuration {
	return configAll
}

// database - all DB variables
func database() DatabaseConfig {
	var databaseConfig DatabaseConfig

	// Load environment variables
	env()

	// RDBMS
	activateRDBMS := os.Getenv("ACTIVATE_RDBMS")
	if activateRDBMS == Activated {
		databaseConfig.RDBMS = databaseRDBMS().RDBMS
	}
	databaseConfig.RDBMS.Activate = activateRDBMS

	// REDIS
	activateRedis := os.Getenv("ACTIVATE_REDIS")
	if activateRedis == Activated {
		databaseConfig.REDIS = databaseRedis().REDIS
	}
	databaseConfig.REDIS.Activate = activateRedis

	// MongoDB
	activateMongo := os.Getenv("ACTIVATE_MONGO")
	if activateMongo == Activated {
		databaseConfig.MongoDB = databaseMongo().MongoDB
	}
	databaseConfig.MongoDB.Activate = activateMongo

	return databaseConfig
}

// databaseRDBMS - all RDBMS variables
func databaseRDBMS() DatabaseConfig {
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

// databaseRedis - all REDIS DB variables
func databaseRedis() DatabaseConfig {
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

// databaseMongo - all MongoDB variables
func databaseMongo() DatabaseConfig {
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

// email - config for using external email services
func email() EmailConfig {
	var emailConfig EmailConfig
	var err error

	// Load environment variables
	env()

	emailConfig.Activate = os.Getenv("ACTIVATE_EMAIL_SERVICE")
	if emailConfig.Activate == Activated {
		emailConfig.Provider = os.Getenv("EMAIL_SERVICE_PROVIDER")
		emailConfig.APIToken = os.Getenv("EMAIL_API_TOKEN")
		emailConfig.AddrFrom = os.Getenv("EMAIL_FROM")

		emailConfig.TrackOpens = false
		trackOpens := os.Getenv("EMAIL_TRACK_OPENS")
		if trackOpens == "yes" {
			emailConfig.TrackOpens = true
		}

		emailConfig.TrackLinks = os.Getenv("EMAIL_TRACK_LINKS")
		emailConfig.DeliveryType = os.Getenv("EMAIL_DELIVERY_TYPE")

		emailConfig.EmailVerificationTemplateID, err = strconv.ParseInt(os.Getenv("EMAIL_VERIFY_TEMPLATE_ID"), 10, 64)
		if err != nil {
			log.WithError(err).Panic("panic code: 141")
		}
		emailConfig.PasswordRecoverTemplateID, err = strconv.ParseInt(os.Getenv("EMAIL_PASS_RECOVER_TEMPLATE_ID"), 10, 64)
		if err != nil {
			log.WithError(err).Panic("panic code: 142")
		}
		emailConfig.EmailVerificationCodeLength, err = strconv.ParseUint(os.Getenv("EMAIL_VERIFY_CODE_LENGTH"), 10, 32)
		if err != nil {
			log.WithError(err).Panic("panic code: 143")
		}
		emailConfig.PasswordRecoverCodeLength, err = strconv.ParseUint(os.Getenv("EMAIL_PASS_RECOVER_CODE_LENGTH"), 10, 32)
		if err != nil {
			log.WithError(err).Panic("panic code: 144")
		}
		emailConfig.EmailVerificationTag = os.Getenv("EMAIL_VERIFY_TAG")
		emailConfig.PasswordRecoverTag = os.Getenv("EMAIL_PASS_RECOVER_TAG")
		emailConfig.HTMLModel = os.Getenv("EMAIL_HTML_MODEL")
		emailConfig.EmailVerifyValidityPeriod, err = strconv.ParseUint(os.Getenv("EMAIL_VERIFY_VALIDITY_PERIOD"), 10, 32)
		if err != nil {
			log.WithError(err).Panic("panic code: 145")
		}
		emailConfig.PassRecoverValidityPeriod, err = strconv.ParseUint(os.Getenv("EMAIL_PASS_RECOVER_VALIDITY_PERIOD"), 10, 32)
		if err != nil {
			log.WithError(err).Panic("panic code: 146")
		}
	}

	return emailConfig
}

// logger - config for sentry.io
func logger() LoggerConfig {
	var loggerConfig LoggerConfig

	// Load environment variables
	env()

	loggerConfig.Activate = os.Getenv("ACTIVATE_SENTRY")
	if loggerConfig.Activate == Activated {
		loggerConfig.SentryDsn = os.Getenv("SentryDSN")
	}

	return loggerConfig
}

// security - configs for generating tokens and hashes
func security() SecurityConfig {
	var securityConfig SecurityConfig

	// Load environment variables
	env()

	// Minimum password length
	userPassMinLength, err := strconv.Atoi(os.Getenv("MIN_PASS_LENGTH"))
	if err != nil {
		log.WithError(err).Panic("panic code: 129")
	}
	securityConfig.UserPassMinLength = userPassMinLength

	// Basic auth
	securityConfig.MustBasicAuth = os.Getenv("ACTIVATE_BASIC_AUTH")
	if securityConfig.MustBasicAuth == Activated {
		securityConfig.BasicAuth.Username = os.Getenv("USERNAME")
		securityConfig.BasicAuth.Password = os.Getenv("PASSWORD")
	}

	// JWT
	securityConfig.MustJWT = os.Getenv("ACTIVATE_JWT")
	if securityConfig.MustJWT == Activated {
		securityConfig.JWT = getParamsJWT()

		// set params globally
		setParamsJWT(securityConfig.JWT)
	}

	// Hashing passwords
	securityConfig.MustHash = os.Getenv("ACTIVATE_HASHING")
	if securityConfig.MustHash == Activated {
		securityConfig.HashPass = getParamsHash()
	}

	// Email verification and password recovery
	securityConfig.VerifyEmail = false
	securityConfig.RecoverPass = false
	if os.Getenv("VERIFY_EMAIL") == "yes" {
		securityConfig.VerifyEmail = true
	}
	if os.Getenv("RECOVER_PASSWORD") == "yes" {
		securityConfig.RecoverPass = true
	}

	// Two-factor authentication
	securityConfig.Must2FA = os.Getenv("ACTIVATE_2FA")
	if securityConfig.Must2FA == Activated {
		securityConfig.TwoFA.Issuer = os.Getenv("TWO_FA_ISSUER")

		cryptoAlg := os.Getenv("TWO_FA_CRYPTO")
		if cryptoAlg == "1" {
			securityConfig.TwoFA.Crypto = crypto.SHA1
		}
		if cryptoAlg == "256" {
			securityConfig.TwoFA.Crypto = crypto.SHA256
		}
		if cryptoAlg == "512" {
			securityConfig.TwoFA.Crypto = crypto.SHA512
		}

		digits, err := strconv.Atoi(os.Getenv("TWO_FA_DIGITS"))
		if err != nil {
			log.WithError(err).Panic("panic code: 130")
		}
		securityConfig.TwoFA.Digits = digits

		// define different statuses of individual user
		securityConfig.TwoFA.Status.Verified = os.Getenv("TWO_FA_VERIFIED")
		securityConfig.TwoFA.Status.On = os.Getenv("TWO_FA_ON")
		securityConfig.TwoFA.Status.Off = os.Getenv("TWO_FA_OFF")
		securityConfig.TwoFA.Status.Invalid = os.Getenv("TWO_FA_INVALID")

		// for saving QR temporarily
		securityConfig.TwoFA.PathQR = strings.TrimSpace(os.Getenv("TWO_FA_QR_PATH"))

		if securityConfig.TwoFA.PathQR != "" {
			// verify directory exists
			if _, err := os.Stat(securityConfig.TwoFA.PathQR); os.IsNotExist(err) {
				// directory does not exist, create the directory
				path := filepath.Join(".", securityConfig.TwoFA.PathQR)
				err := os.MkdirAll(path, os.ModePerm)
				if err != nil {
					log.WithError(err).Panic("panic code: 109")
				}
			}
		}
	}

	// App firewall
	securityConfig.MustFW = os.Getenv("ACTIVATE_FIREWALL")
	if securityConfig.MustFW == Activated {
		securityConfig.Firewall.ListType = os.Getenv("LISTTYPE")
		securityConfig.Firewall.IP = os.Getenv("IP")
	}

	// CORS
	securityConfig.MustCORS = os.Getenv("ACTIVATE_CORS")
	if securityConfig.MustCORS == Activated {
		securityConfig.CORS.Origin = os.Getenv("CORS_ORIGIN")
		securityConfig.CORS.Credentials = os.Getenv("CORS_CREDENTIALS")
		securityConfig.CORS.Headers = os.Getenv("CORS_HEADERS")
		securityConfig.CORS.Methods = os.Getenv("CORS_METHODS")
		securityConfig.CORS.MaxAge = os.Getenv("CORS_MAXAGE")
	}

	// Important for getting real client IP
	securityConfig.TrustedPlatform = os.Getenv("TRUSTED_PLATFORM")

	return securityConfig
}

// server - port and env
func server() ServerConfig {
	var serverConfig ServerConfig

	// Load environment variables
	env()

	serverConfig.ServerPort = os.Getenv("APP_PORT")
	serverConfig.ServerEnv = os.Getenv("APP_ENV")

	return serverConfig
}

// view - HTML renderer
func view() ViewConfig {
	var viewConfig ViewConfig

	// Load environment variables
	env()

	viewConfig.Activate = os.Getenv("ACTIVATE_VIEW")
	if viewConfig.Activate == Activated {
		viewConfig.Directory = strings.TrimSpace(os.Getenv("TEMPLATE_DIR"))

		if viewConfig.Directory != "" {
			// verify directory for templates exists
			if _, err := os.Stat(viewConfig.Directory); os.IsNotExist(err) {
				// directory does not exist, create the directory
				path := filepath.Join(".", viewConfig.Directory)
				err := os.MkdirAll(path, os.ModePerm)
				if err != nil {
					log.WithError(err).Panic("panic code: 110")
				}
			}
		}
	}

	return viewConfig
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
