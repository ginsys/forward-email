package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ginsys/forward-email/internal/client"
	"github.com/ginsys/forward-email/pkg/api"
	"github.com/ginsys/forward-email/pkg/auth"
)

func TestAliasListCommand(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/domains/example.com/aliases" {
			t.Errorf("Expected path /v1/domains/example.com/aliases, got %s", r.URL.Path)
		}

		aliases := []api.Alias{
			{
				ID:          "alias1",
				DomainID:    "example.com",
				Name:        "sales",
				IsEnabled:   true,
				Recipients:  []string{"sales@company.com"},
				Labels:      []string{"business"},
				Description: "Sales inquiries",
				HasIMAP:     false,
				HasPGP:      false,
				HasPassword: false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "alias2",
				DomainID:    "example.com",
				Name:        "support",
				IsEnabled:   true,
				Recipients:  []string{"support1@company.com", "support2@company.com"},
				Labels:      []string{"customer-service", "urgent"},
				Description: "Customer support",
				HasIMAP:     true,
				HasPGP:      false,
				HasPassword: true,
				CreatedAt:   time.Now().Add(-24 * time.Hour),
				UpdatedAt:   time.Now(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(aliases)
	}))
	defer server.Close()

	// Setup test environment
	setupTestEnv(server.URL)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectOut   []string
	}{
		{
			name:        "list with positional domain",
			args:        []string{"alias", "list", "example.com"},
			expectError: false,
			expectOut:   []string{"sales", "support", "business", "customer-service"},
		},
		{
			name:        "list with domain flag",
			args:        []string{"alias", "list", "--domain", "example.com"},
			expectError: false,
			expectOut:   []string{"sales", "support"},
		},
		{
			name:        "list with short domain flag",
			args:        []string{"alias", "list", "-d", "example.com"},
			expectError: false,
			expectOut:   []string{"sales", "support"},
		},
		{
			name:        "list without domain",
			args:        []string{"alias", "list"},
			expectError: true,
			expectOut:   []string{"domain is required"},
		},
		{
			name:        "list with pagination",
			args:        []string{"alias", "list", "example.com", "--page", "2", "--limit", "5"},
			expectError: false,
			expectOut:   []string{"sales", "support"},
		},
		{
			name:        "list with filters",
			args:        []string{"alias", "list", "example.com", "--enabled", "true", "--search", "sales"},
			expectError: false,
			expectOut:   []string{"sales"},
		},
		{
			name:        "list with JSON output",
			args:        []string{"alias", "list", "example.com", "--output", "json"},
			expectError: false,
			expectOut:   []string{`"name": "sales"`, `"name": "support"`},
		},
		{
			name:        "list with short output flag",
			args:        []string{"alias", "list", "example.com", "-o", "json"},
			expectError: false,
			expectOut:   []string{`"name": "sales"`, `"name": "support"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			resetAliasFlags()

			// Create a new command for each test
			cmd := createTestRootCmd()

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			outputStr := output.String()
			for _, expected := range tt.expectOut {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, outputStr)
				}
			}
		})
	}
}

func TestAliasGetCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for both domain + alias ID in path
		if strings.Contains(r.URL.Path, "/v1/domains/example.com/aliases/alias123") {
			alias := api.Alias{
				ID:          "alias123",
				DomainID:    "example.com",
				Name:        "testuser",
				IsEnabled:   true,
				Recipients:  []string{"user@company.com"},
				Labels:      []string{"team"},
				Description: "Test user alias",
				HasIMAP:     true,
				HasPGP:      false,
				HasPassword: true,
				CreatedAt:   time.Now().Add(-7 * 24 * time.Hour),
				UpdatedAt:   time.Now(),
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(alias)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	setupTestEnv(server.URL)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectOut   []string
	}{
		{
			name:        "get with positional domain and alias",
			args:        []string{"alias", "get", "example.com", "alias123"},
			expectError: false,
			expectOut:   []string{"testuser", "user@company.com", "team"},
		},
		{
			name:        "get with domain flag",
			args:        []string{"alias", "get", "alias123", "--domain", "example.com"},
			expectError: false,
			expectOut:   []string{"testuser", "user@company.com"},
		},
		{
			name:        "get with short domain flag",
			args:        []string{"alias", "get", "alias123", "-d", "example.com"},
			expectError: false,
			expectOut:   []string{"testuser", "user@company.com"},
		},
		{
			name:        "get without domain",
			args:        []string{"alias", "get", "alias123"},
			expectError: true,
			expectOut:   []string{"domain is required"},
		},
		{
			name:        "get without alias ID",
			args:        []string{"alias", "get"},
			expectError: true,
			expectOut:   []string{"alias ID is required"},
		},
		{
			name:        "get with JSON output",
			args:        []string{"alias", "get", "example.com", "alias123", "--output", "json"},
			expectError: false,
			expectOut:   []string{`"id": "alias123"`, `"name": "testuser"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAliasFlags()
			cmd := createTestRootCmd()

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			outputStr := output.String()
			for _, expected := range tt.expectOut {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, outputStr)
				}
			}
		})
	}
}

func TestAliasCreateCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		var req api.CreateAliasRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Return created alias
		alias := api.Alias{
			ID:          "new-alias-id",
			DomainID:    "example.com",
			Name:        req.Name,
			IsEnabled:   req.IsEnabled,
			Recipients:  req.Recipients,
			Labels:      req.Labels,
			Description: req.Description,
			HasIMAP:     req.HasIMAP,
			HasPGP:      req.HasPGP,
			HasPassword: false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(alias)
	}))
	defer server.Close()

	setupTestEnv(server.URL)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectOut   []string
	}{
		{
			name:        "create with positional domain",
			args:        []string{"alias", "create", "example.com", "newuser", "--recipients", "user@company.com"},
			expectError: false,
			expectOut:   []string{"created successfully", "newuser"},
		},
		{
			name:        "create with domain flag",
			args:        []string{"alias", "create", "newuser", "--domain", "example.com", "--recipients", "user@company.com"},
			expectError: false,
			expectOut:   []string{"created successfully", "newuser"},
		},
		{
			name:        "create with multiple recipients",
			args:        []string{"alias", "create", "example.com", "sales", "--recipients", "sales1@company.com,sales2@company.com"},
			expectError: false,
			expectOut:   []string{"created successfully", "sales"},
		},
		{
			name: "create with labels and description",
			args: []string{
				"alias", "create", "example.com", "support",
				"--recipients", "support@company.com",
				"--labels", "urgent,customer",
				"--description", "Customer support",
			},
			expectError: false,
			expectOut:   []string{"created successfully", "support"},
		},
		{
			name:        "create without domain",
			args:        []string{"alias", "create", "newuser", "--recipients", "user@company.com"},
			expectError: true,
			expectOut:   []string{"domain is required"},
		},
		{
			name:        "create without recipients",
			args:        []string{"alias", "create", "example.com", "newuser"},
			expectError: true,
			expectOut:   []string{"at least one recipient is required"},
		},
		{
			name:        "create without alias name",
			args:        []string{"alias", "create", "example.com"},
			expectError: true,
			expectOut:   []string{"alias name is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAliasFlags()
			cmd := createTestRootCmd()

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			outputStr := output.String()
			for _, expected := range tt.expectOut {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, outputStr)
				}
			}
		})
	}
}

func TestAliasUpdateCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		var req api.UpdateAliasRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Return updated alias
		alias := api.Alias{
			ID:          "update-alias-id",
			DomainID:    "example.com",
			Name:        "updated",
			IsEnabled:   true,
			Recipients:  []string{"updated@company.com"},
			HasIMAP:     false,
			HasPGP:      false,
			HasPassword: false,
			UpdatedAt:   time.Now(),
		}

		if req.Recipients != nil {
			alias.Recipients = req.Recipients
		}
		if req.IsEnabled != nil {
			alias.IsEnabled = *req.IsEnabled
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alias)
	}))
	defer server.Close()

	setupTestEnv(server.URL)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectOut   []string
	}{
		{
			name:        "update with positional domain",
			args:        []string{"alias", "update", "example.com", "update-alias-id", "--recipients", "new@company.com"},
			expectError: false,
			expectOut:   []string{"updated successfully"},
		},
		{
			name:        "update with domain flag",
			args:        []string{"alias", "update", "update-alias-id", "--domain", "example.com", "--enable"},
			expectError: false,
			expectOut:   []string{"updated successfully"},
		},
		{
			name:        "update without domain",
			args:        []string{"alias", "update", "update-alias-id", "--enable"},
			expectError: true,
			expectOut:   []string{"domain is required"},
		},
		{
			name:        "update without alias ID",
			args:        []string{"alias", "update", "example.com"},
			expectError: true,
			expectOut:   []string{"alias ID is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAliasFlags()
			cmd := createTestRootCmd()

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			outputStr := output.String()
			for _, expected := range tt.expectOut {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, outputStr)
				}
			}
		})
	}
}

func TestAliasDeleteCommand(t *testing.T) {
	// Mock server for get and delete operations
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/aliases/delete-alias-id") {
			// Return alias details for confirmation
			alias := api.Alias{
				ID:       "delete-alias-id",
				DomainID: "example.com",
				Name:     "todelete",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(alias)
		} else if r.Method == "DELETE" && strings.Contains(r.URL.Path, "/aliases/delete-alias-id") {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	setupTestEnv(server.URL)

	tests := []struct {
		name        string
		args        []string
		input       string // Simulated user input for confirmation
		expectError bool
		expectOut   []string
	}{
		{
			name:        "delete with positional domain (canceled)",
			args:        []string{"alias", "delete", "example.com", "delete-alias-id"},
			input:       "no\n",
			expectError: false,
			expectOut:   []string{"Deletion canceled"},
		},
		{
			name:        "delete with domain flag",
			args:        []string{"alias", "delete", "delete-alias-id", "--domain", "example.com"},
			input:       "no\n",
			expectError: false,
			expectOut:   []string{"Deletion canceled"},
		},
		{
			name:        "delete without domain",
			args:        []string{"alias", "delete", "delete-alias-id"},
			expectError: true,
			expectOut:   []string{"domain is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAliasFlags()
			cmd := createTestRootCmd()

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			// Set up stdin simulation if needed
			if tt.input != "" {
				cmd.SetIn(strings.NewReader(tt.input))
			}

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			outputStr := output.String()
			for _, expected := range tt.expectOut {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, outputStr)
				}
			}
		})
	}
}

func TestAliasEnableDisableCommands(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		var req api.UpdateAliasRequest
		json.NewDecoder(r.Body).Decode(&req)

		alias := api.Alias{
			ID:       "toggle-alias-id",
			DomainID: "example.com",
			Name:     "toggle",
		}

		if req.IsEnabled != nil {
			alias.IsEnabled = *req.IsEnabled
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alias)
	}))
	defer server.Close()

	setupTestEnv(server.URL)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectOut   []string
	}{
		{
			name:        "enable with positional domain",
			args:        []string{"alias", "enable", "example.com", "toggle-alias-id"},
			expectError: false,
			expectOut:   []string{"enabled successfully"},
		},
		{
			name:        "disable with domain flag",
			args:        []string{"alias", "disable", "toggle-alias-id", "--domain", "example.com"},
			expectError: false,
			expectOut:   []string{"disabled successfully"},
		},
		{
			name:        "enable without domain",
			args:        []string{"alias", "enable", "toggle-alias-id"},
			expectError: true,
			expectOut:   []string{"domain is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAliasFlags()
			cmd := createTestRootCmd()

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			outputStr := output.String()
			for _, expected := range tt.expectOut {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, outputStr)
				}
			}
		})
	}
}

func TestAliasRecipientsCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		var req api.UpdateAliasRequest
		json.NewDecoder(r.Body).Decode(&req)

		alias := api.Alias{
			ID:         "recipients-alias-id",
			DomainID:   "example.com",
			Name:       "recipients-test",
			Recipients: req.Recipients,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alias)
	}))
	defer server.Close()

	setupTestEnv(server.URL)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectOut   []string
	}{
		{
			name:        "update recipients with positional domain",
			args:        []string{"alias", "recipients", "example.com", "recipients-alias-id", "--recipients", "new1@company.com,new2@company.com"},
			expectError: false,
			expectOut:   []string{"Recipients updated"},
		},
		{
			name:        "update recipients with domain flag",
			args:        []string{"alias", "recipients", "recipients-alias-id", "--domain", "example.com", "--recipients", "updated@company.com"},
			expectError: false,
			expectOut:   []string{"Recipients updated"},
		},
		{
			name:        "update recipients without recipients flag",
			args:        []string{"alias", "recipients", "example.com", "recipients-alias-id"},
			expectError: true,
			expectOut:   []string{"at least one recipient is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAliasFlags()
			cmd := createTestRootCmd()

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			outputStr := output.String()
			for _, expected := range tt.expectOut {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, outputStr)
				}
			}
		})
	}
}

func TestAliasPasswordCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		response := api.GeneratePasswordResponse{
			Password: "generated-secure-password-123",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	setupTestEnv(server.URL)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectOut   []string
	}{
		{
			name:        "generate password with positional domain",
			args:        []string{"alias", "password", "example.com", "password-alias-id"},
			expectError: false,
			expectOut:   []string{"New IMAP password generated", "generated-secure-password-123"},
		},
		{
			name:        "generate password with domain flag",
			args:        []string{"alias", "password", "password-alias-id", "--domain", "example.com"},
			expectError: false,
			expectOut:   []string{"New IMAP password generated", "generated-secure-password-123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAliasFlags()
			cmd := createTestRootCmd()

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			outputStr := output.String()
			for _, expected := range tt.expectOut {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, outputStr)
				}
			}
		})
	}
}

func TestAliasQuotaCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		quota := api.AliasQuota{
			StorageUsed:  1024 * 1024 * 50,  // 50MB
			StorageLimit: 1024 * 1024 * 500, // 500MB
			EmailsSent:   15,
			EmailsLimit:  100,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(quota)
	}))
	defer server.Close()

	setupTestEnv(server.URL)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectOut   []string
	}{
		{
			name:        "show quota with positional domain",
			args:        []string{"alias", "quota", "example.com", "quota-alias-id"},
			expectError: false,
			expectOut:   []string{"Storage", "50.0 MB", "500.0 MB", "Emails", "15", "100"},
		},
		{
			name:        "show quota with JSON output",
			args:        []string{"alias", "quota", "example.com", "quota-alias-id", "--output", "json"},
			expectError: false,
			expectOut:   []string{`"storage_used": 52428800`, `"emails_sent": 15`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAliasFlags()
			cmd := createTestRootCmd()

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			outputStr := output.String()
			for _, expected := range tt.expectOut {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, outputStr)
				}
			}
		})
	}
}

func TestAliasStatsCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		stats := api.AliasStats{
			EmailsReceived: 125,
			EmailsSent:     67,
			StorageUsed:    1024 * 1024 * 35, // 35MB
			LastActivity:   time.Now().Add(-1 * time.Hour),
			RecentSenders:  []string{"client1@company.com", "client2@company.com"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}))
	defer server.Close()

	setupTestEnv(server.URL)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectOut   []string
	}{
		{
			name:        "show stats with positional domain",
			args:        []string{"alias", "stats", "example.com", "stats-alias-id"},
			expectError: false,
			expectOut:   []string{"Emails Received", "125", "Emails Sent", "67", "Storage Used", "35.0 MB"},
		},
		{
			name:        "show stats with JSON output",
			args:        []string{"alias", "stats", "example.com", "stats-alias-id", "--output", "json"},
			expectError: false,
			expectOut:   []string{`"emails_received": 125`, `"emails_sent": 67`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAliasFlags()
			cmd := createTestRootCmd()

			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			outputStr := output.String()
			for _, expected := range tt.expectOut {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, outputStr)
				}
			}
		})
	}
}

// Helper functions

func setupTestEnv(serverURL string) {
	// Set up mock client factory
	client.SetTestMode(serverURL, auth.MockProvider("test-api-key"))
}

func createTestRootCmd() *cobra.Command {
	// Reset viper for each test
	viper.Reset()

	// Create a new root command for testing
	cmd := &cobra.Command{
		Use: "forward-email",
	}

	// Create fresh command instances for testing
	testAliasCmd := &cobra.Command{
		Use:   "alias",
		Short: "Manage alias configurations",
	}

	testAliasListCmd := &cobra.Command{
		Use:   "list [domain]",
		Short: "List aliases",
		RunE:  runAliasList,
	}

	testAliasGetCmd := &cobra.Command{
		Use:   "get [domain] <alias-id>",
		Short: "Get alias details",
		RunE:  runAliasGet,
	}

	testAliasCreateCmd := &cobra.Command{
		Use:   "create [domain] <alias-name>",
		Short: "Create alias",
		RunE:  runAliasCreate,
	}

	testAliasUpdateCmd := &cobra.Command{
		Use:   "update [domain] <alias-id>",
		Short: "Update alias",
		RunE:  runAliasUpdate,
	}

	testAliasDeleteCmd := &cobra.Command{
		Use:   "delete [domain] <alias-id>",
		Short: "Delete alias",
		RunE:  runAliasDelete,
	}

	testAliasEnableCmd := &cobra.Command{
		Use:   "enable [domain] <alias-id>",
		Short: "Enable alias",
		RunE:  runAliasEnable,
	}

	testAliasDisableCmd := &cobra.Command{
		Use:   "disable [domain] <alias-id>",
		Short: "Disable alias",
		RunE:  runAliasDisable,
	}

	testAliasRecipientsCmd := &cobra.Command{
		Use:   "recipients [domain] <alias-id>",
		Short: "Update recipients",
		RunE:  runAliasRecipients,
	}

	testAliasPasswordCmd := &cobra.Command{
		Use:   "password [domain] <alias-id>",
		Short: "Generate password",
		RunE:  runAliasPassword,
	}

	testAliasQuotaCmd := &cobra.Command{
		Use:   "quota [domain] <alias-id>",
		Short: "Show quota",
		RunE:  runAliasQuota,
	}

	testAliasStatsCmd := &cobra.Command{
		Use:   "stats [domain] <alias-id>",
		Short: "Show stats",
		RunE:  runAliasStats,
	}

	// Add subcommands to alias command
	testAliasCmd.AddCommand(testAliasListCmd)
	testAliasCmd.AddCommand(testAliasGetCmd)
	testAliasCmd.AddCommand(testAliasCreateCmd)
	testAliasCmd.AddCommand(testAliasUpdateCmd)
	testAliasCmd.AddCommand(testAliasDeleteCmd)
	testAliasCmd.AddCommand(testAliasEnableCmd)
	testAliasCmd.AddCommand(testAliasDisableCmd)
	testAliasCmd.AddCommand(testAliasRecipientsCmd)
	testAliasCmd.AddCommand(testAliasPasswordCmd)
	testAliasCmd.AddCommand(testAliasQuotaCmd)
	testAliasCmd.AddCommand(testAliasStatsCmd)

	// Add alias command to root
	cmd.AddCommand(testAliasCmd)

	// Set up persistent flags
	cmd.PersistentFlags().StringP("profile", "p", "", "Configuration profile to use")
	cmd.PersistentFlags().StringP("output", "o", "table", "Output format (table|json|yaml|csv)")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Bind flags to viper
	viper.BindPFlag("profile", cmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("output", cmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose"))

	// Set up alias command flags
	testAliasCmd.PersistentFlags().StringVarP(&aliasDomain, "domain", "d", "", "Domain name or ID (required)")

	// List command flags
	testAliasListCmd.Flags().IntVar(&aliasPage, "page", 1, "Page number")
	testAliasListCmd.Flags().IntVar(&aliasLimit, "limit", 25, "Number of aliases per page")
	testAliasListCmd.Flags().StringVar(&aliasSort, "sort", "name", "Sort by (name, created, updated)")
	testAliasListCmd.Flags().StringVar(&aliasOrder, "order", "asc", "Sort order (asc, desc)")
	testAliasListCmd.Flags().StringVar(&aliasSearch, "search", "", "Search alias names")
	testAliasListCmd.Flags().StringVar(&aliasEnabled, "enabled", "", "Filter by enabled status (true/false)")
	testAliasListCmd.Flags().StringVar(&aliasLabels, "labels", "", "Filter by labels (comma-separated)")
	testAliasListCmd.Flags().StringVar(&aliasHasIMAP, "has-imap", "", "Filter by IMAP capability (true/false)")

	// Create command flags
	testAliasCreateCmd.Flags().StringSliceVar(&aliasRecipients, "recipients", nil, "Recipient email addresses")
	testAliasCreateCmd.Flags().StringSliceVar(&aliasLabelsFlag, "labels", nil, "Labels for the alias")
	testAliasCreateCmd.Flags().StringVar(&aliasDescription, "description", "", "Description for the alias")
	testAliasCreateCmd.Flags().BoolVar(&aliasEnableFlag, "enabled", true, "Enable the alias")
	testAliasCreateCmd.Flags().BoolVar(&aliasIMAPFlag, "imap", false, "Enable IMAP access")
	testAliasCreateCmd.Flags().BoolVar(&aliasPGPFlag, "pgp", false, "Enable PGP encryption")
	testAliasCreateCmd.Flags().StringVar(&aliasPublicKey, "public-key", "", "PGP public key")

	// Update command flags
	testAliasUpdateCmd.Flags().StringSliceVar(&aliasRecipients, "recipients", nil, "Update recipient email addresses")
	testAliasUpdateCmd.Flags().StringSliceVar(&aliasLabelsFlag, "labels", nil, "Update labels for the alias")
	testAliasUpdateCmd.Flags().StringVar(&aliasDescription, "description", "", "Update description for the alias")
	testAliasUpdateCmd.Flags().BoolVar(&aliasEnableFlag, "enable", false, "Enable the alias")
	testAliasUpdateCmd.Flags().BoolVar(&aliasDisableFlag, "disable", false, "Disable the alias")
	testAliasUpdateCmd.Flags().BoolVar(&aliasIMAPFlag, "imap", false, "Enable IMAP access")
	testAliasUpdateCmd.Flags().BoolVar(&aliasPGPFlag, "pgp", false, "Enable PGP encryption")
	testAliasUpdateCmd.Flags().StringVar(&aliasPublicKey, "public-key", "", "Update PGP public key")

	// Recipients command flags
	testAliasRecipientsCmd.Flags().StringSliceVar(&aliasRecipients, "recipients", nil, "New recipient email addresses")

	return cmd
}

func resetAliasFlags() {
	// Reset all alias command flags to their default values
	aliasPage = 1
	aliasLimit = 25
	aliasSort = "name"
	aliasOrder = "asc"
	aliasSearch = ""
	aliasEnabled = ""
	aliasLabels = ""
	aliasHasIMAP = ""
	aliasDomain = ""

	// Create/Update flags
	aliasRecipients = nil
	aliasLabelsFlag = nil
	aliasDescription = ""
	aliasEnableFlag = true
	aliasDisableFlag = false
	aliasIMAPFlag = false
	aliasPGPFlag = false
	aliasPublicKey = ""

	// Reset viper values
	viper.Set("output", "table")
	viper.Set("profile", "")
	viper.Set("verbose", false)
}
