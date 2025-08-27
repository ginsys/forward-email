// Package testutil provides testing utilities and helpers for the CLI application.
package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// SetupTempConfig creates a temporary config directory and sets the XDG_CONFIG_HOME
// environment variable to point to it. This ensures tests use isolated config
// directories and don't interfere with the user's actual configuration.
//
// Returns the temporary directory path for additional setup if needed.
// The environment variable and temporary directory are automatically cleaned up
// when the test completes.
func SetupTempConfig(t *testing.T) string {
	t.Helper()

	// Create temporary directory
	tempDir := t.TempDir()

	// Create the forwardemail config directory structure
	configDir := filepath.Join(tempDir, ".config", "forwardemail")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		t.Fatalf("Failed to create temp config directory: %v", err)
	}

	// Set XDG_CONFIG_HOME using t.Setenv for safe cleanup
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))

	return tempDir
}

// WriteTestConfig writes a config file to the specified config directory.
// This is a helper to simplify test setup.
func WriteTestConfig(t *testing.T, tempDir, content string) {
	t.Helper()

	configDir := filepath.Join(tempDir, ".config", "forwardemail")
	configFile := filepath.Join(configDir, "config.yaml")

	if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}
}

// ResetViper resets the viper configuration to ensure tests start with a clean state.
// This should be called at the beginning of each test that uses viper.
func ResetViper() {
	viper.Reset()
}

// SetupTempConfigWithReset sets up temporary config directory and resets viper state.
// This is useful when tests need both viper reset and temp config setup.
func SetupTempConfigWithReset(t *testing.T) string {
	t.Helper()
	ResetViper()
	return SetupTempConfig(t)
}
