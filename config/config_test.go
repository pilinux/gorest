package config_test

import (
	"bytes"
	"crypto"
	"crypto/sha256"
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
	"golang.org/x/crypto/sha3"
)

func TestEnv(t *testing.T) {
	// this should return error
	err := config.Env()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// download a file from a remote location and save it
	fileUrl := strings.TrimSpace(os.Getenv("TEST_ENV_URL"))
	err = downloadFile(".env", fileUrl)
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
	fileUrl := strings.TrimSpace(os.Getenv("TEST_ENV_URL"))
	err = downloadFile(".env", fileUrl)
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
	expected.Security.JWT.AccessKey = []byte("cryptographic_key_1")
	expected.Security.JWT.AccessKeyTTL = 5
	expected.Security.JWT.RefreshKey = []byte("cryptographic_key_2")
	expected.Security.JWT.RefreshKeyTTL = 60
	expected.Security.JWT.PrivKeyECDSA = nil
	expected.Security.JWT.PubKeyECDSA = nil
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
			Key:   "Access-Control-Allow-Origin",
			Value: "*",
		},
	)
	expected.Security.CORS = append(
		expected.Security.CORS, middleware.CORSPolicy{
			Key:   "Access-Control-Allow-Credentials",
			Value: "true",
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
	fileUrl := strings.TrimSpace(os.Getenv("TEST_ENV_URL"))
	err := downloadFile(".env", fileUrl)
	if err != nil {
		t.Error(err)
	}

	err = config.Env()
	if err != nil {
		t.Errorf("got error when calling config.Env(): %v", err)
	}

	for _, tc := range testCases {
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
		FileName  string
		SetValue  string
		ExpErr    error
		ExpValue1 bool
		ExpValue2 http.SameSite
		ExpValue3 []byte
		ExpValue4 crypto.Hash
		ExpValue5 []middleware.CORSPolicy
		ExpValue6 string
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
					Key:   "Access-Control-Allow-Origin",
					Value: "*",
				},
				{
					Key:   "Access-Control-Allow-Credentials",
					Value: "true",
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
			Key:      "PRIV_KEY_FILE_PATH",
			TestNo:   10,
			SetValue: "./wrong_path",
		},
		{
			Key:      "PUB_KEY_FILE_PATH",
			TestNo:   11,
			SetValue: "./wrong_path",
		},
		{
			// test private key for ES256
			Key:      "PRIV_KEY_FILE_PATH",
			TestNo:   12,
			FileName: "private-keyES256.pem",
			SetValue: "./private-keyES256.pem",
		},
		{
			// test private key for RS256
			Key:      "PRIV_KEY_FILE_PATH",
			TestNo:   13,
			FileName: "private-keyRS256.pem",
			SetValue: "./private-keyRS256.pem",
		},
		{
			// test public key for ES256
			Key:      "PUB_KEY_FILE_PATH",
			TestNo:   14,
			FileName: "public-keyES256.pem",
			SetValue: "./public-keyES256.pem",
		},
		{
			// test public key for RS256
			Key:      "PUB_KEY_FILE_PATH",
			TestNo:   15,
			FileName: "public-keyRS256.pem",
			SetValue: "./public-keyRS256.pem",
		},
		{
			// fail test - no private key present for ES256 or RS256
			Key:      "PRIV_KEY_FILE_PATH",
			TestNo:   16,
			SetValue: "./private-keyES256.pem",
		},
		{
			// fail test - no public key present for ES256 or RS256
			Key:      "PUB_KEY_FILE_PATH",
			TestNo:   17,
			SetValue: "./public-keyES256.pem",
		},
		{
			// fail test - wrong private key for ES256
			Key:      "PRIV_KEY_FILE_PATH",
			TestNo:   18,
			FileName: "private-keyRS256.pem",
			SetValue: "./private-keyRS256.pem",
		},
		{
			// fail test - wrong private key for RS256
			Key:      "PRIV_KEY_FILE_PATH",
			TestNo:   19,
			FileName: "private-keyES256.pem",
			SetValue: "./private-keyES256.pem",
		},
		{
			// fail test - wrong public key for ES256
			Key:      "PUB_KEY_FILE_PATH",
			TestNo:   20,
			FileName: "public-keyRS256.pem",
			SetValue: "./public-keyRS256.pem",
		},
		{
			// fail test - wrong public key for RS256
			Key:      "PUB_KEY_FILE_PATH",
			TestNo:   21,
			FileName: "public-keyES256.pem",
			SetValue: "./public-keyES256.pem",
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
	}

	// download a file from a remote location and save it
	fileUrl := strings.TrimSpace(os.Getenv("TEST_ENV_URL"))
	err := downloadFile(".env", fileUrl)
	if err != nil {
		t.Error(err)
	}

	// remote location for private-public key file
	testKeyFilePath := strings.TrimSpace(os.Getenv("TEST_KEY_FILE_LOCATION"))

	err = config.Env()
	if err != nil {
		t.Errorf("got error when calling config.Env(): %v", err)
	}

	for _, tc := range testCases {
		t.Run("Setting "+tc.Key, func(t *testing.T) {
			currentValue := os.Getenv(tc.Key)

			// set new value
			err = os.Setenv(tc.Key, tc.SetValue)
			if err != nil {
				t.Errorf("got error '%v' when setting %v", err, tc.Key)
			}

			if tc.TestNo == 12 || tc.TestNo == 13 || tc.TestNo == 14 || tc.TestNo == 15 ||
				tc.TestNo == 18 || tc.TestNo == 19 || tc.TestNo == 20 || tc.TestNo == 21 {
				// download private-public key file from a remote location and save it
				fmt.Println("downloading...", tc.FileName)
				err := downloadFile(tc.FileName, testKeyFilePath+"/"+tc.FileName)
				if err != nil {
					t.Error(err)
				}
			}

			if tc.TestNo == 12 || tc.TestNo == 14 {
				// test with keys for ES256
				fmt.Println("test with keys for ES256")
				err = os.Setenv("JWT_ALG", "ES256")
				if err != nil {
					t.Errorf("got error '%v' when setting JWT_ALG for test no: '%v'", err, tc.TestNo)
				}
			}

			if tc.TestNo == 13 || tc.TestNo == 15 {
				// test with keys for RS256
				fmt.Println("test with keys for RS256")
				err = os.Setenv("JWT_ALG", "RS256")
				if err != nil {
					t.Errorf("got error '%v' when setting JWT_ALG for test no: '%v'", err, tc.TestNo)
				}
			}

			if tc.TestNo == 18 || tc.TestNo == 20 {
				// test with wrong keys for ES256
				fmt.Println("test with wrong keys for ES256")
				err = os.Setenv("JWT_ALG", "ES256")
				if err != nil {
					t.Errorf("got error '%v' when setting JWT_ALG for test no: '%v'", err, tc.TestNo)
				}
			}

			if tc.TestNo == 19 || tc.TestNo == 21 {
				// test with wrong keys for RS256
				fmt.Println("test with wrong keys for RS256")
				err = os.Setenv("JWT_ALG", "RS256")
				if err != nil {
					t.Errorf("got error '%v' when setting JWT_ALG for test no: '%v'", err, tc.TestNo)
				}
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

			if tc.TestNo == 9 {
				if err == nil {
					t.Errorf("expected error, got nil when setting %v", tc.Key)
				}
			}

			if tc.TestNo == 10 || tc.TestNo == 11 {
				if err == nil {
					t.Errorf("expected error, got nil when setting %v", tc.Key)
				}
			}

			// test with keys for ES256
			if tc.TestNo == 12 || tc.TestNo == 14 {
				if err != nil {
					t.Errorf("got error '%v' when setting '%v' for test no: '%v'", err, tc.Key, tc.TestNo)
				}
				// reset value
				os.Setenv("JWT_ALG", "HS256")
				os.Setenv("PRIV_KEY_FILE_PATH", "")
				os.Setenv("PUB_KEY_FILE_PATH", "")
				// remove the downloaded file at the end of the test
				fmt.Println("deleting...", tc.SetValue)
				err = os.RemoveAll(tc.SetValue)
				if err != nil {
					t.Error(err)
				}
			}

			// test with keys for RS256
			if tc.TestNo == 13 || tc.TestNo == 15 {
				if err != nil {
					t.Errorf("got error '%v' when setting '%v' for test no: '%v'", err, tc.Key, tc.TestNo)
				}
				// reset value
				os.Setenv("JWT_ALG", "HS256")
				os.Setenv("PRIV_KEY_FILE_PATH", "")
				os.Setenv("PUB_KEY_FILE_PATH", "")
				// remove the downloaded file at the end of the test
				fmt.Println("deleting...", tc.SetValue)
				err = os.RemoveAll(tc.SetValue)
				if err != nil {
					t.Error(err)
				}
			}

			// fail test - without keys of ES256 or RS256
			if tc.TestNo == 16 || tc.TestNo == 17 {
				if err == nil {
					t.Errorf("expected error, got nil when setting '%v' for test no: '%v'", tc.Key, tc.TestNo)
				}
			}

			// fail test with wrong keys for ES256
			if tc.TestNo == 18 || tc.TestNo == 20 {
				if err == nil {
					t.Errorf("expected error, got nil when setting '%v' for test no: '%v'", tc.Key, tc.TestNo)
				}
				// reset value
				os.Setenv("JWT_ALG", "HS256")
				os.Setenv("PRIV_KEY_FILE_PATH", "")
				os.Setenv("PUB_KEY_FILE_PATH", "")
				// remove the downloaded file at the end of the test
				fmt.Println("deleting...", tc.SetValue)
				err = os.RemoveAll(tc.SetValue)
				if err != nil {
					t.Error(err)
				}
			}

			// fail test with wrong keys for RS256
			if tc.TestNo == 19 || tc.TestNo == 21 {
				if err == nil {
					t.Errorf("expected error, got nil when setting '%v' for test no: '%v'", tc.Key, tc.TestNo)
				}
				// reset value
				os.Setenv("JWT_ALG", "HS256")
				os.Setenv("PRIV_KEY_FILE_PATH", "")
				os.Setenv("PUB_KEY_FILE_PATH", "")
				// remove the downloaded file at the end of the test
				fmt.Println("deleting...", tc.SetValue)
				err = os.RemoveAll(tc.SetValue)
				if err != nil {
					t.Error(err)
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
	defer out.Close()

	// get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
