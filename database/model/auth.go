// Package model contains all the models required
// for a functional database management system.
package model

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/lib"
)

// Email verification statuses
const (
	EmailNotVerified       int8 = -1
	EmailVerifyNotRequired int8 = 0
	EmailVerified          int8 = 1
)

// Email type
const (
	EmailTypeVerifyEmailNewAcc  int = 1 // verify email of newly registered user
	EmailTypePassRecovery       int = 2 // password recovery code
	EmailTypeVerifyUpdatedEmail int = 3 // verify request of updating user email
)

// Redis key prefixes
const (
	EmailVerificationKeyPrefix string = "gorest-email-verification-"
	EmailUpdateKeyPrefix       string = "gorest-email-update-"
	PasswordRecoveryKeyPrefix  string = "gorest-pass-recover-"     // #nosec G101 —- Redis key prefix only; contains no secret credentials
	Backup2FALimitKeyPrefix    string = "gorest-2fa-backup-limit-" // #nosec G101 —- Redis key prefix only; contains no secret credentials
)

// Password recovery is stored as a Redis hash.
const (
	PasswordRecoveryFieldEmail    string = "email"    // hash field holding the (hashed) email
	PasswordRecoveryFieldAttempts string = "attempts" // hash field holding the failed-attempt count
)

// 2FA backup-code brute-force limiter is stored as a Redis hash.
const (
	Backup2FAFieldAttempts      string = "attempts"       // hash field holding the failed-attempt count
	Backup2FAFieldCooldownUntil string = "cooldown_until" // hash field holding the unix time until which attempts are blocked
)

// Auth represents the auths table.
type Auth struct {
	AuthID      uint64         `gorm:"primaryKey" json:"authID,omitempty"`
	CreatedAt   time.Time      `json:"createdAt,omitzero"`
	UpdatedAt   time.Time      `json:"updatedAt,omitzero"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Email       string         `gorm:"index" json:"email"`
	EmailCipher string         `json:"-"`
	EmailNonce  string         `json:"-"`
	EmailHash   string         `gorm:"index" json:"-"`
	Password    string         `json:"password"` // #nosec G117 -- used for input only, custom MarshalJSON excludes it from output
	VerifyEmail int8           `json:"-"`
}

// UnmarshalJSON implements the json.Unmarshaler interface for Auth.
func (v *Auth) UnmarshalJSON(b []byte) error {
	aux := struct {
		AuthID   uint64 `json:"authID"`
		Email    string `json:"email"`
		Password string `json:"password"` // #nosec G117 -- internal unmarshal helper, password is hashed before storage
	}{}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	configSecurity := config.GetConfig().Security

	// check password length
	// if more checks are required i.e. password pattern,
	// add all conditions here
	if len(aux.Password) < configSecurity.UserPassMinLength {
		return errors.New("short password")
	}

	v.AuthID = aux.AuthID
	v.Email = strings.TrimSpace(aux.Email)

	config := lib.HashPassConfig{
		Memory:      configSecurity.HashPass.Memory,
		Iterations:  configSecurity.HashPass.Iterations,
		Parallelism: configSecurity.HashPass.Parallelism,
		SaltLength:  configSecurity.HashPass.SaltLength,
		KeyLength:   configSecurity.HashPass.KeyLength,
	}
	pass, err := lib.HashPass(config, aux.Password, configSecurity.HashSec)
	if err != nil {
		return err
	}
	v.Password = pass

	return nil
}

// MarshalJSON implements the json.Marshaler interface for Auth.
func (v Auth) MarshalJSON() ([]byte, error) {
	aux := struct {
		AuthID uint64 `json:"authID"`
		Email  string `json:"email"`
	}{
		AuthID: v.AuthID,
		Email:  strings.TrimSpace(v.Email),
	}

	return json.Marshal(aux)
}

// AuthPayload holds all auth-related request data.
type AuthPayload struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"` // #nosec G117 -- request payload only, never stored or returned with sensitive data

	VerificationCode string `json:"verificationCode,omitempty"`

	OTP string `json:"otp,omitempty"`

	SecretCode  string `json:"secretCode,omitempty"`
	RecoveryKey string `json:"recoveryKey,omitempty"`

	PassNew    string `json:"passNew,omitempty"`
	PassRepeat string `json:"passRepeat,omitempty"`
}

// TempEmail represents the temp_emails table used to hold data temporarily
// during the process of replacing a user's email address
// with a new one.
type TempEmail struct {
	ID          uint64    `gorm:"primaryKey" json:"-"`
	CreatedAt   time.Time `json:"createdAt,omitzero"`
	UpdatedAt   time.Time `json:"updatedAt,omitzero"`
	Email       string    `gorm:"index" json:"emailNew"`
	Password    string    `gorm:"-" json:"password,omitempty"` // #nosec G117 -- value not from DB, only used for validation
	EmailCipher string    `json:"-"`
	EmailNonce  string    `json:"-"`
	EmailHash   string    `gorm:"index" json:"-"`
	IDAuth      uint64    `gorm:"index" json:"-"`
}
