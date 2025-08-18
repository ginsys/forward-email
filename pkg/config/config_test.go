package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name         string
		setupConfig  func() string // returns temp dir path
		shouldError  bool
		expectedData *Config
	}{
		{
			name: "load valid config",
			setupConfig: func() string {
				tempDir := t.TempDir()
				configDir := filepath.Join(tempDir, ".config", "forwardemail")
				os.MkdirAll(configDir, 0755)

				configFile := filepath.Join(configDir, "config.yaml")
				configContent := `current_profile: "test"
profiles:
  test:
    base_url: "https://api.forwardemail.net"
    api_key: "test-key"
    username: "testuser"
    password: "testpass"
    timeout: "30s"
    output: "table"
  prod:
    base_url: "https://api.forwardemail.net"
    api_key: ""
    timeout: "60s"
    output: "json"
`
				os.WriteFile(configFile, []byte(configContent), 0600)
				os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
				return tempDir
			},
			shouldError: false,
			expectedData: &Config{
				CurrentProfile: "test",
				Profiles: map[string]Profile{
					"test": {
						BaseURL:  "https://api.forwardemail.net",
						APIKey:   "test-key",
						Username: "testuser",
						Password: "testpass",
						Timeout:  "30s",
						Output:   "table",
					},
					"prod": {
						BaseURL: "https://api.forwardemail.net",
						APIKey:  "",
						Timeout: "60s",
						Output:  "json",
					},
				},
			},
		},
		{
			name: "load with empty config file",
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
			shouldError: false,
			expectedData: &Config{
				CurrentProfile: "",
				Profiles:       map[string]Profile{},
			},
		},
		{
			name: "no config file exists",
			setupConfig: func() string {
				tempDir := t.TempDir()
				os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
				return tempDir
			},
			shouldError: false,
			expectedData: &Config{
				CurrentProfile: "",
				Profiles:       nil,
			},
		},
		{
			name: "malformed config file",
			setupConfig: func() string {
				tempDir := t.TempDir()
				configDir := filepath.Join(tempDir, ".config", "forwardemail")
				os.MkdirAll(configDir, 0755)

				configFile := filepath.Join(configDir, "config.yaml")
				configContent := `current_profile: "test"
profiles:
  test:
    base_url: invalid yaml: [
`
				os.WriteFile(configFile, []byte(configContent), 0600)
				os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
				return tempDir
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := tt.setupConfig()
			defer func() {
				os.Unsetenv("XDG_CONFIG_HOME")
				os.RemoveAll(tempDir)
			}()

			config, err := Load()

			if tt.shouldError {
				if err == nil {
					t.Fatal("Expected error but got success")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if config == nil {
				t.Fatal("Expected config but got nil")
			}

			if config.CurrentProfile != tt.expectedData.CurrentProfile {
				t.Errorf("Expected current profile %q, got %q", tt.expectedData.CurrentProfile, config.CurrentProfile)
			}

			if tt.expectedData.Profiles == nil && config.Profiles != nil {
				t.Errorf("Expected nil profiles, got %v", config.Profiles)
			} else if tt.expectedData.Profiles != nil {
				if len(config.Profiles) != len(tt.expectedData.Profiles) {
					t.Errorf("Expected %d profiles, got %d", len(tt.expectedData.Profiles), len(config.Profiles))
				}

				for name, expectedProfile := range tt.expectedData.Profiles {
					actualProfile, exists := config.Profiles[name]
					if !exists {
						t.Errorf("Expected profile %q not found", name)
						continue
					}

					if actualProfile.BaseURL != expectedProfile.BaseURL {
						t.Errorf("Profile %q: expected base URL %q, got %q", name, expectedProfile.BaseURL, actualProfile.BaseURL)
					}
					if actualProfile.APIKey != expectedProfile.APIKey {
						t.Errorf("Profile %q: expected API key %q, got %q", name, expectedProfile.APIKey, actualProfile.APIKey)
					}
					if actualProfile.Timeout != expectedProfile.Timeout {
						t.Errorf("Profile %q: expected timeout %q, got %q", name, expectedProfile.Timeout, actualProfile.Timeout)
					}
					if actualProfile.Output != expectedProfile.Output {
						t.Errorf("Profile %q: expected output %q, got %q", name, expectedProfile.Output, actualProfile.Output)
					}
				}
			}
		})
	}
}

func TestLoadWithoutDefaults(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "forwardemail")
	os.MkdirAll(configDir, 0755)

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `current_profile: "main"
profiles:
  main:
    base_url: "https://api.forwardemail.net"
    api_key: "main-key"
    timeout: "30s"
    output: "table"
`
	os.WriteFile(configFile, []byte(configContent), 0600)
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
	defer func() {
		os.Unsetenv("XDG_CONFIG_HOME")
		os.RemoveAll(tempDir)
	}()

	config, err := LoadWithoutDefaults()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if config.CurrentProfile != "main" {
		t.Errorf("Expected current profile 'main', got %q", config.CurrentProfile)
	}

	if len(config.Profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(config.Profiles))
	}

	mainProfile, exists := config.Profiles["main"]
	if !exists {
		t.Fatal("Expected 'main' profile not found")
	}

	if mainProfile.APIKey != "main-key" {
		t.Errorf("Expected API key 'main-key', got %q", mainProfile.APIKey)
	}
}

func TestConfig_Save(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
	defer func() {
		os.Unsetenv("XDG_CONFIG_HOME")
		os.RemoveAll(tempDir)
	}()

	config := &Config{
		CurrentProfile: "test",
		Profiles: map[string]Profile{
			"test": {
				BaseURL: "https://api.forwardemail.net",
				APIKey:  "test-key",
				Timeout: "30s",
				Output:  "table",
			},
		},
	}

	err := config.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify the file was created
	configPath := filepath.Join(tempDir, ".config", "forwardemail", "config.yaml")
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		t.Fatal("Config file was not created")
	}

	// Load the config back and verify
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.CurrentProfile != config.CurrentProfile {
		t.Errorf("Expected current profile %q, got %q", config.CurrentProfile, loadedConfig.CurrentProfile)
	}

	if len(loadedConfig.Profiles) != len(config.Profiles) {
		t.Errorf("Expected %d profiles, got %d", len(config.Profiles), len(loadedConfig.Profiles))
	}

	testProfile, exists := loadedConfig.Profiles["test"]
	if !exists {
		t.Fatal("Expected 'test' profile not found")
	}

	if testProfile.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got %q", testProfile.APIKey)
	}
}

func TestConfig_GetProfile(t *testing.T) {
	config := &Config{
		CurrentProfile: "current",
		Profiles: map[string]Profile{
			"current": {
				BaseURL: "https://current.api.url",
				APIKey:  "current-key",
			},
			"other": {
				BaseURL: "https://other.api.url",
				APIKey:  "other-key",
			},
		},
	}

	tests := []struct {
		name         string
		profileName  string
		shouldError  bool
		expectedName string
	}{
		{
			name:         "get current profile with empty name",
			profileName:  "",
			shouldError:  false,
			expectedName: "current",
		},
		{
			name:         "get specific profile",
			profileName:  "other",
			shouldError:  false,
			expectedName: "other",
		},
		{
			name:        "get nonexistent profile",
			profileName: "nonexistent",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := config.GetProfile(tt.profileName)

			if tt.shouldError {
				if err == nil {
					t.Fatal("Expected error but got success")
				}
				expectedError := "not found"
				if !containsString(err.Error(), expectedError) {
					t.Errorf("Expected error to contain %q, got %q", expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			expectedProfile := config.Profiles[tt.expectedName]
			if profile.BaseURL != expectedProfile.BaseURL {
				t.Errorf("Expected base URL %q, got %q", expectedProfile.BaseURL, profile.BaseURL)
			}
			if profile.APIKey != expectedProfile.APIKey {
				t.Errorf("Expected API key %q, got %q", expectedProfile.APIKey, profile.APIKey)
			}
		})
	}
}

func TestConfig_SetProfile(t *testing.T) {
	config := &Config{
		CurrentProfile: "current",
		Profiles: map[string]Profile{
			"current": {
				BaseURL: "https://current.api.url",
				APIKey:  "current-key",
			},
		},
	}

	newProfile := Profile{
		BaseURL: "https://new.api.url",
		APIKey:  "new-key",
		Timeout: "45s",
		Output:  "json",
	}

	// Test setting a new profile
	config.SetProfile("new", &newProfile)

	if len(config.Profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(config.Profiles))
	}

	savedProfile, exists := config.Profiles["new"]
	if !exists {
		t.Fatal("New profile was not saved")
	}

	if savedProfile.BaseURL != newProfile.BaseURL {
		t.Errorf("Expected base URL %q, got %q", newProfile.BaseURL, savedProfile.BaseURL)
	}

	// Test updating existing profile
	updatedProfile := Profile{
		BaseURL: "https://updated.api.url",
		APIKey:  "updated-key",
	}
	config.SetProfile("current", &updatedProfile)

	currentProfile, exists := config.Profiles["current"]
	if !exists {
		t.Fatal("Current profile was lost")
	}

	if currentProfile.BaseURL != updatedProfile.BaseURL {
		t.Errorf("Expected updated base URL %q, got %q", updatedProfile.BaseURL, currentProfile.BaseURL)
	}
}

func TestConfig_DeleteProfile(t *testing.T) {
	config := &Config{
		CurrentProfile: "current",
		Profiles: map[string]Profile{
			"current": {
				BaseURL: "https://current.api.url",
				APIKey:  "current-key",
			},
			"other": {
				BaseURL: "https://other.api.url",
				APIKey:  "other-key",
			},
		},
	}

	tests := []struct {
		name        string
		profileName string
		shouldError bool
		errorText   string
	}{
		{
			name:        "delete non-current profile",
			profileName: "other",
			shouldError: false,
		},
		{
			name:        "delete current profile",
			profileName: "current",
			shouldError: true,
			errorText:   "cannot delete current profile",
		},
		{
			name:        "delete nonexistent profile",
			profileName: "nonexistent",
			shouldError: true,
			errorText:   "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalCount := len(config.Profiles)
			err := config.DeleteProfile(tt.profileName)

			if tt.shouldError {
				if err == nil {
					t.Fatal("Expected error but got success")
				}
				if !containsString(err.Error(), tt.errorText) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorText, err.Error())
				}
				// Profile count should remain the same on error
				if len(config.Profiles) != originalCount {
					t.Errorf("Profile count changed on error: expected %d, got %d", originalCount, len(config.Profiles))
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Profile should be deleted
			if _, exists := config.Profiles[tt.profileName]; exists {
				t.Errorf("Profile %q was not deleted", tt.profileName)
			}

			// Profile count should decrease
			if len(config.Profiles) != originalCount-1 {
				t.Errorf("Expected profile count %d, got %d", originalCount-1, len(config.Profiles))
			}
		})
	}
}

func TestConfig_ListProfiles(t *testing.T) {
	config := &Config{
		CurrentProfile: "current",
		Profiles: map[string]Profile{
			"alpha": {BaseURL: "https://alpha.api.url"},
			"beta":  {BaseURL: "https://beta.api.url"},
			"gamma": {BaseURL: "https://gamma.api.url"},
		},
	}

	profiles := config.ListProfiles()

	if len(profiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(profiles))
	}

	// Check that all expected profiles are present
	expectedProfiles := map[string]bool{
		"alpha": false,
		"beta":  false,
		"gamma": false,
	}

	for _, profile := range profiles {
		if _, exists := expectedProfiles[profile]; exists {
			expectedProfiles[profile] = true
		} else {
			t.Errorf("Unexpected profile %q in list", profile)
		}
	}

	for profile, found := range expectedProfiles {
		if !found {
			t.Errorf("Expected profile %q not found in list", profile)
		}
	}
}

func TestConfig_ListProfiles_Empty(t *testing.T) {
	config := &Config{
		CurrentProfile: "",
		Profiles:       nil,
	}

	profiles := config.ListProfiles()

	if len(profiles) != 0 {
		t.Errorf("Expected 0 profiles, got %d", len(profiles))
	}
}

func Test_getConfigDir(t *testing.T) {
	tests := []struct {
		name       string
		setupEnv   func()
		cleanupEnv func()
		checkPath  func(string) bool
	}{
		{
			name: "uses XDG_CONFIG_HOME when set",
			setupEnv: func() {
				os.Setenv("XDG_CONFIG_HOME", "/custom/config")
			},
			cleanupEnv: func() {
				os.Unsetenv("XDG_CONFIG_HOME")
			},
			checkPath: func(path string) bool {
				return path == "/custom/config/forwardemail"
			},
		},
		{
			name: "uses home directory when XDG_CONFIG_HOME not set",
			setupEnv: func() {
				os.Unsetenv("XDG_CONFIG_HOME")
			},
			cleanupEnv: func() {
				// Nothing to cleanup
			},
			checkPath: func(path string) bool {
				// Should end with /.config/forwardemail
				return containsString(path, "/.config/forwardemail")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			configDir, err := getConfigDir()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !tt.checkPath(configDir) {
				t.Errorf("Config directory path %q does not match expected pattern", configDir)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(haystack[0:len(needle)] == needle ||
			(len(haystack) > len(needle) && containsString(haystack[1:], needle)))
}
