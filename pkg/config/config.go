package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	CurrentProfile string             `yaml:"current_profile" mapstructure:"current_profile"`
	Profiles       map[string]Profile `yaml:"profiles" mapstructure:"profiles"`
}

// Profile represents a configuration profile
type Profile struct {
	BaseURL  string `yaml:"base_url" mapstructure:"base_url"`
	APIKey   string `yaml:"api_key" mapstructure:"api_key"`
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
	Timeout  string `yaml:"timeout" mapstructure:"timeout"`
	Output   string `yaml:"output" mapstructure:"output"`
}

// Load loads the configuration from file and environment
func Load() (*Config, error) {
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

	// Set defaults
	setDefaults()

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
func (c *Config) SetProfile(name string, profile Profile) {
	if c.Profiles == nil {
		c.Profiles = make(map[string]Profile)
	}
	c.Profiles[name] = profile
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
	viper.SetDefault("current_profile", "default")
	viper.SetDefault("profiles.default.base_url", "https://api.forwardemail.net")
	viper.SetDefault("profiles.default.timeout", "30s")
	viper.SetDefault("profiles.default.output", "table")
}
