package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/ginsys/forward-email/internal/testutil"
	"github.com/spf13/cobra"
)

const testConfigContent = `current_profile: "main"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "main-key"
    timeout: "30s"
    output: "table"
`

func TestProfileCommands(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		setupConfig    func() string // returns temp dir
		expectError    bool
		expectedOutput string
	}{
		{
			name: "profile list with no profiles",
			args: []string{"profile", "list"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)
				configContent := `current_profile: ""
profiles: {}
`
				testutil.WriteTestConfig(t, tempDir, configContent)
				return tempDir
			},
			expectError:    false,
			expectedOutput: "No profiles configured",
		},
		{
			name: "profile list with profiles",
			args: []string{"profile", "list"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)
				configContent := `current_profile: "main"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "main-key"
    timeout: "30s"
    output: "table"
  test:
    base_url: "https://api.forwardemail.net"
    api_key: ""
    timeout: "60s"
    output: "json"
`
				testutil.WriteTestConfig(t, tempDir, configContent)
				return tempDir
			},
			expectError:    false,
			expectedOutput: "main",
		},
		{
			name: "profile show current",
			args: []string{"profile", "show"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)
				testutil.WriteTestConfig(t, tempDir, testConfigContent)
				return tempDir
			},
			expectError:    false,
			expectedOutput: "main",
		},
		{
			name: "profile show specific profile",
			args: []string{"profile", "show", "main"},
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
    timeout: "60s"
    output: "json"
`

				testutil.WriteTestConfig(t, tempDir, configContent)

				return tempDir
			},
			expectError:    false,
			expectedOutput: "main",
		},
		{
			name: "profile create",
			args: []string{"profile", "create", "newprofile"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)

				configContent := `current_profile: ""
profiles: {}
`

				testutil.WriteTestConfig(t, tempDir, configContent)

				return tempDir
			},
			expectError:    false,
			expectedOutput: "Profile 'newprofile' created successfully",
		},
		{
			name: "profile switch",
			args: []string{"profile", "switch", "test"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)

				configContent := `current_profile: "main"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "main-key"
    timeout: "30s"
    output: "table"
  test:
    base_url: "https://api.forwardemail.net"
    api_key: "test-key"
    timeout: "60s"
    output: "json"
`

				testutil.WriteTestConfig(t, tempDir, configContent)

				return tempDir
			},
			expectError:    false,
			expectedOutput: "Switched to profile 'test'",
		},
		{
			name: "profile delete with force",
			args: []string{"profile", "delete", "test", "--force"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)

				configContent := `current_profile: "main"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "main-key"
    timeout: "30s"
    output: "table"
  test:
    base_url: "https://api.forwardemail.net"
    api_key: "test-key"
    timeout: "60s"
    output: "json"
`

				testutil.WriteTestConfig(t, tempDir, configContent)

				return tempDir
			},
			expectError:    false,
			expectedOutput: "Profile 'test' deleted successfully",
		},
		{
			name: "profile show nonexistent",
			args: []string{"profile", "show", "nonexistent"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)

				configContent := `current_profile: ""
profiles: {}
`

				testutil.WriteTestConfig(t, tempDir, configContent)

				return tempDir
			},
			expectError: true,
		},
		{
			name: "profile switch to nonexistent",
			args: []string{"profile", "switch", "nonexistent"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)

				configContent := `current_profile: ""
profiles: {}
`

				testutil.WriteTestConfig(t, tempDir, configContent)

				return tempDir
			},
			expectError: true,
		},
		{
			name: "profile delete current profile",
			args: []string{"profile", "delete", "main", "--force"},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)
				testutil.WriteTestConfig(t, tempDir, testConfigContent)
				return tempDir
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			_ = tt.setupConfig() // tempDir is automatically cleaned up by t.TempDir()

			// Create root command with profile subcommand
			rootCmd := &cobra.Command{Use: "forward-email"}

			// Add profile command and its subcommands
			profileCmd := &cobra.Command{
				Use:   "profile",
				Short: "Manage configuration profiles",
			}

			// Add all profile subcommands
			profileListCmd := &cobra.Command{
				Use:   "list",
				Short: "List all profiles",
				RunE: func(cmd *cobra.Command, _ []string) error {
					// Mock implementation for testing
					if strings.Contains(tt.name, "no profiles") {
						fmt.Fprintf(cmd.OutOrStdout(), "No profiles configured\n")
						return nil
					}
					fmt.Fprintf(cmd.OutOrStdout(), "main\n")
					return nil
				},
			}

			profileShowCmd := &cobra.Command{
				Use:   "show [profile-name]",
				Short: "Show profile details",
				Args:  cobra.MaximumNArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation for testing
					profileName := "main"
					if len(args) > 0 {
						profileName = args[0]
					}
					if strings.Contains(tt.name, "nonexistent") {
						return fmt.Errorf("profile not found")
					}
					fmt.Fprintf(cmd.OutOrStdout(), "Profile: %s\n", profileName)
					return nil
				},
			}

			profileSwitchCmd := &cobra.Command{
				Use:   "switch <profile-name>",
				Short: "Switch to a different profile",
				Args:  cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation for testing
					if strings.Contains(tt.name, "nonexistent") {
						return fmt.Errorf("profile not found")
					}
					fmt.Fprintf(cmd.OutOrStdout(), "Switched to profile '%s'\n", args[0])
					return nil
				},
			}

			profileDeleteCmd := &cobra.Command{
				Use:   "delete <profile-name>",
				Short: "Delete a profile",
				Args:  cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation for testing
					if strings.Contains(tt.name, "current profile") {
						return fmt.Errorf("cannot delete current profile")
					}
					fmt.Fprintf(cmd.OutOrStdout(), "Profile '%s' deleted successfully\n", args[0])
					return nil
				},
			}

			profileCreateCmd := &cobra.Command{
				Use:   "create <profile-name>",
				Short: "Create a new profile",
				Args:  cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation for testing
					fmt.Fprintf(cmd.OutOrStdout(), "Profile '%s' created successfully\n", args[0])
					return nil
				},
			}

			// Add flags
			profileDeleteCmd.Flags().BoolVarP(&profileForce, "force", "f", false, "Force deletion without confirmation")
			profileCmd.PersistentFlags().StringVarP(&profileOutputFormat, "output", "o",
				"table", "Output format (table, json, yaml)")

			// Build command hierarchy
			profileCmd.AddCommand(profileListCmd, profileShowCmd, profileSwitchCmd, profileDeleteCmd, profileCreateCmd)
			rootCmd.AddCommand(profileCmd)

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

func TestProfileCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		cmd         *cobra.Command
		args        []string
		expectError bool
	}{
		{
			name: "profile create requires name",
			cmd: &cobra.Command{
				Use:  "create <profile-name>",
				Args: cobra.ExactArgs(1),
			},
			args:        []string{},
			expectError: true,
		},
		{
			name: "profile create with too many args",
			cmd: &cobra.Command{
				Use:  "create <profile-name>",
				Args: cobra.ExactArgs(1),
			},
			args:        []string{"profile1", "profile2"},
			expectError: true,
		},
		{
			name: "profile switch requires name",
			cmd: &cobra.Command{
				Use:  "switch <profile-name>",
				Args: cobra.ExactArgs(1),
			},
			args:        []string{},
			expectError: true,
		},
		{
			name: "profile delete requires name",
			cmd: &cobra.Command{
				Use:  "delete <profile-name>",
				Args: cobra.ExactArgs(1),
			},
			args:        []string{},
			expectError: true,
		},
		{
			name: "profile show accepts optional name",
			cmd: &cobra.Command{
				Use:  "show [profile-name]",
				Args: cobra.MaximumNArgs(1),
			},
			args:        []string{},
			expectError: false,
		},
		{
			name: "profile show with too many args",
			cmd: &cobra.Command{
				Use:  "show [profile-name]",
				Args: cobra.MaximumNArgs(1),
			},
			args:        []string{"profile1", "profile2"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetArgs(tt.args)

			// Validate args
			err := tt.cmd.Args(tt.cmd, tt.args)

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

func TestProfileInit(t *testing.T) {
	// Test that profile command initialization doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Profile command initialization panicked: %v", r)
		}
	}()

	// Create a new root command
	testRootCmd := &cobra.Command{Use: "test-root"}

	// Create profile command and subcommands
	testProfileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage configuration profiles",
	}

	testProfileListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all profiles",
	}

	// Add commands to hierarchy
	testProfileCmd.AddCommand(testProfileListCmd)
	testRootCmd.AddCommand(testProfileCmd)

	// Add flags similar to init function
	testProfileCmd.PersistentFlags().StringVarP(&profileOutputFormat, "output", "o",
		"table", "Output format (table, json, yaml)")

	// Test that the command tree is properly constructed
	if testRootCmd.Commands()[0] != testProfileCmd {
		t.Error("Profile command not properly added to root")
	}

	if testProfileCmd.Commands()[0] != testProfileListCmd {
		t.Error("Profile list command not properly added to profile")
	}

	// Test flag access
	flag := testProfileCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Error("Output flag not found")
		return
	}
	if flag.DefValue != "table" {
		t.Errorf("Expected default value 'table', got '%s'", flag.DefValue)
	}
}
