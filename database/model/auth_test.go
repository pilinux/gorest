package model_test

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/database/model"
)

const authModelHashSecret = "test-hash-secret"

func initAuthModelConfig(t *testing.T) {
	t.Helper()

	if err := os.WriteFile(".env", []byte("# auth model test env\n"), 0600); err != nil {
		t.Fatalf("failed to create .env: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(".env")
	})

	envVars := map[string]string{
		"MIN_PASS_LENGTH":     "6",
		"ACTIVATE_HASHING":    config.Activated,
		"HASHPASSMEMORY":      "64",
		"HASHPASSITERATIONS":  "2",
		"HASHPASSPARALLELISM": "2",
		"HASHPASSSALTLENGTH":  "16",
		"HASHPASSKEYLENGTH":   "32",
		"HASH_SECRET":         authModelHashSecret,
	}

	for key, value := range envVars {
		t.Setenv(key, value)
	}

	if err := config.Config(); err != nil {
		t.Fatalf("config.Config() failed: %v", err)
	}
	if config.GetConfig() == nil {
		t.Fatal("config.GetConfig() returned nil")
	}
	if config.GetConfig().Security.UserPassMinLength != 6 {
		t.Fatalf("unexpected minimum password length: got %d, want 6", config.GetConfig().Security.UserPassMinLength)
	}
}

func TestAuthUnmarshalJSON(t *testing.T) {
	initAuthModelConfig(t)

	input := []byte(`{"authID":42,"email":"  user@example.com  ","password":"s3cr3t!"}`)

	var got model.Auth
	if err := json.Unmarshal(input, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got.AuthID != 42 {
		t.Fatalf("AuthID = %d, want 42", got.AuthID)
	}
	if got.Email != "user@example.com" {
		t.Fatalf("Email = %q, want %q", got.Email, "user@example.com")
	}
	if got.Password == "" {
		t.Fatal("Password is empty, want hashed password")
	}
	if got.Password == "s3cr3t!" {
		t.Fatal("Password was not hashed")
	}

	var again model.Auth
	if err := json.Unmarshal(input, &again); err != nil {
		t.Fatalf("json.Unmarshal() second call error = %v", err)
	}
	if again.Password == got.Password {
		t.Fatal("Password hashes are identical across unmarshals, want different salted hashes")
	}
}

func TestAuthUnmarshalJSONShortPassword(t *testing.T) {
	initAuthModelConfig(t)

	input := []byte(`{"authID":42,"email":"user@example.com","password":"12345"}`)

	var got model.Auth
	err := json.Unmarshal(input, &got)
	if err == nil {
		t.Fatal("json.Unmarshal() error = nil, want short password error")
	}
	if err.Error() != "short password" {
		t.Fatalf("json.Unmarshal() error = %q, want %q", err.Error(), "short password")
	}
	if got != (model.Auth{}) {
		t.Fatalf("Auth after failed unmarshal = %#v, want zero value", got)
	}
}

func TestAuthUnmarshalJSONInvalidJSON(t *testing.T) {
	var got model.Auth
	err := got.UnmarshalJSON([]byte(`{"authID":42`))
	if err == nil {
		t.Fatal("UnmarshalJSON() error = nil, want syntax error")
	}

	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("UnmarshalJSON() error = %T, want *json.SyntaxError", err)
	}
}

func TestAuthMarshalJSON(t *testing.T) {
	auth := model.Auth{
		AuthID:      42,
		Email:       "  user@example.com  ",
		Password:    "s3cr3t!",
		VerifyEmail: model.EmailVerified,
	}

	got, err := json.Marshal(auth)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	want := `{"authID":42,"email":"user@example.com"}`
	if string(got) != want {
		t.Fatalf("json.Marshal() = %s, want %s", got, want)
	}
}
