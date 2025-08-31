// Package keyring provides secure credential storage and retrieval functionality.
package keyring

import (
	"fmt"
	"os"

	"github.com/99designs/keyring"
)

// Constants for keyring service identification and testing.
const (
	ServiceName   = "forward-email" // Service name for OS keyring registration
	testPassConst = "test-password" // Test password for keyring validation
)

// Keyring provides a secure wrapper around the OS keyring functionality.
// It offers cross-platform secure credential storage using the system's native
// keyring service (Windows Credential Manager, macOS Keychain, Linux Secret Service).
type Keyring struct {
	ring keyring.Keyring // The underlying keyring implementation
}

// Config represents configuration options for keyring initialization.
// All fields are optional; defaults will be applied for empty values.
// This configuration is passed through to the underlying keyring library.
type Config struct {
	ServiceName              string                // Service name in the OS keyring (default: "forward-email")
	KeychainName             string                // macOS keychain name (optional)
	FileDir                  string                // Directory for file-based fallback storage
	FilePasswordFunc         keyring.PromptFunc    // Password prompt function for file encryption
	AllowedBackends          []keyring.BackendType // Restrict keyring backend types
	KeychainTrustApplication bool                  // macOS keychain trust setting
}

// New creates a new keyring instance with the specified configuration.
// It initializes the OS keyring service and returns a wrapper that provides
// secure credential storage and retrieval. Falls back to file-based storage
// if the OS keyring is unavailable.
func New(config Config) (*Keyring, error) {
	// Allow environment override to control backend selection without GUI prompts.
	// FORWARDEMAIL_KEYRING_BACKEND values:
	//   os   - use system default keyring (default)
	//   file - use encrypted file backend (requires FORWARDEMAIL_KEYRING_PASSWORD)
	//   none - disable keyring usage (caller should fallback to config/env)
	if backend := os.Getenv("FORWARDEMAIL_KEYRING_BACKEND"); backend != "" && len(config.AllowedBackends) == 0 {
		switch backend {
		case "none":
			return nil, fmt.Errorf("keyring disabled by FORWARDEMAIL_KEYRING_BACKEND=none")
		case "file":
			// Configure file backend from env for non-interactive setups
			pass := os.Getenv("FORWARDEMAIL_KEYRING_PASSWORD")
			config.AllowedBackends = []keyring.BackendType{keyring.FileBackend}
			if config.FilePasswordFunc == nil {
				config.FilePasswordFunc = func(string) (string, error) { return pass, nil }
			}
		}
	}

	if config.ServiceName == "" {
		config.ServiceName = ServiceName
	}

	keyringConfig := keyring.Config{
		ServiceName:              config.ServiceName,
		KeychainName:             config.KeychainName,
		FileDir:                  config.FileDir,
		FilePasswordFunc:         config.FilePasswordFunc,
		AllowedBackends:          config.AllowedBackends,
		KeychainTrustApplication: config.KeychainTrustApplication,
	}

	ring, err := keyring.Open(keyringConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}

	return &Keyring{ring: ring}, nil
}

// SetAPIKey stores an API key for the given profile
func (k *Keyring) SetAPIKey(profile, apiKey string) error {
	key := fmt.Sprintf("api_key_%s", profile)

	item := keyring.Item{
		Key:         key,
		Data:        []byte(apiKey),
		Label:       fmt.Sprintf("Forward Email API Key (%s)", profile),
		Description: fmt.Sprintf("API key for Forward Email CLI profile: %s", profile),
	}

	if err := k.ring.Set(item); err != nil {
		return fmt.Errorf("failed to store API key for profile %s: %w", profile, err)
	}

	return nil
}

// GetAPIKey retrieves an API key for the given profile
func (k *Keyring) GetAPIKey(profile string) (string, error) {
	key := fmt.Sprintf("api_key_%s", profile)

	item, err := k.ring.Get(key)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return "", fmt.Errorf("API key not found for profile %s", profile)
		}
		return "", fmt.Errorf("failed to retrieve API key for profile %s: %w", profile, err)
	}

	return string(item.Data), nil
}

// DeleteAPIKey removes an API key for the given profile
func (k *Keyring) DeleteAPIKey(profile string) error {
	key := fmt.Sprintf("api_key_%s", profile)

	if err := k.ring.Remove(key); err != nil {
		if err == keyring.ErrKeyNotFound {
			return fmt.Errorf("API key not found for profile %s", profile)
		}
		return fmt.Errorf("failed to delete API key for profile %s: %w", profile, err)
	}

	return nil
}

// ListProfiles returns all profiles that have API keys stored
func (k *Keyring) ListProfiles() ([]string, error) {
	keys, err := k.ring.Keys()
	if err != nil {
		return nil, fmt.Errorf("failed to list keyring keys: %w", err)
	}

	var profiles []string
	prefix := "api_key_"

	for _, key := range keys {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			profile := key[len(prefix):]
			profiles = append(profiles, profile)
		}
	}

	return profiles, nil
}

// HasAPIKey checks if an API key exists for the given profile
func (k *Keyring) HasAPIKey(profile string) bool {
	_, err := k.GetAPIKey(profile)
	return err == nil
}

// MockKeyring creates an in-memory keyring for testing
func MockKeyring() (*Keyring, error) {
	// Create a temporary directory for file-based keyring in tests
	tmpDir, err := os.MkdirTemp("", "forwardemail-test-keyring-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	config := Config{
		ServiceName:      ServiceName,
		AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
		FileDir:          tmpDir,
		FilePasswordFunc: func(string) (string, error) { return testPassConst, nil },
	}

	return New(config)
}
