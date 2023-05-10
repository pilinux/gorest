package config_test

import (
	"crypto"
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
	// download a file from a remote location and save it
	fileUrl := strings.TrimSpace(os.Getenv("TEST_ENV_URL"))
	err := downloadFile(".env", fileUrl)
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
	// download a file from a remote location and save it
	fileUrl := strings.TrimSpace(os.Getenv("TEST_ENV_URL"))
	err := downloadFile(".env", fileUrl)
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

	expected.Database.RDBMS.Activate = config.Activated
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
	expected.Database.REDIS.Env.Host = "127.0.0.1"
	expected.Database.REDIS.Env.Port = "6379"
	expected.Database.REDIS.Conn.PoolSize = 10
	expected.Database.REDIS.Conn.ConnTTL = 5

	expected.Database.MongoDB.Activate = config.Activated
	expected.Database.MongoDB.Env.AppName = "gorest"
	expected.Database.MongoDB.Env.URI = "mongodb://user:pass@127.0.0.1:27017/?retryWrites=true&w=majority"
	expected.Database.MongoDB.Env.PoolSize = 50
	expected.Database.MongoDB.Env.PoolMon = "no"
	expected.Database.MongoDB.Env.ConnTTL = 10

	expected.EmailConf.Activate = config.Activated
	expected.EmailConf.Provider = "postmark"
	expected.EmailConf.APIToken = "abcdef"
	expected.EmailConf.AddrFrom = "email@yourdomain.com"
	expected.EmailConf.TrackOpens = false
	expected.EmailConf.TrackLinks = "None"
	expected.EmailConf.DeliveryType = "outbound"

	expected.EmailConf.EmailVerificationTemplateID = 0
	expected.EmailConf.PasswordRecoverTemplateID = 0
	expected.EmailConf.EmailVerificationCodeLength = 8
	expected.EmailConf.PasswordRecoverCodeLength = 12
	expected.EmailConf.EmailVerificationTag = "emailVerification"
	expected.EmailConf.PasswordRecoverTag = "passwordRecover"
	expected.EmailConf.HTMLModel = "product_url:https://github.com/pilinux/gorest;product_name:gorest;company_name:pilinux;company_address:Country"
	expected.EmailConf.EmailVerifyValidityPeriod = 86400
	expected.EmailConf.PassRecoverValidityPeriod = 1800

	expected.Logger.Activate = config.Activated
	expected.Logger.SentryDsn = "https://xyz.ingest.sentry.io/123456"

	expected.Server.ServerPort = "3000"
	expected.Server.ServerEnv = "development"

	expected.Security.UserPassMinLength = 6

	expected.Security.MustBasicAuth = config.Activated
	expected.Security.BasicAuth.Username = "test_username"
	expected.Security.BasicAuth.Password = "secret_password"

	expected.Security.MustJWT = config.Activated
	expected.Security.JWT.AccessKey = []byte("cryptographic_key_1")
	expected.Security.JWT.AccessKeyTTL = 5
	expected.Security.JWT.RefreshKey = []byte("cryptographic_key_2")
	expected.Security.JWT.RefreshKeyTTL = 60

	expected.Security.JWT.Audience = "audience"
	expected.Security.JWT.Issuer = "gorest"
	expected.Security.JWT.AccNbf = 0
	expected.Security.JWT.RefNbf = 0
	expected.Security.JWT.Subject = "subject"

	expected.Security.AuthCookieActivate = true
	expected.Security.AuthCookiePath = "/"
	expected.Security.AuthCookieDomain = "test-domain.com"
	expected.Security.AuthCookieSecure = true
	expected.Security.AuthCookieHTTPOnly = true
	expected.Security.AuthCookieSameSite = http.SameSiteStrictMode
	expected.Security.ServeJwtAsResBody = true

	expected.Security.MustHash = config.Activated
	expected.Security.HashPass.Memory = 64
	expected.Security.HashPass.Iterations = 2
	expected.Security.HashPass.Parallelism = 2
	expected.Security.HashPass.SaltLength = 16
	expected.Security.HashPass.KeyLength = 32

	expected.Security.VerifyEmail = true
	expected.Security.RecoverPass = true

	expected.Security.MustFW = config.Activated
	expected.Security.Firewall.ListType = "whitelist"
	expected.Security.Firewall.IP = "*"

	expected.Security.MustCORS = config.Activated
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

	expected.Security.TrustedPlatform = "X-Real-Ip"

	expected.Security.Must2FA = config.Activated
	expected.Security.TwoFA.Issuer = "gorest"
	expected.Security.TwoFA.Crypto = crypto.SHA1
	expected.Security.TwoFA.Digits = 6
	expected.Security.TwoFA.PathQR = "tmp"

	expected.Security.TwoFA.Status.Verified = "verified"
	expected.Security.TwoFA.Status.On = "on"
	expected.Security.TwoFA.Status.Off = "off"
	expected.Security.TwoFA.Status.Invalid = "invalid"

	expected.ViewConfig.Activate = config.Activated
	expected.ViewConfig.Directory = "templates"

	if !reflect.DeepEqual(configAll, expected) {
		t.Errorf("got: %v, want: %v", configAll, expected)
	}
}

// downloadFile will download a url and save it to a local file.
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
