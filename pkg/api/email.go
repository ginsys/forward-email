package api

import (
	"time"
)

// Email represents a Forward Email message
type Email struct {
	ID          string            `json:"id"`
	MessageID   string            `json:"message_id"`
	From        string            `json:"from"`
	To          []string          `json:"to"`
	CC          []string          `json:"cc,omitempty"`
	BCC         []string          `json:"bcc,omitempty"`
	Subject     string            `json:"subject"`
	Text        string            `json:"text,omitempty"`
	HTML        string            `json:"html,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
	Status      string            `json:"status"` // sent, delivered, bounced, failed
	StatusInfo  string            `json:"status_info,omitempty"`
	SentAt      time.Time         `json:"sent_at"`
	DeliveredAt *time.Time        `json:"delivered_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`          // bytes
	CID         string `json:"cid,omitempty"` // Content-ID for inline images
}

// EmailQuota represents daily email sending limits
type EmailQuota struct {
	EmailsSent     int       `json:"emails_sent"`     // emails sent today
	EmailsLimit    int       `json:"emails_limit"`    // daily email limit
	ResetTime      time.Time `json:"reset_time"`      // when quota resets
	OverageAllowed bool      `json:"overage_allowed"` // can exceed limit
	OverageUsed    int       `json:"overage_used"`    // overage emails used
}

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	From        string            `json:"from"`                  // sender email address
	To          []string          `json:"to"`                    // recipient email addresses
	CC          []string          `json:"cc,omitempty"`          // carbon copy recipients
	BCC         []string          `json:"bcc,omitempty"`         // blind carbon copy recipients
	Subject     string            `json:"subject"`               // email subject
	Text        string            `json:"text,omitempty"`        // plain text content
	HTML        string            `json:"html,omitempty"`        // HTML content
	Headers     map[string]string `json:"headers,omitempty"`     // custom headers
	Attachments []AttachmentData  `json:"attachments,omitempty"` // file attachments
	Template    string            `json:"template,omitempty"`    // template name (future)
	Variables   map[string]string `json:"variables,omitempty"`   // template variables (future)
}

// AttachmentData represents attachment data for sending
type AttachmentData struct {
	Filename    string `json:"filename"`               // file name
	ContentType string `json:"content_type,omitempty"` // MIME type
	Content     []byte `json:"content"`                // file content (base64 encoded)
	CID         string `json:"cid,omitempty"`          // Content-ID for inline images
}

// SendEmailResponse represents the response from sending an email
type SendEmailResponse struct {
	ID        string    `json:"id"`         // email ID
	MessageID string    `json:"message_id"` // SMTP message ID
	Status    string    `json:"status"`     // queued, sent, failed
	SentAt    time.Time `json:"sent_at"`    // when email was sent
}

// ListEmailsOptions represents options for listing emails
type ListEmailsOptions struct {
	Page      int    `json:"page,omitempty"`       // Page number (1-based)
	Limit     int    `json:"limit,omitempty"`      // Items per page
	Sort      string `json:"sort,omitempty"`       // Sort field (sent_at, subject, from, to)
	Order     string `json:"order,omitempty"`      // Sort order (asc, desc)
	Search    string `json:"search,omitempty"`     // Search in subject, from, to
	Status    string `json:"status,omitempty"`     // Filter by status
	From      string `json:"from,omitempty"`       // Filter by sender
	To        string `json:"to,omitempty"`         // Filter by recipient
	DateFrom  string `json:"date_from,omitempty"`  // Filter by date range (YYYY-MM-DD)
	DateTo    string `json:"date_to,omitempty"`    // Filter by date range (YYYY-MM-DD)
	HasAttach *bool  `json:"has_attach,omitempty"` // Filter by attachment presence
}

// ListEmailsResponse represents the response from listing emails
type ListEmailsResponse struct {
	Emails     []Email `json:"emails"`
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	TotalPages int     `json:"total_pages"`
}

// EmailStats represents email usage statistics
type EmailStats struct {
	TotalSent      int64     `json:"total_sent"`
	TotalDelivered int64     `json:"total_delivered"`
	TotalBounced   int64     `json:"total_bounced"`
	TotalFailed    int64     `json:"total_failed"`
	DeliveryRate   float64   `json:"delivery_rate"` // percentage
	BounceRate     float64   `json:"bounce_rate"`   // percentage
	LastSent       time.Time `json:"last_sent"`
	TopRecipients  []string  `json:"top_recipients,omitempty"`
}

// BulkEmailRequest represents a request to send multiple emails
type BulkEmailRequest struct {
	Emails   []SendEmailRequest `json:"emails"`             // list of emails to send
	Template string             `json:"template,omitempty"` // common template (future)
	DryRun   bool               `json:"dry_run,omitempty"`  // validate without sending
}

// BulkEmailResponse represents the response from bulk email sending
type BulkEmailResponse struct {
	JobID       string              `json:"job_id"`       // bulk job ID
	TotalEmails int                 `json:"total_emails"` // total emails in job
	Queued      int                 `json:"queued"`       // successfully queued
	Failed      int                 `json:"failed"`       // failed to queue
	Results     []SendEmailResponse `json:"results"`      // individual results
}

// EmailTemplate represents an email template (future feature)
type EmailTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Subject     string            `json:"subject"`
	TextContent string            `json:"text_content"`
	HTMLContent string            `json:"html_content"`
	Variables   []string          `json:"variables,omitempty"` // available variables
	Headers     map[string]string `json:"headers,omitempty"`   // default headers
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
