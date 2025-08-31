package cmd

import (
	"context"
	"fmt"
	"os"

	buildversion "github.com/ginsys/forward-email/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version information is provided by internal/version (populated via -ldflags).

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "forward-email",
	Short: "A comprehensive CLI for Forward Email API management",
	Long: `Forward Email CLI is a powerful command-line interface for managing
Forward Email accounts, domains, aliases, and email operations through
their public REST API.

Features:
- Complete API coverage with all Forward Email endpoints
- Multi-profile support for different environments
- Security-first design with OS keyring integration
- Developer experience with shell completion and interactive wizards
- Enterprise ready with audit logging and CI/CD integration`,
	Version: buildversion.Version,
}

// Execute is the main entry point for the CLI application.
// It configures the root command with the provided context for cancellation support
// and executes the command tree. This function should be called from main() to
// start the CLI application and handle all command parsing and execution.
func Execute(ctx context.Context) error {
	rootCmd.SetContext(ctx)
	return rootCmd.Execute()
}

// initFlags initializes all persistent flags for the root command and binds them to viper.
// These flags are inherited by all subcommands and provide global configuration options
// including profile selection, output formatting, verbosity levels, and request timeouts.
// All flags are bound to viper for unified configuration management.
func initFlags() {
	// Global flags with short options
	rootCmd.PersistentFlags().StringP("profile", "p", "", "Configuration profile to use")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format (table|json|yaml|csv)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug output")
	rootCmd.PersistentFlags().Duration("timeout", 0, "Request timeout duration")

	// Bind flags to viper
	_ = viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))

	// Version template using internal/version package
	v := buildversion.Get()
	rootCmd.SetVersionTemplate(fmt.Sprintf("forward-email version %s\ncommit: %s\nbuilt: %s\n", v.Version, v.Commit, v.Date))
}

func init() {
	cobra.OnInitialize(initConfig)
	initFlags()
}

// initConfig initializes the configuration system using viper.
// It searches for config files in standard locations (~/.config/forwardemail/, current directory),
// sets up environment variable binding with FORWARDEMAIL_ prefix, and loads the configuration.
// This function is called automatically by cobra.OnInitialize() before any commands run.
func initConfig() {
	// Find home directory
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
		os.Exit(1)
	}

	// Search config in home directory with name ".forwardemail" (without extension)
	viper.AddConfigPath(home + "/.config/forwardemail")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	// Environment variables
	viper.SetEnvPrefix("FORWARDEMAIL")
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil && viper.GetBool("debug") {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}
