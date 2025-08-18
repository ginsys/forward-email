package output

import (
	"fmt"
	"time"

	"github.com/ginsys/forward-email/pkg/api"
)

// FormatDomainList formats a list of domains for display
func FormatDomainList(domains []api.Domain, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		// For JSON/YAML, return the domains directly
		return nil, fmt.Errorf("use direct JSON/YAML encoding for domains")
	}

	headers := []string{"NAME", "VERIFIED", "PLAN", "ALIASES", "MEMBERS", "CREATED"}
	table := NewTableData(headers)

	for i := range domains {
		domain := &domains[i]
		verified := FormatValue(domain.IsVerified)
		plan := FormatValue(domain.Plan)
		aliasCount := "-"
		memberCount := FormatValue(len(domain.Members))
		created := domain.CreatedAt.Format("2006-01-02")

		row := []string{
			domain.Name,
			verified,
			plan,
			aliasCount,
			memberCount,
			created,
		}
		table.AddRow(row)
	}

	return table, nil
}

// FormatDomainDetails formats detailed domain information
func FormatDomainDetails(domain *api.Domain, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for domain details")
	}

	headers := []string{"PROPERTY", "VALUE"}
	table := NewTableData(headers)

	// Basic information
	table.AddRow([]string{"ID", domain.ID})
	table.AddRow([]string{"Name", domain.Name})
	table.AddRow([]string{"Verified", FormatValue(domain.IsVerified)})
	table.AddRow([]string{"Plan", domain.Plan})
	table.AddRow([]string{"Global", FormatValue(domain.IsGlobal)})

	// DNS Records
	table.AddRow([]string{"MX Record", FormatValue(domain.HasMXRecord)})
	table.AddRow([]string{"TXT Record", FormatValue(domain.HasTXTRecord)})
	table.AddRow([]string{"DMARC Record", FormatValue(domain.HasDMARCRecord)})
	table.AddRow([]string{"SPF Record", FormatValue(domain.HasSPFRecord)})
	table.AddRow([]string{"DKIM Record", FormatValue(domain.HasDKIMRecord)})

	// Limits and settings
	table.AddRow([]string{"Max Forwarded Addresses", FormatValue(domain.MaxForwardedAddresses)})
	table.AddRow([]string{"Retention Days", FormatValue(domain.RetentionDays)})

	// Membership
	table.AddRow([]string{"Members", FormatValue(len(domain.Members))})
	table.AddRow([]string{"Pending Invitations", FormatValue(len(domain.Invitations))})

	// Timestamps
	table.AddRow([]string{"Created", domain.CreatedAt.Format(time.RFC3339)})
	table.AddRow([]string{"Updated", domain.UpdatedAt.Format(time.RFC3339)})

	// Verification record if present
	if domain.VerificationRecord != "" {
		table.AddRow([]string{"Verification Record", TruncateString(domain.VerificationRecord, 50)})
	}

	return table, nil
}

// FormatDNSRecords formats DNS records for display
func FormatDNSRecords(records []api.DNSRecord, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for DNS records")
	}

	headers := []string{"TYPE", "NAME", "VALUE", "PRIORITY", "TTL", "REQUIRED", "PURPOSE"}
	table := NewTableData(headers)

	for _, record := range records {
		priority := ""
		if record.Priority > 0 {
			priority = FormatValue(record.Priority)
		} else {
			priority = "-"
		}

		ttl := ""
		if record.TTL > 0 {
			ttl = FormatValue(record.TTL)
		} else {
			ttl = "-"
		}

		row := []string{
			record.Type,
			record.Name,
			TruncateString(record.Value, 40),
			priority,
			ttl,
			FormatValue(record.Required),
			record.Purpose,
		}
		table.AddRow(row)
	}

	return table, nil
}

// FormatDomainVerification formats domain verification status
func FormatDomainVerification(verification *api.DomainVerification, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for domain verification")
	}

	headers := []string{"PROPERTY", "VALUE"}
	table := NewTableData(headers)

	table.AddRow([]string{"Verified", FormatValue(verification.IsVerified)})
	table.AddRow([]string{"DNS Records Found", FormatValue(len(verification.DNSRecords))})
	table.AddRow([]string{"Missing Records", FormatValue(len(verification.MissingRecords))})
	table.AddRow([]string{"Last Checked", verification.LastCheckedAt.Format(time.RFC3339)})

	if verification.VerificationURL != "" {
		table.AddRow([]string{"Verification URL", verification.VerificationURL})
	}

	if len(verification.Errors) > 0 {
		table.AddRow([]string{"Errors", FormatValue(len(verification.Errors))})
		for i, err := range verification.Errors {
			table.AddRow([]string{fmt.Sprintf("Error %d", i+1), err})
		}
	}

	return table, nil
}

// FormatDomainQuota formats domain quota information
func FormatDomainQuota(quota *api.DomainQuota, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for domain quota")
	}

	headers := []string{"RESOURCE", "USED", "LIMIT", "PERCENTAGE"}
	table := NewTableData(headers)

	// Storage
	storageUsed := FormatBytes(quota.StorageUsed)
	storageLimit := FormatBytes(quota.StorageLimit)
	storagePct := FormatPercentage(quota.StorageUsed, quota.StorageLimit)
	table.AddRow([]string{"Storage", storageUsed, storageLimit, storagePct})

	// Aliases
	aliasesPct := FormatPercentage(int64(quota.AliasesUsed), int64(quota.AliasesLimit))
	table.AddRow([]string{"Aliases", FormatValue(quota.AliasesUsed), FormatValue(quota.AliasesLimit), aliasesPct})

	// Forwarding
	forwardingPct := FormatPercentage(int64(quota.ForwardingUsed), int64(quota.ForwardingLimit))
	table.AddRow([]string{"Forwarding", FormatValue(quota.ForwardingUsed), FormatValue(quota.ForwardingLimit), forwardingPct})

	// Bandwidth
	bandwidthUsed := FormatBytes(quota.BandwidthUsed)
	bandwidthLimit := FormatBytes(quota.BandwidthLimit)
	bandwidthPct := FormatPercentage(quota.BandwidthUsed, quota.BandwidthLimit)
	table.AddRow([]string{"Bandwidth", bandwidthUsed, bandwidthLimit, bandwidthPct})

	// Daily emails
	emailsPct := FormatPercentage(int64(quota.EmailsSentToday), int64(quota.EmailsLimitDaily))
	table.AddRow([]string{"Daily Emails", FormatValue(quota.EmailsSentToday), FormatValue(quota.EmailsLimitDaily), emailsPct})

	return table, nil
}

// FormatDomainStats formats domain statistics
func FormatDomainStats(stats *api.DomainStats, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for domain stats")
	}

	headers := []string{"METRIC", "VALUE"}
	table := NewTableData(headers)

	table.AddRow([]string{"Total Aliases", FormatValue(stats.TotalAliases)})
	table.AddRow([]string{"Active Aliases", FormatValue(stats.ActiveAliases)})
	table.AddRow([]string{"Total Members", FormatValue(stats.TotalMembers)})
	table.AddRow([]string{"Emails Sent", FormatValue(stats.EmailsSent)})
	table.AddRow([]string{"Emails Received", FormatValue(stats.EmailsReceived)})
	table.AddRow([]string{"Last Activity", stats.LastActivityAt.Format(time.RFC3339)})
	table.AddRow([]string{"Created", stats.CreatedAt.Format(time.RFC3339)})

	return table, nil
}

// FormatDomainMembers formats domain members list
func FormatDomainMembers(members []api.DomainMember, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for domain members")
	}

	headers := []string{"ID", "EMAIL", "NAME", "GROUP", "JOINED"}
	table := NewTableData(headers)

	for i := range members {
		member := &members[i]
		displayName := member.User.DisplayName
		if displayName == "" {
			displayName = fmt.Sprintf("%s %s", member.User.GivenName, member.User.FamilyName)
		}
		if displayName == " " {
			displayName = "-"
		}

		row := []string{
			TruncateString(member.ID, 8),
			member.User.Email,
			displayName,
			member.Group,
			member.JoinedAt.Format("2006-01-02"),
		}
		table.AddRow(row)
	}

	return table, nil
}
