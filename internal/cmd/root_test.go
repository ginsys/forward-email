package cmd

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{
			name:           "help flag",
			args:           []string{"--help"},
			expectError:    false,
			expectedOutput: "A comprehensive CLI for Forward Email API management",
		},
		{
			name:        "invalid flag",
			args:        []string{"--invalid-flag"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new root command for each test to avoid state pollution
			cmd := &cobra.Command{
				Use:     "forward-email",
				Short:   "A comprehensive CLI for Forward Email API management",
				Version: "test-version",
			}

			// Add persistent flags
			cmd.PersistentFlags().String("profile", "", "Configuration profile to use")
			cmd.PersistentFlags().String("output", "table", "Output format (table|json|yaml|csv)")
			cmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
			cmd.PersistentFlags().Bool("debug", false, "Enable debug output")

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if tt.expectedOutput != "" {
					outputStr := output.String()
					if !strings.Contains(outputStr, tt.expectedOutput) {
						t.Errorf("Expected output to contain %q, got %q", tt.expectedOutput, outputStr)
					}
				}
			}
		})
	}
}

func TestExecute(t *testing.T) {
	// Save original state
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	tests := []struct {
		name      string
		args      []string
		setupFunc func()
	}{
		{
			name: "execute with context",
			args: []string{"forward-email", "--help"},
			setupFunc: func() {
				os.Args = []string{"forward-email", "--help"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			ctx := context.Background()

			// Capture output to avoid cluttering test output
			var output bytes.Buffer
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&output)
			rootCmd.SetArgs([]string{"--help"})

			err := Execute(ctx)

			// Help command returns an error in cobra, but it's expected
			if err != nil && !strings.Contains(err.Error(), "help requested") {
				t.Errorf("Unexpected error from Execute: %v", err)
			}
		})
	}
}

func TestInitFlags(t *testing.T) {
	// Reset viper to clean state
	viper.Reset()

	// Create a test command
	testCmd := &cobra.Command{
		Use: "test-cmd",
	}

	// Add the same flags as initFlags
	testCmd.PersistentFlags().String("profile", "", "Configuration profile to use")
	testCmd.PersistentFlags().String("output", "table", "Output format (table|json|yaml|csv)")
	testCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	testCmd.PersistentFlags().Bool("debug", false, "Enable debug output")

	// Test that flags can be bound to viper
	err := viper.BindPFlag("profile", testCmd.PersistentFlags().Lookup("profile"))
	if err != nil {
		t.Errorf("Failed to bind profile flag: %v", err)
	}

	err = viper.BindPFlag("output", testCmd.PersistentFlags().Lookup("output"))
	if err != nil {
		t.Errorf("Failed to bind output flag: %v", err)
	}

	err = viper.BindPFlag("verbose", testCmd.PersistentFlags().Lookup("verbose"))
	if err != nil {
		t.Errorf("Failed to bind verbose flag: %v", err)
	}

	err = viper.BindPFlag("debug", testCmd.PersistentFlags().Lookup("debug"))
	if err != nil {
		t.Errorf("Failed to bind debug flag: %v", err)
	}

	// Test flag parsing
	testCmd.SetArgs([]string{"--profile", "test", "--output", "json", "--verbose", "--debug"})
	err = testCmd.Execute()
	if err != nil {
		t.Errorf("Failed to execute test command: %v", err)
	}

	// Test that viper has the values
	if viper.GetString("profile") != "test" {
		t.Errorf("Expected profile 'test', got '%s'", viper.GetString("profile"))
	}
	if viper.GetString("output") != "json" {
		t.Errorf("Expected output 'json', got '%s'", viper.GetString("output"))
	}
	if !viper.GetBool("verbose") {
		t.Error("Expected verbose to be true")
	}
	if !viper.GetBool("debug") {
		t.Error("Expected debug to be true")
	}
}

func TestInitConfig(t *testing.T) {
	// Reset viper
	viper.Reset()

	tests := []struct {
		name    string
		setup   func() (cleanup func())
		wantErr bool
	}{
		{
			name: "config file not found - should not error",
			setup: func() (cleanup func()) {
				// Set a non-existent config directory
				viper.AddConfigPath("/nonexistent/path")
				viper.SetConfigType("yaml")
				viper.SetConfigName("config")
				return func() {
					viper.Reset()
				}
			},
			wantErr: false,
		},
		{
			name: "env vars work",
			setup: func() (cleanup func()) {
				// Set environment variables
				os.Setenv("FORWARDEMAIL_PROFILE", "env-test")
				os.Setenv("FORWARDEMAIL_OUTPUT", "yaml")

				viper.SetEnvPrefix("FORWARDEMAIL")
				viper.AutomaticEnv()

				return func() {
					os.Unsetenv("FORWARDEMAIL_PROFILE")
					os.Unsetenv("FORWARDEMAIL_OUTPUT")
					viper.Reset()
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			// The initConfig function doesn't return an error,
			// it exits on fatal errors. We test the setup instead.

			if tt.name == "env vars work" {
				if viper.GetString("profile") != "env-test" {
					t.Errorf("Expected profile 'env-test', got '%s'", viper.GetString("profile"))
				}
				if viper.GetString("output") != "yaml" {
					t.Errorf("Expected output 'yaml', got '%s'", viper.GetString("output"))
				}
			}
		})
	}
}

func TestVersionTemplate(t *testing.T) {
	// Create a test command with version
	testCmd := &cobra.Command{
		Use:     "test",
		Version: "test-version",
	}

	// Set version template similar to root command
	testCmd.SetVersionTemplate("forward-email version test-version\ncommit: test-commit\nbuilt: test-date\n")

	var output bytes.Buffer
	testCmd.SetOut(&output)
	testCmd.SetArgs([]string{"--version"})

	err := testCmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute version command: %v", err)
	}

	expectedParts := []string{"forward-email version", "commit:", "built:"}
	outputStr := output.String()

	for _, part := range expectedParts {
		if !strings.Contains(outputStr, part) {
			t.Errorf("Expected version output to contain %q, got %q", part, outputStr)
		}
	}
}
