package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

    "github.com/ginsys/forward-email/pkg/auth"
)

func TestNewClient(t *testing.T) {
	authProvider := auth.MockProvider("test-key")

	tests := []struct {
		name       string
		baseURL    string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "valid URL",
			baseURL: "https://api.forwardemail.net",
			wantErr: false,
		},
		{
			name:       "invalid URL",
			baseURL:    "://invalid-url",
			wantErr:    true,
			wantErrMsg: "invalid base URL",
		},
		{
			name:    "HTTP URL",
			baseURL: "http://localhost:8080",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL, authProvider)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("NewClient() error = %v, want error containing %v", err, tt.wantErrMsg)
				}
				return
			}

			// Verify client is properly initialized
			if client == nil {
				t.Error("NewClient() returned nil client")
				return
			}

			if client.Auth == nil {
				t.Error("NewClient() client.Auth is nil")
			}

			if client.HTTPClient == nil {
				t.Error("NewClient() client.HTTPClient is nil")
			}

			if client.UserAgent != "forward-email/dev" {
				t.Errorf("NewClient() client.UserAgent = %v, want %v", client.UserAgent, "forward-email/dev")
			}

			// Verify services are initialized
			if client.Account == nil {
				t.Error("NewClient() client.Account is nil")
			}
			if client.Domains == nil {
				t.Error("NewClient() client.Domains is nil")
			}
			if client.Aliases == nil {
				t.Error("NewClient() client.Aliases is nil")
			}
			if client.Emails == nil {
				t.Error("NewClient() client.Emails is nil")
			}
			if client.Logs == nil {
				t.Error("NewClient() client.Logs is nil")
			}
			if client.Crypto == nil {
				t.Error("NewClient() client.Crypto is nil")
			}
		})
	}
}

func TestClient_ValidateAuth(t *testing.T) {
	tests := []struct {
		name       string
		auth       auth.Provider
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "valid auth provider",
			auth:    auth.MockProvider("valid-key"),
			wantErr: false,
		},
		{
			name:       "nil auth provider",
			auth:       nil,
			wantErr:    true,
			wantErrMsg: "no authentication provider configured",
		},
		{
			name:       "invalid auth provider",
			auth:       auth.MockProvider(""), // empty key
			wantErr:    true,
			wantErrMsg: "mock API key is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HTTPClient: &http.Client{Timeout: 10 * time.Second},
				Auth:       tt.auth,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := client.ValidateAuth(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrMsg != "" {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("ValidateAuth() error = %v, want error containing %v", err, tt.wantErrMsg)
				}
			}
		})
	}
}

func TestClient_Do(t *testing.T) {
	tests := []struct {
		name           string
		serverStatus   int
		serverResponse string
		authKey        string
		wantErr        bool
		wantErrMsg     string
	}{
		{
			name:           "successful request",
			serverStatus:   http.StatusOK,
			serverResponse: `{"success": true}`,
			authKey:        "valid-key",
			wantErr:        false,
		},
		{
			name:         "unauthorized",
			serverStatus: http.StatusUnauthorized,
			authKey:      "invalid-key",
			wantErr:      true,
		},
		{
			name:         "server error",
			serverStatus: http.StatusInternalServerError,
			authKey:      "valid-key",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify auth header is set
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" && tt.authKey != "" {
					t.Error("Expected Authorization header to be set")
				}

				// Verify other headers
				if userAgent := r.Header.Get("User-Agent"); userAgent != "forward-email/dev" {
					t.Errorf("Expected User-Agent header to be 'forward-email/dev', got %v", userAgent)
				}

				if accept := r.Header.Get("Accept"); accept != "application/json" {
					t.Errorf("Expected Accept header to be 'application/json', got %v", accept)
				}

				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != "" {
					w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			// Create client
			authProvider := auth.MockProvider(tt.authKey)
			client, err := NewClient(server.URL, authProvider)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			// Create request
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/test", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			// Make request
			var response map[string]interface{}
			err = client.Do(ctx, req, &response)

			if (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("Do() error = %v, want error containing %v", err, tt.wantErrMsg)
				}
			}

			if !tt.wantErr {
				// Verify successful response was decoded
				if tt.serverResponse != "" && response == nil {
					t.Error("Do() should have decoded response")
				}
			}
		})
	}
}

func TestClient_WithOptions(t *testing.T) {
	authProvider := auth.MockProvider("test-key")
	customClient := &http.Client{Timeout: 60 * time.Second}
	customUserAgent := "custom-cli/1.0.0"

	client, err := NewClient("https://api.forwardemail.net", authProvider,
		WithHTTPClient(customClient),
		WithUserAgent(customUserAgent),
	)
	if err != nil {
		t.Fatalf("NewClient() with options error = %v", err)
	}

	if client.HTTPClient != customClient {
		t.Error("WithHTTPClient() option did not set custom HTTP client")
	}

	if client.UserAgent != customUserAgent {
		t.Errorf("WithUserAgent() option did not set custom user agent, got %v", client.UserAgent)
	}
}
