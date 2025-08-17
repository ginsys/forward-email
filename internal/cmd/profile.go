package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ginsys/forwardemail-cli/internal/keyring"
	"github.com/ginsys/forwardemail-cli/pkg/config"
	"github.com/ginsys/forwardemail-cli/pkg/output"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage configuration profiles",
	Long: `Manage configuration profiles for different Forward Email accounts.
	
Profiles allow you to switch between different Forward Email accounts
or environments easily. Each profile stores its own API credentials
and configuration settings.`,
}

// profileListCmd represents the profile list command
var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Long:  `List all available configuration profiles and show the current active profile.`,
	RunE:  runProfileList,
}

// profileShowCmd represents the profile show command
var profileShowCmd = &cobra.Command{
	Use:   "show [profile-name]",
	Short: "Show profile details",
	Long:  `Show detailed information about a specific profile or the current profile.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runProfileShow,
}

// profileSwitchCmd represents the profile switch command
var profileSwitchCmd = &cobra.Command{
	Use:   "switch <profile-name>",
	Short: "Switch to a different profile",
	Long:  `Set the specified profile as the current active profile.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileSwitch,
}

// profileDeleteCmd represents the profile delete command
var profileDeleteCmd = &cobra.Command{
	Use:   "delete <profile-name>",
	Short: "Delete a profile",
	Long:  `Delete a profile and its associated credentials from both the config file and keyring.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileDelete,
}

// profileCreateCmd represents the profile create command
var profileCreateCmd = &cobra.Command{
	Use:   "create <profile-name>",
	Short: "Create a new profile",
	Long:  `Create a new empty profile with default settings.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileCreate,
}

var (
	profileOutputFormat string
	profileForce        bool
)

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileShowCmd)
	profileCmd.AddCommand(profileSwitchCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	profileCmd.AddCommand(profileCreateCmd)

	// Global flags for profile commands
	profileCmd.PersistentFlags().StringVarP(&profileOutputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	
	// Delete command flags
	profileDeleteCmd.Flags().BoolVarP(&profileForce, "force", "f", false, "Force deletion without confirmation")
}

func runProfileList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadWithoutDefaults()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create table data for profiles
	if profileOutputFormat == "table" {
		headers := []string{"PROFILE", "CURRENT", "BASE_URL", "HAS_API_KEY", "OUTPUT", "TIMEOUT"}
		table := output.NewTableData(headers)

		for profileName, profile := range cfg.Profiles {
			current := ""
			if profileName == cfg.CurrentProfile {
				current = "✓"
			}

			// Check if API key exists in keyring or config
			hasAPIKey := "✗"
			if profile.APIKey != "" {
				hasAPIKey = "✓ (config)"
			} else {
				// Check keyring
				kr, err := keyring.New(keyring.Config{})
				if err == nil {
					if apiKey, err := kr.GetAPIKey(profileName); err == nil && apiKey != "" {
						hasAPIKey = "✓ (keyring)"
					}
				}
			}

			row := []string{
				profileName,
				current,
				profile.BaseURL,
				hasAPIKey,
				profile.Output,
				profile.Timeout,
			}
			table.AddRow(row)
		}

		formatter := output.NewFormatter(output.FormatTable, nil)
		return formatter.Format(table)
	}

	// For JSON/YAML output, return the profiles data
	profileData := struct {
		CurrentProfile string                    `json:"current_profile" yaml:"current_profile"`
		Profiles       map[string]config.Profile `json:"profiles" yaml:"profiles"`
	}{
		CurrentProfile: cfg.CurrentProfile,
		Profiles:       cfg.Profiles,
	}

	outputFormat, err := output.ParseFormat(profileOutputFormat)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	formatter := output.NewFormatter(outputFormat, nil)
	return formatter.Format(profileData)
}

func runProfileShow(cmd *cobra.Command, args []string) error {
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

	profile, exists := cfg.Profiles[profileName]
	if !exists {
		return fmt.Errorf("profile '%s' does not exist", profileName)
	}

	// Check API key location
	apiKeyLocation := "none"
	if profile.APIKey != "" {
		apiKeyLocation = "config file"
	} else {
		kr, err := keyring.New(keyring.Config{})
		if err == nil {
			if apiKey, err := kr.GetAPIKey(profileName); err == nil && apiKey != "" {
				apiKeyLocation = "OS keyring"
			}
		}
	}

	if profileOutputFormat == "table" {
		headers := []string{"PROPERTY", "VALUE"}
		table := output.NewTableData(headers)

		isCurrent := "no"
		if profileName == cfg.CurrentProfile {
			isCurrent = "yes"
		}

		table.AddRow([]string{"Profile Name", profileName})
		table.AddRow([]string{"Current Profile", isCurrent})
		table.AddRow([]string{"Base URL", profile.BaseURL})
		table.AddRow([]string{"API Key Location", apiKeyLocation})
		table.AddRow([]string{"Username", output.FormatValue(profile.Username)})
		table.AddRow([]string{"Output Format", profile.Output})
		table.AddRow([]string{"Timeout", profile.Timeout})

		formatter := output.NewFormatter(output.FormatTable, nil)
		return formatter.Format(table)
	}

	// For JSON/YAML, include additional metadata
	profileData := struct {
		Name           string        `json:"name" yaml:"name"`
		IsCurrent      bool          `json:"is_current" yaml:"is_current"`
		BaseURL        string        `json:"base_url" yaml:"base_url"`
		APIKeyLocation string        `json:"api_key_location" yaml:"api_key_location"`
		Username       string        `json:"username" yaml:"username"`
		Output         string        `json:"output" yaml:"output"`
		Timeout        string        `json:"timeout" yaml:"timeout"`
	}{
		Name:           profileName,
		IsCurrent:      profileName == cfg.CurrentProfile,
		BaseURL:        profile.BaseURL,
		APIKeyLocation: apiKeyLocation,
		Username:       profile.Username,
		Output:         profile.Output,
		Timeout:        profile.Timeout,
	}

	outputFormat, err := output.ParseFormat(profileOutputFormat)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	formatter := output.NewFormatter(outputFormat, nil)
	return formatter.Format(profileData)
}

func runProfileSwitch(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if _, exists := cfg.Profiles[profileName]; !exists {
		return fmt.Errorf("profile '%s' does not exist", profileName)
	}

	cfg.CurrentProfile = profileName
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Switched to profile '%s'\n", profileName)
	return nil
}

func runProfileDelete(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if _, exists := cfg.Profiles[profileName]; !exists {
		return fmt.Errorf("profile '%s' does not exist", profileName)
	}

	if !profileForce {
		fmt.Printf("Are you sure you want to delete profile '%s'? This will remove all associated credentials. [y/N]: ", profileName)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Profile deletion cancelled")
			return nil
		}
	}

	// Remove from keyring if it exists
	kr, err := keyring.New(keyring.Config{})
	if err == nil {
		if err := kr.DeleteAPIKey(profileName); err != nil {
			fmt.Printf("Warning: failed to delete credentials from keyring: %v\n", err)
		}
	}

	// Remove from config
	delete(cfg.Profiles, profileName)

	// If this was the current profile, unset it
	if cfg.CurrentProfile == profileName {
		cfg.CurrentProfile = ""
		fmt.Printf("Current profile unset. Use 'forward-email profile switch <name>' to set a new current profile\n")
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Profile '%s' deleted successfully\n", profileName)
	return nil
}

func runProfileCreate(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if _, exists := cfg.Profiles[profileName]; exists {
		return fmt.Errorf("profile '%s' already exists", profileName)
	}

	// Create new profile with default settings
	newProfile := config.Profile{
		BaseURL:  "https://api.forwardemail.net",
		APIKey:   "",
		Username: "",
		Password: "",
		Timeout:  "30s",
		Output:   "table",
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]config.Profile)
	}
	cfg.Profiles[profileName] = newProfile

	// If no current profile is set, make this the current one
	if cfg.CurrentProfile == "" {
		cfg.CurrentProfile = profileName
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Profile '%s' created successfully\n", profileName)
	if cfg.CurrentProfile == profileName {
		fmt.Printf("Set as current profile\n")
	}
	fmt.Printf("Use 'forward-email auth login --profile %s' to add API credentials\n", profileName)
	return nil
}