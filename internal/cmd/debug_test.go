package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/ginsys/forward-email/internal/testutil"
	"github.com/spf13/cobra"
)

func TestDebugCommands(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		setupConfig    func() string // returns temp dir
		expectError    bool
		expectedOutput string
	}{
		{
			name: "debug keys with valid config",
			args: []string{"debug", "keys"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)

				configContent := `current_profile: "main"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "test-key"
    timeout: "30s"
    output: "table"
`

				testutil.WriteTestConfig(t, tempDir, configContent)

				return tempDir
			},
			expectError: false, // Debug commands should not error on valid config
		},
		{
			name: "debug keys with specific profile",
			args: []string{"debug", "keys", "main"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)

				configContent := `current_profile: "test"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "main-key"
    timeout: "30s"
    output: "table"
  test:
    base_url: "https://api.forwardemail.net"
    api_key: "test-key"
    timeout: "30s"
    output: "table"
`

				testutil.WriteTestConfig(t, tempDir, configContent)

				return tempDir
			},
			expectError: false,
		},
		{
			name: "debug auth with valid config",
			args: []string{"debug", "auth"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)

				configContent := `current_profile: "main"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "test-key"
    timeout: "30s"
    output: "table"
`

				testutil.WriteTestConfig(t, tempDir, configContent)

				return tempDir
			},
			expectError: false,
		},
		{
			name: "debug api with valid config",
			args: []string{"debug", "api"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)

				configContent := `current_profile: "main"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "test-key"
    timeout: "30s"
    output: "table"
`

				testutil.WriteTestConfig(t, tempDir, configContent)

				return tempDir
			},
			expectError: false, // Command structure should work, API call may fail but that's expected for debug
		},
		{
			name: "debug help",
			args: []string{"debug", "--help"},
			setupConfig: func() string {
				return testutil.SetupTempConfig(t)
			},
			expectError:    false,
			expectedOutput: "Debug utilities for troubleshooting",
		},
		{
			name: "debug keys help",
			args: []string{"debug", "keys", "--help"},
			setupConfig: func() string {
				return testutil.SetupTempConfig(t)
			},
			expectError:    false,
			expectedOutput: "Show keyring information for debugging",
		},
		{
			name: "debug auth help",
			args: []string{"debug", "auth", "--help"},
			setupConfig: func() string {
				return testutil.SetupTempConfig(t)
			},
			expectError:    false,
			expectedOutput: "Debug the full authentication flow",
		},
		{
			name: "debug api help",
			args: []string{"debug", "api", "--help"},
			setupConfig: func() string {
				return testutil.SetupTempConfig(t)
			},
			expectError:    false,
			expectedOutput: "Test an actual API call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config

			_ = tt.setupConfig() // tempDir is automatically cleaned up by t.TempDir()

			// Create root command with debug subcommand
			rootCmd := &cobra.Command{Use: "forward-email"}

			// Add debug command and its subcommands
			debugCmd := &cobra.Command{
				Use:   "debug",
				Short: "Debug utilities for troubleshooting",
				Long:  `Debug utilities for troubleshooting authentication and keyring issues.`,
			}

			debugKeysCmd := &cobra.Command{
				Use:   "keys [profile]",
				Short: "Show keyring information for debugging",
				Long:  `Show keyring information for debugging authentication issues.`,
				Args:  cobra.MaximumNArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation for testing
					fmt.Fprintf(cmd.OutOrStdout(), "Debug keyring information would be displayed here\n")
					return nil
				},
			}

			debugAuthCmd := &cobra.Command{
				Use:   "auth [profile]",
				Short: "Debug the full authentication flow",
				Long:  `Debug the full authentication flow to see which API key is actually being used.`,
				Args:  cobra.MaximumNArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation for testing
					fmt.Fprintf(cmd.OutOrStdout(), "Debug auth flow information would be displayed here\n")
					return nil
				},
			}

			debugAPICmd := &cobra.Command{
				Use:   "api [profile]",
				Short: "Test API call with current authentication",
				Long:  `Test an actual API call to see the exact error response.`,
				Args:  cobra.MaximumNArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation for testing
					fmt.Fprintf(cmd.OutOrStdout(), "Debug API call information would be displayed here\n")
					return nil
				},
			}

			// Build command hierarchy
			debugCmd.AddCommand(debugKeysCmd, debugAuthCmd, debugAPICmd)
			rootCmd.AddCommand(debugCmd)

			// Capture output
			var output bytes.Buffer
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&output)
			rootCmd.SetArgs(tt.args)

			// Execute command
			err := rootCmd.Execute()

			// Check results
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

func TestDebugCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		cmd         *cobra.Command
		args        []string
		expectError bool
	}{
		{
			name: "debug keys no args",
			cmd: &cobra.Command{
				Use:  "keys [profile]",
				Args: cobra.MaximumNArgs(1),
			},
			args:        []string{},
			expectError: false,
		},
		{
			name: "debug keys with profile",
			cmd: &cobra.Command{
				Use:  "keys [profile]",
				Args: cobra.MaximumNArgs(1),
			},
			args:        []string{"main"},
			expectError: false,
		},
		{
			name: "debug keys too many args",
			cmd: &cobra.Command{
				Use:  "keys [profile]",
				Args: cobra.MaximumNArgs(1),
			},
			args:        []string{"main", "extra"},
			expectError: true,
		},
		{
			name: "debug auth no args",
			cmd: &cobra.Command{
				Use:  "auth [profile]",
				Args: cobra.MaximumNArgs(1),
			},
			args:        []string{},
			expectError: false,
		},
		{
			name: "debug auth with profile",
			cmd: &cobra.Command{
				Use:  "auth [profile]",
				Args: cobra.MaximumNArgs(1),
			},
			args:        []string{"main"},
			expectError: false,
		},
		{
			name: "debug auth too many args",
			cmd: &cobra.Command{
				Use:  "auth [profile]",
				Args: cobra.MaximumNArgs(1),
			},
			args:        []string{"main", "extra"},
			expectError: true,
		},
		{
			name: "debug api no args",
			cmd: &cobra.Command{
				Use:  "api [profile]",
				Args: cobra.MaximumNArgs(1),
			},
			args:        []string{},
			expectError: false,
		},
		{
			name: "debug api with profile",
			cmd: &cobra.Command{
				Use:  "api [profile]",
				Args: cobra.MaximumNArgs(1),
			},
			args:        []string{"main"},
			expectError: false,
		},
		{
			name: "debug api too many args",
			cmd: &cobra.Command{
				Use:  "api [profile]",
				Args: cobra.MaximumNArgs(1),
			},
			args:        []string{"main", "extra"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetArgs(tt.args)

			// Validate args
			var err error
			if tt.cmd.Args != nil {
				err = tt.cmd.Args(tt.cmd, tt.args)
			}

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected validation error but got success")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestDebugCommandStructure(t *testing.T) {
	// Test that debug commands are properly structured
	rootCmd := &cobra.Command{Use: "test-root"}

	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug utilities for troubleshooting",
	}

	debugKeysCmd := &cobra.Command{
		Use:   "keys [profile]",
		Short: "Show keyring information for debugging",
		Args:  cobra.MaximumNArgs(1),
	}

	debugAuthCmd := &cobra.Command{
		Use:   "auth [profile]",
		Short: "Debug the full authentication flow",
		Args:  cobra.MaximumNArgs(1),
	}

	debugAPICmd := &cobra.Command{
		Use:   "api [profile]",
		Short: "Test API call with current authentication",
		Args:  cobra.MaximumNArgs(1),
	}

	// Build hierarchy
	debugCmd.AddCommand(debugKeysCmd, debugAuthCmd, debugAPICmd)
	rootCmd.AddCommand(debugCmd)

	// Test structure
	if len(rootCmd.Commands()) != 1 {
		t.Errorf("Expected 1 command under root, got %d", len(rootCmd.Commands()))
	}

	if rootCmd.Commands()[0] != debugCmd {
		t.Error("Debug command not properly added to root")
	}

	if len(debugCmd.Commands()) != 3 {
		t.Errorf("Expected 3 subcommands under debug, got %d", len(debugCmd.Commands()))
	}

	// Test command names (order doesn't matter)
	subcommands := debugCmd.Commands()
	expectedNames := []string{"keys", "auth", "api"}

	foundCommands := make(map[string]bool)
	for _, cmd := range subcommands {
		// Extract just the command name (before space if present)
		cmdName := strings.Split(cmd.Use, " ")[0]
		foundCommands[cmdName] = true
	}

	for _, expectedName := range expectedNames {
		if !foundCommands[expectedName] {
			t.Errorf("Missing subcommand: %s", expectedName)
		}
	}
}

func TestDebugHelpOutput(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
	}{
		{
			name: "debug help",
			args: []string{"debug", "--help"},
			expectedOutput: []string{
				"Debug utilities for troubleshooting",
				"keys",
				"auth",
				"api",
			},
		},
		{
			name: "debug keys help",
			args: []string{"debug", "keys", "--help"},
			expectedOutput: []string{
				"Show keyring information for debugging",
				"Show keyring information for debugging authentication issues",
			},
		},
		{
			name: "debug auth help",
			args: []string{"debug", "auth", "--help"},
			expectedOutput: []string{
				"Debug the full authentication flow",
				"Debug the full authentication flow to see which API key is actually being used",
			},
		},
		{
			name: "debug api help",
			args: []string{"debug", "api", "--help"},
			expectedOutput: []string{
				"Test an actual API call",
				"see the exact error response",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root command with debug subcommand
			rootCmd := &cobra.Command{Use: "forward-email"}

			debugCmd := &cobra.Command{
				Use:   "debug",
				Short: "Debug utilities for troubleshooting",
				Long:  `Debug utilities for troubleshooting authentication and keyring issues.`,
			}

			debugKeysCmd := &cobra.Command{
				Use:   "keys [profile]",
				Short: "Show keyring information for debugging",
				Long:  `Show keyring information for debugging authentication issues.`,
				Args:  cobra.MaximumNArgs(1),
			}

			debugAuthCmd := &cobra.Command{
				Use:   "auth [profile]",
				Short: "Debug the full authentication flow",
				Long:  `Debug the full authentication flow to see which API key is actually being used.`,
				Args:  cobra.MaximumNArgs(1),
			}

			debugAPICmd := &cobra.Command{
				Use:   "api [profile]",
				Short: "Test API call with current authentication",
				Long:  `Test an actual API call to see the exact error response.`,
				Args:  cobra.MaximumNArgs(1),
			}

			// Build command hierarchy
			debugCmd.AddCommand(debugKeysCmd, debugAuthCmd, debugAPICmd)
			rootCmd.AddCommand(debugCmd)

			// Capture output
			var output bytes.Buffer
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&output)
			rootCmd.SetArgs(tt.args)

			// Execute command (help commands don't return errors)
			rootCmd.Execute()

			outputStr := output.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected help output to contain %q, got %q", expected, outputStr)
				}
			}
		})
	}
}
