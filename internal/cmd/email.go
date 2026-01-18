package cmd

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ginsys/forward-email/internal/client"
	"github.com/ginsys/forward-email/pkg/api"
	"github.com/ginsys/forward-email/pkg/output"
)

var (
	emailPage      int
	emailLimit     int
	emailSort      string
	emailOrder     string
	emailSearch    string
	emailStatus    string
	emailFrom      string
	emailTo        string
	emailDateFrom  string
	emailDateTo    string
	emailHasAttach string

	// Send flags
	emailFromAddr    string
	emailToAddrs     []string
	emailCCAddrs     []string
	emailBCCAddrs    []string
	emailSubject     string
	emailText        string
	emailHTML        string
	emailTextFile    string
	emailHTMLFile    string
	emailHeaders     []string
	emailAttachments []string
	emailInteractive bool
	emailDryRun      bool
)

// emailCmd represents the email command
var emailCmd = &cobra.Command{
	Use:   "email",
	Short: "Send and manage emails",
	Long: `Send emails and manage sent email history through the Forward Email API.
Supports both interactive and command-line email composition.`,
}

// emailSendCmd represents the email send command
var emailSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send an email",
	Long: `Send an email through Forward Email. Can be used interactively or with flags.
If no flags are provided, interactive mode will be used.`,
	RunE: runEmailSend,
}

// emailListCmd represents the email list command
var emailListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sent emails",
	Long:  `List emails that have been sent through your Forward Email account.`,
	RunE:  runEmailList,
}

// emailGetCmd represents the email get command
var emailGetCmd = &cobra.Command{
	Use:   "get <email-id>",
	Short: "Get email details",
	Long:  `Get detailed information about a specific sent email.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEmailGet,
}

// emailDeleteCmd represents the email delete command
var emailDeleteCmd = &cobra.Command{
	Use:   "delete <email-id>",
	Short: "Delete an email",
	Long:  `Delete an email from your sent email history.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEmailDelete,
}

// emailQuotaCmd represents the email quota command
var emailQuotaCmd = &cobra.Command{
	Use:   "quota",
	Short: "Show email quota",
	Long:  `Show your daily email sending quota and usage.`,
	RunE:  runEmailQuota,
}

func init() {
	// Register with root command
	rootCmd.AddCommand(emailCmd)

	// Add subcommands
	emailCmd.AddCommand(emailSendCmd)
	emailCmd.AddCommand(emailListCmd)
	emailCmd.AddCommand(emailGetCmd)
	emailCmd.AddCommand(emailDeleteCmd)
	emailCmd.AddCommand(emailQuotaCmd)

	// Note: output flag now inherited from global root command

	// List command flags
	emailListCmd.Flags().IntVar(&emailPage, "page", 1, "Page number")
	emailListCmd.Flags().IntVar(&emailLimit, "limit", 25, "Number of emails per page")
	emailListCmd.Flags().StringVar(&emailSort, "sort", "sent_at", "Sort by (sent_at, subject, from, to)")
	emailListCmd.Flags().StringVar(&emailOrder, "order", "desc", "Sort order (asc, desc)")
	emailListCmd.Flags().StringVar(&emailSearch, "search", "", "Search in subject, from, to")
	emailListCmd.Flags().StringVar(&emailStatus, "status", "", "Filter by status (sent, delivered, bounced, failed)")
	emailListCmd.Flags().StringVar(&emailFrom, "from", "", "Filter by sender")
	emailListCmd.Flags().StringVar(&emailTo, "to", "", "Filter by recipient")
	emailListCmd.Flags().StringVar(&emailDateFrom, "date-from", "", "Filter by date from (YYYY-MM-DD)")
	emailListCmd.Flags().StringVar(&emailDateTo, "date-to", "", "Filter by date to (YYYY-MM-DD)")
	emailListCmd.Flags().StringVar(&emailHasAttach, "has-attach", "", "Filter by attachment presence (true/false)")

	// Send command flags
	emailSendCmd.Flags().BoolVarP(&emailInteractive, "interactive", "i", false, "Use interactive mode")
	emailSendCmd.Flags().StringVar(&emailFromAddr, "from", "", "Sender email address")
	emailSendCmd.Flags().StringSliceVar(&emailToAddrs, "to", nil, "Recipient email addresses")
	emailSendCmd.Flags().StringSliceVar(&emailCCAddrs, "cc", nil, "CC email addresses")
	emailSendCmd.Flags().StringSliceVar(&emailBCCAddrs, "bcc", nil, "BCC email addresses")
	emailSendCmd.Flags().StringVar(&emailSubject, "subject", "", "Email subject")
	emailSendCmd.Flags().StringVar(&emailText, "text", "", "Plain text content")
	emailSendCmd.Flags().StringVar(&emailHTML, "html", "", "HTML content")
	emailSendCmd.Flags().StringVar(&emailTextFile, "text-file", "", "File containing plain text content")
	emailSendCmd.Flags().StringVar(&emailHTMLFile, "html-file", "", "File containing HTML content")
	emailSendCmd.Flags().StringSliceVar(&emailHeaders, "header", nil, "Custom headers (format: 'Name: Value')")
	emailSendCmd.Flags().StringSliceVar(&emailAttachments, "attach", nil, "Attachment file paths")
	emailSendCmd.Flags().BoolVar(&emailDryRun, "dry-run", false, "Validate email without sending")
}

func runEmailSend(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	var req *api.SendEmailRequest

	// Check if we should use interactive mode
	if emailInteractive || (emailFromAddr == "" && len(emailToAddrs) == 0 && emailSubject == "") {
		req, err = promptForEmail()
		if err != nil {
			return fmt.Errorf("failed to get email input: %v", err)
		}
	} else {
		req, err = buildEmailFromFlags()
		if err != nil {
			return fmt.Errorf("failed to build email from flags: %v", err)
		}
	}

	// Validate the email
	if err2 := validateEmailRequest(req); err2 != nil {
		return fmt.Errorf("email validation failed: %v", err2)
	}

	// Show email preview
	fmt.Println("📧 Email Preview:")
	fmt.Printf("From: %s\n", req.From)
	fmt.Printf("To: %s\n", strings.Join(req.To, ", "))
	if len(req.CC) > 0 {
		fmt.Printf("CC: %s\n", strings.Join(req.CC, ", "))
	}
	if len(req.BCC) > 0 {
		fmt.Printf("BCC: %s\n", strings.Join(req.BCC, ", "))
	}
	fmt.Printf("Subject: %s\n", req.Subject)
	if len(req.Attachments) > 0 {
		fmt.Printf("Attachments: %d files\n", len(req.Attachments))
	}
	fmt.Println()

	if emailDryRun {
		fmt.Println("✅ Email validation successful (dry run mode)")
		return nil
	}

	// Confirm before sending
	fmt.Print("Send this email? [y/N]: ")
	reader := bufio.NewReader(cmd.InOrStdin())
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != yesStr {
		fmt.Println("❌ Email sending canceled")
		return nil
	}

	// Send the email
	result, err := apiClient.Emails.SendEmail(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	cmd.Printf("✅ Email sent successfully!\n")
	fmt.Printf("Email ID: %s\n", result.ID)
	fmt.Printf("Message ID: %s\n", result.MessageID)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Sent at: %s\n", result.SentAt.Format(time.RFC3339))

	return nil
}

func runEmailList(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	// Parse boolean flag
	var hasAttach *bool
	if emailHasAttach != "" {
		if val, parseErr := strconv.ParseBool(emailHasAttach); parseErr == nil {
			hasAttach = &val
		}
	}

	opts := &api.ListEmailsOptions{
		Page:      emailPage,
		Limit:     emailLimit,
		Sort:      emailSort,
		Order:     emailOrder,
		Search:    emailSearch,
		Status:    emailStatus,
		From:      emailFrom,
		To:        emailTo,
		DateFrom:  emailDateFrom,
		DateTo:    emailDateTo,
		HasAttach: hasAttach,
	}

	response, err := apiClient.Emails.ListEmails(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list emails: %v", err)
	}

	format, err := output.ParseFormat(viper.GetString("output"))
	if err != nil {
		return fmt.Errorf("invalid output format: %v", err)
	}

	if format == output.FormatJSON || format == output.FormatYAML {
		formatter := output.NewFormatter(format, cmd.OutOrStdout())
		return formatter.Format(response.Emails)
	}

	// Format as table
	tableData, err := output.FormatEmailList(response.Emails, format)
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	err = formatter.Format(tableData)
	if err != nil {
		return err
	}

	// Show pagination info for non-JSON/YAML formats
	if len(response.Emails) > 0 {
		fmt.Printf("\nShowing %d of %d emails (page %d of %d)\n",
			len(response.Emails), response.TotalCount, response.Page, response.TotalPages)
		if response.Page < response.TotalPages {
			fmt.Printf("Use --page %d to see more results\n", response.Page+1)
		}
	} else {
		fmt.Println("No emails found")
	}

	return nil
}

func runEmailGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	emailID := args[0]

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	email, err := apiClient.Emails.GetEmail(ctx, emailID)
	if err != nil {
		return fmt.Errorf("failed to get email: %v", err)
	}

	format, err := output.ParseFormat(viper.GetString("output"))
	if err != nil {
		return fmt.Errorf("invalid output format: %v", err)
	}

	if format == output.FormatJSON || format == output.FormatYAML {
		formatter := output.NewFormatter(format, cmd.OutOrStdout())
		return formatter.Format(email)
	}

	// Format as table
	tableData, err := output.FormatEmailDetails(email, format)
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	return formatter.Format(tableData)
}

func runEmailDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	emailID := args[0]

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	// Get email info first for confirmation
	email, err := apiClient.Emails.GetEmail(ctx, emailID)
	if err != nil {
		return fmt.Errorf("failed to get email: %v", err)
	}

	fmt.Printf("⚠️  Are you sure you want to delete email '%s'?\n", email.Subject)
	fmt.Printf("Sent to: %s\n", strings.Join(email.To, ", "))
	fmt.Printf("Sent at: %s\n", email.SentAt.Format(time.RFC3339))
	fmt.Print("Type 'yes' to confirm: ")
	reader := bufio.NewReader(cmd.InOrStdin())
	line, _ := reader.ReadString('\n')
	confirmation := strings.TrimSpace(line)

	if !strings.EqualFold(confirmation, "yes") {
		fmt.Println("❌ Deletion canceled")
		return nil
	}

	err = apiClient.Emails.DeleteEmail(ctx, emailID)
	if err != nil {
		return fmt.Errorf("failed to delete email: %v", err)
	}

	cmd.Printf("✅ Email '%s' deleted successfully\n", email.Subject)
	return nil
}

func runEmailQuota(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	quota, err := apiClient.Emails.GetEmailQuota(ctx)
	if err != nil {
		return fmt.Errorf("failed to get email quota: %v", err)
	}

	format, err := output.ParseFormat(viper.GetString("output"))
	if err != nil {
		return fmt.Errorf("invalid output format: %v", err)
	}

	if format == output.FormatJSON || format == output.FormatYAML {
		formatter := output.NewFormatter(format, cmd.OutOrStdout())
		return formatter.Format(quota)
	}

	// Format as table
	tableData, err := output.FormatEmailQuota(quota, format)
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	return formatter.Format(tableData)
}

//nolint:unparam // returns error for future use; currently always nil
func promptForEmail() (*api.SendEmailRequest, error) {
	reader := bufio.NewReader(os.Stdin)
	req := &api.SendEmailRequest{}

	fmt.Println("📧 Interactive Email Composer")
	fmt.Println()

	// From
	fmt.Print("From: ")
	from, _ := reader.ReadString('\n')
	req.From = strings.TrimSpace(from)

	// To
	fmt.Print("To (comma-separated): ")
	to, _ := reader.ReadString('\n')
	req.To = strings.Split(strings.TrimSpace(to), ",")
	for i, addr := range req.To {
		req.To[i] = strings.TrimSpace(addr)
	}

	// CC (optional)
	fmt.Print("CC (comma-separated, optional): ")
	cc, _ := reader.ReadString('\n')
	ccTrimmed := strings.TrimSpace(cc)
	if ccTrimmed != "" {
		req.CC = strings.Split(ccTrimmed, ",")
		for i, addr := range req.CC {
			req.CC[i] = strings.TrimSpace(addr)
		}
	}

	// Subject
	fmt.Print("Subject: ")
	subject, _ := reader.ReadString('\n')
	req.Subject = strings.TrimSpace(subject)

	// Content
	fmt.Println("Content (enter 'END' on a new line to finish):")
	var contentLines []string
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSuffix(line, "\n")
		if line == "END" {
			break
		}
		contentLines = append(contentLines, line)
	}
	req.Text = strings.Join(contentLines, "\n")

	return req, nil
}

func buildEmailFromFlags() (*api.SendEmailRequest, error) {
	req := &api.SendEmailRequest{
		From:    emailFromAddr,
		To:      emailToAddrs,
		CC:      emailCCAddrs,
		BCC:     emailBCCAddrs,
		Subject: emailSubject,
		Text:    emailText,
		HTML:    emailHTML,
	}

	// Read content from files if specified
	if emailTextFile != "" {
		content, err := os.ReadFile(emailTextFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read text file: %v", err)
		}
		req.Text = string(content)
	}

	if emailHTMLFile != "" {
		content, err := os.ReadFile(emailHTMLFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read HTML file: %v", err)
		}
		req.HTML = string(content)
	}

	// Parse custom headers
	if len(emailHeaders) > 0 {
		req.Headers = make(map[string]string)
		for _, header := range emailHeaders {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid header format: %s (expected 'Name: Value')", header)
			}
			req.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Process attachments
	if len(emailAttachments) > 0 {
		for _, attachPath := range emailAttachments {
			attachment, err := processAttachment(attachPath)
			if err != nil {
				return nil, fmt.Errorf("failed to process attachment %s: %v", attachPath, err)
			}
			req.Attachments = append(req.Attachments, *attachment)
		}
	}

	return req, nil
}

func processAttachment(filePath string) (*api.AttachmentData, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Encode content as base64
	encoded := base64.StdEncoding.EncodeToString(content)

	attachment := &api.AttachmentData{
		Filename: filepath.Base(filePath),
		Content:  []byte(encoded),
	}

	// Try to detect content type based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".txt":
		attachment.ContentType = "text/plain"
	case ".pdf":
		attachment.ContentType = "application/pdf"
	case ".jpg", ".jpeg":
		attachment.ContentType = "image/jpeg"
	case ".png":
		attachment.ContentType = "image/png"
	case ".gif":
		attachment.ContentType = "image/gif"
	case ".zip":
		attachment.ContentType = "application/zip"
	default:
		attachment.ContentType = "application/octet-stream"
	}

	return attachment, nil
}

func validateEmailRequest(req *api.SendEmailRequest) error {
	if req.From == "" {
		return fmt.Errorf("from address is required")
	}
	if len(req.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	if req.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if req.Text == "" && req.HTML == "" {
		return fmt.Errorf("either text or HTML content is required")
	}

	// Validate email addresses using RFC-compliant parsing
	allAddrs := req.To
	allAddrs = append(allAddrs, req.CC...)
	allAddrs = append(allAddrs, req.BCC...)
	allAddrs = append(allAddrs, req.From)

	for _, addr := range allAddrs {
		if addr != "" {
			if _, err := mail.ParseAddress(addr); err != nil {
				return fmt.Errorf("invalid email address: %s - %w", addr, err)
			}
		}
	}

	return nil
}
