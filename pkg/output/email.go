package output

import (
	"fmt"
	"strings"
	"time"

    "github.com/ginsys/forward-email/pkg/api"
)

// FormatEmailList formats a list of emails for display
func FormatEmailList(emails []api.Email, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		// For JSON/YAML, return the emails directly
		return nil, fmt.Errorf("use direct JSON/YAML encoding for emails")
	}

	headers := []string{"ID", "FROM", "TO", "SUBJECT", "STATUS", "ATTACHMENTS", "SENT"}
	table := NewTableData(headers)

	for _, email := range emails {
		var id, from, to, subject string

		// For CSV, show full data without truncation
		if format == FormatCSV {
			id = email.ID
			from = email.From
			to = strings.Join(email.To, ", ")
			subject = email.Subject
		} else {
			// For table, truncate for readability
			id = TruncateString(email.ID, 8)
			from = TruncateString(email.From, 25)
			to = strings.Join(email.To, ", ")
			if len(to) > 30 {
				to = TruncateString(to, 27) + "..."
			}
			subject = TruncateString(email.Subject, 40)
		}

		status := FormatEmailStatus(email.Status)
		attachments := fmt.Sprintf("%d", len(email.Attachments))
		if len(email.Attachments) == 0 {
			attachments = "-"
		}

		// Check if SentAt is zero time and format accordingly
		var sent string
		if email.SentAt.IsZero() {
			sent = "-"
		} else {
			sent = email.SentAt.Format("2006-01-02 15:04")
		}

		row := []string{
			id,
			from,
			to,
			subject,
			status,
			attachments,
			sent,
		}
		table.AddRow(row)
	}

	return table, nil
}

// FormatEmailDetails formats detailed email information
func FormatEmailDetails(email *api.Email, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for email details")
	}

	headers := []string{"PROPERTY", "VALUE"}
	table := NewTableData(headers)

	// Basic information
	table.AddRow([]string{"ID", email.ID})
	table.AddRow([]string{"Message ID", email.MessageID})
	table.AddRow([]string{"From", email.From})
	table.AddRow([]string{"To", strings.Join(email.To, ", ")})

	if len(email.CC) > 0 {
		table.AddRow([]string{"CC", strings.Join(email.CC, ", ")})
	}
	if len(email.BCC) > 0 {
		table.AddRow([]string{"BCC", strings.Join(email.BCC, ", ")})
	}

	table.AddRow([]string{"Subject", email.Subject})
	table.AddRow([]string{"Status", FormatEmailStatus(email.Status)})

	if email.StatusInfo != "" {
		table.AddRow([]string{"Status Info", email.StatusInfo})
	}

	table.AddRow([]string{"Sent At", email.SentAt.Format(time.RFC3339)})
	if email.DeliveredAt != nil {
		table.AddRow([]string{"Delivered At", email.DeliveredAt.Format(time.RFC3339)})
	}
	table.AddRow([]string{"Created At", email.CreatedAt.Format(time.RFC3339)})

	// Content preview
	if email.Text != "" {
		preview := email.Text
		if len(preview) > 200 {
			preview = TruncateString(preview, 197) + "..."
		}
		table.AddRow([]string{"Text Preview", preview})
	}
	if email.HTML != "" {
		preview := email.HTML
		if len(preview) > 200 {
			preview = TruncateString(preview, 197) + "..."
		}
		table.AddRow([]string{"HTML Preview", preview})
	}

	// Attachments
	if len(email.Attachments) > 0 {
		table.AddRow([]string{"Attachments", fmt.Sprintf("%d files", len(email.Attachments))})
		for i, attachment := range email.Attachments {
			if i < 5 { // Show first 5 attachments
				size := FormatBytes(attachment.Size)
				table.AddRow([]string{
					fmt.Sprintf("  Attachment %d", i+1),
					fmt.Sprintf("%s (%s, %s)", attachment.Filename, attachment.ContentType, size),
				})
			} else if i == 5 {
				table.AddRow([]string{"  ...", fmt.Sprintf("and %d more", len(email.Attachments)-5)})
				break
			}
		}
	}

	// Custom headers
	if len(email.Headers) > 0 {
		table.AddRow([]string{"Custom Headers", fmt.Sprintf("%d headers", len(email.Headers))})
		count := 0
		for name, value := range email.Headers {
			if count < 3 { // Show first 3 headers
				table.AddRow([]string{fmt.Sprintf("  %s", name), value})
			} else if count == 3 {
				table.AddRow([]string{"  ...", fmt.Sprintf("and %d more", len(email.Headers)-3)})
				break
			}
			count++
		}
	}

	return table, nil
}

// FormatEmailQuota formats email quota information
func FormatEmailQuota(quota *api.EmailQuota, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for email quota")
	}

	headers := []string{"METRIC", "USED", "LIMIT", "PERCENTAGE", "RESET TIME"}
	table := NewTableData(headers)

	// Daily email quota
	emailsUsed := fmt.Sprintf("%d", quota.EmailsSent)
	emailsLimit := fmt.Sprintf("%d", quota.EmailsLimit)
	emailPerc := FormatPercentage(int64(quota.EmailsSent), int64(quota.EmailsLimit))
	resetTime := quota.ResetTime.Format("15:04 MST")

	table.AddRow([]string{"Daily Emails", emailsUsed, emailsLimit, emailPerc, resetTime})

	// Overage information if applicable
	if quota.OverageAllowed {
		overageUsed := fmt.Sprintf("%d", quota.OverageUsed)
		table.AddRow([]string{"Overage Used", overageUsed, "Unlimited", "-", "-"})
	}

	return table, nil
}

// FormatEmailStats formats email usage statistics
func FormatEmailStats(stats *api.EmailStats, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for email stats")
	}

	headers := []string{"STATISTIC", "COUNT", "PERCENTAGE"}
	table := NewTableData(headers)

	total := stats.TotalSent
	if total == 0 {
		total = 1 // Avoid division by zero
	}

	table.AddRow([]string{"Total Sent", fmt.Sprintf("%d", stats.TotalSent), "100%"})
	table.AddRow([]string{"Delivered", fmt.Sprintf("%d", stats.TotalDelivered), fmt.Sprintf("%.1f%%", stats.DeliveryRate)})
	table.AddRow([]string{"Bounced", fmt.Sprintf("%d", stats.TotalBounced), fmt.Sprintf("%.1f%%", stats.BounceRate)})
	table.AddRow([]string{"Failed", fmt.Sprintf("%d", stats.TotalFailed), fmt.Sprintf("%.1f%%", float64(stats.TotalFailed)/float64(total)*100)})

	if !stats.LastSent.IsZero() {
		table.AddRow([]string{"Last Sent", stats.LastSent.Format(time.RFC3339), "-"})
	}

	// Top recipients
	if len(stats.TopRecipients) > 0 {
		recipients := strings.Join(stats.TopRecipients, ", ")
		if len(recipients) > 80 {
			recipients = TruncateString(recipients, 77) + "..."
		}
		table.AddRow([]string{"Top Recipients", recipients, "-"})
	}

	return table, nil
}

// FormatEmailAttachments formats email attachments for display
func FormatEmailAttachments(attachments []api.EmailAttachment, format Format) (*TableData, error) {
	if format != FormatTable && format != FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for attachments")
	}

	headers := []string{"FILENAME", "TYPE", "SIZE", "CID"}
	table := NewTableData(headers)

	for _, attachment := range attachments {
		size := FormatBytes(attachment.Size)
		cid := attachment.CID
		if cid == "" {
			cid = "-"
		}

		table.AddRow([]string{
			attachment.Filename,
			attachment.ContentType,
			size,
			cid,
		})
	}

	return table, nil
}

// FormatEmailStatus formats email status with appropriate styling
func FormatEmailStatus(status string) string {
	switch strings.ToLower(status) {
	case "sent":
		return "✓ Sent"
	case "delivered":
		return "✅ Delivered"
	case "bounced":
		return "↩️ Bounced"
	case "failed":
		return "❌ Failed"
	case "queued":
		return "⏳ Queued"
	default:
		return status
	}
}
