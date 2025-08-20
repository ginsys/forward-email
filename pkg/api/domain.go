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
	MaxForwardedAddresses *int            `json:"max_forwarded_addresses,omitempty"`
	RetentionDays         *int            `json:"retention_days,omitempty"`
	Settings              *DomainSettings `json:"settings,omitempty"`
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
	DomainGroupAdmin DomainGroup = "admin"
	DomainGroupUser  DomainGroup = "user"
)

// DomainPlan represents domain plan types
type DomainPlan string

const (
	DomainPlanFree               DomainPlan = "free"
	DomainPlanEnhancedProtection DomainPlan = "enhanced_protection"
	DomainPlanTeam               DomainPlan = "team"
)

// DomainSortField represents fields that can be used for sorting domains
type DomainSortField string

const (
	DomainSortByName     DomainSortField = "name"
	DomainSortByCreated  DomainSortField = "created_at"
	DomainSortByUpdated  DomainSortField = "updated_at"
	DomainSortByVerified DomainSortField = "is_verified"
	DomainSortByPlan     DomainSortField = "plan"
)

// SortOrder represents sort order options
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)
