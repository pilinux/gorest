package handler_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/handler"
	"github.com/pilinux/gorest/lib/middleware"
)

// TestMain sets up a minimal configuration required by the handler package.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	restoreEnv := writeTestEnv()

	if err := config.Config(); err != nil {
		restoreEnv()
		panic("config.Config() failed: " + err.Error())
	}

	code := m.Run()
	restoreEnv()

	os.Exit(code)
}

// writeTestEnv creates a minimal .env so config.Config() succeeds and returns
// a restore function. os.Exit skips defers, so the caller must invoke it.
func writeTestEnv() func() {
	orig, err := os.ReadFile(".env")
	existed := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		panic("failed to read .env: " + err.Error())
	}

	if err := os.WriteFile(".env", []byte("# minimal test env\n"), 0600); err != nil {
		panic("failed to create .env: " + err.Error())
	}

	return func() {
		if !existed {
			_ = os.Remove(".env")
			return
		}
		_ = os.WriteFile(".env", orig, 0600)
	}
}

// TestLogin_InvalidEmail - a malformed email is rejected before any lookup.
func TestLogin_InvalidEmail(t *testing.T) {
	payload := model.AuthPayload{Email: "x"}
	resp, code := handler.Login(context.Background(), payload)
	if code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", code, http.StatusBadRequest)
	}
	msg, _ := resp.Message.(string)
	if msg != "wrong email address" {
		t.Errorf("message = %q, want %q", msg, "wrong email address")
	}
}

// TestPasswordForgot_InvalidEmail - a malformed email is rejected before any lookup.
func TestPasswordForgot_InvalidEmail(t *testing.T) {
	payload := model.AuthPayload{Email: "bad"}
	resp, code := handler.PasswordForgot(context.Background(), payload)
	if code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", code, http.StatusBadRequest)
	}
	msg, _ := resp.Message.(string)
	if msg != "wrong email address" {
		t.Errorf("message = %q, want %q", msg, "wrong email address")
	}
}

// TestCreateVerificationEmail_InvalidEmail - an empty email is rejected.
func TestCreateVerificationEmail_InvalidEmail(t *testing.T) {
	payload := model.AuthPayload{Email: ""}
	resp, code := handler.CreateVerificationEmail(context.Background(), payload)
	if code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", code, http.StatusBadRequest)
	}
	msg, _ := resp.Message.(string)
	if msg != "wrong email address" {
		t.Errorf("message = %q, want %q", msg, "wrong email address")
	}
}

// TestLogout_RedisDisabled - without Redis there is no JWT blacklist,
// so logout simply succeeds.
func TestLogout_RedisDisabled(t *testing.T) {
	// ensure Redis is not activated
	cfg := config.GetConfig()
	origRedis := cfg.Database.REDIS.Activate
	cfg.Database.REDIS.Activate = ""
	defer func() { cfg.Database.REDIS.Activate = origRedis }()

	resp, code := handler.Logout(context.Background(), "jti-access", "jti-refresh", 9999999999, 9999999999)
	if code != http.StatusOK {
		t.Errorf("status = %d, want %d", code, http.StatusOK)
	}
	msg, _ := resp.Message.(string)
	if msg != "logout successful" {
		t.Errorf("message = %q, want %q", msg, "logout successful")
	}
}

// TestPasswordRecover_PasswordTooShort - a password shorter than the
// configured minimum is rejected.
func TestPasswordRecover_PasswordTooShort(t *testing.T) {
	cfg := config.GetConfig()
	origMinLen := cfg.Security.UserPassMinLength
	cfg.Security.UserPassMinLength = 10
	defer func() { cfg.Security.UserPassMinLength = origMinLen }()

	payload := model.AuthPayload{
		PassNew:    "short",
		PassRepeat: "short",
		SecretCode: "abc123",
	}
	resp, code := handler.PasswordRecover(context.Background(), payload)
	if code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", code, http.StatusBadRequest)
	}
	msg, _ := resp.Message.(string)
	expected := "password length must be greater than or equal to 10"
	if msg != expected {
		t.Errorf("message = %q, want %q", msg, expected)
	}
}

// TestPasswordRecover_PasswordMismatch - the two password fields must match.
func TestPasswordRecover_PasswordMismatch(t *testing.T) {
	cfg := config.GetConfig()
	origMinLen := cfg.Security.UserPassMinLength
	cfg.Security.UserPassMinLength = 1
	defer func() { cfg.Security.UserPassMinLength = origMinLen }()

	payload := model.AuthPayload{
		PassNew:    "password1",
		PassRepeat: "password2",
		SecretCode: "abc123",
	}
	resp, code := handler.PasswordRecover(context.Background(), payload)
	if code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", code, http.StatusBadRequest)
	}
	msg, _ := resp.Message.(string)
	if msg != "password mismatch" {
		t.Errorf("message = %q, want %q", msg, "password mismatch")
	}
}

// TestPasswordRecover_EmptySecretCode - a blank reset code is rejected.
func TestPasswordRecover_EmptySecretCode(t *testing.T) {
	cfg := config.GetConfig()
	origMinLen := cfg.Security.UserPassMinLength
	cfg.Security.UserPassMinLength = 1
	defer func() { cfg.Security.UserPassMinLength = origMinLen }()

	payload := model.AuthPayload{
		PassNew:    "password1",
		PassRepeat: "password1",
		SecretCode: "   ",
	}
	resp, code := handler.PasswordRecover(context.Background(), payload)
	if code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", code, http.StatusBadRequest)
	}
	msg, _ := resp.Message.(string)
	if msg != "password reset code is required" {
		t.Errorf("message = %q, want %q", msg, "password reset code is required")
	}
}

// TestVerifyEmail_EmptyVerificationCode - a blank verification code is rejected.
func TestVerifyEmail_EmptyVerificationCode(t *testing.T) {
	payload := model.AuthPayload{
		VerificationCode: "   ",
	}
	resp, code := handler.VerifyEmail(context.Background(), payload)
	if code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", code, http.StatusBadRequest)
	}
	msg, _ := resp.Message.(string)
	if msg != "required a valid email verification code" {
		t.Errorf("message = %q, want %q", msg, "required a valid email verification code")
	}
}

// TestVerifyUpdatedEmail_EmptyVerificationCode - a blank verification code is rejected.
func TestVerifyUpdatedEmail_EmptyVerificationCode(t *testing.T) {
	payload := model.AuthPayload{
		VerificationCode: "",
	}
	resp, code := handler.VerifyUpdatedEmail(context.Background(), payload)
	if code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", code, http.StatusBadRequest)
	}
	msg, _ := resp.Message.(string)
	if msg != "required a valid email verification code" {
		t.Errorf("message = %q, want %q", msg, "required a valid email verification code")
	}
}

// TestDeactivate2FA_ClaimEmpty - no 2FA claim means there is nothing to turn off.
func TestDeactivate2FA_ClaimEmpty(t *testing.T) {
	cfg := config.GetConfig()
	origStatus := cfg.Security.TwoFA.Status.Off
	cfg.Security.TwoFA.Status.Off = "off"
	defer func() { cfg.Security.TwoFA.Status.Off = origStatus }()

	claims := middleware.MyCustomClaims{TwoFA: ""}
	resp, code := handler.Deactivate2FA(context.Background(), claims, model.AuthPayload{})
	if code != http.StatusOK {
		t.Errorf("status = %d, want %d", code, http.StatusOK)
	}
	msg, _ := resp.Message.(string)
	if msg != "twoFA: off" {
		t.Errorf("message = %q, want %q", msg, "twoFA: off")
	}
}

// TestDeactivate2FA_ClaimOff - 2FA is already off, so the handler returns early.
func TestDeactivate2FA_ClaimOff(t *testing.T) {
	cfg := config.GetConfig()
	origStatus := cfg.Security.TwoFA.Status.Off
	cfg.Security.TwoFA.Status.Off = "off"
	defer func() { cfg.Security.TwoFA.Status.Off = origStatus }()

	claims := middleware.MyCustomClaims{TwoFA: "off"}
	resp, code := handler.Deactivate2FA(context.Background(), claims, model.AuthPayload{})
	if code != http.StatusOK {
		t.Errorf("status = %d, want %d", code, http.StatusOK)
	}
	msg, _ := resp.Message.(string)
	if msg != "twoFA: off" {
		t.Errorf("message = %q, want %q", msg, "twoFA: off")
	}
}

// TestSetup2FA_AccessDenied - a zero auth ID in the claims is denied.
func TestSetup2FA_AccessDenied(t *testing.T) {
	claims := middleware.MyCustomClaims{AuthID: 0}
	resp, code := handler.Setup2FA(context.Background(), claims, model.AuthPayload{})
	if code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", code, http.StatusUnauthorized)
	}
	msg, _ := resp.Message.(string)
	if msg != "access denied" {
		t.Errorf("message = %q, want %q", msg, "access denied")
	}
}

// TestActivate2FA_AccessDenied - a zero auth ID in the claims is denied.
func TestActivate2FA_AccessDenied(t *testing.T) {
	claims := middleware.MyCustomClaims{AuthID: 0}
	resp, code := handler.Activate2FA(context.Background(), claims, model.AuthPayload{})
	if code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", code, http.StatusUnauthorized)
	}
	msg, _ := resp.Message.(string)
	if msg != "access denied" {
		t.Errorf("message = %q, want %q", msg, "access denied")
	}
}

// TestValidate2FA_AccessDenied - a zero auth ID in the claims is denied.
func TestValidate2FA_AccessDenied(t *testing.T) {
	claims := middleware.MyCustomClaims{AuthID: 0}
	resp, code := handler.Validate2FA(context.Background(), claims, model.AuthPayload{})
	if code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", code, http.StatusUnauthorized)
	}
	msg, _ := resp.Message.(string)
	if msg != "access denied" {
		t.Errorf("message = %q, want %q", msg, "access denied")
	}
}

// TestValidateBackup2FA_AccessDenied - a zero auth ID in the claims is denied.
func TestValidateBackup2FA_AccessDenied(t *testing.T) {
	claims := middleware.MyCustomClaims{AuthID: 0}
	resp, code := handler.ValidateBackup2FA(context.Background(), claims, model.AuthPayload{})
	if code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", code, http.StatusUnauthorized)
	}
	msg, _ := resp.Message.(string)
	if msg != "access denied" {
		t.Errorf("message = %q, want %q", msg, "access denied")
	}
}

// TestCreateBackup2FA_AccessDenied - a zero auth ID in the claims is denied.
func TestCreateBackup2FA_AccessDenied(t *testing.T) {
	claims := middleware.MyCustomClaims{AuthID: 0}
	resp, code := handler.CreateBackup2FA(context.Background(), claims, model.AuthPayload{})
	if code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", code, http.StatusUnauthorized)
	}
	msg, _ := resp.Message.(string)
	if msg != "access denied" {
		t.Errorf("message = %q, want %q", msg, "access denied")
	}
}

// TestRefresh_AccessDenied - a zero auth ID in the claims is denied.
func TestRefresh_AccessDenied(t *testing.T) {
	claims := middleware.MyCustomClaims{AuthID: 0}
	resp, code := handler.Refresh(context.Background(), claims)
	if code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", code, http.StatusUnauthorized)
	}
	msg, _ := resp.Message.(string)
	if msg != "access denied" {
		t.Errorf("message = %q, want %q", msg, "access denied")
	}
}

// TestPasswordUpdate_AccessDenied - a zero auth ID in the claims is denied.
func TestPasswordUpdate_AccessDenied(t *testing.T) {
	claims := middleware.MyCustomClaims{AuthID: 0}
	resp, code := handler.PasswordUpdate(context.Background(), claims, model.AuthPayload{})
	if code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", code, http.StatusUnauthorized)
	}
	msg, _ := resp.Message.(string)
	if msg != "access denied" {
		t.Errorf("message = %q, want %q", msg, "access denied")
	}
}

// TestGetUnverifiedEmail_AccessDenied - a zero auth ID in the claims is denied.
func TestGetUnverifiedEmail_AccessDenied(t *testing.T) {
	claims := middleware.MyCustomClaims{AuthID: 0}
	resp, code := handler.GetUnverifiedEmail(context.Background(), claims)
	if code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", code, http.StatusUnauthorized)
	}
	msg, _ := resp.Message.(string)
	if msg != "access denied" {
		t.Errorf("message = %q, want %q", msg, "access denied")
	}
}

// TestResendVerification_AccessDenied - a zero auth ID in the claims is denied.
func TestResendVerification_AccessDenied(t *testing.T) {
	claims := middleware.MyCustomClaims{AuthID: 0}
	resp, code := handler.ResendVerificationCodeToModifyActiveEmail(context.Background(), claims)
	if code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", code, http.StatusUnauthorized)
	}
	msg, _ := resp.Message.(string)
	if msg != "access denied" {
		t.Errorf("message = %q, want %q", msg, "access denied")
	}
}

// TestUpdateEmail_AccessDenied - a zero auth ID in the claims is denied.
func TestUpdateEmail_AccessDenied(t *testing.T) {
	claims := middleware.MyCustomClaims{AuthID: 0}
	resp, code := handler.UpdateEmail(context.Background(), claims, model.TempEmail{})
	if code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", code, http.StatusUnauthorized)
	}
	msg, _ := resp.Message.(string)
	if msg != "access denied" {
		t.Errorf("message = %q, want %q", msg, "access denied")
	}
}
