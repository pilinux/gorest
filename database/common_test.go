package database_test

import (
	"os"
	"testing"

	"github.com/pilinux/gorest/config"
)

// TestMain sets up a minimal configuration required by the database package tests.
func TestMain(m *testing.M) {
	// create a minimal .env so config.Config() succeeds
	envContent := "# minimal test env\n"
	if err := os.WriteFile(".env", []byte(envContent), 0600); err != nil {
		panic("failed to create .env: " + err.Error())
	}
	defer func() {
		_ = os.Remove(".env")
	}()

	if err := config.Config(); err != nil {
		panic("config.Config() failed: " + err.Error())
	}

	os.Exit(m.Run())
}

// helper function to get config with test context; fails the test if config is not initialized
func mustGetConfig(t *testing.T) *config.Configuration {
	t.Helper()

	cfg := config.GetConfig()
	if cfg == nil {
		t.Fatal("config.GetConfig() returned nil; TestMain should have initialized it")
	}

	return cfg
}
