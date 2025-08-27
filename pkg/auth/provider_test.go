package auth

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ginsys/forward-email/internal/keyring"
	"github.com/ginsys/forward-email/pkg/config"
)

func TestForwardEmailAuth_Apply(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		wantAuth string
		wantErr  bool
	}{
		{
			name:     "valid API key",
			apiKey:   "test-api-key",
			wantAuth: "Basic dGVzdC1hcGkta2V5Og==", // base64 encoded "test-api-key:"
			wantErr:  false,
		},
		{
			name:     "empty API key",
			apiKey:   "",
			wantAuth: "Basic Og==", // base64(":")
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := MockProvider(tt.apiKey)

			req, err := http.NewRequestWithContext(context.Background(), "GET", "https://example.com", http.NoBody)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			err = auth.Apply(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got := req.Header.Get("Authorization")
				if got != tt.wantAuth {
					t.Errorf("Apply() Authorization header = %v, want %v", got, tt.wantAuth)
				}
			}
		})
	}
}

func TestForwardEmailAuth_Validate(t *testing.T) {
	tests := []struct {
		name       string
		apiKey     string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "valid credentials",
			apiKey:  "valid-key",
			wantErr: false,
		},
		{
			name:       "empty API key",
			apiKey:     "",
			wantErr:    true,
			wantErrMsg: "mock API key is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := MockProvider(tt.apiKey)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := auth.Validate(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrMsg != "" {
				if err.Error() != tt.wantErrMsg {
					t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.wantErrMsg)
				}
			}
		})
	}
}

func TestForwardEmailAuth_GetAPIKey(t *testing.T) {
	// Setup test configuration
	cfg := &config.Config{
		CurrentProfile: "test",
		Profiles: map[string]config.Profile{
			"test": {
				BaseURL: "https://api.forwardemail.net",
				APIKey:  "config-api-key",
				Timeout: "30s",
				Output:  "table",
			},
		},
	}

	// Setup mock keyring
	kr, err := keyring.MockKeyring()
	if err != nil {
		t.Fatalf("failed to create mock keyring: %v", err)
	}

	tests := []struct {
		name      string
		envVar    string
		envValue  string
		storeInKR bool
		krKey     string
		profile   string
		wantKey   string
		wantErr   bool
	}{
		{
			name:     "env var takes priority",
			envVar:   "FORWARDEMAIL_API_KEY",
			envValue: "env-api-key",
			profile:  "test",
			wantKey:  "env-api-key",
			wantErr:  false,
		},
		{
			name:     "profile-specific env var",
			envVar:   "FORWARDEMAIL_TEST_API_KEY",
			envValue: "profile-env-key",
			profile:  "test",
			wantKey:  "profile-env-key",
			wantErr:  false,
		},
		{
			name:      "keyring fallback",
			storeInKR: true,
			krKey:     "keyring-api-key",
			profile:   "test",
			wantKey:   "keyring-api-key",
			wantErr:   false,
		},
		{
			name:    "config fallback",
			profile: "test",
			wantKey: "config-api-key",
			wantErr: false,
		},
		{
			name:    "no key found",
			profile: "nonexistent",
			wantKey: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.envVar != "" {
				t.Setenv(tt.envVar, tt.envValue)
			}

			// Setup keyring
			testKR := kr
			if tt.storeInKR {
				testKR.SetAPIKey(tt.profile, tt.krKey)
				defer testKR.DeleteAPIKey(tt.profile)
			}

			// Create auth provider
			auth, err := NewProvider(ProviderConfig{
				Profile: tt.profile,
				Config:  cfg,
				Keyring: testKR,
			})
			if err != nil {
				t.Fatalf("failed to create auth provider: %v", err)
			}

			gotKey, err := auth.GetAPIKey()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && gotKey != tt.wantKey {
				t.Errorf("GetAPIKey() = %v, want %v", gotKey, tt.wantKey)
			}
		})
	}
}

func TestForwardEmailAuth_SetAPIKey(t *testing.T) {
	cfg := &config.Config{
		CurrentProfile: "test",
		Profiles:       make(map[string]config.Profile),
	}

	kr, err := keyring.MockKeyring()
	if err != nil {
		t.Fatalf("failed to create mock keyring: %v", err)
	}

	auth, err := NewProvider(ProviderConfig{
		Profile: "test",
		Config:  cfg,
		Keyring: kr,
	})
	if err != nil {
		t.Fatalf("failed to create auth provider: %v", err)
	}

	extAuth, ok := auth.(ExtendedProvider)
	if !ok {
		t.Fatalf("auth provider does not implement ExtendedProvider")
	}

	// Test setting API key
	testKey := "test-api-key-123"
	err = extAuth.SetAPIKey(testKey)
	if err != nil {
		t.Errorf("SetAPIKey() error = %v", err)
	}

	// Verify key is stored
	storedKey, err := extAuth.GetAPIKey()
	if err != nil {
		t.Errorf("GetAPIKey() after SetAPIKey() error = %v", err)
	}
	if storedKey != testKey {
		t.Errorf("GetAPIKey() after SetAPIKey() = %v, want %v", storedKey, testKey)
	}

	// Verify HasAPIKey returns true
	if !extAuth.HasAPIKey() {
		t.Error("HasAPIKey() should return true after SetAPIKey()")
	}
}

func TestForwardEmailAuth_DeleteAPIKey(t *testing.T) {
	cfg := &config.Config{
		CurrentProfile: "test",
		Profiles: map[string]config.Profile{
			"test": {
				BaseURL: "https://api.forwardemail.net",
				APIKey:  "initial-key",
				Timeout: "30s",
				Output:  "table",
			},
		},
	}

	kr, err := keyring.MockKeyring()
	if err != nil {
		t.Fatalf("failed to create mock keyring: %v", err)
	}

	auth, err := NewProvider(ProviderConfig{
		Profile: "test",
		Config:  cfg,
		Keyring: kr,
	})
	if err != nil {
		t.Fatalf("failed to create auth provider: %v", err)
	}

	extAuth, ok := auth.(ExtendedProvider)
	if !ok {
		t.Fatalf("auth provider does not implement ExtendedProvider")
	}

	// Set initial key
	err = extAuth.SetAPIKey("test-key")
	if err != nil {
		t.Fatalf("SetAPIKey() setup error = %v", err)
	}

	// Verify key exists
	if !extAuth.HasAPIKey() {
		t.Fatal("HasAPIKey() should return true before DeleteAPIKey()")
	}

	// Delete key
	err = extAuth.DeleteAPIKey()
	if err != nil {
		t.Errorf("DeleteAPIKey() error = %v", err)
	}

	// Verify key is deleted
	if extAuth.HasAPIKey() {
		t.Error("HasAPIKey() should return false after DeleteAPIKey()")
	}

	// Verify GetAPIKey returns error
	_, err = extAuth.GetAPIKey()
	if err == nil {
		t.Error("GetAPIKey() should return error after DeleteAPIKey()")
	}
}

func TestMockProvider(t *testing.T) {
	testKey := "mock-test-key"
	mock := MockProvider(testKey)

	// Test GetAPIKey
	key, err := mock.GetAPIKey()
	if err != nil {
		t.Errorf("MockProvider.GetAPIKey() error = %v", err)
	}
	if key != testKey {
		t.Errorf("MockProvider.GetAPIKey() = %v, want %v", key, testKey)
	}

	// Test Apply
	req, err := http.NewRequestWithContext(context.Background(), "GET", "https://example.com", http.NoBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	err = mock.Apply(req)
	if err != nil {
		t.Errorf("MockProvider.Apply() error = %v", err)
	}

	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		t.Error("MockProvider.Apply() should set Authorization header")
	}

	// Test Validate
	ctx := context.Background()
	err = mock.Validate(ctx)
	if err != nil {
		t.Errorf("MockProvider.Validate() error = %v", err)
	}

	// Test empty key mock
	emptyMock := MockProvider("")
	_, err = emptyMock.GetAPIKey()
	if err == nil {
		t.Error("MockProvider with empty key should return error")
	}

	err = emptyMock.Validate(ctx)
	if err == nil {
		t.Error("MockProvider.Validate() with empty key should return error")
	}
}
