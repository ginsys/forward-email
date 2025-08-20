package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ListDomains retrieves a list of domains with optional filtering and pagination.
// The opts parameter can be nil to retrieve all domains with default settings.
// Supported filters include verification status, plan type, and search by name.
// Results can be sorted and paginated using the provided options.
func (s *DomainService) ListDomains(ctx context.Context, opts *ListDomainsOptions) (*ListDomainsResponse, error) {
	u := s.client.BaseURL.ResolveReference(&url.URL{Path: "/v1/domains"})

	// Add query parameters
	if opts != nil {
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
		if opts.Verified != nil {
			params.Set("verified", strconv.FormatBool(*opts.Verified))
		}
		if opts.Plan != "" {
			params.Set("plan", opts.Plan)
		}
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var domains []Domain
	if err := s.client.Do(ctx, req, &domains); err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	// Create response with basic pagination (since API returns array directly)
	response := &ListDomainsResponse{
		Domains: domains,
		Pagination: Pagination{
			Page:       1,
			Limit:      len(domains),
			Total:      len(domains),
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
	}

	return response, nil
}

// domainGetHelper is a generic helper function for GET requests to domain endpoints.
// It uses Go generics to handle different response types while providing consistent
// error handling and URL construction. The pathTemplate should contain a %s placeholder
// for the domain identifier, which will be properly URL-escaped.
func domainGetHelper[T any](ctx context.Context, s *DomainService, pathTemplate string, domainIDOrName, errorPrefix string) (*T, error) {
	u := s.client.BaseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf(pathTemplate, url.PathEscape(domainIDOrName)),
	})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var result T
	if err := s.client.Do(ctx, req, &result); err != nil {
		return nil, fmt.Errorf("%s: %w", errorPrefix, err)
	}

	return &result, nil
}

// GetDomain retrieves a specific domain by ID or name.
// The domainIDOrName parameter can be either the domain's UUID or its fully qualified domain name.
// Returns complete domain information including verification status, DNS records, and configuration.
func (s *DomainService) GetDomain(ctx context.Context, domainIDOrName string) (*Domain, error) {
	return domainGetHelper[Domain](ctx, s, "/v1/domains/%s", domainIDOrName, "failed to get domain")
}

// CreateDomain creates a new domain with the specified configuration.
// The request must contain at minimum the domain name. Optional settings include
// SMTP configuration, webhook URLs, and custom plan settings.
// Returns the created domain with initial verification status and required DNS records.
func (s *DomainService) CreateDomain(ctx context.Context, req *CreateDomainRequest) (*Domain, error) {
	if req == nil {
		return nil, fmt.Errorf("create domain request cannot be nil")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: "/v1/domains"})

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var domain Domain
	if err := s.client.Do(ctx, httpReq, &domain); err != nil {
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	return &domain, nil
}

// UpdateDomain updates an existing domain's configuration.
// The domainIDOrName parameter identifies the domain to update (UUID or FQDN).
// Only fields specified in the request will be updated; nil/empty fields are ignored.
// Returns the updated domain with the new configuration applied.
func (s *DomainService) UpdateDomain(ctx context.Context, domainIDOrName string, req *UpdateDomainRequest) (*Domain, error) {
	if req == nil {
		return nil, fmt.Errorf("update domain request cannot be nil")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/v1/domains/%s", url.PathEscape(domainIDOrName)),
	})

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var domain Domain
	if err := s.client.Do(ctx, httpReq, &domain); err != nil {
		return nil, fmt.Errorf("failed to update domain: %w", err)
	}

	return &domain, nil
}

// DeleteDomain permanently deletes a domain and all associated data.
// This operation cannot be undone and will remove all aliases, emails, and configuration
// associated with the domain. The domainIDOrName parameter identifies the domain (UUID or FQDN).
// Returns an error if the domain is not found or deletion fails.
func (s *DomainService) DeleteDomain(ctx context.Context, domainIDOrName string) error {
	u := s.client.BaseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/v1/domains/%s", url.PathEscape(domainIDOrName)),
	})

	req, err := http.NewRequestWithContext(ctx, "DELETE", u.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := s.client.Do(ctx, req, nil); err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	return nil
}

// VerifyDomain initiates DNS verification for a domain's configuration.
// This operation checks that required DNS records (MX, TXT, DMARC, SPF, DKIM) are properly
// configured and validates the domain for email sending and receiving.
// Returns verification results with status for each DNS record type.
func (s *DomainService) VerifyDomain(ctx context.Context, domainIDOrName string) (*DomainVerification, error) {
	u := s.client.BaseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/v1/domains/%s/verify", url.PathEscape(domainIDOrName)),
	})

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var verification DomainVerification
	if err := s.client.Do(ctx, req, &verification); err != nil {
		return nil, fmt.Errorf("failed to verify domain: %w", err)
	}

	return &verification, nil
}

// GetDomainDNSRecords retrieves the required DNS records for a domain.
// Returns a list of DNS records (MX, TXT, CNAME) that must be configured in the
// domain's DNS zone for proper email forwarding functionality. Each record includes
// the type, name, value, and TTL recommendations.
func (s *DomainService) GetDomainDNSRecords(ctx context.Context, domainIDOrName string) ([]DNSRecord, error) {
	u := s.client.BaseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/v1/domains/%s/dns", url.PathEscape(domainIDOrName)),
	})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var records []DNSRecord
	if err := s.client.Do(ctx, req, &records); err != nil {
		return nil, fmt.Errorf("failed to get DNS records: %w", err)
	}

	return records, nil
}

// GetDomainQuota retrieves quota information for a domain.
// Returns current usage statistics including storage used, email count, and bandwidth consumption
// along with the configured limits for the domain's subscription plan.
func (s *DomainService) GetDomainQuota(ctx context.Context, domainIDOrName string) (*DomainQuota, error) {
	return domainGetHelper[DomainQuota](ctx, s, "/v1/domains/%s/quota", domainIDOrName, "failed to get domain quota")
}

// GetDomainStats retrieves usage statistics for a domain.
// Returns metrics including total emails sent/received, bounce rates, spam scores,
// and historical usage data for monitoring and analytics purposes.
func (s *DomainService) GetDomainStats(ctx context.Context, domainIDOrName string) (*DomainStats, error) {
	return domainGetHelper[DomainStats](ctx, s, "/v1/domains/%s/stats", domainIDOrName, "failed to get domain stats")
}

// AddDomainMember adds a new member to a domain with specified permissions.
// The email parameter specifies the member's email address, and group determines their
// access level (e.g., "admin", "user"). The member will receive an invitation to
// access the domain's management interface.
func (s *DomainService) AddDomainMember(ctx context.Context, domainIDOrName, email, group string) (*DomainMember, error) {
	u := s.client.BaseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/v1/domains/%s/members", url.PathEscape(domainIDOrName)),
	})

	reqBody := map[string]string{
		"email": email,
		"group": group,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var member DomainMember
	if err := s.client.Do(ctx, req, &member); err != nil {
		return nil, fmt.Errorf("failed to add domain member: %w", err)
	}

	return &member, nil
}

// RemoveDomainMember removes a member from a domain's access list.
// The memberID parameter identifies the member to remove (UUID from domain member list).
// This operation immediately revokes the member's access to the domain management interface
// and any associated permissions.
func (s *DomainService) RemoveDomainMember(ctx context.Context, domainIDOrName, memberID string) error {
	u := s.client.BaseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/v1/domains/%s/members/%s", url.PathEscape(domainIDOrName), url.PathEscape(memberID)),
	})

	req, err := http.NewRequestWithContext(ctx, "DELETE", u.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := s.client.Do(ctx, req, nil); err != nil {
		return fmt.Errorf("failed to remove domain member: %w", err)
	}

	return nil
}
