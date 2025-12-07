package api

import (
	"time"
)

// Domain represents a Forward Email domain with complete configuration and status information.
// It includes DNS verification status, plan details, security settings, and member management.
type Domain struct {
	// Timestamps
	CreatedAt time.Time `json:"created_at"` // Domain creation timestamp
	UpdatedAt time.Time `json:"updated_at"` // Last modification timestamp

	// Related objects
	Settings    *DomainSettings    `json:"settings,omitempty"`    // Domain-specific configuration settings
	Members     []DomainMember     `json:"members,omitempty"`     // Users with access to this domain
	Invitations []DomainInvitation `json:"invitations,omitempty"` // Pending member invitations

	// Basic properties
	ID                    string `json:"id"`                      // Unique domain identifier (UUID)
	Name                  string `json:"name"`                    // Fully qualified domain name
	VerificationRecord    string `json:"verification_record"`     // TXT record value for domain verification
	Plan                  string `json:"plan"`                    // Subscription plan: free, enhanced_protection, team
	MaxForwardedAddresses int    `json:"max_forwarded_addresses"` // Maximum aliases allowed for this domain
	RetentionDays         int    `json:"retention_days"`          // Email retention period in days

	// Status flags
	IsGlobal       bool `json:"is_global"`        // Whether this is a global Forward Email domain
	HasMXRecord    bool `json:"has_mx_record"`    // MX record verification status
	HasTXTRecord   bool `json:"has_txt_record"`   // TXT record verification status
	HasDMARCRecord bool `json:"has_dmarc_record"` // DMARC policy verification status
	HasSPFRecord   bool `json:"has_spf_record"`   // SPF record verification status
	HasDKIMRecord  bool `json:"has_dkim_record"`  // DKIM signature verification status
	IsVerified     bool `json:"is_verified"`      // Overall domain verification status

	// SMTP Status
	HasSMTP         bool      `json:"has_smtp"`                   // SMTP outbound enabled status
	IsSMTPSuspended bool      `json:"is_smtp_suspended"`          // SMTP suspension status
	SMTPVerifiedAt  time.Time `json:"smtp_verified_at,omitempty"` // SMTP verification timestamp

	// Deliverability
	HasDeliveryLogs bool   `json:"has_delivery_logs"`        // Opt-in for success delivery logs
	BounceWebhook   string `json:"bounce_webhook,omitempty"` // Separate URL for bounce notifications

	// Alias Settings
	HasRegex                bool     `json:"has_regex"`                        // Enable regex alias support
	HasCatchall             bool     `json:"has_catchall"`                     // Enable catch-all aliases
	IsCatchallRegexDisabled bool     `json:"is_catchall_regex_disabled"`       // Disable catch-all on large domains
	AliasCount              int      `json:"alias_count"`                      // Total alias count
	MaxRecipientsPerAlias   int      `json:"max_recipients_per_alias"`         // Per-alias recipient limit (max: 1000)
	MaxQuotaPerAlias        int64    `json:"max_quota_per_alias"`              // Storage quota per alias (max: 100GB)
	Allowlist               []string `json:"allowlist,omitempty"`              // Permitted forwarding destinations
	Denylist                []string `json:"denylist,omitempty"`               // Blocked addresses
	RestrictedAliasNames    []string `json:"restricted_alias_names,omitempty"` // Admin-only alias name restrictions

	// Verification
	HasRecipientVerification bool `json:"has_recipient_verification"` // Enable verification emails to recipients
	HasCustomVerification    bool `json:"has_custom_verification"`    // Enable custom verification templates

	// DNS/DKIM
	HasReturnPathRecord bool   `json:"has_return_path_record"`      // Return-path DNS record status
	DKIMModulusLength   int    `json:"dkim_modulus_length"`         // RSA key length (1024 or 2048)
	DKIMKeySelector     string `json:"dkim_key_selector,omitempty"` // DKIM selector
	ReturnPath          string `json:"return_path,omitempty"`       // Return-path domain
	IgnoreMXCheck       bool   `json:"ignore_mx_check"`             // Bypass MX validation

	// Other
	HasNewsletter bool `json:"has_newsletter"` // Newsletter capability flag
}

// DomainSettings represents configurable domain-specific settings.
// These settings control security features, webhook integration, and service ports.
type DomainSettings struct {
	// Webhook integration
	WebhookURL string `json:"webhook_url,omitempty"` // HTTP endpoint for email notifications
	WebhookKey string `json:"webhook_key,omitempty"` // Authentication key for webhook requests

	// Service ports
	SMTPPort    int `json:"smtp_port"`    // Custom SMTP port (default: 25)
	IMAPPort    int `json:"imap_port"`    // Custom IMAP port (default: 993)
	CalDAVPort  int `json:"caldav_port"`  // Custom CalDAV port (default: 993)
	CardDAVPort int `json:"carddav_port"` // Custom CardDAV port (default: 993)

	// Security protection features
	HasAdultContentProtection bool `json:"has_adult_content_protection"` // Block adult content
	HasPhishingProtection     bool `json:"has_phishing_protection"`      // Anti-phishing filtering
	HasExecutableProtection   bool `json:"has_executable_protection"`    // Block executable attachments
	HasVirusProtection        bool `json:"has_virus_protection"`         // Virus scanning enabled
}

// DomainMember represents a user with access to a domain.
// Members can have different permission levels and access rights.
type DomainMember struct {
	JoinedAt time.Time `json:"joined_at"` // When the user joined this domain
	User     User      `json:"user"`      // User information and profile
	ID       string    `json:"id"`        // Unique member identifier
	Group    string    `json:"group"`     // Permission group: admin, user
}

// DomainInvitation represents a pending domain invitation
type DomainInvitation struct {
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Group     string    `json:"group"`
}

// User represents a Forward Email user
type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	GivenName   string `json:"given_name"`
	FamilyName  string `json:"family_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// DNSRecord represents a DNS record required for domain setup
type DNSRecord struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Purpose  string `json:"purpose"`
	Priority int    `json:"priority,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Required bool   `json:"required"`
}

// DomainVerification represents the verification status of a domain
type DomainVerification struct {
	LastCheckedAt   time.Time   `json:"last_checked_at"`
	DNSRecords      []DNSRecord `json:"dns_records"`
	MissingRecords  []DNSRecord `json:"missing_records"`
	Errors          []string    `json:"errors,omitempty"`
	VerificationURL string      `json:"verification_url,omitempty"`
	IsVerified      bool        `json:"is_verified"`
}

// CreateDomainRequest represents a request to create a new domain
type CreateDomainRequest struct {
	Name string `json:"name" validate:"required,fqdn"`
	Plan string `json:"plan,omitempty"`
}

// UpdateDomainRequest represents a request to update domain settings
type UpdateDomainRequest struct {
	// Existing fields
	MaxForwardedAddresses *int            `json:"max_forwarded_addresses,omitempty"`
	RetentionDays         *int            `json:"retention_days,omitempty"`
	Settings              *DomainSettings `json:"settings,omitempty"`

	// New fields
	HasDeliveryLogs          *bool    `json:"has_delivery_logs,omitempty"`
	BounceWebhook            *string  `json:"bounce_webhook,omitempty"`
	HasRegex                 *bool    `json:"has_regex,omitempty"`
	HasCatchall              *bool    `json:"has_catchall,omitempty"`
	IsCatchallRegexDisabled  *bool    `json:"is_catchall_regex_disabled,omitempty"`
	MaxRecipientsPerAlias    *int     `json:"max_recipients_per_alias,omitempty"`
	MaxQuotaPerAlias         *int64   `json:"max_quota_per_alias,omitempty"`
	Allowlist                []string `json:"allowlist,omitempty"`
	Denylist                 []string `json:"denylist,omitempty"`
	HasRecipientVerification *bool    `json:"has_recipient_verification,omitempty"`
	IgnoreMXCheck            *bool    `json:"ignore_mx_check,omitempty"`
}

// ListDomainsOptions represents options for listing domains
type ListDomainsOptions struct {
	Verified *bool  `json:"verified,omitempty"`
	Sort     string `json:"sort,omitempty"`
	Order    string `json:"order,omitempty"`
	Search   string `json:"search,omitempty"`
	Plan     string `json:"plan,omitempty"`
	Page     int    `json:"page,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// ListDomainsResponse represents the response from listing domains
type ListDomainsResponse struct {
	Domains    []Domain   `json:"domains"`
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination information
type Pagination struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// DomainQuota represents domain quota information
type DomainQuota struct {
	StorageUsed      int64 `json:"storage_used"`
	StorageLimit     int64 `json:"storage_limit"`
	AliasesUsed      int   `json:"aliases_used"`
	AliasesLimit     int   `json:"aliases_limit"`
	ForwardingUsed   int   `json:"forwarding_used"`
	ForwardingLimit  int   `json:"forwarding_limit"`
	BandwidthUsed    int64 `json:"bandwidth_used"`
	BandwidthLimit   int64 `json:"bandwidth_limit"`
	EmailsSentToday  int   `json:"emails_sent_today"`
	EmailsLimitDaily int   `json:"emails_limit_daily"`
}

// DomainStats represents domain statistics
type DomainStats struct {
	LastActivityAt time.Time `json:"last_activity_at"`
	CreatedAt      time.Time `json:"created_at"`
	TotalAliases   int       `json:"total_aliases"`
	ActiveAliases  int       `json:"active_aliases"`
	TotalMembers   int       `json:"total_members"`
	EmailsSent     int       `json:"emails_sent"`
	EmailsReceived int       `json:"emails_received"`
}

// DomainGroup represents domain permission groups
type DomainGroup string

const (
	// DomainGroupAdmin represents administrative access level
	DomainGroupAdmin DomainGroup = "admin"
	// DomainGroupUser represents user access level
	DomainGroupUser DomainGroup = "user"
)

// DomainPlan represents domain plan types
type DomainPlan string

const (
	// DomainPlanFree represents the free plan tier
	DomainPlanFree DomainPlan = "free"
	// DomainPlanEnhancedProtection represents the enhanced protection plan tier
	DomainPlanEnhancedProtection DomainPlan = "enhanced_protection"
	// DomainPlanTeam represents the team plan tier
	DomainPlanTeam DomainPlan = "team"
)

// DomainSortField represents fields that can be used for sorting domains
type DomainSortField string

const (
	// DomainSortByName sorts domains by name
	DomainSortByName DomainSortField = "name"
	// DomainSortByCreated sorts domains by creation date
	DomainSortByCreated DomainSortField = "created_at"
	// DomainSortByUpdated sorts domains by last update date
	DomainSortByUpdated DomainSortField = "updated_at"
	// DomainSortByVerified sorts domains by verification status
	DomainSortByVerified DomainSortField = "is_verified"
	// DomainSortByPlan sorts domains by plan type
	DomainSortByPlan DomainSortField = "plan"
)

// SortOrder represents sort order options
type SortOrder string

const (
	// SortOrderAsc represents ascending sort order
	SortOrderAsc SortOrder = "asc"
	// SortOrderDesc represents descending sort order
	SortOrderDesc SortOrder = "desc"
)
