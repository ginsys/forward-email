package client

import (
	"fmt"

	"github.com/spf13/viper"

    "github.com/ginsys/forward-email/internal/keyring"
    "github.com/ginsys/forward-email/pkg/api"
    "github.com/ginsys/forward-email/pkg/auth"
    "github.com/ginsys/forward-email/pkg/config"
)

// Test mode variables
var (
	testMode    bool
	testBaseURL string
	testAuth    auth.Provider
)

// SetTestMode configures the client for testing with a mock server
func SetTestMode(baseURL string, authProvider auth.Provider) {
	testMode = true
	testBaseURL = baseURL
	testAuth = authProvider
}

// ResetTestMode disables test mode
func ResetTestMode() {
	testMode = false
	testBaseURL = ""
	testAuth = nil
}

// NewAPIClient creates a new API client with proper authentication
// This centralizes the authentication logic that was duplicated across commands
func NewAPIClient() (*api.Client, error) {
	// If in test mode, return test client
	if testMode {
		return api.NewClient(testBaseURL, testAuth)
	}

	profile := viper.GetString("profile")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// If no profile specified via flag, use current profile from config
	if profile == "" {
		profile = cfg.CurrentProfile
		if profile == "" {
			return nil, fmt.Errorf("no profile configured. Use 'forward-email profile create <name>' to create a profile and " +
				"'forward-email profile switch <name>' to set it as current")
		}
	}

	// Initialize keyring
	kr, err := keyring.New(keyring.Config{})
	if err != nil {
		// Continue without keyring, auth will fall back to config file
		kr = nil
	}

	authProvider, err := auth.NewProvider(auth.ProviderConfig{
		Profile: profile,
		Config:  cfg,
		Keyring: kr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create auth provider: %w", err)
	}

	baseURL := viper.GetString("api_base_url")
	if baseURL == "" {
		baseURL = "https://api.forwardemail.net"
	}

	return api.NewClient(baseURL, authProvider)
}
