// Package config is responsible for reading all environment
// variables and set up the base configuration for a
// functional application
package config

import (
	"crypto"
	"crypto/sha256"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/sha3"

	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/lib/middleware"
)

// Activated - "yes" keyword to activate a service
const Activated string = "yes"

// PrefixJtiBlacklist - to manage JWT blacklist in Redis database
const PrefixJtiBlacklist string = "gorest-blacklist-jti:"

// Configuration - server and db configuration variables
type Configuration struct {
	Version    string
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
	// load environment variables
	err = Env()
	if err != nil {
		return
	}

	var configuration Configuration

	configuration.Version = strings.TrimSpace(os.Getenv("RELEASE_VERSION_OR_COMMIT_NUMBER"))

	configuration.Database, err = database()
	if err != nil {
		return
	}
	configuration.EmailConf, err = email()
	if err != nil {
		return
	}
	configuration.Logger = logger()

	configuration.Security, err = security()
	if err != nil {
		return
	}
	configuration.Server = server()

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
	// RDBMS
	activateRDBMS := strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_RDBMS")))
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
	activateRedis := strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_REDIS")))
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
	activateMongo := strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_MONGO")))
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
	// Env
	databaseConfig.RDBMS.Env.Driver = strings.ToLower(strings.TrimSpace(os.Getenv("DBDRIVER")))
	databaseConfig.RDBMS.Env.Host = strings.TrimSpace(os.Getenv("DBHOST"))
	databaseConfig.RDBMS.Env.Port = strings.TrimSpace(os.Getenv("DBPORT"))
	databaseConfig.RDBMS.Env.TimeZone = strings.TrimSpace(os.Getenv("DBTIMEZONE"))
	// Access
	databaseConfig.RDBMS.Access.DbName = strings.TrimSpace(os.Getenv("DBNAME"))
	databaseConfig.RDBMS.Access.User = strings.TrimSpace(os.Getenv("DBUSER"))
	databaseConfig.RDBMS.Access.Pass = strings.TrimSpace(os.Getenv("DBPASS"))
	// SSL
	databaseConfig.RDBMS.Ssl.Sslmode = strings.TrimSpace(os.Getenv("DBSSLMODE"))
	databaseConfig.RDBMS.Ssl.MinTLS = strings.TrimSpace(os.Getenv("DBSSL_TLS_MIN"))
	databaseConfig.RDBMS.Ssl.RootCA = strings.TrimSpace(os.Getenv("DBSSL_ROOT_CA"))
	databaseConfig.RDBMS.Ssl.ServerCert = strings.TrimSpace(os.Getenv("DBSSL_SERVER_CERT"))
	databaseConfig.RDBMS.Ssl.ClientCert = strings.TrimSpace(os.Getenv("DBSSL_CLIENT_CERT"))
	databaseConfig.RDBMS.Ssl.ClientKey = strings.TrimSpace(os.Getenv("DBSSL_CLIENT_KEY"))
	// Conn
	dbMaxIdleConns := strings.TrimSpace(os.Getenv("DBMAXIDLECONNS"))
	dbMaxOpenConns := strings.TrimSpace(os.Getenv("DBMAXOPENCONNS"))
	dbConnMaxLifetime := strings.TrimSpace(os.Getenv("DBCONNMAXLIFETIME"))
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
	dbLogLevel := strings.TrimSpace(os.Getenv("DBLOGLEVEL"))
	databaseConfig.RDBMS.Log.LogLevel, err = strconv.Atoi(dbLogLevel)
	if err != nil {
		return
	}

	return
}

// databaseRedis - all REDIS DB variables
func databaseRedis() (databaseConfig DatabaseConfig, err error) {
	// REDIS
	poolSize, errThis := strconv.Atoi(strings.TrimSpace(os.Getenv("POOLSIZE")))
	if errThis != nil {
		err = errThis
		return
	}
	connTTL, errThis := strconv.Atoi(strings.TrimSpace(os.Getenv("CONNTTL")))
	if errThis != nil {
		err = errThis
		return
	}

	databaseConfig.REDIS.Env.Host = strings.TrimSpace(os.Getenv("REDISHOST"))
	databaseConfig.REDIS.Env.Port = strings.TrimSpace(os.Getenv("REDISPORT"))
	databaseConfig.REDIS.Conn.PoolSize = poolSize
	databaseConfig.REDIS.Conn.ConnTTL = connTTL

	return
}

// databaseMongo - all MongoDB variables
func databaseMongo() (databaseConfig DatabaseConfig, err error) {
	// MongoDB
	poolSize, errThis := strconv.ParseUint(strings.TrimSpace(os.Getenv("MONGO_POOLSIZE")), 10, 64)
	if errThis != nil {
		err = errThis
		return
	}
	connTTL, errThis := strconv.Atoi(strings.TrimSpace(os.Getenv("MONGO_CONNTTL")))
	if errThis != nil {
		err = errThis
		return
	}

	databaseConfig.MongoDB.Env.URI = strings.TrimSpace(os.Getenv("MONGO_URI"))
	databaseConfig.MongoDB.Env.AppName = strings.TrimSpace(os.Getenv("MONGO_APP"))
	databaseConfig.MongoDB.Env.PoolSize = poolSize
	databaseConfig.MongoDB.Env.PoolMon = strings.TrimSpace(os.Getenv("MONGO_MONITOR_POOL"))
	databaseConfig.MongoDB.Env.ConnTTL = connTTL

	return
}

// email - config for using external email services
func email() (emailConfig EmailConfig, err error) {
	emailConfig.Activate = strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_EMAIL_SERVICE")))
	if emailConfig.Activate == Activated {
		emailConfig.Provider = strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_SERVICE_PROVIDER")))
		emailConfig.APIToken = strings.TrimSpace(os.Getenv("EMAIL_API_TOKEN"))
		emailConfig.AddrFrom = strings.TrimSpace(os.Getenv("EMAIL_FROM"))

		emailConfig.TrackOpens = false
		trackOpens := strings.TrimSpace(os.Getenv("EMAIL_TRACK_OPENS"))
		if trackOpens == Activated {
			emailConfig.TrackOpens = true
		}

		emailConfig.TrackLinks = strings.TrimSpace(os.Getenv("EMAIL_TRACK_LINKS"))
		emailConfig.DeliveryType = strings.TrimSpace(os.Getenv("EMAIL_DELIVERY_TYPE"))

		emailConfig.EmailVerificationTemplateID, err = strconv.ParseInt(strings.TrimSpace(os.Getenv("EMAIL_VERIFY_TEMPLATE_ID")), 10, 64)
		if err != nil {
			return
		}
		emailConfig.PasswordRecoverTemplateID, err = strconv.ParseInt(strings.TrimSpace(os.Getenv("EMAIL_PASS_RECOVER_TEMPLATE_ID")), 10, 64)
		if err != nil {
			return
		}
		emailConfig.EmailUpdateVerifyTemplateID, err = strconv.ParseInt(strings.TrimSpace(os.Getenv("EMAIL_UPDATE_VERIFY_TEMPLATE_ID")), 10, 64)
		if err != nil {
			return
		}

		useUUIDv4EmailVerificationCode := strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_VERIFY_USE_UUIDv4")))
		if useUUIDv4EmailVerificationCode == Activated {
			emailConfig.EmailVerificationCodeUUIDv4 = true
		} else {
			emailConfig.EmailVerificationCodeLength, err = strconv.ParseUint(strings.TrimSpace(os.Getenv("EMAIL_VERIFY_CODE_LENGTH")), 10, 32)
			if err != nil {
				return
			}
		}
		useUUIDv4PasswordRecoverCode := strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_PASS_RECOVER_USE_UUIDv4")))
		if useUUIDv4PasswordRecoverCode == Activated {
			emailConfig.PasswordRecoverCodeUUIDv4 = true
		} else {
			emailConfig.PasswordRecoverCodeLength, err = strconv.ParseUint(strings.TrimSpace(os.Getenv("EMAIL_PASS_RECOVER_CODE_LENGTH")), 10, 32)
			if err != nil {
				return
			}
		}
		emailConfig.EmailVerificationTag = strings.TrimSpace(os.Getenv("EMAIL_VERIFY_TAG"))
		emailConfig.PasswordRecoverTag = strings.TrimSpace(os.Getenv("EMAIL_PASS_RECOVER_TAG"))
		emailConfig.HTMLModel = strings.TrimSpace(os.Getenv("EMAIL_HTML_MODEL"))
		emailConfig.EmailVerifyValidityPeriod, err = strconv.ParseUint(strings.TrimSpace(os.Getenv("EMAIL_VERIFY_VALIDITY_PERIOD")), 10, 32)
		if err != nil {
			return
		}
		emailConfig.PassRecoverValidityPeriod, err = strconv.ParseUint(strings.TrimSpace(os.Getenv("EMAIL_PASS_RECOVER_VALIDITY_PERIOD")), 10, 32)
		if err != nil {
			return
		}
	}

	return
}

// logger - config for sentry.io
func logger() (loggerConfig LoggerConfig) {
	loggerConfig.Activate = strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_SENTRY")))
	if loggerConfig.Activate == Activated {
		loggerConfig.SentryDsn = strings.TrimSpace(os.Getenv("SentryDSN"))
		loggerConfig.PerformanceTracing = strings.ToLower(strings.TrimSpace(os.Getenv("SENTRY_ENABLE_TRACING")))
		loggerConfig.TracesSampleRate = strings.TrimSpace(os.Getenv("SENTRY_TRACES_SAMPLE_RATE"))
	}

	return
}

// security - configs for generating tokens and hashes
func security() (securityConfig SecurityConfig, err error) {
	// Minimum password length
	minPassLength := strings.TrimSpace(os.Getenv("MIN_PASS_LENGTH"))
	if minPassLength != "" {
		userPassMinLength, errThis := strconv.Atoi(minPassLength)
		if errThis != nil {
			err = errThis
			return
		}
		securityConfig.UserPassMinLength = userPassMinLength
	}

	// Basic auth
	securityConfig.MustBasicAuth = strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_BASIC_AUTH")))
	if securityConfig.MustBasicAuth == Activated {
		securityConfig.BasicAuth.Username = strings.TrimSpace(os.Getenv("USERNAME"))
		securityConfig.BasicAuth.Password = strings.TrimSpace(os.Getenv("PASSWORD"))
	}

	// JWT
	securityConfig.MustJWT = strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_JWT")))
	if securityConfig.MustJWT == Activated {
		securityConfig.JWT, err = getParamsJWT()
		if err != nil {
			return
		}

		// set params globally
		setParamsJWT(securityConfig.JWT)
	}

	// When user logs off, invalidate the tokens
	securityConfig.InvalidateJWT = strings.ToLower(strings.TrimSpace(os.Getenv("INVALIDATE_JWT")))

	// Cookie for authentication and authorization
	authCookieActivate := strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_COOKIE_ACTIVATE")))
	if authCookieActivate == Activated {
		securityConfig.AuthCookieActivate = true
		securityConfig.AuthCookiePath = strings.TrimSpace(os.Getenv("AUTH_COOKIE_PATH"))
		securityConfig.AuthCookieDomain = strings.TrimSpace(os.Getenv("AUTH_COOKIE_DOMAIN"))

		if strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_COOKIE_SECURE"))) == Activated {
			securityConfig.AuthCookieSecure = true
		}

		if strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_COOKIE_HttpOnly"))) == Activated {
			securityConfig.AuthCookieHTTPOnly = true
		}

		authCookieSameSite := strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_COOKIE_SameSite")))
		if len(authCookieSameSite) > 0 {
			if authCookieSameSite == "strict" {
				securityConfig.AuthCookieSameSite = http.SameSiteStrictMode
			}
			if authCookieSameSite == "lax" {
				securityConfig.AuthCookieSameSite = http.SameSiteLaxMode
			}
			if authCookieSameSite == "none" {
				securityConfig.AuthCookieSameSite = http.SameSiteNoneMode
			}
		}

		if strings.ToLower(strings.TrimSpace(os.Getenv("SERVE_JWT_AS_RESPONSE_BODY"))) != "no" {
			securityConfig.ServeJwtAsResBody = true
		}
	}

	// Hashing passwords
	securityConfig.MustHash = strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_HASHING")))
	if securityConfig.MustHash == Activated {
		securityConfig.HashPass, err = getParamsHash()
		if err != nil {
			return
		}
		securityConfig.HashSec = strings.TrimSpace(os.Getenv("HASH_SECRET"))
	}

	// config for ChaCha20-Poly1305 encryption
	activateCipher := strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_CIPHER")))
	if activateCipher == Activated {
		securityConfig.MustCipher = true

		cipherKey := strings.TrimSpace(os.Getenv("CIPHER_KEY"))
		if cipherKey == "" {
			err = errors.New("CIPHER_KEY is missing")
			return
		}
		cipherKeyHash2 := sha256.Sum256([]byte(cipherKey)) // sha2-256
		cipherKeyHash3 := sha3.Sum256(cipherKeyHash2[:])   // sha3-256
		securityConfig.CipherKey = cipherKeyHash3[:]

	}

	// config for blake2b hashing
	blake2bSec := strings.TrimSpace(os.Getenv("BLAKE2B_SECRET"))
	if blake2bSec == "" {
		securityConfig.Blake2bSec = nil
	} else {
		securityConfig.Blake2bSec = []byte(blake2bSec)
	}

	// Email verification and password recovery
	securityConfig.VerifyEmail = false
	securityConfig.RecoverPass = false
	if strings.ToLower(strings.TrimSpace(os.Getenv("VERIFY_EMAIL"))) == Activated {
		securityConfig.VerifyEmail = true
	}
	if strings.ToLower(strings.TrimSpace(os.Getenv("RECOVER_PASSWORD"))) == Activated {
		securityConfig.RecoverPass = true
	}

	// Two-factor authentication
	securityConfig.Must2FA = strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_2FA")))
	if securityConfig.Must2FA == Activated {
		securityConfig.TwoFA.Issuer = strings.TrimSpace(os.Getenv("TWO_FA_ISSUER"))

		cryptoAlg := strings.TrimSpace(os.Getenv("TWO_FA_CRYPTO"))
		if cryptoAlg == "1" {
			securityConfig.TwoFA.Crypto = crypto.SHA1
		}
		if cryptoAlg == "256" {
			securityConfig.TwoFA.Crypto = crypto.SHA256
		}
		if cryptoAlg == "512" {
			securityConfig.TwoFA.Crypto = crypto.SHA512
		}

		digits, errThis := strconv.Atoi(strings.TrimSpace(os.Getenv("TWO_FA_DIGITS")))
		if errThis != nil {
			err = errThis
			return
		}
		securityConfig.TwoFA.Digits = digits

		// define different statuses of individual user
		securityConfig.TwoFA.Status.Verified = strings.TrimSpace(os.Getenv("TWO_FA_VERIFIED"))
		securityConfig.TwoFA.Status.On = strings.TrimSpace(os.Getenv("TWO_FA_ON"))
		securityConfig.TwoFA.Status.Off = strings.TrimSpace(os.Getenv("TWO_FA_OFF"))
		securityConfig.TwoFA.Status.Invalid = strings.TrimSpace(os.Getenv("TWO_FA_INVALID"))

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

		// false: sha2_256()
		// true: blake2b(sha2_256())
		doubleHashTwoFA := strings.ToLower(strings.TrimSpace(os.Getenv("TWO_FA_DOUBLE_HASH")))
		if doubleHashTwoFA == Activated {
			securityConfig.TwoFA.DoubleHash = true
		}
	}

	// App firewall
	securityConfig.MustFW = strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_FIREWALL")))
	if securityConfig.MustFW == Activated {
		securityConfig.Firewall.ListType = strings.TrimSpace(os.Getenv("LISTTYPE"))
		securityConfig.Firewall.IP = strings.TrimSpace(os.Getenv("IP"))
	}

	// CORS
	securityConfig.MustCORS = strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_CORS")))
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

	// Validate origin of the request
	securityConfig.CheckOrigin = strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_ORIGIN_VALIDATION")))

	// IP-based rate limiter
	securityConfig.RateLimit = strings.ToUpper(strings.TrimSpace(os.Getenv("RATE_LIMIT")))

	// Important for getting real client IP
	securityConfig.TrustedPlatform = strings.TrimSpace(os.Getenv("TRUSTED_PLATFORM"))

	return
}

// server - port and env
func server() (serverConfig ServerConfig) {
	serverConfig.ServerHost = strings.TrimSpace(os.Getenv("APP_HOST"))
	serverConfig.ServerPort = strings.TrimSpace(os.Getenv("APP_PORT"))
	serverConfig.ServerEnv = strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))

	return
}

// view - HTML renderer
func view() (viewConfig ViewConfig, err error) {
	viewConfig.Activate = strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVATE_VIEW")))
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
	alg := strings.TrimSpace(os.Getenv("JWT_ALG"))
	if alg == "" {
		alg = "HS256" // default algorithm
	}
	// list of accepted algorithms
	// HS256: HMAC-SHA256
	// HS384: HMAC-SHA384
	// HS512: HMAC-SHA512
	// ES256: ECDSA Signature with SHA-256
	// ES384: ECDSA Signature with SHA-384
	// ES512: ECDSA Signature with SHA-512
	// RS256: RSA Signature with SHA-256
	// RS384: RSA Signature with SHA-384
	// RS512: RSA Signature with SHA-512
	if alg != "HS256" && alg != "HS384" && alg != "HS512" &&
		alg != "ES256" && alg != "ES384" && alg != "ES512" &&
		alg != "RS256" && alg != "RS384" && alg != "RS512" {
		err = errors.New("unsupported algorithm for JWT")
		return
	}
	params.Algorithm = alg
	params.AccessKey = []byte(strings.TrimSpace(os.Getenv("ACCESS_KEY")))
	params.AccessKeyTTL, err = strconv.Atoi(strings.TrimSpace(os.Getenv("ACCESS_KEY_TTL")))
	if err != nil {
		return
	}
	params.RefreshKey = []byte(strings.TrimSpace(os.Getenv("REFRESH_KEY")))
	params.RefreshKeyTTL, err = strconv.Atoi(strings.TrimSpace(os.Getenv("REFRESH_KEY_TTL")))
	if err != nil {
		return
	}

	privateKeyFile := strings.TrimSpace(os.Getenv("PRIV_KEY_FILE_PATH"))
	if privateKeyFile != "" {
		// load the private key
		privateKeyBytes, errThis := os.ReadFile(privateKeyFile)
		if errThis != nil {
			err = errThis
			return
		}

		// ECDSA
		if alg == "ES256" || alg == "ES384" || alg == "ES512" {
			privateKey, errThis := jwt.ParseECPrivateKeyFromPEM(privateKeyBytes)
			if errThis != nil {
				err = errThis
				return
			}
			params.PrivKeyECDSA = privateKey
		}

		// RSA
		if alg == "RS256" || alg == "RS384" || alg == "RS512" {
			privateKey, errThis := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
			if errThis != nil {
				err = errThis
				return
			}
			params.PrivKeyRSA = privateKey
		}
	}

	publicKeyFile := strings.TrimSpace(os.Getenv("PUB_KEY_FILE_PATH"))
	if publicKeyFile != "" {
		// load the public key
		publicKeyBytes, errThis := os.ReadFile(publicKeyFile)
		if errThis != nil {
			err = errThis
			return
		}

		// ECDSA
		if alg == "ES256" || alg == "ES384" || alg == "ES512" {
			publicKey, errThis := jwt.ParseECPublicKeyFromPEM(publicKeyBytes)
			if errThis != nil {
				err = errThis
				return
			}
			params.PubKeyECDSA = publicKey
		}

		// RSA
		if alg == "RS256" || alg == "RS384" || alg == "RS512" {
			publicKey, errThis := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
			if errThis != nil {
				err = errThis
				return
			}
			params.PubKeyRSA = publicKey
		}
	}

	params.Audience = strings.TrimSpace(os.Getenv("AUDIENCE"))
	params.Issuer = strings.TrimSpace(os.Getenv("ISSUER"))
	params.AccNbf, err = strconv.Atoi(strings.TrimSpace(os.Getenv("NOT_BEFORE_ACC")))
	if err != nil {
		return
	}
	params.RefNbf, err = strconv.Atoi(strings.TrimSpace(os.Getenv("NOT_BEFORE_REF")))
	if err != nil {
		return
	}
	params.Subject = strings.TrimSpace(os.Getenv("SUBJECT"))

	return
}

// setParamsJWT - set parameters for JWT
func setParamsJWT(c middleware.JWTParameters) {
	middleware.JWTParams.Algorithm = c.Algorithm
	middleware.JWTParams.AccessKey = c.AccessKey
	middleware.JWTParams.AccessKeyTTL = c.AccessKeyTTL
	middleware.JWTParams.RefreshKey = c.RefreshKey
	middleware.JWTParams.RefreshKeyTTL = c.RefreshKeyTTL
	middleware.JWTParams.PrivKeyECDSA = c.PrivKeyECDSA
	middleware.JWTParams.PubKeyECDSA = c.PubKeyECDSA
	middleware.JWTParams.PrivKeyRSA = c.PrivKeyRSA
	middleware.JWTParams.PubKeyRSA = c.PubKeyRSA

	middleware.JWTParams.Audience = c.Audience
	middleware.JWTParams.Issuer = c.Issuer
	middleware.JWTParams.AccNbf = c.AccNbf
	middleware.JWTParams.RefNbf = c.RefNbf
	middleware.JWTParams.Subject = c.Subject
}

// getParamsHash - read parameters from env
func getParamsHash() (params lib.HashPassConfig, err error) {
	hashPassMemory64, errThis := strconv.ParseUint((strings.TrimSpace(os.Getenv("HASHPASSMEMORY"))), 10, 32)
	if errThis != nil {
		err = errThis
		return
	}
	hashPassIterations64, errThis := strconv.ParseUint((strings.TrimSpace(os.Getenv("HASHPASSITERATIONS"))), 10, 32)
	if errThis != nil {
		err = errThis
		return
	}
	hashPassParallelism64, errThis := strconv.ParseUint((strings.TrimSpace(os.Getenv("HASHPASSPARALLELISM"))), 10, 8)
	if errThis != nil {
		err = errThis
		return
	}
	hashPassSaltLength64, errThis := strconv.ParseUint((strings.TrimSpace(os.Getenv("HASHPASSSALTLENGTH"))), 10, 32)
	if errThis != nil {
		err = errThis
		return
	}
	hashPassKeyLength64, errThis := strconv.ParseUint((strings.TrimSpace(os.Getenv("HASHPASSKEYLENGTH"))), 10, 32)
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
