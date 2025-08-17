package api

import (
	"time"
)

// Alias represents a Forward Email alias
type Alias struct {
	ID          string             `json:"id"`
	DomainID    string             `json:"domain_id"`
	Name        string             `json:"name"`
	IsEnabled   bool               `json:"is_enabled"`
	Recipients  []string           `json:"recipients"`
	Labels      []string           `json:"labels,omitempty"`
	Description string             `json:"description,omitempty"`
	HasIMAP     bool               `json:"has_imap"`
	HasPGP      bool               `json:"has_pgp"`
	PublicKey   string             `json:"public_key,omitempty"`
	HasPassword bool               `json:"has_password"`
	Quota       *AliasQuota        `json:"quota,omitempty"`
	Vacation    *VacationResponder `json:"vacation,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// AliasQuota represents alias storage and email quotas
type AliasQuota struct {
	StorageUsed  int64 `json:"storage_used"`  // bytes used
	StorageLimit int64 `json:"storage_limit"` // bytes limit
	EmailsSent   int   `json:"emails_sent"`   // emails sent today
	EmailsLimit  int   `json:"emails_limit"`  // daily email limit
}

// VacationResponder represents auto-responder configuration
type VacationResponder struct {
	IsEnabled   bool      `json:"is_enabled"`
	Subject     string    `json:"subject,omitempty"`
	Message     string    `json:"message,omitempty"`
	StartDate   time.Time `json:"start_date,omitempty"`
	EndDate     time.Time `json:"end_date,omitempty"`
	LastUpdated time.Time `json:"last_updated,omitempty"`
}

// ListAliasesOptions represents options for listing aliases
type ListAliasesOptions struct {
	Domain  string `json:"domain,omitempty"`   // Domain name or ID to filter by
	Page    int    `json:"page,omitempty"`     // Page number (1-based)
	Limit   int    `json:"limit,omitempty"`    // Items per page
	Sort    string `json:"sort,omitempty"`     // Sort field (name, created, updated)
	Order   string `json:"order,omitempty"`    // Sort order (asc, desc)
	Search  string `json:"search,omitempty"`   // Search term for alias names
	Enabled *bool  `json:"enabled,omitempty"`  // Filter by enabled status
	Labels  string `json:"labels,omitempty"`   // Filter by labels (comma-separated)
	HasIMAP *bool  `json:"has_imap,omitempty"` // Filter by IMAP capability
}

// ListAliasesResponse represents the response from listing aliases
type ListAliasesResponse struct {
	Aliases    []Alias `json:"aliases"`
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	TotalPages int     `json:"total_pages"`
}

// CreateAliasRequest represents a request to create an alias
type CreateAliasRequest struct {
	Name        string   `json:"name"`                  // Alias name (local part)
	Recipients  []string `json:"recipients"`            // List of recipients
	Labels      []string `json:"labels,omitempty"`      // Optional labels
	Description string   `json:"description,omitempty"` // Optional description
	IsEnabled   bool     `json:"is_enabled"`            // Default enabled status
	HasIMAP     bool     `json:"has_imap,omitempty"`    // Enable IMAP access
	HasPGP      bool     `json:"has_pgp,omitempty"`     // Enable PGP encryption
	PublicKey   string   `json:"public_key,omitempty"`  // PGP public key
}

// UpdateAliasRequest represents a request to update an alias
type UpdateAliasRequest struct {
	Recipients  []string `json:"recipients,omitempty"`  // Update recipients
	Labels      []string `json:"labels,omitempty"`      // Update labels
	Description *string  `json:"description,omitempty"` // Update description (nil to clear)
	IsEnabled   *bool    `json:"is_enabled,omitempty"`  // Update enabled status
	HasIMAP     *bool    `json:"has_imap,omitempty"`    // Update IMAP access
	HasPGP      *bool    `json:"has_pgp,omitempty"`     // Update PGP encryption
	PublicKey   *string  `json:"public_key,omitempty"`  // Update PGP public key
}

// GeneratePasswordResponse represents the response from generating an IMAP password
type GeneratePasswordResponse struct {
	Password string `json:"password"`
}

// AliasStats represents alias usage statistics
type AliasStats struct {
	EmailsReceived int64     `json:"emails_received"`
	EmailsSent     int64     `json:"emails_sent"`
	StorageUsed    int64     `json:"storage_used"`
	LastActivity   time.Time `json:"last_activity"`
	RecentSenders  []string  `json:"recent_senders,omitempty"`
}
