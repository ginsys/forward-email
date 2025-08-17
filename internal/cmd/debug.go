package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ginsys/forwardemail-cli/internal/client"
	"github.com/ginsys/forwardemail-cli/internal/keyring"
	"github.com/ginsys/forwardemail-cli/pkg/auth"
	"github.com/ginsys/forwardemail-cli/pkg/config"
)

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug utilities for troubleshooting",
	Long:  `Debug utilities for troubleshooting authentication and keyring issues.`,
}

// debugKeysCmd represents the debug keys command
var debugKeysCmd = &cobra.Command{
	Use:   "keys [profile]",
	Short: "Show keyring information for debugging",
	Long:  `Show keyring information for debugging authentication issues.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDebugKeys,
}

// debugAuthCmd represents the debug auth command
var debugAuthCmd = &cobra.Command{
	Use:   "auth [profile]",
	Short: "Debug the full authentication flow",
	Long:  `Debug the full authentication flow to see which API key is actually being used.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDebugAuth,
}

// debugAPICmd represents the debug api command
var debugAPICmd = &cobra.Command{
	Use:   "api [profile]",
	Short: "Test API call with current authentication",
	Long:  `Test an actual API call to see the exact error response.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDebugAPI,
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.AddCommand(debugKeysCmd)
	debugCmd.AddCommand(debugAuthCmd)
	debugCmd.AddCommand(debugAPICmd)
}

func runDebugKeys(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	profileName := cfg.CurrentProfile
	if len(args) > 0 {
		profileName = args[0]
	}

	if profileName == "" {
		return fmt.Errorf("no profile specified and no current profile set")
	}

	fmt.Printf("Debug information for profile: %s\n", profileName)
	fmt.Printf("Current profile in config: %s\n", cfg.CurrentProfile)

	// Check if profile exists in config
	profile, exists := cfg.Profiles[profileName]
	if !exists {
		fmt.Printf("❌ Profile '%s' does not exist in config\n", profileName)
		return nil
	}

	fmt.Printf("✅ Profile exists in config\n")
	fmt.Printf("   Base URL: %s\n", profile.BaseURL)
	fmt.Printf("   Config API Key: %s\n", func() string {
		if profile.APIKey == "" {
			return "(empty - should be in keyring)"
		}
		return fmt.Sprintf("(set - %d characters)", len(profile.APIKey))
	}())

	// Check keyring
	kr, err := keyring.New(keyring.Config{})
	if err != nil {
		fmt.Printf("❌ Failed to initialize keyring: %v\n", err)
		return nil
	}

	fmt.Printf("✅ Keyring initialized\n")

	// Try to get API key from keyring
	apiKey, err := kr.GetAPIKey(profileName)
	if err != nil {
		fmt.Printf("❌ Failed to get API key from keyring: %v\n", err)
	} else {
		fmt.Printf("✅ API key found in keyring\n")
		fmt.Printf("   Key length: %d characters\n", len(apiKey))
		if len(apiKey) > 10 {
			fmt.Printf("   Key preview: %s...%s\n", apiKey[:5], apiKey[len(apiKey)-5:])
		} else {
			fmt.Printf("   Key preview: %s\n", apiKey)
		}

		// Basic validation
		if len(apiKey) < 10 {
			fmt.Printf("⚠️  API key seems too short (< 10 characters)\n")
		}
		// Check for non-printable characters in API key
		for _, r := range apiKey {
			if r < 32 || r > 126 {
				fmt.Printf("⚠️  API key contains non-printable characters\n")
				break
			}
		}
	}

	return nil
}

func runDebugAuth(cmd *cobra.Command, args []string) error {
	// Simulate the exact same flow as NewAPIClient
	profile := viper.GetString("profile")
	if len(args) > 0 {
		profile = args[0]
		// Temporarily set this profile for testing
		viper.Set("profile", profile)
	}

	fmt.Printf("🔍 Debugging authentication flow\n")
	fmt.Printf("   Requested profile: %s\n", profile)

	// Load configuration (same as NewAPIClient)
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("   Config current_profile: %s\n", cfg.CurrentProfile)

	// Profile resolution logic (same as NewAPIClient)
	if profile == "" {
		profile = cfg.CurrentProfile
		if profile == "" {
			return fmt.Errorf("no profile configured")
		}
	}

	fmt.Printf("✅ Resolved profile: %s\n", profile)

	// Check if profile exists in config
	profileConfig, exists := cfg.Profiles[profile]
	if !exists {
		return fmt.Errorf("profile '%s' does not exist in config", profile)
	}

	fmt.Printf("✅ Profile exists in config\n")
	fmt.Printf("   Base URL: %s\n", profileConfig.BaseURL)

	// Initialize keyring (same as NewAPIClient)
	kr, err := keyring.New(keyring.Config{})
	if err != nil {
		fmt.Printf("⚠️  Keyring init failed: %v (will use config file)\n", err)
		kr = nil
	} else {
		fmt.Printf("✅ Keyring initialized\n")
	}

	// Create auth provider (same as NewAPIClient)
	fmt.Printf("🔐 Creating auth provider...\n")
	_, err = auth.NewProvider(auth.ProviderConfig{
		Profile: profile,
		Config:  cfg,
		Keyring: kr,
	})
	if err != nil {
		return fmt.Errorf("failed to create auth provider: %w", err)
	}

	fmt.Printf("✅ Auth provider created\n")

	// Now let's see what API key the auth provider actually uses
	// We'll need to add a debug method to the auth provider or check its behavior

	// For now, let's manually check what the auth provider would find
	var apiKey string
	var source string

	// Check environment variables first (same order as auth provider)
	if envKey := viper.GetString("api_key"); envKey != "" {
		apiKey = envKey
		source = "environment variable"
	} else if kr != nil {
		// Check keyring
		if keyringKey, err := kr.GetAPIKey(profile); err == nil {
			apiKey = keyringKey
			source = "keyring"
		}
	}

	// Fallback to config file
	if apiKey == "" && profileConfig.APIKey != "" {
		apiKey = profileConfig.APIKey
		source = "config file"
	}

	if apiKey == "" {
		fmt.Printf("❌ No API key found for profile '%s'\n", profile)
		return fmt.Errorf("no API key found")
	}

	fmt.Printf("✅ API key found\n")
	fmt.Printf("   Source: %s\n", source)
	fmt.Printf("   Length: %d characters\n", len(apiKey))
	if len(apiKey) > 10 {
		fmt.Printf("   Preview: %s...%s\n", apiKey[:5], apiKey[len(apiKey)-5:])
	} else {
		fmt.Printf("   Preview: %s\n", apiKey)
	}

	// Basic validation
	if len(apiKey) < 10 {
		fmt.Printf("⚠️  API key seems too short (< 10 characters)\n")
	}

	baseURL := viper.GetString("api_base_url")
	if baseURL == "" {
		baseURL = "https://api.forwardemail.net"
	}
	fmt.Printf("📡 API Base URL: %s\n", baseURL)

	return nil
}

func runDebugAPI(cmd *cobra.Command, args []string) error {
	profile := ""
	if len(args) > 0 {
		profile = args[0]
		// Temporarily set this profile for testing
		viper.Set("profile", profile)
	}

	fmt.Printf("🌐 Testing actual API call\n")
	if profile != "" {
		fmt.Printf("   Using profile: %s\n", profile)
	}

	// Use the same client creation as the real commands
	apiClient, err := client.NewAPIClient()
	if err != nil {
		fmt.Printf("❌ Failed to create API client: %v\n", err)
		return err
	}

	fmt.Printf("✅ API client created successfully\n")

	// Try a simple API call
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Printf("📞 Making API call to list domains...\n")

	response, err := apiClient.Domains.ListDomains(ctx, nil)
	if err != nil {
		fmt.Printf("❌ API call failed: %v\n", err)
		return fmt.Errorf("API call failed: %w", err)
	}

	fmt.Printf("✅ API call successful!\n")
	fmt.Printf("   Found %d domains\n", len(response.Domains))
	if len(response.Domains) > 0 {
		fmt.Printf("   First domain: %s\n", response.Domains[0].Name)
	}

	return nil
}
