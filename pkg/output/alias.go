package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/ginsys/forward-email/pkg/api"
)

// FormatAliasList formats a list of aliases for display
func FormatAliasList(aliases []api.Alias, format Format, domain string) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		// For JSON/YAML, return the aliases directly
		return nil, fmt.Errorf("use direct JSON/YAML encoding for aliases")
	}

	headers := []string{"NAME", "DOMAIN", "RECIPIENTS", "ENABLED", "IMAP", "LABELS", "CREATED"}
	table := NewTableData(headers)

	for _, alias := range aliases {
		enabled := FormatValue(alias.IsEnabled)
		imap := FormatValue(alias.HasIMAP)

		var recipients, labels string

		// For CSV, show full data without truncation
		if format == FormatCSV {
			recipients = strings.Join(alias.Recipients, ", ")
			labels = strings.Join(alias.Labels, ", ")
		} else {
			// For table, use intelligent text wrapping
			recipients = strings.Join(alias.Recipients, ", ")
			// Don't truncate - let table wrapper handle long content
			labels = strings.Join(alias.Labels, ", ")
			// Don't truncate - let table wrapper handle long content
		}

		var created string
		if alias.CreatedAt.IsZero() {
			created = "-"
		} else {
			created = alias.CreatedAt.Format("2006-01-02")
		}

		row := []string{
			alias.Name,
			domain, // Use the domain name provided to the list command
			recipients,
			enabled,
			imap,
			labels,
			created,
		}
		table.AddRow(row)
	}

	return table, nil
}

// FormatAliasDetails formats detailed alias information
func FormatAliasDetails(alias *api.Alias, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for alias details")
	}

	headers := []string{"PROPERTY", "VALUE"}
	table := NewTableData(headers)

	// Basic information
	table.AddRow([]string{"ID", alias.ID})
	table.AddRow([]string{"Name", alias.Name})
	table.AddRow([]string{"Domain ID", alias.DomainID})
	table.AddRow([]string{"Enabled", FormatValue(alias.IsEnabled)})
	table.AddRow([]string{"Created", alias.CreatedAt.Format(time.RFC3339)})
	table.AddRow([]string{"Updated", alias.UpdatedAt.Format(time.RFC3339)})

	// Recipients
	if len(alias.Recipients) > 0 {
		table.AddRow([]string{"Recipients", strings.Join(alias.Recipients, ", ")})
	}

	// Labels
	if len(alias.Labels) > 0 {
		table.AddRow([]string{"Labels", strings.Join(alias.Labels, ", ")})
	}

	// Description
	if alias.Description != "" {
		table.AddRow([]string{"Description", alias.Description})
	}

	// IMAP/PGP settings
	table.AddRow([]string{"IMAP Enabled", FormatValue(alias.HasIMAP)})
	table.AddRow([]string{"PGP Enabled", FormatValue(alias.HasPGP)})
	table.AddRow([]string{"Has Password", FormatValue(alias.HasPassword)})

	if alias.PublicKey != "" {
		key := alias.PublicKey
		if len(key) > 100 {
			key = TruncateString(key, 97) + "..."
		}
		table.AddRow([]string{"PGP Public Key", key})
	}

	// Quota information
	if alias.Quota != nil {
		table.AddRow([]string{"Storage Used", FormatBytes(alias.Quota.StorageUsed)})
		table.AddRow([]string{"Storage Limit", FormatBytes(alias.Quota.StorageLimit)})
		table.AddRow([]string{"Emails Sent Today", fmt.Sprintf("%d/%d", alias.Quota.EmailsSent, alias.Quota.EmailsLimit)})
	}

	// Vacation responder
	if alias.Vacation != nil && alias.Vacation.IsEnabled {
		table.AddRow([]string{"Vacation Enabled", "Yes"})
		if alias.Vacation.Subject != "" {
			table.AddRow([]string{"Vacation Subject", alias.Vacation.Subject})
		}
		if !alias.Vacation.StartDate.IsZero() {
			table.AddRow([]string{"Vacation Start", alias.Vacation.StartDate.Format("2006-01-02")})
		}
		if !alias.Vacation.EndDate.IsZero() {
			table.AddRow([]string{"Vacation End", alias.Vacation.EndDate.Format("2006-01-02")})
		}
	}

	return table, nil
}

// FormatAliasQuota formats alias quota information
func FormatAliasQuota(quota *api.AliasQuota, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for alias quota")
	}

	headers := []string{"METRIC", "USED", "LIMIT", "PERCENTAGE"}
	table := NewTableData(headers)

	// Storage quota
	storageUsed := FormatBytes(quota.StorageUsed)
	storageLimit := FormatBytes(quota.StorageLimit)
	storagePerc := FormatPercentage(quota.StorageUsed, quota.StorageLimit)
	table.AddRow([]string{"Storage", storageUsed, storageLimit, storagePerc})

	// Email quota
	emailsUsed := fmt.Sprintf("%d", quota.EmailsSent)
	emailsLimit := fmt.Sprintf("%d", quota.EmailsLimit)
	emailPerc := FormatPercentage(int64(quota.EmailsSent), int64(quota.EmailsLimit))
	table.AddRow([]string{"Emails (Daily)", emailsUsed, emailsLimit, emailPerc})

	return table, nil
}

// FormatAliasStats formats alias usage statistics
func FormatAliasStats(stats *api.AliasStats, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for alias stats")
	}

	headers := []string{"STATISTIC", "VALUE"}
	table := NewTableData(headers)

	table.AddRow([]string{"Emails Received", fmt.Sprintf("%d", stats.EmailsReceived)})
	table.AddRow([]string{"Emails Sent", fmt.Sprintf("%d", stats.EmailsSent)})
	table.AddRow([]string{"Storage Used", FormatBytes(stats.StorageUsed)})

	if !stats.LastActivity.IsZero() {
		table.AddRow([]string{"Last Activity", stats.LastActivity.Format(time.RFC3339)})
	}

	if len(stats.RecentSenders) > 0 {
		senders := strings.Join(stats.RecentSenders, ", ")
		if len(senders) > 80 {
			senders = TruncateString(senders, 77) + "..."
		}
		table.AddRow([]string{"Recent Senders", senders})
	}

	return table, nil
}

// FormatAliasRecipients formats alias recipients for display
func FormatAliasRecipients(recipients []string, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for recipients")
	}

	headers := []string{"RECIPIENT", "TYPE"}
	table := NewTableData(headers)

	for _, recipient := range recipients {
		recipientType := "Email"
		if strings.Contains(recipient, "://") {
			recipientType = "Webhook"
		} else if !strings.Contains(recipient, "@") {
			recipientType = "FQDN/IP"
		}

		table.AddRow([]string{recipient, recipientType})
	}

	return table, nil
}
