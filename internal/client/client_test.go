package client

import (
	"testing"

	"github.com/ginsys/forward-email/internal/testutil"
	"github.com/spf13/viper"
)

func TestNewAPIClient(t *testing.T) {
	// Save original state
	originalConfig := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalConfig {
			viper.Set(k, v)
		}
	}()

	tests := []struct {
		name          string
		setupEnv      func()
		setupConfig   func() string // returns config dir
		expectedError string
		shouldSucceed bool
	}{
		{
			name: "no profile configured",
			setupEnv: func() {
				testutil.ResetViper()
			},
			setupConfig: func() string {
				// Create empty config
				tempDir := testutil.SetupTempConfig(t)
				configContent := `current_profile: ""
profiles: {}
`
				testutil.WriteTestConfig(t, tempDir, configContent)
				return tempDir
			},
			expectedError: "no profile configured",
			shouldSucceed: false,
		},
		{
			name: "profile specified via viper flag",
			setupEnv: func() {
				testutil.ResetViper()
				viper.Set("profile", "test")
			},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)
				configContent := `current_profile: "default"
profiles:
  test:
    base_url: "https://api.forwardemail.net"
    api_key: "test-key"
    timeout: "30s"
    output: "table"
  default:
    base_url: "https://api.forwardemail.net"
    api_key: ""
    timeout: "30s"
    output: "table"
`
				testutil.WriteTestConfig(t, tempDir, configContent)
				return tempDir
			},
			shouldSucceed: true,
		},
		{
			name: "uses current profile from config",
			setupEnv: func() {
				testutil.ResetViper()
			},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)
				configContent := `current_profile: "main"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "main-key"
    timeout: "30s"
    output: "table"
`
				testutil.WriteTestConfig(t, tempDir, configContent)
				return tempDir
			},
			shouldSucceed: true,
		},
		{
			name: "profile with no API key",
			setupEnv: func() {
				testutil.ResetViper()
				viper.Set("profile", "empty")
			},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)
				configContent := `current_profile: "main"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "main-key"
    timeout: "30s"
    output: "table"
  empty:
    base_url: "https://api.forwardemail.net"
    api_key: ""
    timeout: "30s"
    output: "table"
`
				testutil.WriteTestConfig(t, tempDir, configContent)
				return tempDir
			},
			shouldSucceed: true, // Client creation succeeds, auth fails later
		},
		{
			name: "custom base URL",
			setupEnv: func() {
				testutil.ResetViper()
				viper.Set("api_base_url", "https://custom.api.url")
			},
			setupConfig: func() string {
				tempDir := testutil.SetupTempConfig(t)
				configContent := `current_profile: "main"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "main-key"
    timeout: "30s"
    output: "table"
`
				testutil.WriteTestConfig(t, tempDir, configContent)
				return tempDir
			},
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			tt.setupEnv()
			_ = tt.setupConfig() // tempDir is automatically cleaned up by t.TempDir()

			// Test the function
			client, err := NewAPIClient()

			if tt.shouldSucceed {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
				if client == nil {
					t.Fatal("Expected client but got nil")
				}
				// Verify that client was created with expected base URL
				if tt.name == "custom base URL" {
					expectedURL := "https://custom.api.url"
					if client.BaseURL.String() != expectedURL {
						t.Errorf("Expected base URL %s, got %s", expectedURL, client.BaseURL.String())
					}
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got success")
				}
				if tt.expectedError != "" && !containsString(err.Error(), tt.expectedError) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedError, err.Error())
				}
				if client != nil {
					t.Error("Expected nil client on error")
				}
			}
		})
	}
}

func TestNewAPIClient_ConfigLoadFailure(t *testing.T) {
	// Save original state
	originalConfig := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalConfig {
			viper.Set(k, v)
		}
	}()

	testutil.ResetViper()

	// This shouldn't fail because config.Load() handles missing config gracefully
	// But let's test with a malformed config file
	tempDir := testutil.SetupTempConfig(t)

	// Create malformed YAML config
	malformedContent := `current_profile: "test"
profiles:
  test:
    base_url: "https://api.forwardemail.net"
    api_key: malformed yaml content here: [
`
	testutil.WriteTestConfig(t, tempDir, malformedContent)

	_, err := NewAPIClient()
	if err == nil {
		t.Fatal("Expected error due to malformed config, but got success")
	}

	expectedError := "failed to load config"
	if !containsString(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

// Helper function to check if a string contains a substring
func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(haystack[0:len(needle)] == needle ||
			(len(haystack) > len(needle) && containsString(haystack[1:], needle)))
}
