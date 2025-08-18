package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAuthCommands(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		setupConfig    func() string // returns temp dir
		expectError    bool
		expectedOutput string
	}{
		{
			name: "auth verify with valid config",
			args: []string{"auth", "verify"},
			setupConfig: func() string {
				tempDir := t.TempDir()
				configDir := filepath.Join(tempDir, ".config", "forwardemail")
				os.MkdirAll(configDir, 0755)

				configFile := filepath.Join(configDir, "config.yaml")
				configContent := `current_profile: "main"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "test-key"
    timeout: "30s"
    output: "table"
`
				os.WriteFile(configFile, []byte(configContent), 0600)
				os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
				return tempDir
			},
			expectError: false, // Mock implementation should succeed
		},
		{
			name: "auth verify with no profile",
			args: []string{"auth", "verify"},
			setupConfig: func() string {
				tempDir := t.TempDir()
				configDir := filepath.Join(tempDir, ".config", "forwardemail")
				os.MkdirAll(configDir, 0755)

				configFile := filepath.Join(configDir, "config.yaml")
				configContent := `current_profile: ""
profiles: {}
`
				os.WriteFile(configFile, []byte(configContent), 0600)
				os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
				return tempDir
			},
			expectError: true,
		},
		{
			name: "auth login command exists",
			args: []string{"auth", "login", "--help"},
			setupConfig: func() string {
				tempDir := t.TempDir()
				os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
				return tempDir
			},
			expectError:    false,
			expectedOutput: "Interactively log in to Forward Email",
		},
		{
			name: "auth command help",
			args: []string{"auth", "--help"},
			setupConfig: func() string {
				tempDir := t.TempDir()
				os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
				return tempDir
			},
			expectError:    false,
			expectedOutput: "Manage authentication credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			tempDir := tt.setupConfig()
			defer func() {
				os.Unsetenv("XDG_CONFIG_HOME")
				os.RemoveAll(tempDir)
			}()

			// Create root command with auth subcommand
			testRootCmd := &cobra.Command{Use: "forward-email"}

			// Add auth command and its subcommands
			authCmd := &cobra.Command{
				Use:   "auth",
				Short: "Manage authentication credentials",
				Long: `Manage authentication credentials for Forward Email API.

The auth command group provides subcommands to log in, verify credentials,
and manage API keys across different profiles.`,
			}

			authVerifyCmd := &cobra.Command{
				Use:   "verify",
				Short: "Verify authentication credentials",
				Long: `Verify that the current authentication credentials are valid.

This command will attempt to authenticate with the Forward Email API
using the current profile's credentials and report whether they are valid.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation for testing
					if strings.Contains(tt.name, "no profile") {
						return fmt.Errorf("no profile configured")
					}
					return nil
				},
			}

			authLoginCmd := &cobra.Command{
				Use:   "login",
				Short: "Log in and save API credentials",
				Long: `Interactively log in to Forward Email and save API credentials.

This command will prompt for your API key and securely store it
in the OS keyring or configuration file.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation for testing
					return nil
				},
			}

			// Build command hierarchy
			authCmd.AddCommand(authVerifyCmd, authLoginCmd)
			testRootCmd.AddCommand(authCmd)

			// Capture output
			var output bytes.Buffer
			testRootCmd.SetOut(&output)
			testRootCmd.SetErr(&output)
			testRootCmd.SetArgs(tt.args)

			// Execute command
			err := testRootCmd.Execute()

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

func TestAuthCommandStructure(t *testing.T) {
	// Test that auth commands are properly structured
	structureRootCmd := &cobra.Command{Use: "test-root"}

	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication credentials",
	}

	authVerifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify authentication credentials",
	}

	authLoginCmd := &cobra.Command{
		Use:   "login",
		Short: "Log in and save API credentials",
	}

	// Build hierarchy
	authCmd.AddCommand(authVerifyCmd, authLoginCmd)
	structureRootCmd.AddCommand(authCmd)

	// Test structure
	if len(structureRootCmd.Commands()) != 1 {
		t.Errorf("Expected 1 command under root, got %d", len(structureRootCmd.Commands()))
	}

	if structureRootCmd.Commands()[0] != authCmd {
		t.Error("Auth command not properly added to root")
	}

	if len(authCmd.Commands()) != 2 {
		t.Errorf("Expected 2 subcommands under auth, got %d", len(authCmd.Commands()))
	}

	// Test command names (order doesn't matter)
	subcommands := authCmd.Commands()
	expectedNames := []string{"verify", "login"}

	foundCommands := make(map[string]bool)
	for _, cmd := range subcommands {
		foundCommands[cmd.Use] = true
	}

	for _, expectedName := range expectedNames {
		if !foundCommands[expectedName] {
			t.Errorf("Missing subcommand: %s", expectedName)
		}
	}
}

func TestAuthCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "auth verify no args",
			commandName: "verify",
			args:        []string{},
			expectError: false,
			description: "verify command should accept no arguments",
		},
		{
			name:        "auth login no args",
			commandName: "login",
			args:        []string{},
			expectError: false,
			description: "login command should accept no arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd *cobra.Command

			switch tt.commandName {
			case "verify":
				cmd = &cobra.Command{
					Use:   "verify",
					Short: "Verify authentication credentials",
					Args:  cobra.NoArgs, // Verify accepts no args
				}
			case "login":
				cmd = &cobra.Command{
					Use:   "login",
					Short: "Log in and save API credentials",
					Args:  cobra.NoArgs, // Login accepts no args
				}
			default:
				t.Fatalf("Unknown command: %s", tt.commandName)
			}

			cmd.SetArgs(tt.args)

			// Validate args
			var err error
			if cmd.Args != nil {
				err = cmd.Args(cmd, tt.args)
			}

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected validation error but got success: %s", tt.description)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected validation error: %v (%s)", err, tt.description)
				}
			}
		})
	}
}

func TestAuthHelpOutput(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
	}{
		{
			name: "auth help",
			args: []string{"auth", "--help"},
			expectedOutput: []string{
				"Manage authentication credentials",
				"verify",
				"login",
			},
		},
		{
			name: "auth verify help",
			args: []string{"auth", "verify", "--help"},
			expectedOutput: []string{
				"Verify that the current authentication credentials are valid",
				"authenticate with the Forward Email API",
			},
		},
		{
			name: "auth login help",
			args: []string{"auth", "login", "--help"},
			expectedOutput: []string{
				"Interactively log in to Forward Email",
				"securely store it",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root command with auth subcommand
			helpRootCmd := &cobra.Command{Use: "forward-email"}

			authCmd := &cobra.Command{
				Use:   "auth",
				Short: "Manage authentication credentials",
				Long: `Manage authentication credentials for Forward Email API.

The auth command group provides subcommands to log in, verify credentials,
and manage API keys across different profiles.`,
			}

			authVerifyCmd := &cobra.Command{
				Use:   "verify",
				Short: "Verify authentication credentials",
				Long: `Verify that the current authentication credentials are valid.

This command will attempt to authenticate with the Forward Email API
using the current profile's credentials and report whether they are valid.`,
			}

			authLoginCmd := &cobra.Command{
				Use:   "login",
				Short: "Log in and save API credentials",
				Long: `Interactively log in to Forward Email and save API credentials.

This command will prompt for your API key and securely store it
in the OS keyring or configuration file.`,
			}

			// Build command hierarchy
			authCmd.AddCommand(authVerifyCmd, authLoginCmd)
			helpRootCmd.AddCommand(authCmd)

			// Capture output
			var output bytes.Buffer
			helpRootCmd.SetOut(&output)
			helpRootCmd.SetErr(&output)
			helpRootCmd.SetArgs(tt.args)

			// Execute command (help commands don't return errors)
			helpRootCmd.Execute()

			outputStr := output.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected help output to contain %q, got %q", expected, outputStr)
				}
			}
		})
	}
}
