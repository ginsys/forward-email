package keyring

import (
	"fmt"
	"os"

	"github.com/99designs/keyring"
)

const (
	ServiceName = "forward-email"
)

// Keyring provides a wrapper around the keyring library
type Keyring struct {
	ring keyring.Keyring
}

// Config represents keyring configuration
type Config struct {
	ServiceName              string
	KeychainName             string
	FileDir                  string
	FilePasswordFunc         keyring.PromptFunc
	AllowedBackends          []keyring.BackendType
	KeychainTrustApplication bool
}

// New creates a new keyring instance
func New(config Config) (*Keyring, error) {
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
		ServiceName:     ServiceName,
		AllowedBackends: []keyring.BackendType{keyring.FileBackend},
		FileDir:         tmpDir,
		FilePasswordFunc: func(string) (string, error) {
			return "test-password", nil
		},
	}

	return New(config)
}
