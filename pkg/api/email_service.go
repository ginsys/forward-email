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

// SendEmail sends a single email
func (s *EmailService) SendEmail(ctx context.Context, req *SendEmailRequest) (*SendEmailResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("send request is required")
	}
	if req.From == "" {
		return nil, fmt.Errorf("from address is required")
	}
	if len(req.To) == 0 {
		return nil, fmt.Errorf("at least one recipient is required")
	}
	if req.Subject == "" {
		return nil, fmt.Errorf("subject is required")
	}
	if req.Text == "" && req.HTML == "" {
		return nil, fmt.Errorf("either text or HTML content is required")
	}

	// Validate email addresses
	allRecipients := append([]string{}, req.To...)
	allRecipients = append(allRecipients, req.CC...)
	allRecipients = append(allRecipients, req.BCC...)
	for _, email := range allRecipients {
		if strings.TrimSpace(email) == "" {
			return nil, fmt.Errorf("recipient email cannot be empty")
		}
		if !strings.Contains(email, "@") {
			return nil, fmt.Errorf("invalid email address: %s", email)
		}
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: "/v1/emails"})

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var response SendEmailResponse
	if err := s.client.Do(ctx, httpReq, &response); err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	return &response, nil
}

// SendBulkEmails sends multiple emails in a batch
func (s *EmailService) SendBulkEmails(ctx context.Context, req *BulkEmailRequest) (*BulkEmailResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("bulk request is required")
	}
	if len(req.Emails) == 0 {
		return nil, fmt.Errorf("at least one email is required")
	}

	// Validate each email in the batch
	for i, email := range req.Emails {
		if email.From == "" {
			return nil, fmt.Errorf("email %d: from address is required", i+1)
		}
		if len(email.To) == 0 {
			return nil, fmt.Errorf("email %d: at least one recipient is required", i+1)
		}
		if email.Subject == "" {
			return nil, fmt.Errorf("email %d: subject is required", i+1)
		}
		if email.Text == "" && email.HTML == "" {
			return nil, fmt.Errorf("email %d: either text or HTML content is required", i+1)
		}
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: "/v1/emails/bulk"})

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var response BulkEmailResponse
	if err := s.client.Do(ctx, httpReq, &response); err != nil {
		return nil, fmt.Errorf("failed to send bulk emails: %w", err)
	}

	return &response, nil
}

// ListEmails retrieves a list of sent emails
func (s *EmailService) ListEmails(ctx context.Context, opts *ListEmailsOptions) (*ListEmailsResponse, error) {
	u := s.client.BaseURL.ResolveReference(&url.URL{Path: "/v1/emails"})

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
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.From != "" {
			params.Set("from", opts.From)
		}
		if opts.To != "" {
			params.Set("to", opts.To)
		}
		if opts.DateFrom != "" {
			params.Set("date_from", opts.DateFrom)
		}
		if opts.DateTo != "" {
			params.Set("date_to", opts.DateTo)
		}
		if opts.HasAttach != nil {
			params.Set("has_attach", strconv.FormatBool(*opts.HasAttach))
		}
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var emails []Email
	if err := s.client.Do(ctx, req, &emails); err != nil {
		return nil, fmt.Errorf("failed to list emails: %w", err)
	}

	// Calculate pagination info (since API returns array directly)
	totalCount := len(emails)
	page := 1
	if opts != nil && opts.Page > 0 {
		page = opts.Page
	}
	limit := 25
	if opts != nil && opts.Limit > 0 {
		limit = opts.Limit
	}
	totalPages := (totalCount + limit - 1) / limit

	return &ListEmailsResponse{
		Emails:     emails,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetEmail retrieves a specific email by ID
func (s *EmailService) GetEmail(ctx context.Context, emailID string) (*Email, error) {
	if emailID == "" {
		return nil, fmt.Errorf("email ID is required")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/emails/%s", emailID)})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var email Email
	if err := s.client.Do(ctx, req, &email); err != nil {
		return nil, fmt.Errorf("failed to get email: %w", err)
	}

	return &email, nil
}

// DeleteEmail deletes an email from the sent history
func (s *EmailService) DeleteEmail(ctx context.Context, emailID string) error {
	if emailID == "" {
		return fmt.Errorf("email ID is required")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/emails/%s", emailID)})

	req, err := http.NewRequestWithContext(ctx, "DELETE", u.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := s.client.Do(ctx, req, nil); err != nil {
		return fmt.Errorf("failed to delete email: %w", err)
	}

	return nil
}

// GetEmailQuota retrieves daily email sending quota information
func (s *EmailService) GetEmailQuota(ctx context.Context) (*EmailQuota, error) {
	u := s.client.BaseURL.ResolveReference(&url.URL{Path: "/v1/emails/quota"})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var quota EmailQuota
	if err := s.client.Do(ctx, req, &quota); err != nil {
		return nil, fmt.Errorf("failed to get email quota: %w", err)
	}

	return &quota, nil
}

// GetEmailStats retrieves email usage statistics
func (s *EmailService) GetEmailStats(ctx context.Context) (*EmailStats, error) {
	u := s.client.BaseURL.ResolveReference(&url.URL{Path: "/v1/emails/stats"})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var stats EmailStats
	if err := s.client.Do(ctx, req, &stats); err != nil {
		return nil, fmt.Errorf("failed to get email stats: %w", err)
	}

	return &stats, nil
}

// ValidateRecipients validates email addresses before sending
func (s *EmailService) ValidateRecipients(_ context.Context, recipients []string) error {
	if len(recipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	for _, email := range recipients {
		email = strings.TrimSpace(email)
		if email == "" {
			return fmt.Errorf("recipient email cannot be empty")
		}
		if !strings.Contains(email, "@") {
			return fmt.Errorf("invalid email address: %s", email)
		}
		// Basic email validation - could be enhanced
		parts := strings.Split(email, "@")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid email address format: %s", email)
		}
	}

	return nil
}

// GetBulkJobStatus retrieves the status of a bulk email job
func (s *EmailService) GetBulkJobStatus(ctx context.Context, jobID string) (*BulkEmailResponse, error) {
	if jobID == "" {
		return nil, fmt.Errorf("job ID is required")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/emails/bulk/%s", jobID)})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var response BulkEmailResponse
	if err := s.client.Do(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("failed to get bulk job status: %w", err)
	}

	return &response, nil
}

// GetAttachment retrieves an email attachment
func (s *EmailService) GetAttachment(ctx context.Context, emailID, attachmentID string) (*EmailAttachment, error) {
	if emailID == "" {
		return nil, fmt.Errorf("email ID is required")
	}
	if attachmentID == "" {
		return nil, fmt.Errorf("attachment ID is required")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/emails/%s/attachments/%s", emailID, attachmentID)})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var attachment EmailAttachment
	if err := s.client.Do(ctx, req, &attachment); err != nil {
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}

	return &attachment, nil
}

// DownloadAttachment downloads an email attachment content
func (s *EmailService) DownloadAttachment(ctx context.Context, emailID, attachmentID string) ([]byte, error) {
	if emailID == "" {
		return nil, fmt.Errorf("email ID is required")
	}
	if attachmentID == "" {
		return nil, fmt.Errorf("attachment ID is required")
	}

	u := s.client.BaseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/emails/%s/attachments/%s/download", emailID, attachmentID)})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var content []byte
	if err := s.client.Do(ctx, req, &content); err != nil {
		return nil, fmt.Errorf("failed to download attachment: %w", err)
	}

	return content, nil
}
