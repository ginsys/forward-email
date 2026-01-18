package api

import (
	"time"
)

// Email represents a Forward Email message
// Note: From, To, and Message-ID are available in the Headers map, not as separate fields
type Email struct {
	SentAt      time.Time         `json:"sent_at"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeliveredAt *time.Time        `json:"delivered_at,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"` // Contains From, To, Message-ID, etc.
	Attachments []EmailAttachment `json:"attachments,omitempty"`
	ID          string            `json:"id"`
	Subject     string            `json:"subject"`
	Text        string            `json:"text,omitempty"`
	HTML        string            `json:"html,omitempty"`
	Status      string            `json:"status"` // sent, delivered, bounced, failed
	StatusInfo  string            `json:"status_info,omitempty"`
}

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	Size        int64  `json:"size"` // bytes
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	CID         string `json:"cid,omitempty"` // Content-ID for inline images
}

// EmailQuota represents daily email sending limits
type EmailQuota struct {
	ResetTime      time.Time `json:"reset_time"`      // when quota resets
	EmailsSent     int       `json:"emails_sent"`     // emails sent today
	EmailsLimit    int       `json:"emails_limit"`    // daily email limit
	OverageUsed    int       `json:"overage_used"`    // overage emails used
	OverageAllowed bool      `json:"overage_allowed"` // can exceed limit
}

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	Headers     map[string]string `json:"headers,omitempty"`     // custom headers
	Variables   map[string]string `json:"variables,omitempty"`   // template variables (future)
	To          []string          `json:"to"`                    // recipient email addresses
	CC          []string          `json:"cc,omitempty"`          // carbon copy recipients
	BCC         []string          `json:"bcc,omitempty"`         // blind carbon copy recipients
	Attachments []AttachmentData  `json:"attachments,omitempty"` // file attachments
	From        string            `json:"from"`                  // sender email address
	Subject     string            `json:"subject"`               // email subject
	Text        string            `json:"text,omitempty"`        // plain text content
	HTML        string            `json:"html,omitempty"`        // HTML content
	Template    string            `json:"template,omitempty"`    // template name (future)
}

// AttachmentData represents attachment data for sending
type AttachmentData struct {
	Content     []byte `json:"content"`                // file content (base64 encoded)
	Filename    string `json:"filename"`               // file name
	ContentType string `json:"content_type,omitempty"` // MIME type
	CID         string `json:"cid,omitempty"`          // Content-ID for inline images
}

// SendEmailResponse represents the response from sending an email
type SendEmailResponse struct {
	SentAt    time.Time `json:"sent_at"`    // when email was sent
	ID        string    `json:"id"`         // email ID
	MessageID string    `json:"message_id"` // SMTP message ID
	Status    string    `json:"status"`     // queued, sent, failed
}

// ListEmailsOptions represents options for listing emails
type ListEmailsOptions struct {
	HasAttach *bool  `json:"has_attach,omitempty"` // Filter by attachment presence
	Sort      string `json:"sort,omitempty"`       // Sort field (sent_at, subject, from, to)
	Order     string `json:"order,omitempty"`      // Sort order (asc, desc)
	Search    string `json:"search,omitempty"`     // Search in subject, from, to
	Status    string `json:"status,omitempty"`     // Filter by status
	From      string `json:"from,omitempty"`       // Filter by sender
	To        string `json:"to,omitempty"`         // Filter by recipient
	DateFrom  string `json:"date_from,omitempty"`  // Filter by date range (YYYY-MM-DD)
	DateTo    string `json:"date_to,omitempty"`    // Filter by date range (YYYY-MM-DD)
	Page      int    `json:"page,omitempty"`       // Page number (1-based)
	Limit     int    `json:"limit,omitempty"`      // Items per page
}

// ListEmailsResponse represents the response from listing emails
type ListEmailsResponse struct {
	Emails     []Email `json:"emails"`
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	TotalPages int     `json:"total_pages"`
}

// BulkEmailRequest represents a request to send multiple emails
type BulkEmailRequest struct {
	Emails   []SendEmailRequest `json:"emails"`             // list of emails to send
	Template string             `json:"template,omitempty"` // common template (future)
	DryRun   bool               `json:"dry_run,omitempty"`  // validate without sending
}

// BulkEmailResponse represents the response from bulk email sending
type BulkEmailResponse struct {
	Results     []SendEmailResponse `json:"results"`      // individual results
	JobID       string              `json:"job_id"`       // bulk job ID
	TotalEmails int                 `json:"total_emails"` // total emails in job
	Queued      int                 `json:"queued"`       // successfully queued
	Failed      int                 `json:"failed"`       // failed to queue
}

// EmailTemplate represents an email template (future feature)
type EmailTemplate struct {
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Headers     map[string]string `json:"headers,omitempty"`   // default headers
	Variables   []string          `json:"variables,omitempty"` // available variables
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Subject     string            `json:"subject"`
	TextContent string            `json:"text_content"`
	HTMLContent string            `json:"html_content"`
}
