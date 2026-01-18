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

	// === Basic Information ===
	table.AddRow([]string{"ID", domain.ID})
	table.AddRow([]string{"Name", domain.Name})
	table.AddRow([]string{"Verified", FormatValue(domain.IsVerified)})
	table.AddRow([]string{"Plan", domain.Plan})
	table.AddRow([]string{"Global", FormatValue(domain.IsGlobal)})
	if domain.VerificationRecord != "" {
		table.AddRow([]string{"Verification Record", TruncateString(domain.VerificationRecord, 50)})
	}

	// === DNS Records ===
	table.AddRow([]string{"MX Record", FormatValue(domain.HasMXRecord)})
	table.AddRow([]string{"TXT Record", FormatValue(domain.HasTXTRecord)})
	table.AddRow([]string{"DMARC Record", FormatValue(domain.HasDMARCRecord)})
	table.AddRow([]string{"SPF Record", FormatValue(domain.HasSPFRecord)})
	table.AddRow([]string{"DKIM Record", FormatValue(domain.HasDKIMRecord)})
	table.AddRow([]string{"Return Path Record", FormatValue(domain.HasReturnPathRecord)})

	// === SMTP Status ===
	table.AddRow([]string{"SMTP Enabled", FormatValue(domain.HasSMTP)})
	if domain.HasSMTP {
		table.AddRow([]string{"SMTP Suspended", FormatValue(domain.IsSMTPSuspended)})
		if !domain.SMTPVerifiedAt.IsZero() {
			table.AddRow([]string{"SMTP Verified", domain.SMTPVerifiedAt.Format(time.RFC3339)})
		}
	}

	// === Deliverability Settings ===
	table.AddRow([]string{"Delivery Logs", FormatValue(domain.HasDeliveryLogs)})
	if domain.BounceWebhook != "" {
		table.AddRow([]string{"Bounce Webhook", TruncateString(domain.BounceWebhook, 50)})
	}

	// === Protection Settings ===
	if domain.Settings != nil {
		table.AddRow([]string{"Adult Content Protection", FormatValue(domain.Settings.HasAdultContentProtection)})
		table.AddRow([]string{"Phishing Protection", FormatValue(domain.Settings.HasPhishingProtection)})
		table.AddRow([]string{"Executable Protection", FormatValue(domain.Settings.HasExecutableProtection)})
		table.AddRow([]string{"Virus Protection", FormatValue(domain.Settings.HasVirusProtection)})
	}

	// === Alias Settings ===
	table.AddRow([]string{"Alias Count", FormatValue(domain.AliasCount)})
	table.AddRow([]string{"Max Forwarded Addresses", FormatValue(domain.MaxForwardedAddresses)})
	table.AddRow([]string{"Regex Aliases", FormatValue(domain.HasRegex)})
	table.AddRow([]string{"Catch-All Aliases", FormatValue(domain.HasCatchall)})
	if domain.IsCatchallRegexDisabled {
		table.AddRow([]string{"Catch-All Regex Disabled", FormatValue(domain.IsCatchallRegexDisabled)})
	}
	if domain.MaxRecipientsPerAlias > 0 {
		table.AddRow([]string{"Max Recipients Per Alias", FormatValue(domain.MaxRecipientsPerAlias)})
	}
	if domain.MaxQuotaPerAlias > 0 {
		table.AddRow([]string{"Max Quota Per Alias", FormatBytes(domain.MaxQuotaPerAlias)})
	}
	if len(domain.Allowlist) > 0 {
		table.AddRow([]string{"Allowlist", FormatValue(len(domain.Allowlist)) + " addresses"})
	}
	if len(domain.Denylist) > 0 {
		table.AddRow([]string{"Denylist", FormatValue(len(domain.Denylist)) + " addresses"})
	}

	// === Verification Settings ===
	table.AddRow([]string{"Recipient Verification", FormatValue(domain.HasRecipientVerification)})
	table.AddRow([]string{"Custom Verification", FormatValue(domain.HasCustomVerification)})

	// === Webhook & Port Settings ===
	if domain.Settings != nil {
		if domain.Settings.WebhookURL != "" {
			table.AddRow([]string{"Webhook URL", TruncateString(domain.Settings.WebhookURL, 50)})
		}
		if domain.Settings.SMTPPort > 0 {
			table.AddRow([]string{"SMTP Port", FormatValue(domain.Settings.SMTPPort)})
		}
		if domain.Settings.IMAPPort > 0 {
			table.AddRow([]string{"IMAP Port", FormatValue(domain.Settings.IMAPPort)})
		}
		if domain.Settings.CalDAVPort > 0 {
			table.AddRow([]string{"CalDAV Port", FormatValue(domain.Settings.CalDAVPort)})
		}
		if domain.Settings.CardDAVPort > 0 {
			table.AddRow([]string{"CardDAV Port", FormatValue(domain.Settings.CardDAVPort)})
		}
	}

	// === DKIM Settings ===
	if domain.DKIMModulusLength > 0 {
		table.AddRow([]string{"DKIM Modulus Length", FormatValue(domain.DKIMModulusLength)})
	}
	if domain.DKIMKeySelector != "" {
		table.AddRow([]string{"DKIM Key Selector", domain.DKIMKeySelector})
	}
	if domain.ReturnPath != "" {
		table.AddRow([]string{"Return Path", domain.ReturnPath})
	}

	// === Other Settings ===
	table.AddRow([]string{"Retention Days", FormatValue(domain.RetentionDays)})
	table.AddRow([]string{"Newsletter", FormatValue(domain.HasNewsletter)})
	if domain.IgnoreMXCheck {
		table.AddRow([]string{"Ignore MX Check", FormatValue(domain.IgnoreMXCheck)})
	}

	// === Membership ===
	table.AddRow([]string{"Members", FormatValue(len(domain.Members))})
	table.AddRow([]string{"Pending Invitations", FormatValue(len(domain.Invitations))})

	// === Timestamps ===
	table.AddRow([]string{"Created", domain.CreatedAt.Format(time.RFC3339)})
	table.AddRow([]string{"Updated", domain.UpdatedAt.Format(time.RFC3339)})

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
			TruncateString(member.User.ID, 8),
			member.User.Email,
			displayName,
			member.Group,
			member.JoinedAt.Format("2006-01-02"),
		}
		table.AddRow(row)
	}

	return table, nil
}
