# AGENTS.md

Instructions for AI coding agents working in this repository.

## Project Overview

Go RESTful API starter kit built with Gin, GORM, JWT, Redis, MongoDB,
2FA, email verification, and password recovery. Module path:
`github.com/pilinux/gorest`. Requires Go 1.24.1+.

## Build and Run Commands

### Build

```bash
go build -v ./...
```

### Tidy Dependencies

```bash
go mod tidy
```

### Format

```bash
go fmt ./...
```

### Vet

```bash
go vet -v ./...
```

### Lint

```bash
golangci-lint run ./...
revive ./...
```

CI uses golangci-lint v2.8.0 with `--timeout 5m --verbose`.

### Test - All

```bash
go test -v -cover ./...
```

### Test - Single Package

```bash
go test -v -cover ./lib/...
go test -v -cover ./config/...
```

### Test - Single Test Function

```bash
go test -v -run TestHashPass ./lib/...
go test -v -run TestFirewall ./lib/middleware/...
```

### Test Environment

Tests require environment variables. In CI these come from secrets.
Locally, source `setTestEnv.sh` before running tests:

```bash
source setTestEnv.sh
```

### Cross-Platform Vet (CI runs all six)

```bash
GOOS=linux GOARCH=amd64 go vet -v ./...
GOOS=darwin GOARCH=arm64 go vet -v ./...
```

## Project Structure

- `config/` - Configuration loading from `.env` files
- `controller/` - HTTP request handlers (thin layer, calls handler)
- `handler/` - Business logic layer (called by controller)
- `service/` - Utility services (auth, security, email)
- `database/` - DB connection management and models
- `database/model/` - Shared data models (Auth, HTTPResponse)
- `database/migrate/` - GORM auto-migration
- `lib/` - Core utility library (hashing, encryption, validation)
- `lib/middleware/` - Gin middleware (JWT, CORS, firewall, rate limit)
- `lib/renderer/` - HTTP response rendering (JSON and HTML)
- `lib/server/` - Graceful server shutdown
- `example/` - Legacy example app
- `example2/` - Recommended example app (interface-driven, DI)

### example2/ Structure (Recommended)

```text
example2/
├── cmd/app/main.go          # Entry point
└── internal/
    ├── database/
    │   ├── migrate/migrate.go  # Auto-migration
    │   └── model/models.go     # App-specific GORM models
    ├── handler/                # HTTP handlers (Gin context binding)
    ├── repo/                   # Repository pattern (data access)
    ├── router/router.go        # Route definitions + middleware setup
    └── service/                # Business logic
```

### Layer Responsibilities

| Layer | Package | Responsibility |
| ----- | ------- | -------------- |
| Controller | `controller/` | Bind request, call handler, render response |
| Handler | `handler/` | Business logic, validation, returns `(HTTPResponse, int)` |
| Service | `service/` | Shared utilities (auth, email, crypto, JWT blacklist) |
| Repository | `example2/internal/repo/` | Data access abstraction (interface-driven) |
| Database | `database/` | Connection management (RDBMS, Redis, MongoDB) |
| Config | `config/` | Load `.env`, expose `GetConfig()`, feature checks |

### Controller-Handler 1:1 Mapping

Every controller function in `controller/` is a thin wrapper that binds
the HTTP request and calls the corresponding function in `handler/`.
Controllers use `renderer.Render()` to send the response. There are
19 exported functions in each package with a 1:1 correspondence:

| Controller | Handler | Description |
| ---------- | ------- | ----------- |
| `CreateUserAuth` | `CreateUserAuth(auth model.Auth)` | User registration |
| `Login` | `Login(payload model.AuthPayload)` | User login |
| `Refresh` | `Refresh(claims middleware.MyCustomClaims)` | Refresh JWT |
| `Logout` | `Logout(jtiAccess, jtiRefresh string, expAccess, expRefresh int64)` | Invalidate tokens |
| `PasswordForgot` | `PasswordForgot(authPayload model.AuthPayload)` | Send recovery email |
| `PasswordRecover` | `PasswordRecover(authPayload model.AuthPayload)` | Reset password |
| `PasswordUpdate` | `PasswordUpdate(claims, authPayload)` | Change password |
| `Setup2FA` | `Setup2FA(claims, authPayload)` | Generate 2FA secret |
| `Activate2FA` | `Activate2FA(claims, authPayload)` | Enable 2FA |
| `Validate2FA` | `Validate2FA(claims, authPayload)` | Verify OTP |
| `Deactivate2FA` | `Deactivate2FA(claims, authPayload)` | Disable 2FA |
| `CreateBackup2FA` | `CreateBackup2FA(claims, authPayload)` | Generate backup codes |
| `ValidateBackup2FA` | `ValidateBackup2FA(claims, authPayload)` | Use backup code |
| `VerifyEmail` | `VerifyEmail(payload model.AuthPayload)` | Verify email |
| `CreateVerificationEmail` | `CreateVerificationEmail(payload model.AuthPayload)` | Resend verification |
| `VerifyUpdatedEmail` | `VerifyUpdatedEmail(payload model.AuthPayload)` | Verify email change |
| `UpdateEmail` | `UpdateEmail(claims, req model.TempEmail)` | Submit new email |
| `GetUnverifiedEmail` | `GetUnverifiedEmail(claims)` | Get pending email |
| `ResendVerificationCodeToModifyActiveEmail` | `ResendVerificationCodeToModifyActiveEmail(claims)` | Resend email change code |

All handler functions return `(httpResponse model.HTTPResponse, httpStatusCode int)`.

## Import Alias Conventions

Always use these aliases when importing gorest packages:

```go
import (
    gconfig "github.com/pilinux/gorest/config"
    gcontroller "github.com/pilinux/gorest/controller"
    gdb "github.com/pilinux/gorest/database"
    gmodel "github.com/pilinux/gorest/database/model"
    ghandler "github.com/pilinux/gorest/handler"
    glib "github.com/pilinux/gorest/lib"
    gmiddleware "github.com/pilinux/gorest/lib/middleware"
    grenderer "github.com/pilinux/gorest/lib/renderer"
    gserver "github.com/pilinux/gorest/lib/server"
    gservice "github.com/pilinux/gorest/service"
)
```

Also alias logrus: `log "github.com/sirupsen/logrus"`

## Code Style Guidelines

### Imports

Organize imports in three groups separated by blank lines:

1. Standard library
2. External packages (use aliases where conventional)
3. Internal packages (`github.com/pilinux/gorest/...`)

```go
import (
    "errors"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"

    "github.com/pilinux/gorest/config"
    "github.com/pilinux/gorest/database/model"
)
```

### Naming Conventions

- **Exported** functions/types: `PascalCase` (e.g., `CreateUserAuth`)
- **Unexported** functions/vars: `camelCase` (e.g., `authFinal`)
- **Acronyms** stay uppercase: `JWT`, `CORS`, `RDBMS`, `HTTP`, `2FA`
- **Packages**: lowercase, single word (e.g., `middleware`, `renderer`)
- **Test structs**: camelCase (e.g., `hashPassTest`, `hashVerifyTest`)
- **Error variables**: prefix with `err` (e.g., `errThis`, `errHash`)

### Struct Tags

Use `json:"fieldName,omitempty"` for JSON tags and
`gorm:"primaryKey"` style for GORM tags. Hide sensitive fields with
`json:"-"`. Mark non-DB fields with `gorm:"-"`.

```go
type MyCustomClaims struct {
    AuthID  uint64 `json:"authID,omitempty"`
    Email   string `json:"email,omitempty"`
    Role    string `json:"role,omitempty"`
}
```

### Function Signatures

Handlers return `(model.HTTPResponse, int)` where the int is the
HTTP status code. Use named return values with bare returns for
early exits:

```go
func CreateUserAuth(
    auth model.Auth,
) (httpResponse model.HTTPResponse, httpStatusCode int) {
    // ...
    httpResponse.Message = "wrong email address"
    httpStatusCode = http.StatusBadRequest
    return
}
```

### Error Handling

- Log errors with `log.WithError(err).Error("error code: XXXX.X")`
- Use numbered error codes (e.g., `1002.1`, `1002.2`) for tracing
- Return user-facing messages via `httpResponse.Message`
- Never expose internal errors to API consumers
- Use `errors.New()` for simple errors, `fmt.Errorf` with `%w` for
  wrapping

```go
if err != nil {
    log.WithError(err).Error("error code: 1002.1")
    httpResponse.Message = "internal server error"
    httpStatusCode = http.StatusInternalServerError
    return
}
```

### HTTP Responses

Always use `model.HTTPResponse` for API responses. Use
`renderer.Render()` to send responses in controllers, which handles
both JSON and HTML template rendering:

```go
renderer.Render(c, data, http.StatusOK)
renderer.Render(c, data, http.StatusBadRequest)
```

### Comments

- Package doc comments on every package: `// Package name - description`
- Function doc comments: `// FuncName - description`
- Copyright headers where applicable:
  `// The MIT License (MIT)`
  `// Copyright (c) 20XX pilinux`
- Use numbered step comments for complex logic

### Testing

- Use external test packages (`package lib_test`, not `package lib`)
- Table-driven tests with named struct types
- Use `t.Run()` for subtests with descriptive names
- Use `t.Errorf` for assertions (no external assertion library)
- Test helper functions should be unexported

```go
func TestHashPass(t *testing.T) {
    tests := []hashPassTest{
        {name: "blank password", ...},
        {name: "with password", ...},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            // ...
        })
    }
}
```

## Core Models and Types

### Key Models (database/model)

```go
// Auth - users table (has custom JSON marshal/unmarshal)
type Auth struct {
    AuthID    uint64         `gorm:"primaryKey" json:"authID,omitempty"`
    Email     string         `gorm:"index" json:"email"`
    Password  string         `json:"password"`
    // EmailCipher, EmailNonce, EmailHash: hidden from JSON (json:"-")
    // VerifyEmail: int8 status field (json:"-")
}

// AuthPayload - registration/login request body
type AuthPayload struct {
    Email            string `json:"email,omitempty"`
    Password         string `json:"password,omitempty"`
    VerificationCode string `json:"verificationCode,omitempty"`
    OTP              string `json:"otp,omitempty"`
    SecretCode       string `json:"secretCode,omitempty"`
    RecoveryKey      string `json:"recoveryKey,omitempty"`
    PassNew          string `json:"passNew,omitempty"`
    PassRepeat       string `json:"passRepeat,omitempty"`
}

// HTTPResponse - standard API response wrapper
type HTTPResponse struct {
    Message any `json:"message,omitempty"`
}

// JWTPayload - token pair response
type JWTPayload struct {
    AccessJWT   string `json:"accessJWT,omitempty"`
    RefreshJWT  string `json:"refreshJWT,omitempty"`
    TwoAuth     string `json:"twoFA,omitempty"`
    RecoveryKey string `json:"recoveryKey,omitempty"`
}

// TempEmail - holds data during email replacement
type TempEmail struct {
    Email    string `gorm:"index" json:"emailNew"`
    Password string `gorm:"-" json:"password,omitempty"`
    // + ID, timestamps, cipher fields, IDAuth
}
```

### Key Constants (database/model)

```go
// Email verification statuses
const EmailNotVerified int8 = -1
const EmailVerifyNotRequired int8 = 0
const EmailVerified int8 = 1

// Email type constants (for SendEmail)
const EmailTypeVerifyEmailNewAcc int = 1
const EmailTypePassRecovery int = 2
const EmailTypeVerifyUpdatedEmail int = 3
```

### Key Constants (config)

```go
const Activated string = "yes"
const PrefixJtiBlacklist string = "gorest-blacklist-jti:"
```

### Key Constants (database)

```go
const RecordNotFound string = "record not found"
var RedisConnTTL int // context deadline in seconds for Redis connections
```

## Configuration System

### Loading

```go
gconfig.Env()       // loads .env file via godotenv.Load()
gconfig.Config()    // loads all config from env vars into Configuration struct
gconfig.GetConfig() // returns *Configuration
```

### Feature Check Functions

These return `bool` and are used to conditionally enable features:

```go
gconfig.IsProd()       gconfig.IsRDBMS()      gconfig.IsRedis()
gconfig.IsMongo()      gconfig.IsJWT()        gconfig.Is2FA()
gconfig.IsCORS()       gconfig.IsWAF()        gconfig.IsRateLimit()
gconfig.IsSentry()     gconfig.IsCipher()      gconfig.IsHashPass()
gconfig.IsBasicAuth()  gconfig.IsAuthCookie()  gconfig.InvalidateJWT()
gconfig.IsOriginCheck()          gconfig.IsTemplatingEngine()
gconfig.IsEmailService()         gconfig.IsEmailVerificationService()
gconfig.IsPassRecoveryService()  gconfig.Is2FADoubleHash()
gconfig.IsEmailVerificationCodeUUIDv4()
gconfig.IsPasswordRecoverCodeUUIDv4()
```

## Database Initialization Patterns

### RDBMS

```go
// Production pattern uses retry loop
for {
    if err := gdb.InitDB().Error; err != nil {
        fmt.Println(err)
        time.Sleep(10 * time.Second)
        continue
    }
    break
}
db := gdb.GetDB()        // get *gorm.DB connection
gdb.InitTLSMySQL()       // TLS setup for MySQL
```

### Redis

```go
for {
    if _, err := gdb.InitRedis(); err != nil {
        fmt.Println(err)
        time.Sleep(10 * time.Second)
        continue
    }
    break
}
client := gdb.GetRedis() // get radix.Client
```

### MongoDB

```go
for {
    if _, err := gdb.InitMongo(); err != nil {
        fmt.Println(err)
        time.Sleep(10 * time.Second)
        continue
    }
    break
}
client := gdb.GetMongo() // get *qmgo.Client

// Index management
gdb.MongoCreateIndex("db", "collection", index)
gdb.MongoCreateIndexes("db", "collection", indexes)
gdb.MongoDropIndex("db", "collection", []string{"field"})
gdb.MongoDropAllIndexes("db", "collection")
```

### Closing

```go
gdb.CloseSQL()     // close RDBMS
gdb.CloseRedis()   // close Redis
gdb.CloseMongo()   // close MongoDB
gdb.CloseAllDB()   // close all (returns error)
```

## Middleware Patterns

### JWT

```go
router.Use(gmiddleware.JWT())           // validate access token
router.Use(gmiddleware.JWT("__session")) // custom cookie name
router.Use(gmiddleware.RefreshJWT())    // validate refresh token
```

### CORS

```go
router.Use(gmiddleware.CORS(cfg.Security.CORS))
gmiddleware.GetCORS()    // returns CORSConfig struct
gmiddleware.ResetCORS()  // reset configuration
```

### Firewall

```go
router.Use(gmiddleware.Firewall("whitelist", "192.168.1.0/24, 10.0.0.1"))
router.Use(gmiddleware.Firewall("blacklist", "192.168.100.0/24"))
router.Use(gmiddleware.Firewall("whitelist", "*")) // allow all
gmiddleware.ResetFirewallState() // reset (useful for testing)
```

### 2FA

```go
// Requires JWT middleware first
router.Use(gmiddleware.TwoFA("on", "off", "verified"))
```

### Rate Limiting

```go
limiter, _ := glib.InitRateLimiter("100-M", "X-Real-Ip") // 100/minute
router.Use(gmiddleware.RateLimit(limiter))
```

### Origin Check

```go
router.Use(gmiddleware.CheckOrigin([]string{"https://example.com"}))
```

### Sentry

```go
gmiddleware.InitSentry(dsn, env, version, enableTracing, sampleRate)
gmiddleware.NewSentryHook(dsn, env, ...) // for goroutine-specific loggers
gmiddleware.DestroySentry()
router.Use(gmiddleware.SentryCapture())
```

### Template Rendering

```go
router.Use(gmiddleware.Pongo2("templates/"))
```

## Service Layer Quick Reference

```go
// Auth
gservice.GetClaims(c)                    // extract JWT claims from Gin context
gservice.GetUserByEmail(email, bool)     // look up user
gservice.GetEmailByAuthID(id)            // get email by auth ID
gservice.IsAuthIDValid(id)               // check existence in DB
gservice.ValidateAuthID(id)              // checks zero + DB existence
gservice.ValidateUserID(id, email)       // checks authID != 0 and email != ""

// JWT blacklist
gservice.JWTBlacklistChecker()           // middleware
gservice.IsTokenAllowed(jti)             // check if token is blacklisted

// Email
gservice.SendEmail(email, emailType, opts...)  // send via Postmark
gservice.Postmark(params)               // low-level Postmark delivery

// Crypto
gservice.DecryptEmail(nonce, ciphertext) // decrypt stored email
gservice.CalcHash(data, secret)          // BLAKE2b hash
gservice.GetHash(password)              // hash for 2FA key encryption
gservice.RandomByte(n)                  // secure random bytes
gservice.GenerateCode(n)               // random alphanumeric code

// 2FA
gservice.Validate2FA(secret, issuer, otp)
gservice.DelMem2FA(authID)
```

## Utility Library Quick Reference

```go
// Encryption (AES-GCM)
glib.Encrypt(plaintext, key)
glib.Decrypt(ciphertext, key)

// Password hashing (Argon2id)
glib.HashPass(config, password, pepper)

// Key derivation
glib.GetArgon2Key(password, salt, keyLen)

// TOTP / QR
glib.NewTOTP(account, issuer, hash, digits)
glib.NewQR(otpBytes, issuer)
glib.ValidateTOTP(otpBytes, issuer, otp)
glib.ByteToPNG(qrPNG, path)

// Validation
glib.ValidateEmail(email)               // MX record check
glib.ValidatePath(path, allowedDir)     // prevent traversal
glib.FileExist(path)

// Random
glib.SecureRandomNumber(digits)

// String helpers
glib.RemoveAllSpace(s)
glib.StrArrHTMLModel(s)                // split "k:v;k:v" into []string
glib.HTMLModel(strArr)                  // []string to map[string]any

// Rate limiter
glib.InitRateLimiter(rate, trustedPlatform)
```

## Graceful Shutdown

```go
// Full signature:
// GracefulShutdown(srv *http.Server, timeout time.Duration, done chan struct{}, closeDB ...func() error) error
done := make(chan struct{})
go func() {
    err := gserver.GracefulShutdown(srv, 30*time.Second, done, gdb.CloseAllDB)
    if err != nil {
        fmt.Println(err)
    }
}()
```

## HTTP Response Rendering

```go
// JSON response
grenderer.Render(c, data, http.StatusOK)

// HTML template response (requires Pongo2 middleware)
grenderer.Render(c, data, http.StatusOK, "template.html")
```

## Files to Never Read or Modify

Per `.gitignore`, avoid these paths:

- `*.log`, `.env`, `*.bak`
- `vendor/`, `keys/`, `tmp/`, `.vscode/`, `.build/`
- `coverage.txt`
- `pri*key*.pem`, `pub*key*.pem`
- `crosscompile*.sh`

## CI Pipeline

GitHub Actions runs on push/PR to `main`:

- `go vet` on 6 platform combinations
- `gosec` security scanner
- `govulncheck` for known vulnerabilities
- Build on 6 platforms (linux/darwin/windows x amd64/arm64)
- Tests with coverage (push only), uploaded to Codecov
- `golangci-lint` v2.8.0

## Contributing

- PRs target the `main` branch
- One commit per unique task
- Test and validate code before committing
- Document new features in `README.md`
