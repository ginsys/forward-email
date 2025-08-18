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

// ListDomains retrieves a list of domains
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

// domainGetHelper is a generic helper for GET requests to domain endpoints
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

// GetDomain retrieves a specific domain by ID or name
func (s *DomainService) GetDomain(ctx context.Context, domainIDOrName string) (*Domain, error) {
	return domainGetHelper[Domain](ctx, s, "/v1/domains/%s", domainIDOrName, "failed to get domain")
}

// CreateDomain creates a new domain
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

// UpdateDomain updates an existing domain
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

// DeleteDomain deletes a domain
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

// VerifyDomain verifies a domain's DNS configuration
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

// GetDomainDNSRecords retrieves the required DNS records for a domain
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

// GetDomainQuota retrieves quota information for a domain
func (s *DomainService) GetDomainQuota(ctx context.Context, domainIDOrName string) (*DomainQuota, error) {
	return domainGetHelper[DomainQuota](ctx, s, "/v1/domains/%s/quota", domainIDOrName, "failed to get domain quota")
}

// GetDomainStats retrieves statistics for a domain
func (s *DomainService) GetDomainStats(ctx context.Context, domainIDOrName string) (*DomainStats, error) {
	return domainGetHelper[DomainStats](ctx, s, "/v1/domains/%s/stats", domainIDOrName, "failed to get domain stats")
}

// AddDomainMember adds a member to a domain
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

// RemoveDomainMember removes a member from a domain
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
