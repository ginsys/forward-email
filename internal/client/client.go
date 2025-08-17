package client

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/ginsys/forwardemail-cli/internal/keyring"
	"github.com/ginsys/forwardemail-cli/pkg/api"
	"github.com/ginsys/forwardemail-cli/pkg/auth"
	"github.com/ginsys/forwardemail-cli/pkg/config"
)

// NewAPIClient creates a new API client with proper authentication
// This centralizes the authentication logic that was duplicated across commands
func NewAPIClient() (*api.Client, error) {
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
			return nil, fmt.Errorf("no profile configured. Use 'forward-email profile create <name>' to create a profile and 'forward-email profile switch <name>' to set it as current")
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