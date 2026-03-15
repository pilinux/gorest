package database_test

import (
	"os"
	"testing"

	"github.com/pilinux/gorest/config"
)

// helper function to get config with test context; fails the test if config is not initialized
func mustGetConfig(t *testing.T) *config.Configuration {
	t.Helper()

	// create a minimal .env so config.Config() succeeds
	envContent := "# minimal test env\n"
	if err := os.WriteFile(".env", []byte(envContent), 0600); err != nil {
		panic("failed to create .env: " + err.Error())
	}
	// ensure .env is cleaned up after the test
	defer func() {
		_ = os.Remove(".env")
	}()

	// initialize config
	if err := config.Config(); err != nil {
		panic("config.Config() failed: " + err.Error())
	}

	// get the config; should not be nil if Config() succeeded
	cfg := config.GetConfig()
	if cfg == nil {
		t.Fatal("config.GetConfig() returned nil; TestMain should have initialized it")
	}

	return cfg
}
