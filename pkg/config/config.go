package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the main application configuration structure.
// It manages multiple profiles for different Forward Email accounts or environments,
// and tracks which profile is currently active for CLI operations.
type Config struct {
	Profiles       map[string]Profile `yaml:"profiles" mapstructure:"profiles"`               // Map of profile name to profile configuration
	CurrentProfile string             `yaml:"current_profile" mapstructure:"current_profile"` // Name of the currently active profile
}

// Profile represents a configuration profile for a specific Forward Email account or environment.
// Each profile stores API connection details, authentication credentials, and user preferences.
// Credentials stored here should be considered less secure than OS keyring storage.
type Profile struct {
	BaseURL  string `yaml:"base_url" mapstructure:"base_url"` // Forward Email API base URL
	APIKey   string `yaml:"api_key" mapstructure:"api_key"`   // API key (prefer keyring storage)
	Username string `yaml:"username" mapstructure:"username"` // Username (legacy, not used)
	Password string `yaml:"password" mapstructure:"password"` // Password (legacy, not used)
	Timeout  string `yaml:"timeout" mapstructure:"timeout"`   // Request timeout duration
	Output   string `yaml:"output" mapstructure:"output"`     // Default output format (table/json/yaml/csv)
}

// Load loads the complete application configuration from file and environment variables.
// It applies default profile settings if no configuration exists, making it safe
// to call on first run. This is the standard method for loading configuration.
func Load() (*Config, error) {
	return loadConfig(true)
}

// LoadWithoutDefaults loads configuration without creating default profiles.
// This is useful for profile management operations where we need to see only
// explicitly configured profiles. Returns empty configuration if no config file exists.
func LoadWithoutDefaults() (*Config, error) {
	return loadConfig(false)
}

// loadConfig is the internal implementation for configuration loading.
// It handles file discovery, parsing, environment variable integration,
// and optional default profile creation. The withDefaults parameter controls
// whether to create a default profile when no configuration exists.
func loadConfig(withDefaults bool) (*Config, error) {
	// Set config file location
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	// Set environment variable prefix
	viper.SetEnvPrefix("FORWARDEMAIL")
	viper.AutomaticEnv()

	// Set defaults only if requested
	if withDefaults {
		setDefaults()
	}

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// Set the values in viper
	viper.Set("current_profile", c.CurrentProfile)
	viper.Set("profiles", c.Profiles)

	// Write config file
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetProfile returns the specified profile or the current profile
func (c *Config) GetProfile(name string) (Profile, error) {
	if name == "" {
		name = c.CurrentProfile
	}

	profile, exists := c.Profiles[name]
	if !exists {
		return Profile{}, fmt.Errorf("profile %q not found", name)
	}

	return profile, nil
}

// SetProfile sets or updates a profile
func (c *Config) SetProfile(name string, profile *Profile) {
	if c.Profiles == nil {
		c.Profiles = make(map[string]Profile)
	}
	c.Profiles[name] = *profile
}

// DeleteProfile removes a profile
func (c *Config) DeleteProfile(name string) error {
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile %q not found", name)
	}

	if name == c.CurrentProfile {
		return fmt.Errorf("cannot delete current profile %q", name)
	}

	delete(c.Profiles, name)
	return nil
}

// ListProfiles returns all profile names
func (c *Config) ListProfiles() []string {
	profiles := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		profiles = append(profiles, name)
	}
	return profiles
}

// getConfigDir returns the configuration directory
func getConfigDir() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}

	return filepath.Join(configDir, "forwardemail"), nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Don't set a default current_profile - let users choose explicitly
	// Don't create any default profiles - users should create what they need
}
