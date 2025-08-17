package api

import (
	"time"
)

// Domain represents a Forward Email domain
type Domain struct {
	ID                    string             `json:"id"`
	Name                  string             `json:"name"`
	IsGlobal              bool               `json:"is_global"`
	HasMXRecord           bool               `json:"has_mx_record"`
	HasTXTRecord          bool               `json:"has_txt_record"`
	HasDMARCRecord        bool               `json:"has_dmarc_record"`
	HasSPFRecord          bool               `json:"has_spf_record"`
	HasDKIMRecord         bool               `json:"has_dkim_record"`
	IsVerified            bool               `json:"is_verified"`
	VerificationRecord    string             `json:"verification_record"`
	MaxForwardedAddresses int                `json:"max_forwarded_addresses"`
	RetentionDays         int                `json:"retention_days"`
	Plan                  string             `json:"plan"`
	Settings              *DomainSettings    `json:"settings,omitempty"`
	Members               []DomainMember     `json:"members,omitempty"`
	Invitations           []DomainInvitation `json:"invitations,omitempty"`
	CreatedAt             time.Time          `json:"created_at"`
	UpdatedAt             time.Time          `json:"updated_at"`
}

// DomainSettings represents domain-specific settings
type DomainSettings struct {
	HasAdultContentProtection bool   `json:"has_adult_content_protection"`
	HasPhishingProtection     bool   `json:"has_phishing_protection"`
	HasExecutableProtection   bool   `json:"has_executable_protection"`
	HasVirusProtection        bool   `json:"has_virus_protection"`
	SMTPPort                  int    `json:"smtp_port"`
	IMAPPort                  int    `json:"imap_port"`
	CalDAVPort                int    `json:"caldav_port"`
	CardDAVPort               int    `json:"carddav_port"`
	WebhookURL                string `json:"webhook_url,omitempty"`
	WebhookKey                string `json:"webhook_key,omitempty"`
}

// DomainMember represents a domain member
type DomainMember struct {
	ID       string    `json:"id"`
	User     User      `json:"user"`
	Group    string    `json:"group"`
	JoinedAt time.Time `json:"joined_at"`
}

// DomainInvitation represents a pending domain invitation
type DomainInvitation struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Group     string    `json:"group"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
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
	Priority int    `json:"priority,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Required bool   `json:"required"`
	Purpose  string `json:"purpose"`
}

// DomainVerification represents the verification status of a domain
type DomainVerification struct {
	IsVerified      bool        `json:"is_verified"`
	DNSRecords      []DNSRecord `json:"dns_records"`
	MissingRecords  []DNSRecord `json:"missing_records"`
	LastCheckedAt   time.Time   `json:"last_checked_at"`
	VerificationURL string      `json:"verification_url,omitempty"`
	Errors          []string    `json:"errors,omitempty"`
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
	Page     int    `json:"page,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Sort     string `json:"sort,omitempty"`
	Order    string `json:"order,omitempty"`
	Search   string `json:"search,omitempty"`
	Verified *bool  `json:"verified,omitempty"`
	Plan     string `json:"plan,omitempty"`
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
	TotalAliases   int       `json:"total_aliases"`
	ActiveAliases  int       `json:"active_aliases"`
	TotalMembers   int       `json:"total_members"`
	EmailsSent     int       `json:"emails_sent"`
	EmailsReceived int       `json:"emails_received"`
	LastActivityAt time.Time `json:"last_activity_at"`
	CreatedAt      time.Time `json:"created_at"`
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
