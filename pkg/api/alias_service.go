package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ListAliases retrieves a list of aliases for a domain with optional filtering and pagination.
// The domain parameter is required. Supports filtering by enabled status, labels, IMAP configuration,
// and text search across alias names. Results can be sorted and paginated using the provided options.
func (s *AliasService) ListAliases(ctx context.Context, opts *ListAliasesOptions) (*ListAliasesResponse, error) {
	if opts == nil || opts.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/domains/%s/aliases", opts.Domain)})

	// Add query parameters
	params := url.Values{}
	if opts.Page > 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Sort != "" {
		params.Set("sort", opts.Sort)
	}
	if opts.Order != "" {
		params.Set("order", opts.Order)
	}
	if opts.Search != "" {
		params.Set("search", opts.Search)
	}
	if opts.Enabled != nil {
		params.Set("enabled", strconv.FormatBool(*opts.Enabled))
	}
	if opts.Labels != "" {
		params.Set("labels", opts.Labels)
	}
	if opts.HasIMAP != nil {
		params.Set("has_imap", strconv.FormatBool(*opts.HasIMAP))
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var aliases []Alias
	if err := s.client.Do(ctx, req, &aliases); err != nil {
		return nil, fmt.Errorf("failed to list aliases: %w", err)
	}

	// Calculate pagination info (since API returns array directly)
	totalCount := len(aliases)
	page := 1
	if opts.Page > 0 {
		page = opts.Page
	}
	limit := 25
	if opts.Limit > 0 {
		limit = opts.Limit
	}
	totalPages := (totalCount + limit - 1) / limit

	return &ListAliasesResponse{
		Aliases:    aliases,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetAlias retrieves a specific alias by ID within a domain.
// Returns complete alias information including recipients, IMAP configuration, labels,
// vacation responder settings, and PGP encryption status.
func (s *AliasService) GetAlias(ctx context.Context, domain, aliasID string) (*Alias, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	if aliasID == "" {
		return nil, fmt.Errorf("alias ID is required")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/domains/%s/aliases/%s", domain, aliasID)})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var alias Alias
	if err := s.client.Do(ctx, req, &alias); err != nil {
		return nil, fmt.Errorf("failed to get alias: %w", err)
	}

	return &alias, nil
}

// CreateAlias creates a new alias within a domain.
// Requires a name and at least one recipient. Optional settings include labels,
// description, IMAP configuration, PGP encryption, and vacation responder.
// Recipients can be email addresses, webhooks, or FQDN forwarding targets.
func (s *AliasService) CreateAlias(ctx context.Context, domain string, req *CreateAliasRequest) (*Alias, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	if req == nil {
		return nil, fmt.Errorf("create request is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("alias name is required")
	}
	if len(req.Recipients) == 0 {
		return nil, fmt.Errorf("at least one recipient is required")
	}

	// Validate recipients
	for _, recipient := range req.Recipients {
		if strings.TrimSpace(recipient) == "" {
			return nil, fmt.Errorf("recipient cannot be empty")
		}
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/domains/%s/aliases", domain)})

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var alias Alias
	if err := s.client.Do(ctx, httpReq, &alias); err != nil {
		return nil, fmt.Errorf("failed to create alias: %w", err)
	}

	return &alias, nil
}

// UpdateAlias updates an existing alias configuration.
// Only fields specified in the request will be updated; nil/empty fields are ignored.
// Can modify recipients, labels, IMAP settings, vacation responder, and enabled status.
// Returns the updated alias with the new configuration applied.
func (s *AliasService) UpdateAlias(ctx context.Context, domain, aliasID string, req *UpdateAliasRequest) (*Alias, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	if aliasID == "" {
		return nil, fmt.Errorf("alias ID is required")
	}
	if req == nil {
		return nil, fmt.Errorf("update request is required")
	}

	// Validate recipients if provided
	if req.Recipients != nil {
		for _, recipient := range req.Recipients {
			if strings.TrimSpace(recipient) == "" {
				return nil, fmt.Errorf("recipient cannot be empty")
			}
		}
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/domains/%s/aliases/%s", domain, aliasID)})

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var alias Alias
	if err := s.client.Do(ctx, httpReq, &alias); err != nil {
		return nil, fmt.Errorf("failed to update alias: %w", err)
	}

	return &alias, nil
}

// DeleteAlias permanently deletes an alias and all associated data.
// This operation cannot be undone and will remove all stored emails, configuration,
// and forwarding rules for the alias. IMAP access will be immediately revoked.
func (s *AliasService) DeleteAlias(ctx context.Context, domain, aliasID string) error {
	if domain == "" {
		return fmt.Errorf("domain is required")
	}
	if aliasID == "" {
		return fmt.Errorf("alias ID is required")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/domains/%s/aliases/%s", domain, aliasID)})

	req, err := http.NewRequestWithContext(ctx, "DELETE", u.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := s.client.Do(ctx, req, nil); err != nil {
		return fmt.Errorf("failed to delete alias: %w", err)
	}

	return nil
}

// GeneratePassword generates a new IMAP password for an alias.
// This invalidates any existing IMAP password and generates a new secure password
// for accessing the alias via IMAP clients. The alias must have IMAP enabled.
// Returns the new password which should be stored securely by the client.
func (s *AliasService) GeneratePassword(ctx context.Context, domain, aliasID string) (*GeneratePasswordResponse, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	if aliasID == "" {
		return nil, fmt.Errorf("alias ID is required")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/domains/%s/aliases/%s/generate-password", domain, aliasID)})

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var resp GeneratePasswordResponse
	if err := s.client.Do(ctx, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	return &resp, nil
}

// EnableAlias enables an alias for email forwarding (convenience method).
// This is a shortcut for UpdateAlias with IsEnabled set to true.
// Once enabled, the alias will start receiving and forwarding emails.
func (s *AliasService) EnableAlias(ctx context.Context, domain, aliasID string) (*Alias, error) {
	enabled := true
	req := &UpdateAliasRequest{
		IsEnabled: &enabled,
	}
	return s.UpdateAlias(ctx, domain, aliasID, req)
}

// DisableAlias disables an alias to stop email forwarding (convenience method).
// This is a shortcut for UpdateAlias with IsEnabled set to false.
// When disabled, incoming emails to the alias will be rejected.
func (s *AliasService) DisableAlias(ctx context.Context, domain, aliasID string) (*Alias, error) {
	enabled := false
	req := &UpdateAliasRequest{
		IsEnabled: &enabled,
	}
	return s.UpdateAlias(ctx, domain, aliasID, req)
}

// UpdateRecipients updates only the recipients of an alias (convenience method).
// This is a shortcut for UpdateAlias that modifies only the recipients list.
// Recipients can be email addresses, webhooks (starting with http/https),
// or FQDN forwarding targets. At least one recipient is required.
func (s *AliasService) UpdateRecipients(ctx context.Context, domain, aliasID string, recipients []string) (*Alias, error) {
	if len(recipients) == 0 {
		return nil, fmt.Errorf("at least one recipient is required")
	}

	req := &UpdateAliasRequest{
		Recipients: recipients,
	}
	return s.UpdateAlias(ctx, domain, aliasID, req)
}

// GetAliasQuota retrieves quota information for an alias.
// Returns current storage usage, email count, and bandwidth consumption
// along with the configured limits for the alias based on the domain's plan.
func (s *AliasService) GetAliasQuota(ctx context.Context, domain, aliasID string) (*AliasQuota, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	if aliasID == "" {
		return nil, fmt.Errorf("alias ID is required")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/domains/%s/aliases/%s/quota", domain, aliasID)})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var quota AliasQuota
	if err := s.client.Do(ctx, req, &quota); err != nil {
		return nil, fmt.Errorf("failed to get alias quota: %w", err)
	}

	return &quota, nil
}

// GetAliasStats retrieves usage statistics for an alias.
// Returns metrics including emails received/forwarded, bounce rates, spam scores,
// IMAP access statistics, and historical usage data for monitoring purposes.
func (s *AliasService) GetAliasStats(ctx context.Context, domain, aliasID string) (*AliasStats, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	if aliasID == "" {
		return nil, fmt.Errorf("alias ID is required")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/domains/%s/aliases/%s/stats", domain, aliasID)})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var stats AliasStats
	if err := s.client.Do(ctx, req, &stats); err != nil {
		return nil, fmt.Errorf("failed to get alias stats: %w", err)
	}

	return &stats, nil
}
