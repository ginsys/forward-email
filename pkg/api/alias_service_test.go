package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ginsys/forwardemail-cli/pkg/auth"
)

func TestAliasService_ListAliases(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/domains/example.com/aliases" {
			t.Errorf("Expected path /v1/domains/example.com/aliases, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("limit") != "10" {
			t.Errorf("Expected limit=10, got %s", query.Get("limit"))
		}
		if query.Get("sort") != "name" {
			t.Errorf("Expected sort=name, got %s", query.Get("sort"))
		}
		if query.Get("order") != "asc" {
			t.Errorf("Expected order=asc, got %s", query.Get("order"))
		}
		if query.Get("search") != "test" {
			t.Errorf("Expected search=test, got %s", query.Get("search"))
		}
		if query.Get("enabled") != "true" {
			t.Errorf("Expected enabled=true, got %s", query.Get("enabled"))
		}
		if query.Get("labels") != "important,work" {
			t.Errorf("Expected labels=important,work, got %s", query.Get("labels"))
		}
		if query.Get("has_imap") != "false" {
			t.Errorf("Expected has_imap=false, got %s", query.Get("has_imap"))
		}

		// Return mock response - API returns array directly
		aliases := []Alias{
			{
				ID:          "alias1",
				DomainID:    "example.com",
				Name:        "test",
				IsEnabled:   true,
				Recipients:  []string{"user@example.com"},
				Labels:      []string{"important", "work"},
				Description: "Test alias",
				HasIMAP:     false,
				HasPGP:      false,
				HasPassword: false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          "alias2",
				DomainID:    "example.com",
				Name:        "sales",
				IsEnabled:   true,
				Recipients:  []string{"sales@company.com", "backup@company.com"},
				Labels:      []string{"business"},
				Description: "Sales inquiries",
				HasIMAP:     true,
				HasPGP:      true,
				HasPassword: true,
				CreatedAt:   time.Now().Add(-24 * time.Hour),
				UpdatedAt:   time.Now().Add(-1 * time.Hour),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(aliases)
	}))
	defer server.Close()

	// Create client
	client, err := createTestAliasClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test ListAliases
	ctx := context.Background()
	enabled := true
	hasIMAP := false
	opts := &ListAliasesOptions{
		Domain:  "example.com",
		Page:    2,
		Limit:   10,
		Sort:    "name",
		Order:   "asc",
		Search:  "test",
		Enabled: &enabled,
		Labels:  "important,work",
		HasIMAP: &hasIMAP,
	}

	result, err := client.Aliases.ListAliases(ctx, opts)
	if err != nil {
		t.Fatalf("ListAliases failed: %v", err)
	}

	if len(result.Aliases) != 2 {
		t.Errorf("Expected 2 aliases, got %d", len(result.Aliases))
	}

	// Check first alias
	alias := result.Aliases[0]
	if alias.Name != "test" {
		t.Errorf("Expected alias name 'test', got '%s'", alias.Name)
	}
	if !alias.IsEnabled {
		t.Error("Expected alias to be enabled")
	}
	if len(alias.Recipients) != 1 {
		t.Errorf("Expected 1 recipient, got %d", len(alias.Recipients))
	}
	if alias.Recipients[0] != "user@example.com" {
		t.Errorf("Expected recipient 'user@example.com', got '%s'", alias.Recipients[0])
	}

	// Check second alias
	alias2 := result.Aliases[1]
	if alias2.Name != "sales" {
		t.Errorf("Expected alias name 'sales', got '%s'", alias2.Name)
	}
	if len(alias2.Recipients) != 2 {
		t.Errorf("Expected 2 recipients, got %d", len(alias2.Recipients))
	}
	if !alias2.HasIMAP {
		t.Error("Expected alias to have IMAP enabled")
	}
	if !alias2.HasPGP {
		t.Error("Expected alias to have PGP enabled")
	}

	// Check pagination
	if result.Page != 2 {
		t.Errorf("Expected page 2, got %d", result.Page)
	}
	if result.Limit != 10 {
		t.Errorf("Expected limit 10, got %d", result.Limit)
	}
	if result.TotalCount != 2 {
		t.Errorf("Expected total count 2, got %d", result.TotalCount)
	}
}

func TestAliasService_ListAliases_RequiredDomain(t *testing.T) {
	client, err := createTestAliasClient("http://example.com")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with nil options
	_, err = client.Aliases.ListAliases(ctx, nil)
	if err == nil || err.Error() != "domain is required" {
		t.Errorf("Expected 'domain is required' error, got: %v", err)
	}

	// Test with empty domain
	opts := &ListAliasesOptions{}
	_, err = client.Aliases.ListAliases(ctx, opts)
	if err == nil || err.Error() != "domain is required" {
		t.Errorf("Expected 'domain is required' error, got: %v", err)
	}
}

func TestAliasService_GetAlias(t *testing.T) {
	aliasID := "test-alias-id"
	domain := "example.com"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domain + "/aliases/" + aliasID
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		alias := Alias{
			ID:          aliasID,
			DomainID:    domain,
			Name:        "support",
			IsEnabled:   true,
			Recipients:  []string{"support@company.com"},
			Labels:      []string{"customer-service"},
			Description: "Customer support",
			HasIMAP:     true,
			HasPGP:      false,
			HasPassword: true,
			Quota: &AliasQuota{
				StorageUsed:  1024 * 1024 * 50,  // 50MB
				StorageLimit: 1024 * 1024 * 500, // 500MB
				EmailsSent:   5,
				EmailsLimit:  100,
			},
			CreatedAt: time.Now().Add(-7 * 24 * time.Hour), // 1 week ago
			UpdatedAt: time.Now().Add(-1 * time.Hour),      // 1 hour ago
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alias)
	}))
	defer server.Close()

	client, err := createTestAliasClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	result, err := client.Aliases.GetAlias(ctx, domain, aliasID)
	if err != nil {
		t.Fatalf("GetAlias failed: %v", err)
	}

	if result.ID != aliasID {
		t.Errorf("Expected alias ID '%s', got '%s'", aliasID, result.ID)
	}
	if result.Name != "support" {
		t.Errorf("Expected alias name 'support', got '%s'", result.Name)
	}
	if result.DomainID != domain {
		t.Errorf("Expected domain ID '%s', got '%s'", domain, result.DomainID)
	}
	if !result.IsEnabled {
		t.Error("Expected alias to be enabled")
	}
	if !result.HasIMAP {
		t.Error("Expected alias to have IMAP enabled")
	}
	if result.HasPGP {
		t.Error("Expected alias to not have PGP enabled")
	}
	if !result.HasPassword {
		t.Error("Expected alias to have password")
	}

	// Check quota
	if result.Quota == nil {
		t.Fatal("Expected quota to be present")
	}
	if result.Quota.StorageUsed != 1024*1024*50 {
		t.Errorf("Expected storage used 52428800, got %d", result.Quota.StorageUsed)
	}
	if result.Quota.EmailsSent != 5 {
		t.Errorf("Expected emails sent 5, got %d", result.Quota.EmailsSent)
	}
}

func TestAliasService_GetAlias_RequiredParameters(t *testing.T) {
	client, err := createTestAliasClient("http://example.com")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with empty domain
	_, err = client.Aliases.GetAlias(ctx, "", "alias-id")
	if err == nil || err.Error() != "domain is required" {
		t.Errorf("Expected 'domain is required' error, got: %v", err)
	}

	// Test with empty alias ID
	_, err = client.Aliases.GetAlias(ctx, "example.com", "")
	if err == nil || err.Error() != "alias ID is required" {
		t.Errorf("Expected 'alias ID is required' error, got: %v", err)
	}
}

func TestAliasService_CreateAlias(t *testing.T) {
	domain := "example.com"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domain + "/aliases"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Parse request body
		var req CreateAliasRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "newuser" {
			t.Errorf("Expected alias name 'newuser', got '%s'", req.Name)
		}
		if len(req.Recipients) != 2 {
			t.Errorf("Expected 2 recipients, got %d", len(req.Recipients))
		}
		if req.Recipients[0] != "user@company.com" {
			t.Errorf("Expected first recipient 'user@company.com', got '%s'", req.Recipients[0])
		}
		if req.Recipients[1] != "backup@company.com" {
			t.Errorf("Expected second recipient 'backup@company.com', got '%s'", req.Recipients[1])
		}
		if len(req.Labels) != 1 || req.Labels[0] != "team" {
			t.Errorf("Expected labels ['team'], got %v", req.Labels)
		}
		if req.Description != "New team member" {
			t.Errorf("Expected description 'New team member', got '%s'", req.Description)
		}
		if !req.IsEnabled {
			t.Error("Expected alias to be enabled")
		}
		if req.HasIMAP {
			t.Error("Expected IMAP to be disabled")
		}
		if req.HasPGP {
			t.Error("Expected PGP to be disabled")
		}

		// Return created alias
		alias := Alias{
			ID:          "new-alias-id",
			DomainID:    domain,
			Name:        req.Name,
			IsEnabled:   req.IsEnabled,
			Recipients:  req.Recipients,
			Labels:      req.Labels,
			Description: req.Description,
			HasIMAP:     req.HasIMAP,
			HasPGP:      req.HasPGP,
			PublicKey:   req.PublicKey,
			HasPassword: false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(alias)
	}))
	defer server.Close()

	client, err := createTestAliasClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	req := &CreateAliasRequest{
		Name:        "newuser",
		Recipients:  []string{"user@company.com", "backup@company.com"},
		Labels:      []string{"team"},
		Description: "New team member",
		IsEnabled:   true,
		HasIMAP:     false,
		HasPGP:      false,
		PublicKey:   "",
	}

	result, err := client.Aliases.CreateAlias(ctx, domain, req)
	if err != nil {
		t.Fatalf("CreateAlias failed: %v", err)
	}

	if result.Name != "newuser" {
		t.Errorf("Expected alias name 'newuser', got '%s'", result.Name)
	}
	if result.DomainID != domain {
		t.Errorf("Expected domain ID '%s', got '%s'", domain, result.DomainID)
	}
	if len(result.Recipients) != 2 {
		t.Errorf("Expected 2 recipients, got %d", len(result.Recipients))
	}
}

func TestAliasService_CreateAlias_ValidationErrors(t *testing.T) {
	client, err := createTestAliasClient("http://example.com")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with empty domain
	req := &CreateAliasRequest{Name: "test", Recipients: []string{"user@example.com"}}
	_, err = client.Aliases.CreateAlias(ctx, "", req)
	if err == nil || err.Error() != "domain is required" {
		t.Errorf("Expected 'domain is required' error, got: %v", err)
	}

	// Test with nil request
	_, err = client.Aliases.CreateAlias(ctx, "example.com", nil)
	if err == nil || err.Error() != "create request is required" {
		t.Errorf("Expected 'create request is required' error, got: %v", err)
	}

	// Test with empty name
	req = &CreateAliasRequest{Recipients: []string{"user@example.com"}}
	_, err = client.Aliases.CreateAlias(ctx, "example.com", req)
	if err == nil || err.Error() != "alias name is required" {
		t.Errorf("Expected 'alias name is required' error, got: %v", err)
	}

	// Test with empty recipients
	req = &CreateAliasRequest{Name: "test"}
	_, err = client.Aliases.CreateAlias(ctx, "example.com", req)
	if err == nil || err.Error() != "at least one recipient is required" {
		t.Errorf("Expected 'at least one recipient is required' error, got: %v", err)
	}

	// Test with empty recipient
	req = &CreateAliasRequest{Name: "test", Recipients: []string{"user@example.com", " "}}
	_, err = client.Aliases.CreateAlias(ctx, "example.com", req)
	if err == nil || err.Error() != "recipient cannot be empty" {
		t.Errorf("Expected 'recipient cannot be empty' error, got: %v", err)
	}
}

func TestAliasService_UpdateAlias(t *testing.T) {
	domain := "example.com"
	aliasID := "update-alias-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domain + "/aliases/" + aliasID
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Parse request body
		var req UpdateAliasRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if len(req.Recipients) != 1 || req.Recipients[0] != "newemail@company.com" {
			t.Errorf("Expected recipients ['newemail@company.com'], got %v", req.Recipients)
		}
		if req.IsEnabled == nil || !*req.IsEnabled {
			t.Errorf("Expected IsEnabled to be true, got %v", req.IsEnabled)
		}

		// Return updated alias
		alias := Alias{
			ID:          aliasID,
			DomainID:    domain,
			Name:        "updated",
			IsEnabled:   *req.IsEnabled,
			Recipients:  req.Recipients,
			HasIMAP:     false,
			HasPGP:      false,
			HasPassword: false,
			UpdatedAt:   time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alias)
	}))
	defer server.Close()

	client, err := createTestAliasClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	enabled := true
	req := &UpdateAliasRequest{
		Recipients: []string{"newemail@company.com"},
		IsEnabled:  &enabled,
	}

	result, err := client.Aliases.UpdateAlias(ctx, domain, aliasID, req)
	if err != nil {
		t.Fatalf("UpdateAlias failed: %v", err)
	}

	if result.ID != aliasID {
		t.Errorf("Expected alias ID '%s', got '%s'", aliasID, result.ID)
	}
	if !result.IsEnabled {
		t.Error("Expected alias to be enabled")
	}
	if len(result.Recipients) != 1 || result.Recipients[0] != "newemail@company.com" {
		t.Errorf("Expected recipients ['newemail@company.com'], got %v", result.Recipients)
	}
}

func TestAliasService_UpdateAlias_ValidationErrors(t *testing.T) {
	client, err := createTestAliasClient("http://example.com")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	req := &UpdateAliasRequest{}

	// Test with empty domain
	_, err = client.Aliases.UpdateAlias(ctx, "", "alias-id", req)
	if err == nil || err.Error() != "domain is required" {
		t.Errorf("Expected 'domain is required' error, got: %v", err)
	}

	// Test with empty alias ID
	_, err = client.Aliases.UpdateAlias(ctx, "example.com", "", req)
	if err == nil || err.Error() != "alias ID is required" {
		t.Errorf("Expected 'alias ID is required' error, got: %v", err)
	}

	// Test with nil request
	_, err = client.Aliases.UpdateAlias(ctx, "example.com", "alias-id", nil)
	if err == nil || err.Error() != "update request is required" {
		t.Errorf("Expected 'update request is required' error, got: %v", err)
	}

	// Test with empty recipient in update
	req = &UpdateAliasRequest{Recipients: []string{"user@example.com", " "}}
	_, err = client.Aliases.UpdateAlias(ctx, "example.com", "alias-id", req)
	if err == nil || err.Error() != "recipient cannot be empty" {
		t.Errorf("Expected 'recipient cannot be empty' error, got: %v", err)
	}
}

func TestAliasService_DeleteAlias(t *testing.T) {
	domain := "example.com"
	aliasID := "delete-alias-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domain + "/aliases/" + aliasID
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := createTestAliasClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.Aliases.DeleteAlias(ctx, domain, aliasID)
	if err != nil {
		t.Fatalf("DeleteAlias failed: %v", err)
	}
}

func TestAliasService_DeleteAlias_ValidationErrors(t *testing.T) {
	client, err := createTestAliasClient("http://example.com")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with empty domain
	err = client.Aliases.DeleteAlias(ctx, "", "alias-id")
	if err == nil || err.Error() != "domain is required" {
		t.Errorf("Expected 'domain is required' error, got: %v", err)
	}

	// Test with empty alias ID
	err = client.Aliases.DeleteAlias(ctx, "example.com", "")
	if err == nil || err.Error() != "alias ID is required" {
		t.Errorf("Expected 'alias ID is required' error, got: %v", err)
	}
}

func TestAliasService_GeneratePassword(t *testing.T) {
	domain := "example.com"
	aliasID := "password-alias-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domain + "/aliases/" + aliasID + "/generate-password"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		response := GeneratePasswordResponse{
			Password: "secure-generated-password-123",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := createTestAliasClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	result, err := client.Aliases.GeneratePassword(ctx, domain, aliasID)
	if err != nil {
		t.Fatalf("GeneratePassword failed: %v", err)
	}

	if result.Password != "secure-generated-password-123" {
		t.Errorf("Expected password 'secure-generated-password-123', got '%s'", result.Password)
	}
}

func TestAliasService_EnableDisableAlias(t *testing.T) {
	domain := "example.com"
	aliasID := "toggle-alias-id"

	// Test enable
	enableServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		var req UpdateAliasRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.IsEnabled == nil || !*req.IsEnabled {
			t.Error("Expected IsEnabled to be true for enable operation")
		}

		alias := Alias{
			ID:        aliasID,
			DomainID:  domain,
			Name:      "toggle-test",
			IsEnabled: true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alias)
	}))
	defer enableServer.Close()

	enableClient, err := createTestAliasClient(enableServer.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	result, err := enableClient.Aliases.EnableAlias(ctx, domain, aliasID)
	if err != nil {
		t.Fatalf("EnableAlias failed: %v", err)
	}

	if !result.IsEnabled {
		t.Error("Expected alias to be enabled")
	}

	// Test disable
	disableServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req UpdateAliasRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.IsEnabled == nil || *req.IsEnabled {
			t.Error("Expected IsEnabled to be false for disable operation")
		}

		alias := Alias{
			ID:        aliasID,
			DomainID:  domain,
			Name:      "toggle-test",
			IsEnabled: false,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alias)
	}))
	defer disableServer.Close()

	disableClient, err := createTestAliasClient(disableServer.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	result, err = disableClient.Aliases.DisableAlias(ctx, domain, aliasID)
	if err != nil {
		t.Fatalf("DisableAlias failed: %v", err)
	}

	if result.IsEnabled {
		t.Error("Expected alias to be disabled")
	}
}

func TestAliasService_UpdateRecipients(t *testing.T) {
	domain := "example.com"
	aliasID := "recipients-alias-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req UpdateAliasRequest
		json.NewDecoder(r.Body).Decode(&req)

		expectedRecipients := []string{"new1@example.com", "new2@example.com"}
		if len(req.Recipients) != 2 || req.Recipients[0] != expectedRecipients[0] || req.Recipients[1] != expectedRecipients[1] {
			t.Errorf("Expected recipients %v, got %v", expectedRecipients, req.Recipients)
		}

		alias := Alias{
			ID:         aliasID,
			DomainID:   domain,
			Name:       "recipients-test",
			Recipients: req.Recipients,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alias)
	}))
	defer server.Close()

	client, err := createTestAliasClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	newRecipients := []string{"new1@example.com", "new2@example.com"}
	result, err := client.Aliases.UpdateRecipients(ctx, domain, aliasID, newRecipients)
	if err != nil {
		t.Fatalf("UpdateRecipients failed: %v", err)
	}

	if len(result.Recipients) != 2 {
		t.Errorf("Expected 2 recipients, got %d", len(result.Recipients))
	}
	for i, expected := range newRecipients {
		if result.Recipients[i] != expected {
			t.Errorf("Expected recipient %d to be '%s', got '%s'", i, expected, result.Recipients[i])
		}
	}
}

func TestAliasService_UpdateRecipients_ValidationError(t *testing.T) {
	client, err := createTestAliasClient("http://example.com")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with empty recipients
	_, err = client.Aliases.UpdateRecipients(ctx, "example.com", "alias-id", []string{})
	if err == nil || err.Error() != "at least one recipient is required" {
		t.Errorf("Expected 'at least one recipient is required' error, got: %v", err)
	}
}

func TestAliasService_GetAliasQuota(t *testing.T) {
	domain := "example.com"
	aliasID := "quota-alias-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domain + "/aliases/" + aliasID + "/quota"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		quota := AliasQuota{
			StorageUsed:  1024 * 1024 * 75,  // 75MB
			StorageLimit: 1024 * 1024 * 500, // 500MB
			EmailsSent:   12,
			EmailsLimit:  100,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(quota)
	}))
	defer server.Close()

	client, err := createTestAliasClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	result, err := client.Aliases.GetAliasQuota(ctx, domain, aliasID)
	if err != nil {
		t.Fatalf("GetAliasQuota failed: %v", err)
	}

	if result.StorageUsed != 1024*1024*75 {
		t.Errorf("Expected storage used 78643200, got %d", result.StorageUsed)
	}
	if result.StorageLimit != 1024*1024*500 {
		t.Errorf("Expected storage limit 524288000, got %d", result.StorageLimit)
	}
	if result.EmailsSent != 12 {
		t.Errorf("Expected emails sent 12, got %d", result.EmailsSent)
	}
	if result.EmailsLimit != 100 {
		t.Errorf("Expected emails limit 100, got %d", result.EmailsLimit)
	}
}

func TestAliasService_GetAliasStats(t *testing.T) {
	domain := "example.com"
	aliasID := "stats-alias-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domain + "/aliases/" + aliasID + "/stats"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		stats := AliasStats{
			EmailsReceived: 156,
			EmailsSent:     89,
			StorageUsed:    1024 * 1024 * 45, // 45MB
			LastActivity:   time.Now().Add(-2 * time.Hour),
			RecentSenders:  []string{"client1@company.com", "client2@company.com", "newsletter@service.com"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}))
	defer server.Close()

	client, err := createTestAliasClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	result, err := client.Aliases.GetAliasStats(ctx, domain, aliasID)
	if err != nil {
		t.Fatalf("GetAliasStats failed: %v", err)
	}

	if result.EmailsReceived != 156 {
		t.Errorf("Expected emails received 156, got %d", result.EmailsReceived)
	}
	if result.EmailsSent != 89 {
		t.Errorf("Expected emails sent 89, got %d", result.EmailsSent)
	}
	if result.StorageUsed != 1024*1024*45 {
		t.Errorf("Expected storage used 47185920, got %d", result.StorageUsed)
	}
	if len(result.RecentSenders) != 3 {
		t.Errorf("Expected 3 recent senders, got %d", len(result.RecentSenders))
	}
	if result.RecentSenders[0] != "client1@company.com" {
		t.Errorf("Expected first sender 'client1@company.com', got '%s'", result.RecentSenders[0])
	}
}

func TestAliasService_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Alias not found",
			"code":    "ALIAS_NOT_FOUND",
		})
	}))
	defer server.Close()

	client, err := createTestAliasClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	_, err = client.Aliases.GetAlias(ctx, "example.com", "nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent alias")
	}

	// Check that the error message contains expected information
	expectedMessage := "Alias not found"
	if !containsString(err.Error(), expectedMessage) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedMessage, err.Error())
	}
}

// Helper functions

func createTestAliasClient(serverURL string) (*Client, error) {
	mockAuth := auth.MockProvider("test-api-key")

	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	client := &Client{
		HTTPClient: &http.Client{},
		BaseURL:    u,
		Auth:       mockAuth,
		UserAgent:  "test-client",
	}

	client.Aliases = &AliasService{client: client}

	return client, nil
}
