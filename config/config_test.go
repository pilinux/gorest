package config_test

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"crypto/sha3"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/lib/middleware"
)

func TestEnv(t *testing.T) {
	// this should return error
	err := config.Env()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// download a file from a remote location and save it
	fileURL := strings.TrimSpace(os.Getenv("TEST_ENV_URL"))
	err = downloadFile(".env", fileURL)
	if err != nil {
		t.Error(err)
	}

	err = config.Env()
	if err != nil {
		t.Errorf("got error when calling config.Env(): %v", err)
		return
	}

	// remove the downloaded file at the end of the test
	err = os.RemoveAll(".env")
	if err != nil {
		t.Error(err)
	}
}

func TestConfig(t *testing.T) {
	// this should return error
	err := config.Config()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// download a file from a remote location and save it
	fileURL := strings.TrimSpace(os.Getenv("TEST_ENV_URL"))
	err = downloadFile(".env", fileURL)
	if err != nil {
		t.Error(err)
	}

	err = config.Config()
	if err != nil {
		t.Errorf("got error when calling config.Config(): %v", err)
		return
	}

	// remove the downloaded file at the end of the test
	err = os.RemoveAll(".env")
	if err != nil {
		t.Error(err)
	}
}

func TestGetConfig(t *testing.T) {
	configAll := config.GetConfig()

	expected := &config.Configuration{}

	if config.IsProd() {
		t.Errorf("expected IsProd() to return false, but got true")
	}
	expected.Version = "d591c29"

	expected.Database.RDBMS.Activate = config.Activated
	if !config.IsRDBMS() {
		t.Errorf("expected IsRDBMS() to return true, but got false")
	}
	expected.Database.RDBMS.Env.Driver = "mysql"
	expected.Database.RDBMS.Env.URI = ""
	expected.Database.RDBMS.Env.Host = "127.0.0.1"
	expected.Database.RDBMS.Env.Port = "3306"
	expected.Database.RDBMS.Env.TimeZone = "Europe/Berlin"
	expected.Database.RDBMS.Access.DbName = "test_database"
	expected.Database.RDBMS.Access.User = "test_user"
	expected.Database.RDBMS.Access.Pass = "test_password"
	expected.Database.RDBMS.Ssl.Sslmode = "disable"
	expected.Database.RDBMS.Ssl.MinTLS = "1.2"
	expected.Database.RDBMS.Ssl.RootCA = "/path/to/ca.pem"
	expected.Database.RDBMS.Ssl.ServerCert = "/path/to/server-cert.pem"
	expected.Database.RDBMS.Ssl.ClientCert = "/path/to/client-cert.pem"
	expected.Database.RDBMS.Ssl.ClientKey = "/path/to/client-key.pem"
	expected.Database.RDBMS.Conn.MaxIdleConns = 10
	expected.Database.RDBMS.Conn.MaxOpenConns = 100
	expected.Database.RDBMS.Conn.ConnMaxLifetime = time.Duration(1 * time.Hour)
	expected.Database.RDBMS.Log.LogLevel = 1

	expected.Database.REDIS.Activate = config.Activated
	if !config.IsRedis() {
		t.Errorf("expected IsRedis() to return true, but got false")
	}
	expected.Database.REDIS.Env.URI = ""
	expected.Database.REDIS.Env.Host = "127.0.0.1"
	expected.Database.REDIS.Env.Port = "6379"
	expected.Database.REDIS.Conn.PoolSize = 10
	expected.Database.REDIS.Conn.ConnTTL = 5

	expected.Database.MongoDB.Activate = config.Activated
	if !config.IsMongo() {
		t.Errorf("expected IsMongo() to return true, but got false")
	}
	expected.Database.MongoDB.Env.AppName = "gorest"
	expected.Database.MongoDB.Env.URI = "mongodb://user:pass@127.0.0.1:27017/?retryWrites=true&w=majority"
	expected.Database.MongoDB.Env.PoolSize = 50
	expected.Database.MongoDB.Env.PoolMon = "no"
	expected.Database.MongoDB.Env.ConnTTL = 10

	expected.EmailConf.Activate = config.Activated
	if !config.IsEmailService() {
		t.Errorf("expected IsEmailService() to return true, but got false")
	}
	expected.EmailConf.Provider = "postmark"
	expected.EmailConf.APIToken = "abcdef"
	expected.EmailConf.AddrFrom = "email@yourdomain.com"
	expected.EmailConf.TrackOpens = false
	expected.EmailConf.TrackLinks = "None"
	expected.EmailConf.DeliveryType = "outbound"

	expected.EmailConf.EmailVerificationTemplateID = 0
	expected.EmailConf.PasswordRecoverTemplateID = 0
	expected.EmailConf.EmailUpdateVerifyTemplateID = 0
	expected.EmailConf.EmailVerificationCodeLength = 8
	expected.EmailConf.PasswordRecoverCodeLength = 12
	expected.EmailConf.EmailVerificationTag = "emailVerification"
	expected.EmailConf.PasswordRecoverTag = "passwordRecover"
	expected.EmailConf.HTMLModel = "product_url:https://github.com/pilinux/gorest;product_name:gorest;company_name:pilinux;company_address:Country"
	expected.EmailConf.EmailVerifyValidityPeriod = 86400
	expected.EmailConf.PassRecoverValidityPeriod = 1800

	expected.Logger.Activate = config.Activated
	if !config.IsSentry() {
		t.Errorf("expected IsSentry() to return true, but got false")
	}
	expected.Logger.SentryDsn = "https://xyz.ingest.sentry.io/123456"
	expected.Logger.PerformanceTracing = "yes"
	expected.Logger.TracesSampleRate = "1.0"

	expected.Server.ServerHost = "localhost"
	expected.Server.ServerPort = "3000"
	expected.Server.ServerEnv = "development"

	expected.Security.UserPassMinLength = 6

	expected.Security.MustBasicAuth = config.Activated
	if !config.IsBasicAuth() {
		t.Errorf("expected IsBasicAuth() to return true, but got false")
	}
	expected.Security.BasicAuth.Username = "test_username"
	expected.Security.BasicAuth.Password = "secret_password"

	expected.Security.MustJWT = config.Activated
	if !config.IsJWT() {
		t.Errorf("expected IsJWT() to return true, but got false")
	}
	expected.Security.JWT.Algorithm = "HS256"
	expected.Security.JWT.AccessKey = []byte("this_is_a_long_cryptographic_key_1")
	expected.Security.JWT.AccessKeyTTL = 5
	expected.Security.JWT.RefreshKey = []byte("this_is_a_long_cryptographic_key_2")
	expected.Security.JWT.RefreshKeyTTL = 60
	expected.Security.JWT.PrivKeyECDSA = nil
	expected.Security.JWT.PubKeyECDSA = nil
	expected.Security.JWT.PrivKeyEdDSA = nil
	expected.Security.JWT.PubKeyEdDSA = nil
	expected.Security.JWT.PrivKeyRSA = nil
	expected.Security.JWT.PubKeyRSA = nil

	expected.Security.JWT.Audience = "audience"
	expected.Security.JWT.Issuer = "gorest"
	expected.Security.JWT.AccNbf = 0
	expected.Security.JWT.RefNbf = 0
	expected.Security.JWT.Subject = "subject"

	expected.Security.InvalidateJWT = config.Activated
	if !config.InvalidateJWT() {
		t.Errorf("expected InvalidateJWT() to return true, but got false")
	}

	expected.Security.AuthCookieActivate = true
	if !config.IsAuthCookie() {
		t.Errorf("expected IsAuthCookie() to return true, but got false")
	}
	expected.Security.AuthCookiePath = "/"
	expected.Security.AuthCookieDomain = "test-domain.com"
	expected.Security.AuthCookieSecure = true
	expected.Security.AuthCookieHTTPOnly = true
	expected.Security.AuthCookieSameSite = http.SameSiteStrictMode
	expected.Security.ServeJwtAsResBody = true

	expected.Security.MustHash = config.Activated
	if !config.IsHashPass() {
		t.Errorf("expected IsHashPass() to return true, but got false")
	}
	expected.Security.HashPass.Memory = 64
	expected.Security.HashPass.Iterations = 2
	expected.Security.HashPass.Parallelism = 2
	expected.Security.HashPass.SaltLength = 16
	expected.Security.HashPass.KeyLength = 32
	expected.Security.HashSec = "s€cr$t"

	expected.Security.MustCipher = true
	if !config.IsCipher() {
		t.Errorf("expected IsCipher() to return true, but got false")
	}
	exCipherKey := "cipher_key_secret"
	exCipherKeyHash2 := sha256.Sum256([]byte(exCipherKey)) // sha2-256
	exCipherKeyHash3 := sha3.Sum256(exCipherKeyHash2[:])   // sha3-256
	expected.Security.CipherKey = exCipherKeyHash3[:]
	expected.Security.Blake2bSec = []byte("blake2b_secret")

	expected.Security.VerifyEmail = true
	if !config.IsEmailVerificationService() {
		t.Errorf("expected IsEmailVerificationService() to return true, but got false")
	}
	expected.Security.RecoverPass = true
	if !config.IsPassRecoveryService() {
		t.Errorf("expected IsPassRecoveryService() to return true, but got false")
	}

	expected.Security.MustFW = config.Activated
	if !config.IsWAF() {
		t.Errorf("expected IsWAF() to return true, but got false")
	}
	expected.Security.Firewall.ListType = "whitelist"
	expected.Security.Firewall.IP = "*"

	expected.Security.MustCORS = config.Activated
	if !config.IsCORS() {
		t.Errorf("expected IsCORS() to return true, but got false")
	}
	expected.Security.CORS = append(
		expected.Security.CORS, middleware.CORSPolicy{
			Key:   "Access-Control-Allow-Credentials",
			Value: "true",
		},
	)
	expected.Security.CORS = append(
		expected.Security.CORS, middleware.CORSPolicy{
			Key:   "Access-Control-Allow-Origin",
			Value: "https://example.com",
		},
	)
	expected.Security.CORS = append(
		expected.Security.CORS, middleware.CORSPolicy{
			Key:   "Access-Control-Allow-Headers",
			Value: "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With",
		},
	)
	expected.Security.CORS = append(
		expected.Security.CORS, middleware.CORSPolicy{
			Key:   "Access-Control-Expose-Headers",
			Value: "Content-Length",
		},
	)
	expected.Security.CORS = append(
		expected.Security.CORS, middleware.CORSPolicy{
			Key:   "Access-Control-Allow-Methods",
			Value: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		},
	)
	expected.Security.CORS = append(
		expected.Security.CORS, middleware.CORSPolicy{
			Key:   "Access-Control-Max-Age",
			Value: "3600",
		},
	)
	expected.Security.CORS = append(
		expected.Security.CORS, middleware.CORSPolicy{
			Key:   "X-Content-Type-Options",
			Value: "nosniff",
		},
	)
	expected.Security.CORS = append(
		expected.Security.CORS, middleware.CORSPolicy{
			Key:   "X-Frame-Options",
			Value: "DENY",
		},
	)
	expected.Security.CORS = append(
		expected.Security.CORS, middleware.CORSPolicy{
			Key:   "Referrer-Policy",
			Value: "strict-origin-when-cross-origin",
		},
	)
	expected.Security.CORS = append(
		expected.Security.CORS, middleware.CORSPolicy{
			Key:   "Content-Security-Policy",
			Value: "default-src 'none'; script-src 'self'; connect-src 'self'; img-src 'self'; style-src 'self'; base-uri 'self'; form-action 'self'",
		},
	)
	expected.Security.CORS = append(
		expected.Security.CORS, middleware.CORSPolicy{
			Key:   "Timing-Allow-Origin",
			Value: "*",
		},
	)

	expected.Security.CheckOrigin = config.Activated
	if !config.IsOriginCheck() {
		t.Errorf("expected IsOriginCheck() to return true, but got false")
	}
	expected.Security.RateLimit = "100-M"
	if !config.IsRateLimit() {
		t.Errorf("expected IsRateLimit() to return true, but got false")
	}
	expected.Security.TrustedPlatform = "X-Real-Ip"

	expected.Security.Must2FA = config.Activated
	if !config.Is2FA() {
		t.Errorf("expected Is2FA() to return true, but got false")
	}
	expected.Security.TwoFA.Issuer = "gorest"
	expected.Security.TwoFA.Crypto = crypto.SHA1
	expected.Security.TwoFA.Digits = 6
	expected.Security.TwoFA.PathQR = "tmp"
	expected.Security.TwoFA.DoubleHash = true
	if !config.Is2FADoubleHash() {
		t.Errorf("expected Is2FADoubleHash() to return true, but got false")
	}

	expected.Security.TwoFA.Status.Verified = "verified"
	expected.Security.TwoFA.Status.On = "on"
	expected.Security.TwoFA.Status.Off = "off"
	expected.Security.TwoFA.Status.Invalid = "invalid"

	expected.ViewConfig.Activate = config.Activated
	if !config.IsTemplatingEngine() {
		t.Errorf("expected IsTemplatingEngine() to return true, but got false")
	}
	expected.ViewConfig.Directory = "templates"

	if !reflect.DeepEqual(configAll, expected) {
		t.Errorf("got: %v, want: %v", configAll, expected)
	}
}

func TestErrorGetConfig(t *testing.T) {
	testCases := []struct {
		Key   string
		Value string
	}{
		{
			Key:   "MIN_PASS_LENGTH",
			Value: "text",
		},
		{
			Key: "ACCESS_KEY_TTL",
		},
		{
			Key: "REFRESH_KEY_TTL",
		},
		{
			Key: "NOT_BEFORE_ACC",
		},
		{
			Key: "NOT_BEFORE_REF",
		},
		{
			Key: "HASHPASSMEMORY",
		},
		{
			Key: "HASHPASSITERATIONS",
		},
		{
			Key: "HASHPASSPARALLELISM",
		},
		{
			Key: "HASHPASSSALTLENGTH",
		},
		{
			Key: "HASHPASSKEYLENGTH",
		},
		{
			Key: "CIPHER_KEY",
		},
		{
			Key: "TWO_FA_DIGITS",
		},
		{
			Key: "DBMAXIDLECONNS",
		},
		{
			Key: "DBMAXOPENCONNS",
		},
		{
			Key: "DBCONNMAXLIFETIME",
		},
		{
			Key: "DBLOGLEVEL",
		},
		{
			Key: "POOLSIZE",
		},
		{
			Key: "CONNTTL",
		},
		{
			Key: "MONGO_POOLSIZE",
		},
		{
			Key: "MONGO_CONNTTL",
		},
		{
			Key: "EMAIL_VERIFY_TEMPLATE_ID",
		},
		{
			Key: "EMAIL_PASS_RECOVER_TEMPLATE_ID",
		},
		{
			Key: "EMAIL_UPDATE_VERIFY_TEMPLATE_ID",
		},
		{
			Key: "EMAIL_VERIFY_CODE_LENGTH",
		},
		{
			Key: "EMAIL_PASS_RECOVER_CODE_LENGTH",
		},
		{
			Key: "EMAIL_VERIFY_VALIDITY_PERIOD",
		},
		{
			Key: "EMAIL_PASS_RECOVER_VALIDITY_PERIOD",
		},
	}

	// download a file from a remote location and save it
	fileURL := strings.TrimSpace(os.Getenv("TEST_ENV_URL"))
	err := downloadFile(".env", fileURL)
	if err != nil {
		t.Error(err)
	}

	err = config.Env()
	if err != nil {
		t.Errorf("got error when calling config.Env(): %v", err)
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run("When missing: "+tc.Key, func(t *testing.T) {
			currentValue := os.Getenv(tc.Key)

			// set new value
			err = os.Setenv(tc.Key, tc.Value)
			if err != nil {
				t.Errorf("got error '%v' when setting %v", err, tc.Key)
			}

			err = config.Config()
			if err == nil {
				t.Errorf("expected error, got nil")
			}

			// set old value
			err = os.Setenv(tc.Key, currentValue)
			if err != nil {
				t.Errorf("got error '%v' when setting %v", err, tc.Key)
			}
		})
	}

	// remove the downloaded file at the end of the test
	err = os.RemoveAll(".env")
	if err != nil {
		t.Error(err)
	}
}

func TestConfigWithDifferentExpectedValueTypes(t *testing.T) {
	testCases := []struct {
		Key       string
		TestNo    int
		SetValue  string
		ExpErr    error
		ExpValue1 bool
		ExpValue2 http.SameSite
		ExpValue3 []byte
		ExpValue4 crypto.Hash
		ExpValue5 []middleware.CORSPolicy
		ExpValue6 string

		// fields below drive the JWT key-file tests (asymmetric algorithms).
		// Alg is the JWT_ALG to set; PrivFile/PubFile are remote key files to
		// download (empty = skip); PrivPath/PubPath are the values for
		// PRIV_KEY_FILE_PATH/PUB_KEY_FILE_PATH; ExpKeyErr marks a failure case.
		Alg       string
		PrivFile  string
		PubFile   string
		PrivPath  string
		PubPath   string
		ExpKeyErr bool
	}{
		{
			Key:       "EMAIL_TRACK_OPENS",
			TestNo:    1,
			SetValue:  "yes",
			ExpValue1: true,
		},
		{
			Key:       "AUTH_COOKIE_SameSite",
			TestNo:    2,
			SetValue:  "lax",
			ExpValue2: http.SameSiteLaxMode,
		},
		{
			Key:       "AUTH_COOKIE_SameSite",
			TestNo:    3,
			SetValue:  "none",
			ExpValue2: http.SameSiteNoneMode,
		},
		{
			Key:       "BLAKE2B_SECRET",
			TestNo:    4,
			SetValue:  "",
			ExpValue3: nil,
		},
		{
			Key:       "TWO_FA_CRYPTO",
			TestNo:    5,
			SetValue:  "256",
			ExpValue4: crypto.SHA256,
		},
		{
			Key:       "TWO_FA_CRYPTO",
			TestNo:    6,
			SetValue:  "512",
			ExpValue4: crypto.SHA512,
		},
		{
			Key:      "CORS_HSTS",
			TestNo:   7,
			SetValue: "max-age=63072000; includeSubDomains; preload",
			ExpValue5: []middleware.CORSPolicy{
				{
					Key:   "Access-Control-Allow-Credentials",
					Value: "true",
				},
				{
					Key:   "Access-Control-Allow-Origin",
					Value: "https://example.com",
				},
				{
					Key:   "Access-Control-Allow-Headers",
					Value: "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With",
				},
				{
					Key:   "Access-Control-Expose-Headers",
					Value: "Content-Length",
				},
				{
					Key:   "Access-Control-Allow-Methods",
					Value: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
				},
				{
					Key:   "Access-Control-Max-Age",
					Value: "3600",
				},
				{
					Key:   "X-Content-Type-Options",
					Value: "nosniff",
				},
				{
					Key:   "X-Frame-Options",
					Value: "DENY",
				},
				{
					Key:   "Referrer-Policy",
					Value: "strict-origin-when-cross-origin",
				},
				{
					Key:   "Content-Security-Policy",
					Value: "default-src 'none'; script-src 'self'; connect-src 'self'; img-src 'self'; style-src 'self'; base-uri 'self'; form-action 'self'",
				},
				{
					Key:   "Timing-Allow-Origin",
					Value: "*",
				},
				{
					Key:   "Strict-Transport-Security",
					Value: "max-age=63072000; includeSubDomains; preload",
				},
			},
		},
		{
			Key:       "JWT_ALG",
			TestNo:    8,
			SetValue:  "",
			ExpValue6: "HS256",
		},
		{
			Key:      "JWT_ALG",
			TestNo:   9,
			SetValue: "any",
		},
		{
			// fail test - asymmetric alg with a non-existent private key path
			TestNo:    10,
			Alg:       "ES256",
			PrivPath:  "./wrong_path",
			PubPath:   "",
			ExpKeyErr: true,
		},
		{
			// fail test - asymmetric alg, valid private key but non-existent public key path
			TestNo:    11,
			Alg:       "ES256",
			PrivFile:  "private-keyES256.pem",
			PrivPath:  "./private-keyES256.pem",
			PubPath:   "./wrong_path",
			ExpKeyErr: true,
		},
		{
			// test valid ES256 key pair
			TestNo:   12,
			Alg:      "ES256",
			PrivFile: "private-keyES256.pem",
			PubFile:  "public-keyES256.pem",
			PrivPath: "./private-keyES256.pem",
			PubPath:  "./public-keyES256.pem",
		},
		{
			// test valid RS256 key pair
			TestNo:   13,
			Alg:      "RS256",
			PrivFile: "private-keyRS256.pem",
			PubFile:  "public-keyRS256.pem",
			PrivPath: "./private-keyRS256.pem",
			PubPath:  "./public-keyRS256.pem",
		},
		{
			// fail test - asymmetric alg, private key file missing on disk
			TestNo:    14,
			Alg:       "ES256",
			PrivPath:  "./private-keyES256.pem",
			PubPath:   "./public-keyES256.pem",
			ExpKeyErr: true,
		},
		{
			// fail test - asymmetric alg, valid private key but public key file missing on disk
			TestNo:    15,
			Alg:       "ES256",
			PrivFile:  "private-keyES256.pem",
			PrivPath:  "./private-keyES256.pem",
			PubPath:   "./public-keyES256.pem",
			ExpKeyErr: true,
		},
		{
			// fail test - wrong private key for ES256 (RSA key supplied)
			TestNo:    16,
			Alg:       "ES256",
			PrivFile:  "private-keyRS256.pem",
			PrivPath:  "./private-keyRS256.pem",
			ExpKeyErr: true,
		},
		{
			// fail test - wrong private key for RS256 (ECDSA key supplied)
			TestNo:    17,
			Alg:       "RS256",
			PrivFile:  "private-keyES256.pem",
			PrivPath:  "./private-keyES256.pem",
			ExpKeyErr: true,
		},
		{
			// fail test - valid private key but wrong public key for ES256 (RSA key supplied)
			TestNo:    18,
			Alg:       "ES256",
			PrivFile:  "private-keyES256.pem",
			PubFile:   "public-keyRS256.pem",
			PrivPath:  "./private-keyES256.pem",
			PubPath:   "./public-keyRS256.pem",
			ExpKeyErr: true,
		},
		{
			// fail test - valid private key but wrong public key for RS256 (ECDSA key supplied)
			TestNo:    19,
			Alg:       "RS256",
			PrivFile:  "private-keyRS256.pem",
			PubFile:   "public-keyES256.pem",
			PrivPath:  "./private-keyRS256.pem",
			PubPath:   "./public-keyES256.pem",
			ExpKeyErr: true,
		},
		{
			// test valid EdDSA Ed25519 key pair
			TestNo:   20,
			Alg:      "EdDSA",
			PrivFile: "private-keyEdDSA.pem",
			PubFile:  "public-keyEdDSA.pem",
			PrivPath: "./private-keyEdDSA.pem",
			PubPath:  "./public-keyEdDSA.pem",
		},
		{
			// fail test - wrong private key for EdDSA Ed25519 (RSA key supplied)
			TestNo:    21,
			Alg:       "EdDSA",
			PrivFile:  "private-keyRS256.pem",
			PrivPath:  "./private-keyRS256.pem",
			ExpKeyErr: true,
		},
		{
			// fail test - valid private key but wrong public key for EdDSA Ed25519 (RSA key supplied)
			TestNo:    29,
			Alg:       "EdDSA",
			PrivFile:  "private-keyEdDSA.pem",
			PubFile:   "public-keyRS256.pem",
			PrivPath:  "./private-keyEdDSA.pem",
			PubPath:   "./public-keyRS256.pem",
			ExpKeyErr: true,
		},
		{
			// fail test - asymmetric alg with an empty private key path
			// (PRIV_KEY_FILE_PATH is required for asymmetric JWT algorithms)
			TestNo:    30,
			Alg:       "ES256",
			PrivPath:  "",
			PubPath:   "",
			ExpKeyErr: true,
		},
		{
			// fail test - asymmetric alg, valid private key but empty public key path
			// (PUB_KEY_FILE_PATH is required for asymmetric JWT algorithms)
			TestNo:    31,
			Alg:       "ES256",
			PrivFile:  "private-keyES256.pem",
			PrivPath:  "./private-keyES256.pem",
			PubPath:   "",
			ExpKeyErr: true,
		},
		{
			Key:       "EMAIL_VERIFY_USE_UUIDv4",
			TestNo:    22,
			SetValue:  "yes",
			ExpValue1: true,
		},
		{
			Key:       "EMAIL_PASS_RECOVER_USE_UUIDv4",
			TestNo:    23,
			SetValue:  "yes",
			ExpValue1: true,
		},
		{
			// invalid: credentials true with wildcard in comma-separated list for Access-Control-Allow-Origin
			Key:      "CORS_ORIGIN",
			TestNo:   24,
			SetValue: "https://example.com, * ,https://test.com",
			ExpErr:   errors.New("credentialed request cannot have '*' as Access-Control-Allow-Origin"),
		},
		{
			// invalid: credentials true with wildcard in comma-separated list for Access-Control-Allow-Headers
			Key:      "CORS_HEADERS",
			TestNo:   25,
			SetValue: "Content-Type, Content-Length, *",
			ExpErr:   errors.New("credentialed request cannot have '*' as Access-Control-Allow-Headers"),
		},
		{
			// invalid: credentials true with wildcard in comma-separated list for Access-Control-Expose-Headers
			Key:      "CORS_EXPOSE_HEADERS",
			TestNo:   26,
			SetValue: "Content-Length, *",
			ExpErr:   errors.New("credentialed request cannot have '*' as Access-Control-Expose-Headers"),
		},
		{
			// invalid: credentials true with wildcard in comma-separated list for Access-Control-Allow-Methods
			Key:      "CORS_METHODS",
			TestNo:   27,
			SetValue: "GET, POST, PUT, *",
			ExpErr:   errors.New("credentialed request cannot have '*' as Access-Control-Allow-Methods"),
		},
		{
			// invalid template directory (traversal)
			Key:      "TEMPLATE_DIR",
			TestNo:   33,
			SetValue: "../escape",
			ExpErr:   errors.New("invalid template directory"),
		},
		{
			// invalid 2FA QR path (traversal)
			Key:      "TWO_FA_QR_PATH",
			TestNo:   34,
			SetValue: "../escape",
			ExpErr:   errors.New("invalid 2FA QR path"),
		},
		{
			// SERVE_JWT_AS_RESPONSE_BODY set to "no"
			Key:       "SERVE_JWT_AS_RESPONSE_BODY",
			TestNo:    35,
			SetValue:  "no",
			ExpValue1: false,
		},
		{
			// AUTH_COOKIE_SameSite with unknown value (not strict/lax/none)
			Key:       "AUTH_COOKIE_SameSite",
			TestNo:    36,
			SetValue:  "unknown",
			ExpValue2: 0,
		},
		{
			Key:       "DB_URI",
			TestNo:    37,
			SetValue:  "postgres://db_user:db_pass@127.0.0.1:5432/db_name?sslmode=disable",
			ExpValue6: "postgres://db_user:db_pass@127.0.0.1:5432/db_name?sslmode=disable",
		},
		{
			Key:       "REDIS_URI",
			TestNo:    38,
			SetValue:  "redis://127.0.0.1:6379/0",
			ExpValue6: "redis://127.0.0.1:6379/0",
		},
		{
			// cors is activated but empty
			Key:      "ACTIVATE_CORS",
			TestNo:   28,
			SetValue: "yes",
			ExpErr:   errors.New(("empty CORS header")),
		},
		{
			// fail test - ACCESS_KEY shorter than 32 bytes is rejected for HMAC algorithms
			Key:      "ACCESS_KEY",
			TestNo:   39,
			SetValue: "short_access_key",
			ExpErr:   errors.New("ACCESS_KEY and REFRESH_KEY must each be at least 32 bytes for HMAC algorithms"),
		},
		{
			// fail test - REFRESH_KEY shorter than 32 bytes is rejected for HMAC algorithms
			Key:      "REFRESH_KEY",
			TestNo:   40,
			SetValue: "short_refresh_key",
			ExpErr:   errors.New("ACCESS_KEY and REFRESH_KEY must each be at least 32 bytes for HMAC algorithms"),
		},
	}

	// download a file from a remote location and save it
	fileURL := strings.TrimSpace(os.Getenv("TEST_ENV_URL"))
	err := downloadFile(".env", fileURL)
	if err != nil {
		t.Error(err)
	}

	// remote location for private-public key file
	testKeyFilePath := strings.TrimSpace(os.Getenv("TEST_KEY_FILE_LOCATION"))

	err = config.Env()
	if err != nil {
		t.Errorf("got error when calling config.Env(): %v", err)
	}

	// isKeyTest reports whether a test exercises the asymmetric JWT key files.
	// These tests manage JWT_ALG, PRIV_KEY_FILE_PATH and PUB_KEY_FILE_PATH
	// themselves, so the generic env setter/reset is skipped for them.
	isKeyTest := func(no int) bool {
		switch no {
		case 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 29, 30, 31:
			return true
		}
		return false
	}

	for i := range testCases {
		tc := testCases[i]
		name := fmt.Sprintf("Setting TestNo %d", tc.TestNo)
		if tc.Key != "" {
			name = fmt.Sprintf("Setting %s TestNo %d", tc.Key, tc.TestNo)
		}
		t.Run(name, func(t *testing.T) {
			currentValue := os.Getenv(tc.Key)

			// set new value
			if !isKeyTest(tc.TestNo) {
				err = os.Setenv(tc.Key, tc.SetValue)
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
			}

			if isKeyTest(tc.TestNo) {
				// download the required private/public key files (if any)
				if tc.PrivFile != "" {
					fmt.Println("downloading...", tc.PrivFile)
					if e := downloadFile(tc.PrivFile, testKeyFilePath+"/"+tc.PrivFile); e != nil {
						t.Error(e)
					}
				}
				if tc.PubFile != "" {
					fmt.Println("downloading...", tc.PubFile)
					if e := downloadFile(tc.PubFile, testKeyFilePath+"/"+tc.PubFile); e != nil {
						t.Error(e)
					}
				}

				// set the algorithm and key paths for this test
				if e := os.Setenv("JWT_ALG", tc.Alg); e != nil {
					t.Errorf("got error '%v' when setting JWT_ALG for test no: '%v'", e, tc.TestNo)
				}
				if e := os.Setenv("PRIV_KEY_FILE_PATH", tc.PrivPath); e != nil {
					t.Errorf("got error '%v' when setting PRIV_KEY_FILE_PATH for test no: '%v'", e, tc.TestNo)
				}
				if e := os.Setenv("PUB_KEY_FILE_PATH", tc.PubPath); e != nil {
					t.Errorf("got error '%v' when setting PUB_KEY_FILE_PATH for test no: '%v'", e, tc.TestNo)
				}
			}

			if tc.TestNo >= 24 && tc.TestNo <= 27 {
				// test with invalid CORS settings
				fmt.Println("test with invalid CORS settings")
				fmt.Println("test no:", tc.TestNo)
				err = os.Setenv(tc.Key, tc.SetValue)
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
			}

			if tc.TestNo == 28 {
				// test with empty CORS settings
				fmt.Println("test with empty CORS settings")
				_ = os.Setenv("CORS_CREDENTIALS", "")
				_ = os.Setenv("CORS_ORIGIN", "")
				_ = os.Setenv("CORS_HEADERS", "")
				_ = os.Setenv("CORS_EXPOSE_HEADERS", "")
				_ = os.Setenv("CORS_METHODS", "")
				_ = os.Setenv("CORS_MAXAGE", "")
				_ = os.Setenv("CORS_X_CONTENT_TYPE", "")
				_ = os.Setenv("CORS_X_FRAME", "")
				_ = os.Setenv("CORS_REFERRER", "")
				_ = os.Setenv("CORS_CONTENT_SECURITY", "")
				_ = os.Setenv("CORS_TIMING_ALLOW_ORIGIN", "")
				_ = os.Setenv("CORS_HSTS", "")
			}

			err = config.Config()

			if tc.TestNo == 1 {
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
				if !config.GetConfig().EmailConf.TrackOpens {
					t.Errorf("expected true, got false when setting %v", tc.Key)
				}
			}

			if tc.TestNo == 2 || tc.TestNo == 3 {
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
				got := config.GetConfig().Security.AuthCookieSameSite
				if got != tc.ExpValue2 {
					t.Errorf("expected %v, got %v when setting %v", tc.ExpValue2, got, tc.Key)
				}
			}
			if tc.TestNo == 4 {
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
				got := config.GetConfig().Security.Blake2bSec
				if !bytes.Equal(got, tc.ExpValue3) {
					t.Errorf("expected %v, got %v when setting %v", tc.ExpValue3, got, tc.Key)
				}
			}

			if tc.TestNo == 5 || tc.TestNo == 6 {
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
				got := config.GetConfig().Security.TwoFA.Crypto
				if got != tc.ExpValue4 {
					t.Errorf("expected %v, got %v when setting %v", tc.ExpValue4, got, tc.Key)
				}
			}

			if tc.TestNo == 7 {
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
				got := config.GetConfig().Security.CORS
				if !compareSlice(got, tc.ExpValue5) {
					t.Errorf("expected\n %v\n, got\n %v\n when setting %v", tc.ExpValue5, got, tc.Key)
				}
			}

			if tc.TestNo == 8 {
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
				got := config.GetConfig().Security.JWT.Algorithm
				if got != tc.ExpValue6 {
					t.Errorf("expected %v, got %v when setting %v", tc.ExpValue6, got, tc.Key)
				}
			}

			if tc.TestNo == 37 {
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
				got := config.GetConfig().Database.RDBMS.Env.URI
				if got != tc.ExpValue6 {
					t.Errorf("expected %v, got %v when setting %v", tc.ExpValue6, got, tc.Key)
				}
			}

			if tc.TestNo == 38 {
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
				got := config.GetConfig().Database.REDIS.Env.URI
				if got != tc.ExpValue6 {
					t.Errorf("expected %v, got %v when setting %v", tc.ExpValue6, got, tc.Key)
				}
			}

			if tc.TestNo == 9 {
				if err == nil {
					t.Errorf("expected error, got nil when setting %v", tc.Key)
				}
			}

			// JWT key-file tests (asymmetric algorithms): asymmetric algorithms
			// require both a private and a public key. ExpKeyErr marks the cases
			// that must fail (missing/non-existent/mismatched keys).
			if isKeyTest(tc.TestNo) {
				if tc.ExpKeyErr {
					if err == nil {
						t.Errorf("expected error, got nil for test no: '%v'", tc.TestNo)
					}
				} else {
					if err != nil {
						t.Errorf("got error '%v' for test no: '%v'", err, tc.TestNo)
					}
				}

				// reset algorithm and key paths
				if e := os.Setenv("JWT_ALG", "HS256"); e != nil {
					t.Errorf("got error '%v' when resetting JWT_ALG for test no: '%v'", e, tc.TestNo)
				}
				if e := os.Setenv("PRIV_KEY_FILE_PATH", ""); e != nil {
					t.Errorf("got error '%v' when resetting PRIV_KEY_FILE_PATH for test no: '%v'", e, tc.TestNo)
				}
				if e := os.Setenv("PUB_KEY_FILE_PATH", ""); e != nil {
					t.Errorf("got error '%v' when resetting PUB_KEY_FILE_PATH for test no: '%v'", e, tc.TestNo)
				}

				// remove the downloaded key files at the end of the test
				if tc.PrivFile != "" {
					fmt.Println("deleting...", tc.PrivPath)
					if e := os.RemoveAll(tc.PrivPath); e != nil {
						t.Error(e)
					}
				}
				if tc.PubFile != "" {
					fmt.Println("deleting...", tc.PubPath)
					if e := os.RemoveAll(tc.PubPath); e != nil {
						t.Error(e)
					}
				}
			}

			if tc.TestNo == 22 {
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
				if !config.IsEmailVerificationCodeUUIDv4() {
					t.Errorf("expected true, got false when setting %v", tc.Key)
				}
			}

			if tc.TestNo == 23 {
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
				if !config.IsPasswordRecoverCodeUUIDv4() {
					t.Errorf("expected true, got false when setting %v", tc.Key)
				}
			}

			if tc.TestNo >= 24 && tc.TestNo <= 27 {
				if err == nil {
					t.Errorf("expected error, got nil for test no: '%v'", tc.TestNo)
				}
			}

			if tc.TestNo == 28 {
				if err == nil {
					t.Errorf("expected error, got nil for test no: '%v'", tc.TestNo)
				}
			}

			// invalid template directory (traversal)
			if tc.TestNo == 33 {
				if err == nil {
					t.Errorf("expected error, got nil for test no: '%v'", tc.TestNo)
				}
				if err != nil && !strings.Contains(err.Error(), "invalid template directory") {
					t.Errorf("expected 'invalid template directory' error, got: %v", err)
				}
			}

			// invalid 2FA QR path (traversal)
			if tc.TestNo == 34 {
				if err == nil {
					t.Errorf("expected error, got nil for test no: '%v'", tc.TestNo)
				}
				if err != nil && !strings.Contains(err.Error(), "invalid 2FA QR path") {
					t.Errorf("expected 'invalid 2FA QR path' error, got: %v", err)
				}
			}

			// SERVE_JWT_AS_RESPONSE_BODY set to "no"
			if tc.TestNo == 35 {
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
				if config.GetConfig().Security.ServeJwtAsResBody {
					t.Errorf("expected ServeJwtAsResBody false, got true")
				}
			}

			// AUTH_COOKIE_SameSite with unknown value
			if tc.TestNo == 36 {
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
				got := config.GetConfig().Security.AuthCookieSameSite
				if got != tc.ExpValue2 {
					t.Errorf("expected %v, got %v when setting %v", tc.ExpValue2, got, tc.Key)
				}
			}

			// empty or too-short HMAC ACCESS_KEY/REFRESH_KEY must be rejected
			if tc.TestNo == 39 || tc.TestNo == 40 {
				if err == nil {
					t.Errorf("expected error, got nil when setting %v to '%v'", tc.Key, tc.SetValue)
				}
				if err != nil && err.Error() != tc.ExpErr.Error() {
					t.Errorf("expected error '%v', got '%v' when setting %v", tc.ExpErr, err, tc.Key)
				}
			}

			// set old value (key tests manage their own env, handled above)
			if !isKeyTest(tc.TestNo) {
				err = os.Setenv(tc.Key, currentValue)
				if err != nil {
					t.Errorf("got error '%v' when setting %v", err, tc.Key)
				}
			}
		})
	}

	// remove the downloaded file at the end of the test
	err = os.RemoveAll(".env")
	if err != nil {
		t.Error(err)
	}
}

// compareSlice compares two slices of middleware.CORSPolicy
// for equality, ignoring order
func compareSlice(a, b []middleware.CORSPolicy) bool {
	if len(a) != len(b) {
		return false
	}

	// create maps to store elements from both slices
	aMap := make(map[middleware.CORSPolicy]struct{}, len(a))
	bMap := make(map[middleware.CORSPolicy]struct{}, len(b))

	// populate maps with elements from slices
	for _, policy := range a {
		aMap[policy] = struct{}{}
	}
	for _, policy := range b {
		bMap[policy] = struct{}{}
	}

	// compare maps to check for equality
	return reflect.DeepEqual(aMap, bMap)
}

// downloadFile will download from a url and save it to a local file.
// It's efficient because it will write as it downloads and not
// load the whole file into memory.
func downloadFile(filepath string, url string) error {

	// create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() {
		if e := out.Close(); e != nil && err == nil {
			err = e
		}
	}()

	// get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil && err == nil {
			err = e
		}
	}()

	// write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
