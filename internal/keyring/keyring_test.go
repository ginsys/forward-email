package keyring

import (
	"fmt"
	"testing"

	"github.com/99designs/keyring"
)

const (
	testProfile = "test-profile"
	testAPIKey  = "test-api-key-12345"
)

func TestKeyring_SetGetAPIKey(t *testing.T) {
	kr, err := MockKeyring()
	if err != nil {
		t.Fatalf("failed to create mock keyring: %v", err)
	}

	profile := testProfile
	apiKey := testAPIKey

	// Test SetAPIKey
	err = kr.SetAPIKey(profile, apiKey)
	if err != nil {
		t.Errorf("SetAPIKey() error = %v", err)
	}

	// Test GetAPIKey
	retrievedKey, err := kr.GetAPIKey(profile)
	if err != nil {
		t.Errorf("GetAPIKey() error = %v", err)
	}

	if retrievedKey != apiKey {
		t.Errorf("GetAPIKey() = %v, want %v", retrievedKey, apiKey)
	}

	// Test HasAPIKey
	if !kr.HasAPIKey(profile) {
		t.Error("HasAPIKey() should return true for existing key")
	}

	// Test with non-existent profile
	if kr.HasAPIKey("non-existent") {
		t.Error("HasAPIKey() should return false for non-existent key")
	}
}

func TestKeyring_DeleteAPIKey(t *testing.T) {
	kr, err := MockKeyring()
	if err != nil {
		t.Fatalf("failed to create mock keyring: %v", err)
	}

	profile := testProfile
	apiKey := testAPIKey

	// Set initial key
	err = kr.SetAPIKey(profile, apiKey)
	if err != nil {
		t.Fatalf("SetAPIKey() setup error = %v", err)
	}

	// Verify key exists
	if !kr.HasAPIKey(profile) {
		t.Fatal("HasAPIKey() should return true before delete")
	}

	// Delete key
	err = kr.DeleteAPIKey(profile)
	if err != nil {
		t.Errorf("DeleteAPIKey() error = %v", err)
	}

	// Verify key is deleted
	if kr.HasAPIKey(profile) {
		t.Error("HasAPIKey() should return false after delete")
	}

	// Try to get deleted key
	_, err = kr.GetAPIKey(profile)
	if err == nil {
		t.Error("GetAPIKey() should return error for deleted key")
	}

	// Try to delete non-existent key
	err = kr.DeleteAPIKey("non-existent")
	if err == nil {
		t.Error("DeleteAPIKey() should return error for non-existent key")
	}
}

func TestKeyring_ListProfiles(t *testing.T) {
	kr, err := MockKeyring()
	if err != nil {
		t.Fatalf("failed to create mock keyring: %v", err)
	}

	// Initially should be empty
	profiles, err := kr.ListProfiles()
	if err != nil {
		t.Errorf("ListProfiles() error = %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("ListProfiles() initial length = %v, want 0", len(profiles))
	}

	// Add some profiles
	testProfiles := []string{"profile1", "profile2", "profile3"}
	for i, profile := range testProfiles {
		err = kr.SetAPIKey(profile, fmt.Sprintf("key%d", i+1))
		if err != nil {
			t.Fatalf("SetAPIKey() setup error = %v", err)
		}
	}

	// List profiles
	profiles, err = kr.ListProfiles()
	if err != nil {
		t.Errorf("ListProfiles() error = %v", err)
	}

	if len(profiles) != len(testProfiles) {
		t.Errorf("ListProfiles() length = %v, want %v", len(profiles), len(testProfiles))
	}

	// Check all profiles are present
	profileMap := make(map[string]bool)
	for _, profile := range profiles {
		profileMap[profile] = true
	}

	for _, expectedProfile := range testProfiles {
		if !profileMap[expectedProfile] {
			t.Errorf("ListProfiles() missing profile %v", expectedProfile)
		}
	}
}

func TestKeyring_MultipleProfiles(t *testing.T) {
	kr, err := MockKeyring()
	if err != nil {
		t.Fatalf("failed to create mock keyring: %v", err)
	}

	profiles := map[string]string{
		"development": "dev-api-key",
		"staging":     "staging-api-key",
		"production":  "prod-api-key",
	}

	// Set multiple profiles
	for profile, apiKey := range profiles {
		err = kr.SetAPIKey(profile, apiKey)
		if err != nil {
			t.Errorf("SetAPIKey(%v) error = %v", profile, err)
		}
	}

	// Verify all profiles
	for profile, expectedKey := range profiles {
		retrievedKey, getErr := kr.GetAPIKey(profile)
		if getErr != nil {
			t.Errorf("GetAPIKey(%v) error = %v", profile, getErr)
			continue
		}

		if retrievedKey != expectedKey {
			t.Errorf("GetAPIKey(%v) = %v, want %v", profile, retrievedKey, expectedKey)
		}
	}

	// Update one profile
	newKey := "updated-staging-key"
	err = kr.SetAPIKey("staging", newKey)
	if err != nil {
		t.Errorf("SetAPIKey(staging, updated) error = %v", err)
	}

	// Verify update
	retrievedKey, err := kr.GetAPIKey("staging")
	if err != nil {
		t.Errorf("GetAPIKey(staging) after update error = %v", err)
	}
	if retrievedKey != newKey {
		t.Errorf("GetAPIKey(staging) after update = %v, want %v", retrievedKey, newKey)
	}

	// Verify other profiles unchanged
	for profile, expectedKey := range profiles {
		if profile == "staging" {
			continue
		}
		retrievedKey, err := kr.GetAPIKey(profile)
		if err != nil {
			t.Errorf("GetAPIKey(%v) after staging update error = %v", profile, err)
			continue
		}
		if retrievedKey != expectedKey {
			t.Errorf("GetAPIKey(%v) after staging update = %v, want %v", profile, retrievedKey, expectedKey)
		}
	}
}

func TestNew_WithConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "default config",
			config: Config{
				ServiceName: "test-service",
			},
			wantErr: false,
		},
		{
			name: "empty service name uses default",
			config: Config{
				ServiceName: "",
			},
			wantErr: false,
		},
		{
			name: "with allowed backends",
			config: Config{
				ServiceName:     "test-service",
				AllowedBackends: []keyring.BackendType{keyring.FileBackend},
				FilePasswordFunc: func(string) (string, error) {
					return testPassConst, nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For testing, always use file backend
			tt.config.AllowedBackends = []keyring.BackendType{keyring.FileBackend}
			if tt.config.FilePasswordFunc == nil {
				tt.config.FilePasswordFunc = func(string) (string, error) {
					return testPassConst, nil
				}
			}

			kr, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && kr == nil {
				t.Error("New() returned nil keyring without error")
			}
		})
	}
}

func TestNew_DisableKeyringViaEnv(t *testing.T) {
	t.Setenv("FORWARDEMAIL_KEYRING_BACKEND", "none")
	k, err := New(Config{})
	if err == nil {
		t.Fatalf("expected error when keyring disabled, got nil")
	}
	if k != nil {
		t.Fatalf("expected nil keyring when disabled via env")
	}
}
