package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ginsys/forward-email/pkg/auth"
)

func TestDomainService_ListDomains(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/domains" {
			t.Errorf("Expected path /v1/domains, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", query.Get("page"))
		}
		if query.Get("limit") != "10" {
			t.Errorf("Expected limit=10, got %s", query.Get("limit"))
		}

		// Return mock response - API returns array directly, not wrapped in response object
		domains := []Domain{
			{
				ID:         "domain1",
				Name:       "example.com",
				IsVerified: true,
				Plan:       "free",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(domains)
	}))
	defer server.Close()

	// Create client
	client, err := createTestClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test ListDomains
	ctx := context.Background()
	opts := &ListDomainsOptions{
		Page:  2,
		Limit: 10,
	}

	result, err := client.Domains.ListDomains(ctx, opts)
	if err != nil {
		t.Fatalf("ListDomains failed: %v", err)
	}

	if len(result.Domains) != 1 {
		t.Errorf("Expected 1 domain, got %d", len(result.Domains))
	}

	domain := result.Domains[0]
	if domain.Name != "example.com" {
		t.Errorf("Expected domain name 'example.com', got '%s'", domain.Name)
	}
	if !domain.IsVerified {
		t.Error("Expected domain to be verified")
	}
}

func TestDomainService_GetDomain(t *testing.T) {
	domainID := "test-domain-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domainID
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		domain := Domain{
			ID:         domainID,
			Name:       "test.com",
			IsVerified: false,
			Plan:       "enhanced_protection",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(domain)
	}))
	defer server.Close()

	client, err := createTestClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	result, err := client.Domains.GetDomain(ctx, domainID)
	if err != nil {
		t.Fatalf("GetDomain failed: %v", err)
	}

	if result.ID != domainID {
		t.Errorf("Expected domain ID '%s', got '%s'", domainID, result.ID)
	}
	if result.Name != "test.com" {
		t.Errorf("Expected domain name 'test.com', got '%s'", result.Name)
	}
	if result.IsVerified {
		t.Error("Expected domain to not be verified")
	}
}

func TestDomainService_CreateDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/domains" {
			t.Errorf("Expected path /v1/domains, got %s", r.URL.Path)
		}

		// Parse request body
		var req CreateDomainRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Name != "newdomain.com" {
			t.Errorf("Expected domain name 'newdomain.com', got '%s'", req.Name)
		}
		if req.Plan != "team" {
			t.Errorf("Expected plan 'team', got '%s'", req.Plan)
		}

		// Return created domain
		domain := Domain{
			ID:         "new-domain-id",
			Name:       req.Name,
			Plan:       req.Plan,
			IsVerified: false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(domain)
	}))
	defer server.Close()

	client, err := createTestClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	req := &CreateDomainRequest{
		Name: "newdomain.com",
		Plan: "team",
	}

	result, err := client.Domains.CreateDomain(ctx, req)
	if err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	if result.Name != "newdomain.com" {
		t.Errorf("Expected domain name 'newdomain.com', got '%s'", result.Name)
	}
	if result.Plan != "team" {
		t.Errorf("Expected plan 'team', got '%s'", result.Plan)
	}
}

func TestDomainService_UpdateDomain(t *testing.T) {
	domainID := "update-domain-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domainID
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Parse request body
		var req UpdateDomainRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.MaxForwardedAddresses == nil || *req.MaxForwardedAddresses != 100 {
			t.Errorf("Expected max forwarded addresses 100, got %v", req.MaxForwardedAddresses)
		}

		// Return updated domain
		domain := Domain{
			ID:                    domainID,
			Name:                  "updated.com",
			MaxForwardedAddresses: 100,
			IsVerified:            true,
			UpdatedAt:             time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(domain)
	}))
	defer server.Close()

	client, err := createTestClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	maxAddresses := 100
	req := &UpdateDomainRequest{
		MaxForwardedAddresses: &maxAddresses,
	}

	result, err := client.Domains.UpdateDomain(ctx, domainID, req)
	if err != nil {
		t.Fatalf("UpdateDomain failed: %v", err)
	}

	if result.ID != domainID {
		t.Errorf("Expected domain ID '%s', got '%s'", domainID, result.ID)
	}
	if result.MaxForwardedAddresses != 100 {
		t.Errorf("Expected max forwarded addresses 100, got %d", result.MaxForwardedAddresses)
	}
}

func TestDomainService_DeleteDomain(t *testing.T) {
	domainID := "delete-domain-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domainID
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := createTestClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.Domains.DeleteDomain(ctx, domainID)
	if err != nil {
		t.Fatalf("DeleteDomain failed: %v", err)
	}
}

func TestDomainService_VerifyDomain(t *testing.T) {
	domainID := "verify-domain-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domainID + "/verify-records"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// API returns plain text, not JSON
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Domain's DNS records have been verified."))
	}))
	defer server.Close()

	client, err := createTestClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	result, err := client.Domains.VerifyDomain(ctx, domainID)
	if err != nil {
		t.Fatalf("VerifyDomain failed: %v", err)
	}

	// API returns plain text, so we only get IsVerified=true
	if !result.IsVerified {
		t.Error("Expected domain to be verified")
	}
}

func TestDomainService_AddDomainMember(t *testing.T) {
	domainID := "member-domain-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domainID + "/members"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Parse request body
		var reqBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if reqBody["email"] != "newuser@example.com" {
			t.Errorf("Expected email 'newuser@example.com', got '%s'", reqBody["email"])
		}
		if reqBody["group"] != "user" {
			t.Errorf("Expected group 'user', got '%s'", reqBody["group"])
		}

		member := DomainMember{
			Group: reqBody["group"],
			User: User{
				ID:          "user-id",
				Email:       reqBody["email"],
				DisplayName: "New User",
			},
			JoinedAt: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(member)
	}))
	defer server.Close()

	client, err := createTestClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	result, err := client.Domains.AddDomainMember(ctx, domainID, "newuser@example.com", "user")
	if err != nil {
		t.Fatalf("AddDomainMember failed: %v", err)
	}

	if result.User.Email != "newuser@example.com" {
		t.Errorf("Expected user email 'newuser@example.com', got '%s'", result.User.Email)
	}
	if result.Group != "user" {
		t.Errorf("Expected group 'user', got '%s'", result.Group)
	}
}

func TestDomainService_RemoveDomainMember(t *testing.T) {
	domainID := "member-domain-id"
	memberID := "member-to-remove"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		expectedPath := "/v1/domains/" + domainID + "/members/" + memberID
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := createTestClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.Domains.RemoveDomainMember(ctx, domainID, memberID)
	if err != nil {
		t.Fatalf("RemoveDomainMember failed: %v", err)
	}
}

func TestDomainService_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Domain not found",
			"code": "DOMAIN_NOT_FOUND",
		})
	}))
	defer server.Close()

	client, err := createTestClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	_, err = client.Domains.GetDomain(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent domain")
	}

	// Check that the error message contains expected information
	expectedMessage := "Domain not found"
	if !containsString(err.Error(), expectedMessage) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedMessage, err.Error())
	}
}

// Helper functions

func createTestClient(serverURL string) (*Client, error) {
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

	client.Domains = &DomainService{client: client}

	return client, nil
}

func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		haystack[0:len(needle)] == needle ||
		len(haystack) > len(needle) &&
			containsString(haystack[1:], needle)
}
