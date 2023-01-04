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

// Env - load the configurations from .env
func Env() error {
	// Load environment variables
	return godotenv.Load()
}

// Config - load all the configurations
func Config() (err error) {
	var configuration Configuration

	configuration.Database, err = database()
	if err != nil {
		return
	}
	configuration.EmailConf, err = email()
	if err != nil {
		return
	}
	configuration.Logger, err = logger()
	if err != nil {
		return
	}
	configuration.Security, err = security()
	if err != nil {
		return
	}
	configuration.Server, err = server()
	if err != nil {
		return
	}
	configuration.ViewConfig, err = view()
	if err != nil {
		return
	}

	configAll = &configuration

	return
}

// GetConfig - return all the config variables
func GetConfig() *Configuration {
	return configAll
}

// database - all DB variables
func database() (databaseConfig DatabaseConfig, err error) {
	// Load environment variables
	err = Env()
	if err != nil {
		return
	}

	// RDBMS
	activateRDBMS := os.Getenv("ACTIVATE_RDBMS")
	if activateRDBMS == Activated {
		dbRDBMS, errThis := databaseRDBMS()
		if errThis != nil {
			err = errThis
			return
		}
		databaseConfig.RDBMS = dbRDBMS.RDBMS
	}
	databaseConfig.RDBMS.Activate = activateRDBMS

	// REDIS
	activateRedis := os.Getenv("ACTIVATE_REDIS")
	if activateRedis == Activated {
		dbRedis, errThis := databaseRedis()
		if errThis != nil {
			err = errThis
			return
		}
		databaseConfig.REDIS = dbRedis.REDIS
	}
	databaseConfig.REDIS.Activate = activateRedis

	// MongoDB
	activateMongo := os.Getenv("ACTIVATE_MONGO")
	if activateMongo == Activated {
		dbMongo, errThis := databaseMongo()
		if errThis != nil {
			err = errThis
			return
		}
		databaseConfig.MongoDB = dbMongo.MongoDB
	}
	databaseConfig.MongoDB.Activate = activateMongo

	return
}

// databaseRDBMS - all RDBMS variables
func databaseRDBMS() (databaseConfig DatabaseConfig, err error) {
	// Load environment variables
	err = Env()
	if err != nil {
		return
	}

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
		return
	}
	databaseConfig.RDBMS.Conn.MaxOpenConns, err = strconv.Atoi(dbMaxOpenConns)
	if err != nil {
		return
	}
	databaseConfig.RDBMS.Conn.ConnMaxLifetime, err = time.ParseDuration(dbConnMaxLifetime)
	if err != nil {
		return
	}

	// Logger
	dbLogLevel := os.Getenv("DBLOGLEVEL")
	databaseConfig.RDBMS.Log.LogLevel, err = strconv.Atoi(dbLogLevel)
	if err != nil {
		return
	}

	return
}

// databaseRedis - all REDIS DB variables
func databaseRedis() (databaseConfig DatabaseConfig, err error) {
	// Load environment variables
	err = Env()
	if err != nil {
		return
	}

	// REDIS
	poolSize, errThis := strconv.Atoi(os.Getenv("POOLSIZE"))
	if errThis != nil {
		err = errThis
		return
	}
	connTTL, errThis := strconv.Atoi(os.Getenv("CONNTTL"))
	if errThis != nil {
		err = errThis
		return
	}

	databaseConfig.REDIS.Env.Host = os.Getenv("REDISHOST")
	databaseConfig.REDIS.Env.Port = os.Getenv("REDISPORT")
	databaseConfig.REDIS.Conn.PoolSize = poolSize
	databaseConfig.REDIS.Conn.ConnTTL = connTTL

	return
}

// databaseMongo - all MongoDB variables
func databaseMongo() (databaseConfig DatabaseConfig, err error) {
	// Load environment variables
	err = Env()
	if err != nil {
		return
	}

	// MongoDB
	poolSize, errThis := strconv.ParseUint(os.Getenv("MONGO_POOLSIZE"), 10, 64)
	if errThis != nil {
		err = errThis
		return
	}
	connTTL, errThis := strconv.Atoi(os.Getenv("MONGO_CONNTTL"))
	if errThis != nil {
		err = errThis
		return
	}

	databaseConfig.MongoDB.Env.URI = os.Getenv("MONGO_URI")
	databaseConfig.MongoDB.Env.AppName = os.Getenv("MONGO_APP")
	databaseConfig.MongoDB.Env.PoolSize = poolSize
	databaseConfig.MongoDB.Env.PoolMon = os.Getenv("MONGO_MONITOR_POOL")
	databaseConfig.MongoDB.Env.ConnTTL = connTTL

	return
}

// email - config for using external email services
func email() (emailConfig EmailConfig, err error) {
	// Load environment variables
	err = Env()
	if err != nil {
		return
	}

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
			return
		}
		emailConfig.PasswordRecoverTemplateID, err = strconv.ParseInt(os.Getenv("EMAIL_PASS_RECOVER_TEMPLATE_ID"), 10, 64)
		if err != nil {
			return
		}
		emailConfig.EmailVerificationCodeLength, err = strconv.ParseUint(os.Getenv("EMAIL_VERIFY_CODE_LENGTH"), 10, 32)
		if err != nil {
			return
		}
		emailConfig.PasswordRecoverCodeLength, err = strconv.ParseUint(os.Getenv("EMAIL_PASS_RECOVER_CODE_LENGTH"), 10, 32)
		if err != nil {
			return
		}
		emailConfig.EmailVerificationTag = os.Getenv("EMAIL_VERIFY_TAG")
		emailConfig.PasswordRecoverTag = os.Getenv("EMAIL_PASS_RECOVER_TAG")
		emailConfig.HTMLModel = os.Getenv("EMAIL_HTML_MODEL")
		emailConfig.EmailVerifyValidityPeriod, err = strconv.ParseUint(os.Getenv("EMAIL_VERIFY_VALIDITY_PERIOD"), 10, 32)
		if err != nil {
			return
		}
		emailConfig.PassRecoverValidityPeriod, err = strconv.ParseUint(os.Getenv("EMAIL_PASS_RECOVER_VALIDITY_PERIOD"), 10, 32)
		if err != nil {
			return
		}
	}

	return
}

// logger - config for sentry.io
func logger() (loggerConfig LoggerConfig, err error) {
	// Load environment variables
	err = Env()
	if err != nil {
		return
	}

	loggerConfig.Activate = os.Getenv("ACTIVATE_SENTRY")
	if loggerConfig.Activate == Activated {
		loggerConfig.SentryDsn = os.Getenv("SentryDSN")
	}

	return
}

// security - configs for generating tokens and hashes
func security() (securityConfig SecurityConfig, err error) {
	// Load environment variables
	err = Env()
	if err != nil {
		return
	}

	// Minimum password length
	userPassMinLength, errThis := strconv.Atoi(os.Getenv("MIN_PASS_LENGTH"))
	if errThis != nil {
		err = errThis
		return
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
		securityConfig.JWT, err = getParamsJWT()
		if err != nil {
			return
		}

		// set params globally
		setParamsJWT(securityConfig.JWT)
	}

	// Hashing passwords
	securityConfig.MustHash = os.Getenv("ACTIVATE_HASHING")
	if securityConfig.MustHash == Activated {
		securityConfig.HashPass, err = getParamsHash()
		if err != nil {
			return
		}
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

		digits, errThis := strconv.Atoi(os.Getenv("TWO_FA_DIGITS"))
		if errThis != nil {
			err = errThis
			return
		}
		securityConfig.TwoFA.Digits = digits

		// define different statuses of individual user
		securityConfig.TwoFA.Status.Verified = os.Getenv("TWO_FA_VERIFIED")
		securityConfig.TwoFA.Status.On = os.Getenv("TWO_FA_ON")
		securityConfig.TwoFA.Status.Off = os.Getenv("TWO_FA_OFF")
		securityConfig.TwoFA.Status.Invalid = os.Getenv("TWO_FA_INVALID")

		// for saving QR temporarily
		securityConfig.TwoFA.PathQR = strings.TrimRight(strings.TrimSpace(os.Getenv("TWO_FA_QR_PATH")), "/")

		if securityConfig.TwoFA.PathQR != "" {
			// verify directory exists
			if _, errThis = os.Stat(securityConfig.TwoFA.PathQR); os.IsNotExist(errThis) {
				// directory does not exist, create the directory
				path := filepath.Join(".", securityConfig.TwoFA.PathQR)
				err = os.MkdirAll(path, os.ModePerm)
				if err != nil {
					return
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
		cp := middleware.CORSPolicy{}

		// Access-Control-Allow-Origin
		// Indicates whether the response can be shared with requesting code from the given origin
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
		cp.Value = strings.TrimSpace(os.Getenv("CORS_ORIGIN"))
		if cp.Value != "" {
			cp.Key = "Access-Control-Allow-Origin"
			securityConfig.CORS = append(securityConfig.CORS, cp)
		}

		// Access-Control-Allow-Credentials
		// Indicates whether or not the actual request can be made using credentials
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
		cp.Value = strings.TrimSpace(os.Getenv("CORS_CREDENTIALS"))
		if cp.Value != "" {
			cp.Key = "Access-Control-Allow-Credentials"
			securityConfig.CORS = append(securityConfig.CORS, cp)
		}

		// Access-Control-Allow-Headers
		// Indicate which HTTP headers can be used during the actual request
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
		cp.Value = strings.TrimSpace(os.Getenv("CORS_HEADERS"))
		if cp.Value != "" {
			cp.Key = "Access-Control-Allow-Headers"
			securityConfig.CORS = append(securityConfig.CORS, cp)
		}

		// Access-Control-Expose-Headers
		// Which response headers should be made available to scripts running in the browser
		// in response to a cross-origin request
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Expose-Headers
		cp.Value = strings.TrimSpace(os.Getenv("CORS_EXPOSE_HEADERS"))
		if cp.Value != "" {
			cp.Key = "Access-Control-Expose-Headers"
			securityConfig.CORS = append(securityConfig.CORS, cp)
		}

		// Access-Control-Allow-Methods
		// Specifies one or more allowed methods
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
		cp.Value = strings.TrimSpace(os.Getenv("CORS_METHODS"))
		if cp.Value != "" {
			cp.Key = "Access-Control-Allow-Methods"
			securityConfig.CORS = append(securityConfig.CORS, cp)
		}

		// Access-Control-Max-Age
		// Indicates how long the results of a preflight request can be cached
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age
		cp.Value = strings.TrimSpace(os.Getenv("CORS_MAXAGE"))
		if cp.Value != "" {
			cp.Key = "Access-Control-Max-Age"
			securityConfig.CORS = append(securityConfig.CORS, cp)
		}

		// X-Content-Type-Options
		// Prevent some browsers from MIME-sniffing the response
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
		cp.Value = strings.TrimSpace(os.Getenv("CORS_X_CONTENT_TYPE"))
		if cp.Value != "" {
			cp.Key = "X-Content-Type-Options"
			securityConfig.CORS = append(securityConfig.CORS, cp)
		}

		// X-Frame-Options
		// Protect website against clickjacking
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options
		// https://tools.ietf.org/html/rfc7034#section-2.1
		// X-Frame-Options: DENY, SAMEORIGIN
		cp.Value = strings.TrimSpace(os.Getenv("CORS_X_FRAME"))
		if cp.Value != "" {
			cp.Key = "X-Frame-Options"
			securityConfig.CORS = append(securityConfig.CORS, cp)
		}

		// Referrer-Policy
		// Set a strict Referrer Policy to mitigate information leakage
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referrer-Policy
		cp.Value = strings.TrimSpace(os.Getenv("CORS_REFERRER"))
		if cp.Value != "" {
			cp.Key = "Referrer-Policy"
			securityConfig.CORS = append(securityConfig.CORS, cp)
		}

		// Content-Security-Policy
		// Mitigate the risk of cross-site scripting and other content-injection attacks
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
		// https://content-security-policy.com/
		// https://developers.google.com/web/fundamentals/security/csp
		cp.Value = strings.TrimSpace(os.Getenv("CORS_CONTENT_SECURITY"))
		if cp.Value != "" {
			cp.Key = "Content-Security-Policy"
			securityConfig.CORS = append(securityConfig.CORS, cp)
		}

		// Timing-Allow-Origin
		// Allow cross-origin access to the timing information for all resources
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Timing-Allow-Origin
		cp.Value = strings.TrimSpace(os.Getenv("CORS_TIMING_ALLOW_ORIGIN"))
		if cp.Value != "" {
			cp.Key = "Timing-Allow-Origin"
			securityConfig.CORS = append(securityConfig.CORS, cp)
		}

		// Strict-Transport-Security
		// HTTP Strict Transport Security (HSTS)
		// https://tools.ietf.org/html/rfc6797#section-6.1
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security
		// Strict-Transport-Security: max-age=63072000; includeSubDomains
		// To enable HSTS preload inclusion: https://hstspreload.org/#deployment-recommendations
		// Strict-Transport-Security: max-age=63072000; includeSubDomains; preload
		cp.Value = strings.TrimSpace(os.Getenv("CORS_HSTS"))
		if cp.Value != "" {
			cp.Key = "Strict-Transport-Security"
			securityConfig.CORS = append(securityConfig.CORS, cp)
		}
	}

	// Important for getting real client IP
	securityConfig.TrustedPlatform = os.Getenv("TRUSTED_PLATFORM")

	return
}

// server - port and env
func server() (serverConfig ServerConfig, err error) {
	// Load environment variables
	err = Env()
	if err != nil {
		return
	}

	serverConfig.ServerPort = os.Getenv("APP_PORT")
	serverConfig.ServerEnv = os.Getenv("APP_ENV")

	return
}

// view - HTML renderer
func view() (viewConfig ViewConfig, err error) {
	// Load environment variables
	err = Env()
	if err != nil {
		return
	}

	viewConfig.Activate = os.Getenv("ACTIVATE_VIEW")
	if viewConfig.Activate == Activated {
		viewConfig.Directory = strings.TrimRight(strings.TrimSpace(os.Getenv("TEMPLATE_DIR")), "/")

		if viewConfig.Directory != "" {
			// verify directory for templates exists
			if _, errThis := os.Stat(viewConfig.Directory); os.IsNotExist(errThis) {
				// directory does not exist, create the directory
				path := filepath.Join(".", viewConfig.Directory)
				err = os.MkdirAll(path, os.ModePerm)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

// getParamsJWT - read parameters from env
func getParamsJWT() (params middleware.JWTParameters, err error) {
	err = Env()
	if err != nil {
		return
	}

	params.AccessKey = []byte(os.Getenv("ACCESS_KEY"))
	params.AccessKeyTTL, err = strconv.Atoi(os.Getenv("ACCESS_KEY_TTL"))
	if err != nil {
		return
	}
	params.RefreshKey = []byte(os.Getenv("REFRESH_KEY"))
	params.RefreshKeyTTL, err = strconv.Atoi(os.Getenv("REFRESH_KEY_TTL"))
	if err != nil {
		return
	}
	params.Audience = os.Getenv("AUDIENCE")
	params.Issuer = os.Getenv("ISSUER")
	params.AccNbf, err = strconv.Atoi(os.Getenv("NOT_BEFORE_ACC"))
	if err != nil {
		return
	}
	params.RefNbf, err = strconv.Atoi(os.Getenv("NOT_BEFORE_REF"))
	if err != nil {
		return
	}
	params.Subject = os.Getenv("SUBJECT")

	return
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
func getParamsHash() (params lib.HashPassConfig, err error) {
	err = Env()
	if err != nil {
		return
	}

	hashPassMemory64, errThis := strconv.ParseUint((os.Getenv("HASHPASSMEMORY")), 10, 32)
	if errThis != nil {
		err = errThis
		return
	}
	hashPassIterations64, errThis := strconv.ParseUint((os.Getenv("HASHPASSITERATIONS")), 10, 32)
	if errThis != nil {
		err = errThis
		return
	}
	hashPassParallelism64, errThis := strconv.ParseUint((os.Getenv("HASHPASSPARALLELISM")), 10, 8)
	if errThis != nil {
		err = errThis
		return
	}
	hashPassSaltLength64, errThis := strconv.ParseUint((os.Getenv("HASHPASSSALTLENGTH")), 10, 32)
	if errThis != nil {
		err = errThis
		return
	}
	hashPassKeyLength64, errThis := strconv.ParseUint((os.Getenv("HASHPASSKEYLENGTH")), 10, 32)
	if errThis != nil {
		err = errThis
		return
	}

	params.Memory = uint32(hashPassMemory64)
	params.Iterations = uint32(hashPassIterations64)
	params.Parallelism = uint8(hashPassParallelism64)
	params.SaltLength = uint32(hashPassSaltLength64)
	params.KeyLength = uint32(hashPassKeyLength64)

	return
}
