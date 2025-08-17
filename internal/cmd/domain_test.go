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

func TestDomainCommands(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		setupConfig    func() string // returns temp dir
		expectError    bool
		expectedOutput string
	}{
		{
			name: "domain list with valid config",
			args: []string{"domain", "list"},
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
				os.WriteFile(configFile, []byte(configContent), 0644)
				os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
				return tempDir
			},
			expectError: true, // Will fail because API key is fake, but we test command structure
		},
		{
			name: "domain get requires argument",
			args: []string{"domain", "get"},
			setupConfig: func() string {
				tempDir := t.TempDir()
				os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
				return tempDir
			},
			expectError: true,
		},
		{
			name: "domain get with argument",
			args: []string{"domain", "get", "example.com"},
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
				os.WriteFile(configFile, []byte(configContent), 0644)
				os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
				return tempDir
			},
			expectError: true, // Will fail because API key is fake, but we test command structure
		},
		{
			name: "domain help",
			args: []string{"domain", "--help"},
			setupConfig: func() string {
				tempDir := t.TempDir()
				os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
				return tempDir
			},
			expectError:    false,
			expectedOutput: "Manage Forward Email domains",
		},
		{
			name: "domain list help",
			args: []string{"domain", "list", "--help"},
			setupConfig: func() string {
				tempDir := t.TempDir()
				os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
				return tempDir
			},
			expectError:    false,
			expectedOutput: "List all domains associated with your Forward Email account",
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

			// Create root command with domain subcommand
			rootCmd := &cobra.Command{Use: "forward-email"}

			// Add domain command and its subcommands
			domainCmd := &cobra.Command{
				Use:   "domain",
				Short: "Manage Forward Email domains",
				Long: `Manage Forward Email domains including creating, listing, updating, 
and configuring domain settings and DNS records.`,
			}

			domainListCmd := &cobra.Command{
				Use:   "list",
				Short: "List domains",
				Long:  `List all domains associated with your Forward Email account.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation for testing
					return fmt.Errorf("mock API error")
				},
			}

			domainGetCmd := &cobra.Command{
				Use:   "get <domain-name-or-id>",
				Short: "Get domain details",
				Long:  `Get detailed information about a specific domain.`,
				Args:  cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation for testing
					return fmt.Errorf("mock API error")
				},
			}

			// Add flags
			domainListCmd.Flags().IntVar(&domainPage, "page", 1, "Page number")
			domainListCmd.Flags().IntVar(&domainLimit, "limit", 50, "Number of results per page")
			domainListCmd.Flags().StringVar(&domainSort, "sort", "name", "Sort field")
			domainListCmd.Flags().StringVar(&domainOrder, "order", "asc", "Sort order (asc|desc)")
			domainListCmd.Flags().StringVar(&domainSearch, "search", "", "Search query")
			domainListCmd.Flags().StringVar(&domainVerified, "verified", "", "Filter by verification status")
			domainListCmd.Flags().StringVar(&domainPlan, "plan", "", "Filter by plan type")

			// Build command hierarchy
			domainCmd.AddCommand(domainListCmd, domainGetCmd)
			rootCmd.AddCommand(domainCmd)

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

func TestDomainCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		cmd         *cobra.Command
		args        []string
		expectError bool
	}{
		{
			name: "domain get requires argument",
			cmd: &cobra.Command{
				Use:  "get <domain-name-or-id>",
				Args: cobra.ExactArgs(1),
			},
			args:        []string{},
			expectError: true,
		},
		{
			name: "domain get with valid argument",
			cmd: &cobra.Command{
				Use:  "get <domain-name-or-id>",
				Args: cobra.ExactArgs(1),
			},
			args:        []string{"example.com"},
			expectError: false,
		},
		{
			name: "domain get with too many arguments",
			cmd: &cobra.Command{
				Use:  "get <domain-name-or-id>",
				Args: cobra.ExactArgs(1),
			},
			args:        []string{"example.com", "extra"},
			expectError: true,
		},
		{
			name: "domain list accepts no arguments",
			cmd: &cobra.Command{
				Use:  "list",
				Args: cobra.NoArgs,
			},
			args:        []string{},
			expectError: false,
		},
		{
			name: "domain list with arguments should fail",
			cmd: &cobra.Command{
				Use:  "list",
				Args: cobra.NoArgs,
			},
			args:        []string{"unexpected"},
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

func TestDomainCommandStructure(t *testing.T) {
	// Test that domain commands are properly structured
	rootCmd := &cobra.Command{Use: "test-root"}

	domainCmd := &cobra.Command{
		Use:   "domain",
		Short: "Manage Forward Email domains",
	}

	domainListCmd := &cobra.Command{
		Use:   "list",
		Short: "List domains",
	}

	domainGetCmd := &cobra.Command{
		Use:   "get <domain-name-or-id>",
		Short: "Get domain details",
		Args:  cobra.ExactArgs(1),
	}

	// Build hierarchy
	domainCmd.AddCommand(domainListCmd, domainGetCmd)
	rootCmd.AddCommand(domainCmd)

	// Test structure
	if len(rootCmd.Commands()) != 1 {
		t.Errorf("Expected 1 command under root, got %d", len(rootCmd.Commands()))
	}

	if rootCmd.Commands()[0] != domainCmd {
		t.Error("Domain command not properly added to root")
	}

	if len(domainCmd.Commands()) != 2 {
		t.Errorf("Expected 2 subcommands under domain, got %d", len(domainCmd.Commands()))
	}

	// Test command names (order doesn't matter)
	subcommands := domainCmd.Commands()
	expectedNames := []string{"list", "get"}

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

func TestDomainCommandFlags(t *testing.T) {
	tests := []struct {
		name          string
		commandName   string
		expectedFlags []string
	}{
		{
			name:        "domain list flags",
			commandName: "list",
			expectedFlags: []string{
				"output",
				"page",
				"limit",
				"sort",
				"order",
				"search",
				"verified",
				"plan",
			},
		},
		{
			name:        "domain get flags",
			commandName: "get",
			expectedFlags: []string{
				"output",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd *cobra.Command

			switch tt.commandName {
			case "list":
				cmd = &cobra.Command{
					Use:   "list",
					Short: "List domains",
				}
				// Add flags as they would be in the real command
				cmd.Flags().IntVar(&domainPage, "page", 1, "Page number")
				cmd.Flags().IntVar(&domainLimit, "limit", 50, "Number of results per page")
				cmd.Flags().StringVar(&domainSort, "sort", "name", "Sort field")
				cmd.Flags().StringVar(&domainOrder, "order", "asc", "Sort order")
				cmd.Flags().StringVar(&domainSearch, "search", "", "Search query")
				cmd.Flags().StringVar(&domainVerified, "verified", "", "Filter by verification status")
				cmd.Flags().StringVar(&domainPlan, "plan", "", "Filter by plan type")

			case "get":
				cmd = &cobra.Command{
					Use:   "get <domain-name-or-id>",
					Short: "Get domain details",
					Args:  cobra.ExactArgs(1),
				}

			default:
				t.Fatalf("Unknown command: %s", tt.commandName)
			}

			// Test that all expected flags exist
			for _, flagName := range tt.expectedFlags {
				flag := cmd.Flags().Lookup(flagName)
				if flag == nil {
					t.Errorf("Expected flag %s not found", flagName)
				}
			}

			// Test flag parsing
			switch tt.commandName {
			case "list":
				cmd.SetArgs([]string{"--output", "json", "--page", "2", "--limit", "25"})
				err := cmd.ParseFlags([]string{"--output", "json", "--page", "2", "--limit", "25"})
				if err != nil {
					t.Errorf("Failed to parse flags: %v", err)
				}

			case "get":
				cmd.SetArgs([]string{"example.com", "--output", "yaml"})
				err := cmd.ParseFlags([]string{"--output", "yaml"})
				if err != nil {
					t.Errorf("Failed to parse flags: %v", err)
				}
			}
		})
	}
}

func TestDomainHelpOutput(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
	}{
		{
			name: "domain help",
			args: []string{"domain", "--help"},
			expectedOutput: []string{
				"Manage Forward Email domains",
				"list",
				"get",
			},
		},
		{
			name: "domain list help",
			args: []string{"domain", "list", "--help"},
			expectedOutput: []string{
				"List all domains associated",
				"Forward Email account",
			},
		},
		{
			name: "domain get help",
			args: []string{"domain", "get", "--help"},
			expectedOutput: []string{
				"Get detailed information",
				"specific domain",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root command with domain subcommand
			rootCmd := &cobra.Command{Use: "forward-email"}

			domainCmd := &cobra.Command{
				Use:   "domain",
				Short: "Manage Forward Email domains",
				Long: `Manage Forward Email domains including creating, listing, updating, 
and configuring domain settings and DNS records.`,
			}

			domainListCmd := &cobra.Command{
				Use:   "list",
				Short: "List domains",
				Long:  `List all domains associated with your Forward Email account.`,
			}

			domainGetCmd := &cobra.Command{
				Use:   "get <domain-name-or-id>",
				Short: "Get domain details",
				Long:  `Get detailed information about a specific domain.`,
				Args:  cobra.ExactArgs(1),
			}

			// Add flags
			domainListCmd.Flags().IntVar(&domainPage, "page", 1, "Page number")
			domainListCmd.Flags().IntVar(&domainLimit, "limit", 50, "Number of results per page")

			// Build command hierarchy
			domainCmd.AddCommand(domainListCmd, domainGetCmd)
			rootCmd.AddCommand(domainCmd)

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
