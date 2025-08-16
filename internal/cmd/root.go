package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

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
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(ctx context.Context) error {
	rootCmd.SetContext(ctx)
	return rootCmd.Execute()
}

// initFlags initializes the root command flags and viper bindings
func initFlags() {
	// Global flags
	rootCmd.PersistentFlags().String("profile", "", "Configuration profile to use")
	rootCmd.PersistentFlags().String("output", "table", "Output format (table|json|yaml|csv)")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug output")
	rootCmd.PersistentFlags().Duration("timeout", 0, "Request timeout duration")

	// Bind flags to viper
	_ = viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))

	// Version template
	rootCmd.SetVersionTemplate(fmt.Sprintf("forward-email version %s\ncommit: %s\nbuilt: %s\n", version, commit, date))
}

func init() {
	cobra.OnInitialize(initConfig)
	initFlags()
}

// initConfig reads in config file and ENV variables if set.
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
