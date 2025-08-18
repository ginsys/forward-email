package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ginsys/forward-email/internal/keyring"
	"github.com/ginsys/forward-email/pkg/config"
)

// Provider defines the interface for authentication
type Provider interface {
	Apply(req *http.Request) error
	Validate(ctx context.Context) error
	GetAPIKey() (string, error)
}

// ExtendedProvider extends Provider with credential management
type ExtendedProvider interface {
	Provider
	SetAPIKey(apiKey string) error
	DeleteAPIKey() error
	HasAPIKey() bool
}

// ForwardEmailAuth implements authentication for Forward Email API
type ForwardEmailAuth struct {
	profile string
	config  *config.Config
	keyring *keyring.Keyring
}

// ProviderConfig holds configuration for creating auth providers
type ProviderConfig struct {
	Profile string
	Config  *config.Config
	Keyring *keyring.Keyring
}

// NewProvider creates a new authentication provider
func NewProvider(cfg ProviderConfig) (Provider, error) {
	if cfg.Profile == "" {
		cfg.Profile = "default"
	}

	if cfg.Config == nil {
		var err error
		cfg.Config, err = config.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return &ForwardEmailAuth{
		profile: cfg.Profile,
		config:  cfg.Config,
		keyring: cfg.Keyring,
	}, nil
}

// Apply adds authentication headers to the request
func (f *ForwardEmailAuth) Apply(req *http.Request) error {
	apiKey, err := f.GetAPIKey()
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}

	// Forward Email uses HTTP Basic Auth with API key as username and empty password
	auth := base64.StdEncoding.EncodeToString([]byte(apiKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)

	return nil
}

// Validate checks if the current credentials are valid
func (f *ForwardEmailAuth) Validate(ctx context.Context) error {
	apiKey, err := f.GetAPIKey()
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}

	if apiKey == "" {
		return fmt.Errorf("API key is empty")
	}

	// Create a test request to validate credentials
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.forwardemail.net/v1/account", http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	// Apply authentication
	if err2 := f.Apply(req); err2 != nil {
		return fmt.Errorf("failed to apply authentication: %w", err2)
	}

	// Make the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("validation failed with status %d", resp.StatusCode)
	}

	return nil
}

// GetAPIKey retrieves the API key from the configured sources in priority order:
// 1. Environment variable (FORWARDEMAIL_API_KEY or FORWARDEMAIL_<PROFILE>_API_KEY)
// 2. OS Keyring
// 3. Configuration file
func (f *ForwardEmailAuth) GetAPIKey() (string, error) {
	// 1. Check environment variables (highest priority)
	if apiKey := f.getAPIKeyFromEnv(); apiKey != "" {
		return apiKey, nil
	}

	// 2. Check OS keyring (medium priority)
	if f.keyring != nil {
		if apiKey, err := f.keyring.GetAPIKey(f.profile); err == nil {
			return apiKey, nil
		}
	}

	// 3. Check configuration file (lowest priority)
	if apiKey, err := f.getAPIKeyFromConfig(); err == nil && apiKey != "" {
		return apiKey, nil
	}

	return "", fmt.Errorf("no API key found for profile %s", f.profile)
}

// getAPIKeyFromEnv retrieves API key from environment variables
func (f *ForwardEmailAuth) getAPIKeyFromEnv() string {
	// Try profile-specific environment variable first
	profileEnv := fmt.Sprintf("FORWARDEMAIL_%s_API_KEY", strings.ToUpper(f.profile))
	if apiKey := os.Getenv(profileEnv); apiKey != "" {
		return apiKey
	}

	// Try generic environment variable
	return os.Getenv("FORWARDEMAIL_API_KEY")
}

// getAPIKeyFromConfig retrieves API key from configuration file
func (f *ForwardEmailAuth) getAPIKeyFromConfig() (string, error) {
	profile, err := f.config.GetProfile(f.profile)
	if err != nil {
		return "", fmt.Errorf("failed to get profile %s: %w", f.profile, err)
	}

	return profile.APIKey, nil
}

// SetAPIKey stores an API key for the current profile
func (f *ForwardEmailAuth) SetAPIKey(apiKey string) error {
	// Store in keyring if available
	if f.keyring != nil {
		if err := f.keyring.SetAPIKey(f.profile, apiKey); err != nil {
			return fmt.Errorf("failed to store API key in keyring: %w", err)
		}
		return nil
	}

	// Fallback to config file
	profile, err := f.config.GetProfile(f.profile)
	if err != nil {
		// Create new profile if it doesn't exist
		profile = config.Profile{
			BaseURL: "https://api.forwardemail.net",
			Timeout: "30s",
			Output:  "table",
		}
	}

	profile.APIKey = apiKey
	f.config.SetProfile(f.profile, &profile)

	if err := f.config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// DeleteAPIKey removes the API key for the current profile
func (f *ForwardEmailAuth) DeleteAPIKey() error {
	// Remove from keyring if available
	if f.keyring != nil {
		if err := f.keyring.DeleteAPIKey(f.profile); err != nil {
			// Don't fail if key doesn't exist in keyring
			fmt.Printf("Warning: failed to delete API key from keyring: %v\n", err)
		}
	}

	// Remove from config file
	profile, err := f.config.GetProfile(f.profile)
	if err != nil {
		return fmt.Errorf("failed to get profile %s: %w", f.profile, err)
	}

	profile.APIKey = ""
	f.config.SetProfile(f.profile, &profile)

	if err := f.config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// HasAPIKey checks if an API key is available for the current profile
func (f *ForwardEmailAuth) HasAPIKey() bool {
	_, err := f.GetAPIKey()
	return err == nil
}

// MockProvider creates a mock auth provider for testing
func MockProvider(apiKey string) Provider {
	return &mockAuth{apiKey: apiKey}
}

type mockAuth struct {
	apiKey string
}

func (m *mockAuth) Apply(req *http.Request) error {
	auth := base64.StdEncoding.EncodeToString([]byte(m.apiKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)
	return nil
}

func (m *mockAuth) Validate(_ context.Context) error {
	if m.apiKey == "" {
		return fmt.Errorf("mock API key is empty")
	}
	return nil
}

func (m *mockAuth) GetAPIKey() (string, error) {
	if m.apiKey == "" {
		return "", fmt.Errorf("mock API key not set")
	}
	return m.apiKey, nil
}
