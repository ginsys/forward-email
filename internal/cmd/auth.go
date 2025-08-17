package cmd

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"

	"github.com/ginsys/forwardemail-cli/internal/client"
	"github.com/ginsys/forwardemail-cli/internal/keyring"
	"github.com/ginsys/forwardemail-cli/pkg/auth"
	"github.com/ginsys/forwardemail-cli/pkg/config"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication credentials",
	Long: `Manage authentication credentials for Forward Email API.

The auth command group provides subcommands to log in, verify credentials,
and manage API keys across different profiles.`,
}

// authVerifyCmd represents the auth verify command
var authVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify authentication credentials",
	Long: `Verify that the current authentication credentials are valid.

This command will attempt to authenticate with the Forward Email API
using the current profile's credentials and report whether they are valid.`,
	RunE: runAuthVerify,
}

// authLoginCmd represents the auth login command
var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in and save API credentials",
	Long: `Interactively log in to Forward Email and save API credentials.

This command will prompt for your API key and securely store it
in the OS keyring or configuration file.`,
	RunE: runAuthLogin,
}

// authLogoutCmd represents the auth logout command
var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out and clear stored credentials",
	Long: `Clear stored credentials for the current or specified profile.

This command will remove the API key from the OS keyring and
configuration file for the specified profile.`,
	RunE: runAuthLogout,
}

// authStatusCmd represents the auth status command
var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long: `Show the current authentication status for all profiles.

This command displays which profiles have API keys configured
and from which sources they are loaded.`,
	RunE: runAuthStatus,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authVerifyCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)

	// Add flags for profile specification
	authVerifyCmd.Flags().String("profile", "", "Profile to verify (defaults to current profile)")
	authLoginCmd.Flags().String("profile", "", "Profile to log in to (defaults to current profile)")
	authLogoutCmd.Flags().String("profile", "", "Profile to log out from (defaults to current profile)")
	authLogoutCmd.Flags().Bool("all", false, "Log out from all profiles")
}

func runAuthVerify(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	profile := cmd.Flag("profile").Value.String()
	if profile != "" {
		// Temporarily set the profile for this verification
		viper.Set("profile", profile)
	}

	// Use centralized client creation which handles all the auth setup
	apiClient, err := client.NewAPIClient()
	if err != nil {
		fmt.Printf("❌ Authentication failed: %v\n", err)
		return fmt.Errorf("authentication verification failed")
	}

	// If we got here, authentication was successful
	currentProfile := viper.GetString("profile")
	if currentProfile == "" {
		// Load config to get current profile
		cfg, _ := config.Load()
		if cfg != nil {
			currentProfile = cfg.CurrentProfile
		}
		if currentProfile == "" {
			currentProfile = "default"
		}
	}
	
	fmt.Printf("✅ Authentication successful for profile '%s'\n", currentProfile)
	
	// Make a simple API call to double-check
	_, err = apiClient.Domains.ListDomains(ctx, nil)
	if err != nil {
		fmt.Printf("⚠️  Authentication succeeded but API call failed: %v\n", err)
		return fmt.Errorf("API verification failed")
	}
	
	fmt.Printf("✅ API access verified\n")
	return nil
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	profile := cmd.Flag("profile").Value.String()
	if profile == "" {
		profile = viper.GetString("profile")
		if profile == "" {
			profile = "default"
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize keyring
	kr, err := keyring.New(keyring.Config{})
	if err != nil {
		fmt.Printf("Warning: failed to initialize keyring: %v\n", err)
		fmt.Println("Credentials will be stored in configuration file.")
	}

	// Prompt for API key
	fmt.Printf("Forward Email CLI Login\n")
	fmt.Printf("Profile: %s\n\n", profile)
	fmt.Printf("Please enter your Forward Email API key.\n")
	fmt.Printf("You can find this in your Forward Email account under Security settings.\n\n")
	fmt.Print("API Key: ")

	// Read API key securely
	apiKeyBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read API key: %w", err)
	}
	fmt.Println() // Add newline after password input

	apiKey := string(apiKeyBytes)
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Create auth provider and validate
	authProvider, err := auth.NewProvider(auth.ProviderConfig{
		Profile: profile,
		Config:  cfg,
		Keyring: kr,
	})
	if err != nil {
		return fmt.Errorf("failed to create auth provider: %w", err)
	}

	// Create a temporary auth provider for validation
	tempAuth := auth.MockProvider(apiKey)
	if err := tempAuth.Validate(ctx); err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}

	// Store the API key
	if extAuth, ok := authProvider.(auth.ExtendedProvider); ok {
		if err := extAuth.SetAPIKey(apiKey); err != nil {
			return fmt.Errorf("failed to store API key: %w", err)
		}
	} else {
		return fmt.Errorf("auth provider does not support credential management")
	}

	fmt.Printf("✅ Successfully logged in to profile '%s'\n", profile)

	// Set as current profile if it's not already
	if cfg.CurrentProfile != profile {
		cfg.CurrentProfile = profile
		if err := cfg.Save(); err != nil {
			fmt.Printf("Warning: failed to set current profile: %v\n", err)
		} else {
			fmt.Printf("Set '%s' as the current profile\n", profile)
		}
	}

	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	logoutAll := cmd.Flag("all").Value.String() == "true"
	profile := cmd.Flag("profile").Value.String()

	if !logoutAll && profile == "" {
		profile = viper.GetString("profile")
		if profile == "" {
			profile = "default"
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize keyring
	kr, err := keyring.New(keyring.Config{})
	if err != nil {
		fmt.Printf("Warning: failed to initialize keyring: %v\n", err)
	}

	if logoutAll {
		// Logout from all profiles
		profiles := cfg.ListProfiles()
		if kr != nil {
			keyringProfiles, err := kr.ListProfiles()
			if err == nil {
				// Merge profiles from keyring
				profileMap := make(map[string]bool)
				for _, p := range profiles {
					profileMap[p] = true
				}
				for _, p := range keyringProfiles {
					if !profileMap[p] {
						profiles = append(profiles, p)
					}
				}
			}
		}

		if len(profiles) == 0 {
			fmt.Println("No profiles found to log out from")
			return nil
		}

		for _, p := range profiles {
			if err := logoutProfile(cfg, kr, p); err != nil {
				fmt.Printf("Warning: failed to logout from profile '%s': %v\n", p, err)
			} else {
				fmt.Printf("✅ Logged out from profile '%s'\n", p)
			}
		}
	} else {
		// Logout from specific profile
		if err := logoutProfile(cfg, kr, profile); err != nil {
			return fmt.Errorf("failed to logout from profile '%s': %w", profile, err)
		}
		fmt.Printf("✅ Logged out from profile '%s'\n", profile)
	}

	return nil
}

func logoutProfile(cfg *config.Config, kr *keyring.Keyring, profile string) error {
	// Create auth provider
	authProvider, err := auth.NewProvider(auth.ProviderConfig{
		Profile: profile,
		Config:  cfg,
		Keyring: kr,
	})
	if err != nil {
		return fmt.Errorf("failed to create auth provider: %w", err)
	}

	// Delete API key
	if extAuth, ok := authProvider.(auth.ExtendedProvider); ok {
		return extAuth.DeleteAPIKey()
	}

	return fmt.Errorf("auth provider does not support credential management")
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize keyring
	kr, err := keyring.New(keyring.Config{})
	if err != nil {
		fmt.Printf("Warning: failed to initialize keyring: %v\n", err)
	}

	fmt.Printf("Authentication Status\n")
	fmt.Printf("====================\n\n")

	// Get all profiles
	profiles := cfg.ListProfiles()
	if kr != nil {
		keyringProfiles, err := kr.ListProfiles()
		if err == nil {
			// Merge profiles from keyring
			profileMap := make(map[string]bool)
			for _, p := range profiles {
				profileMap[p] = true
			}
			for _, p := range keyringProfiles {
				if !profileMap[p] {
					profiles = append(profiles, p)
				}
			}
		}
	}

	if len(profiles) == 0 {
		fmt.Println("No profiles configured")
		return nil
	}

	for _, profile := range profiles {
		current := ""
		if profile == cfg.CurrentProfile {
			current = " (current)"
		}

		fmt.Printf("Profile: %s%s\n", profile, current)

		// Check credential sources
		authProvider, err := auth.NewProvider(auth.ProviderConfig{
			Profile: profile,
			Config:  cfg,
			Keyring: kr,
		})
		if err != nil {
			fmt.Printf("  Status: ❌ Error creating auth provider: %v\n", err)
			continue
		}

		if extAuth, ok := authProvider.(auth.ExtendedProvider); ok {
			if extAuth.HasAPIKey() {
				fmt.Printf("  Status: ✅ API key configured\n")

				// Determine source
				source := "unknown"
				if envKey := os.Getenv("FORWARDEMAIL_API_KEY"); envKey != "" {
					source = "environment variable (FORWARDEMAIL_API_KEY)"
				} else if envKey := os.Getenv(fmt.Sprintf("FORWARDEMAIL_%s_API_KEY", profile)); envKey != "" {
					source = fmt.Sprintf("environment variable (FORWARDEMAIL_%s_API_KEY)", profile)
				} else if kr != nil && kr.HasAPIKey(profile) {
					source = "OS keyring"
				} else {
					source = "configuration file"
				}
				fmt.Printf("  Source: %s\n", source)
			} else {
				fmt.Printf("  Status: ❌ No API key configured\n")
			}
		}
		fmt.Println()
	}

	return nil
}
