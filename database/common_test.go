package database_test

import (
	"testing"

	"github.com/pilinux/gorest/config"
)

// helper function to get config with test context; fails the test if config is not initialized
func mustGetConfig(t *testing.T) *config.Configuration {
	t.Helper()

	cfg := config.GetConfig()
	if cfg == nil {
		if err := config.Config(); err != nil {
			t.Fatalf("config.Config() failed: %v", err)
		}
		cfg = config.GetConfig()
	}

	if cfg == nil {
		t.Fatal("config.GetConfig() returned nil")
	}

	return cfg
}
