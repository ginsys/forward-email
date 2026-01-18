package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ginsys/forward-email/internal/client"
	"github.com/ginsys/forward-email/pkg/api"
	"github.com/ginsys/forward-email/pkg/output"
)

// Global variables for domain command flags.
// These store the parsed command-line arguments for domain list filtering and pagination.
var (
	domainPage     int    // Page number for pagination (default: 1)
	domainLimit    int    // Number of results per page (default: 25)
	domainSort     string // Sort field: name, created_at, updated_at, is_verified, plan
	domainOrder    string // Sort order: asc or desc
	domainSearch   string // Search filter for domain names
	domainVerified string // Verification status filter: "true" or "false"
	domainPlan     string // Plan filter: free, enhanced_protection, team
)

// domainCmd represents the domain command
var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Manage Forward Email domains",
	Long: `Manage Forward Email domains including creating, listing, updating, 
and configuring domain settings and DNS records.`,
}

// domainListCmd represents the domain list command
var domainListCmd = &cobra.Command{
	Use:   "list",
	Short: "List domains",
	Long:  `List all domains associated with your Forward Email account.`,
	RunE:  runDomainList,
}

// domainGetCmd represents the domain get command
var domainGetCmd = &cobra.Command{
	Use:   "get <domain-name-or-id>",
	Short: "Get domain details",
	Long:  `Get detailed information about a specific domain.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDomainGet,
}

// domainCreateCmd represents the domain create command
var domainCreateCmd = &cobra.Command{
	Use:   "create <domain-name>",
	Short: "Create a new domain",
	Long:  `Create a new domain in your Forward Email account.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDomainCreate,
}

// domainUpdateCmd represents the domain update command
var domainUpdateCmd = &cobra.Command{
	Use:   "update <domain-name-or-id>",
	Short: "Update domain settings",
	Long: `Update settings for an existing domain.

Note: Some fields shown in 'domain get' are read-only and cannot be updated:
  - plan (view only, use separate plan management)
  - DKIM settings (managed automatically by Forward Email)
  - return_path (configured automatically)
  - created_at, updated_at (system timestamps)
  - id, name (immutable identifiers)`,
	Args: cobra.ExactArgs(1),
	RunE: runDomainUpdate,
}

// domainDeleteCmd represents the domain delete command
var domainDeleteCmd = &cobra.Command{
	Use:   "delete <domain-name-or-id>",
	Short: "Delete a domain",
	Long:  `Delete a domain from your Forward Email account.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDomainDelete,
}

// domainVerifyCmd represents the domain verify command
var domainVerifyCmd = &cobra.Command{
	Use:   "verify <domain-name-or-id>",
	Short: "Verify domain DNS configuration",
	Long: `Verify that the DNS records for a domain are correctly configured.

By default, verifies DNS records. Use --smtp to verify SMTP outbound configuration.`,
	Args: cobra.ExactArgs(1),
	RunE: runDomainVerify,
}

// domainDNSCmd represents the domain dns command
var domainDNSCmd = &cobra.Command{
	Use:   "dns <domain-name-or-id>",
	Short: "Show required DNS records",
	Long:  `Show the DNS records required for domain configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDomainDNS,
}

// domainQuotaCmd represents the domain quota command
var domainQuotaCmd = &cobra.Command{
	Use:   "quota <domain-name-or-id>",
	Short: "Show domain quota information",
	Long:  `Show quota usage and limits for a domain.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDomainQuota,
}

// domainStatsCmd represents the domain stats command
var domainStatsCmd = &cobra.Command{
	Use:   "stats <domain-name-or-id>",
	Short: "Show domain statistics",
	Long:  `Show statistics and metrics for a domain.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDomainStats,
}

// domainMembersCmd represents the domain members command group
var domainMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "Manage domain members",
	Long:  `Manage members of a domain.`,
}

// domainMembersListCmd represents the domain members list command
var domainMembersListCmd = &cobra.Command{
	Use:   "list <domain-name-or-id>",
	Short: "List domain members",
	Long:  `List all members of a domain.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDomainMembersList,
}

// domainMembersAddCmd represents the domain members add command
var domainMembersAddCmd = &cobra.Command{
	Use:   "add <domain-name-or-id> <email> [group]",
	Short: "Add domain member",
	Long:  `Add a new member to a domain.`,
	Args:  cobra.RangeArgs(2, 3),
	RunE:  runDomainMembersAdd,
}

// domainMembersRemoveCmd represents the domain members remove command
var domainMembersRemoveCmd = &cobra.Command{
	Use:   "remove <domain-name-or-id> <member-id>",
	Short: "Remove domain member",
	Long:  `Remove a member from a domain.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runDomainMembersRemove,
}

func init() {
	rootCmd.AddCommand(domainCmd)

	// Add subcommands
	domainCmd.AddCommand(domainListCmd)
	domainCmd.AddCommand(domainGetCmd)
	domainCmd.AddCommand(domainCreateCmd)
	domainCmd.AddCommand(domainUpdateCmd)
	domainCmd.AddCommand(domainDeleteCmd)
	domainCmd.AddCommand(domainVerifyCmd)
	domainCmd.AddCommand(domainDNSCmd)
	domainCmd.AddCommand(domainQuotaCmd)
	domainCmd.AddCommand(domainStatsCmd)
	domainCmd.AddCommand(domainMembersCmd)

	// Add members subcommands
	domainMembersCmd.AddCommand(domainMembersListCmd)
	domainMembersCmd.AddCommand(domainMembersAddCmd)
	domainMembersCmd.AddCommand(domainMembersRemoveCmd)

	// Note: output flag now inherited from global root command

	// List command specific flags
	domainListCmd.Flags().IntVar(&domainPage, "page", 1, "Page number")
	domainListCmd.Flags().IntVar(&domainLimit, "limit", 25, "Number of results per page")
	domainListCmd.Flags().StringVar(&domainSort, "sort", "name",
		"Sort field (name, created_at, updated_at, is_verified, plan)")
	domainListCmd.Flags().StringVar(&domainOrder, "order", "asc", "Sort order (asc, desc)")
	domainListCmd.Flags().StringVar(&domainSearch, "search", "", "Search domains by name")
	domainListCmd.Flags().StringVar(&domainVerified, "verified", "", "Filter by verification status (true, false)")
	domainListCmd.Flags().StringVar(&domainPlan, "plan", "", "Filter by plan (free, enhanced_protection, team)")

	// Create command flags
	domainCreateCmd.Flags().String("plan", "", "Domain plan (free, enhanced_protection, team)")

	// Update command flags
	domainUpdateCmd.Flags().Int("max-forwarded-addresses", 0, "Maximum forwarded addresses")
	domainUpdateCmd.Flags().Int("retention-days", 0, "Email retention period in days")
	domainUpdateCmd.Flags().Bool("adult-content-protection", false, "Enable adult content protection")
	domainUpdateCmd.Flags().Bool("phishing-protection", false, "Enable phishing protection")
	domainUpdateCmd.Flags().Bool("executable-protection", false, "Enable executable protection")
	domainUpdateCmd.Flags().Bool("virus-protection", false, "Enable virus protection")
	domainUpdateCmd.Flags().Int("smtp-port", 0, "SMTP port")
	domainUpdateCmd.Flags().Int("imap-port", 0, "IMAP port")
	domainUpdateCmd.Flags().Int("caldav-port", 0, "CalDAV port")
	domainUpdateCmd.Flags().Int("carddav-port", 0, "CardDAV port")
	domainUpdateCmd.Flags().String("webhook-url", "", "Webhook URL")
	// New flags for additional settings
	domainUpdateCmd.Flags().Bool("delivery-logs", false, "Enable deliverability logs for successful emails")
	domainUpdateCmd.Flags().String("bounce-webhook", "", "URL for bounce notifications")
	domainUpdateCmd.Flags().Bool("regex", false, "Enable regex alias support")
	domainUpdateCmd.Flags().Bool("catchall", false, "Enable catch-all aliases")
	domainUpdateCmd.Flags().Bool("disable-catchall-regex", false, "Disable catch-all regex on large domains")
	domainUpdateCmd.Flags().Int("max-recipients", 0, "Max recipients per alias (0-1000)")
	domainUpdateCmd.Flags().Int64("max-quota", 0, "Max storage quota per alias in bytes")
	domainUpdateCmd.Flags().String("allowlist", "", "Comma-separated list of allowed forwarding destinations")
	domainUpdateCmd.Flags().String("denylist", "", "Comma-separated list of blocked addresses")
	domainUpdateCmd.Flags().Bool("recipient-verification", false, "Enable recipient verification emails")
	domainUpdateCmd.Flags().Bool("ignore-mx-check", false, "Bypass MX record validation")

	// Delete command flags
	domainDeleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")

	// Verify command flags
	domainVerifyCmd.Flags().Bool("smtp", false, "Verify SMTP configuration instead of DNS records")

	// Members add command flags
	domainMembersAddCmd.Flags().String("group", "user", "Member group (admin, user)")
}

// runDomainList implements the 'domain list' command.
// It retrieves domains from the API with filtering and pagination options,
// formats the output according to the user's preference (table/JSON/YAML/CSV),
// and displays pagination information for non-structured formats.
func runDomainList(_ *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	// Parse verified filter
	var verified *bool
	if domainVerified != "" {
		v, parseErr := strconv.ParseBool(domainVerified)
		if parseErr != nil {
			return fmt.Errorf("invalid verified filter: %s", domainVerified)
		}
		verified = &v
	}

	opts := &api.ListDomainsOptions{
		Page:     domainPage,
		Limit:    domainLimit,
		Sort:     domainSort,
		Order:    domainOrder,
		Search:   domainSearch,
		Verified: verified,
		Plan:     domainPlan,
	}

	response, err := apiClient.Domains.ListDomains(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list domains: %w", err)
	}

	// Handle output formatting
	outputFormat, err := output.ParseFormat(viper.GetString("output"))
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	formatter := output.NewFormatter(outputFormat, nil)

	if outputFormat == output.FormatJSON || outputFormat == output.FormatYAML {
		return formatter.Format(response.Domains)
	}

	// Format as table/CSV
	tableData, err := output.FormatDomainList(response.Domains, outputFormat)
	if err != nil {
		return err
	}

	err = formatter.Format(tableData)
	if err != nil {
		return err
	}

	// Show pagination info for non-JSON/YAML formats
	if len(response.Domains) > 0 {
		fmt.Printf("\nShowing %d of %d domains (page %d of %d)\n",
			len(response.Domains), response.Pagination.Total, response.Pagination.Page, response.Pagination.TotalPages)
		if response.Pagination.HasNext {
			fmt.Printf("Use --page %d to see more results\n", response.Pagination.Page+1)
		}
	} else {
		fmt.Println("No domains found")
	}

	return nil
}

// domainOperationRunner is a generic helper function that reduces code duplication
// across domain operations. It handles common patterns like API client creation,
// error handling, and output formatting. The function uses Go generics to work
// with different return types while maintaining type safety.
func domainOperationRunner[T any](
	args []string,
	operation func(context.Context, *api.DomainService, string) (T, error),
	errorMessage string,
	formatter func(T, output.Format) (interface{}, error),
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	result, err := operation(ctx, apiClient.Domains, args[0])
	if err != nil {
		return fmt.Errorf("%s: %w", errorMessage, err)
	}

	return formatOutput(result, viper.GetString("output"), func(format output.Format) (interface{}, error) {
		return formatter(result, format)
	})
}

// runDomainGet implements the 'domain get' command.
// It retrieves detailed information for a specific domain by ID or name
// using the generic domainOperationRunner helper for consistent error handling and output formatting.
func runDomainGet(_ *cobra.Command, args []string) error {
	return domainOperationRunner(
		args,
		func(ctx context.Context, domains *api.DomainService, domainID string) (*api.Domain, error) {
			return domains.GetDomain(ctx, domainID)
		},
		"failed to get domain",
		func(domain *api.Domain, format output.Format) (interface{}, error) {
			if format == output.FormatTable || format == output.FormatCSV {
				return output.FormatDomainDetails(domain, format)
			}
			return domain, nil
		},
	)
}

// runDomainCreate implements the 'domain create' command.
// It creates a new domain with the specified name and optional plan setting.
// The domain name is validated by the API, and the response includes initial
// DNS configuration requirements for domain verification.
func runDomainCreate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	planFlag, _ := cmd.Flags().GetString("plan")

	req := &api.CreateDomainRequest{
		Name: args[0],
		Plan: planFlag,
	}

	domain, err := apiClient.Domains.CreateDomain(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create domain: %w", err)
	}

	cmd.Printf("Domain '%s' created successfully\n", domain.Name)

	return formatOutput(domain, viper.GetString("output"), func(format output.Format) (interface{}, error) {
		if format == output.FormatTable || format == output.FormatCSV {
			return output.FormatDomainDetails(domain, format)
		}
		return domain, nil
	})
}

// runDomainUpdate implements the 'domain update' command.
// It builds an update request from the changed command flags and applies the updates
// to the specified domain. Only fields that were explicitly set via flags are updated,
// allowing for partial updates without affecting other domain settings.
func runDomainUpdate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	req := &api.UpdateDomainRequest{}

	// Parse flags and build update request
	if cmd.Flags().Changed("max-forwarded-addresses") {
		maxAddresses, _ := cmd.Flags().GetInt("max-forwarded-addresses")
		req.MaxForwardedAddresses = &maxAddresses
	}

	if cmd.Flags().Changed("retention-days") {
		retentionDays, _ := cmd.Flags().GetInt("retention-days")
		req.RetentionDays = &retentionDays
	}

	// Build settings if any protection flags are set
	if cmd.Flags().Changed("adult-content-protection") ||
		cmd.Flags().Changed("phishing-protection") ||
		cmd.Flags().Changed("executable-protection") ||
		cmd.Flags().Changed("virus-protection") ||
		cmd.Flags().Changed("smtp-port") ||
		cmd.Flags().Changed("imap-port") ||
		cmd.Flags().Changed("caldav-port") ||
		cmd.Flags().Changed("carddav-port") ||
		cmd.Flags().Changed("webhook-url") {
		settings := &api.DomainSettings{}

		if cmd.Flags().Changed("adult-content-protection") {
			settings.HasAdultContentProtection, _ = cmd.Flags().GetBool("adult-content-protection")
		}
		if cmd.Flags().Changed("phishing-protection") {
			settings.HasPhishingProtection, _ = cmd.Flags().GetBool("phishing-protection")
		}
		if cmd.Flags().Changed("executable-protection") {
			settings.HasExecutableProtection, _ = cmd.Flags().GetBool("executable-protection")
		}
		if cmd.Flags().Changed("virus-protection") {
			settings.HasVirusProtection, _ = cmd.Flags().GetBool("virus-protection")
		}
		if cmd.Flags().Changed("smtp-port") {
			settings.SMTPPort, _ = cmd.Flags().GetInt("smtp-port")
		}
		if cmd.Flags().Changed("imap-port") {
			settings.IMAPPort, _ = cmd.Flags().GetInt("imap-port")
		}
		if cmd.Flags().Changed("caldav-port") {
			settings.CalDAVPort, _ = cmd.Flags().GetInt("caldav-port")
		}
		if cmd.Flags().Changed("carddav-port") {
			settings.CardDAVPort, _ = cmd.Flags().GetInt("carddav-port")
		}
		if cmd.Flags().Changed("webhook-url") {
			settings.WebhookURL, _ = cmd.Flags().GetString("webhook-url")
		}

		req.Settings = settings
	}

	// Handle new flags
	if cmd.Flags().Changed("delivery-logs") {
		deliveryLogs, _ := cmd.Flags().GetBool("delivery-logs")
		req.HasDeliveryLogs = &deliveryLogs
	}

	if cmd.Flags().Changed("bounce-webhook") {
		bounceWebhook, _ := cmd.Flags().GetString("bounce-webhook")
		req.BounceWebhook = &bounceWebhook
	}

	if cmd.Flags().Changed("regex") {
		hasRegex, _ := cmd.Flags().GetBool("regex")
		req.HasRegex = &hasRegex
	}

	if cmd.Flags().Changed("catchall") {
		hasCatchall, _ := cmd.Flags().GetBool("catchall")
		req.HasCatchall = &hasCatchall
	}

	if cmd.Flags().Changed("disable-catchall-regex") {
		disableCatchallRegex, _ := cmd.Flags().GetBool("disable-catchall-regex")
		req.IsCatchallRegexDisabled = &disableCatchallRegex
	}

	if cmd.Flags().Changed("max-recipients") {
		maxRecipients, _ := cmd.Flags().GetInt("max-recipients")
		req.MaxRecipientsPerAlias = &maxRecipients
	}

	if cmd.Flags().Changed("max-quota") {
		maxQuota, _ := cmd.Flags().GetInt64("max-quota")
		req.MaxQuotaPerAlias = &maxQuota
	}

	if cmd.Flags().Changed("allowlist") {
		allowlistStr, _ := cmd.Flags().GetString("allowlist")
		if allowlistStr != "" {
			req.Allowlist = strings.Split(allowlistStr, ",")
			// Trim whitespace from each entry
			for i := range req.Allowlist {
				req.Allowlist[i] = strings.TrimSpace(req.Allowlist[i])
			}
		}
	}

	if cmd.Flags().Changed("denylist") {
		denylistStr, _ := cmd.Flags().GetString("denylist")
		if denylistStr != "" {
			req.Denylist = strings.Split(denylistStr, ",")
			// Trim whitespace from each entry
			for i := range req.Denylist {
				req.Denylist[i] = strings.TrimSpace(req.Denylist[i])
			}
		}
	}

	if cmd.Flags().Changed("recipient-verification") {
		recipientVerification, _ := cmd.Flags().GetBool("recipient-verification")
		req.HasRecipientVerification = &recipientVerification
	}

	if cmd.Flags().Changed("ignore-mx-check") {
		ignoreMXCheck, _ := cmd.Flags().GetBool("ignore-mx-check")
		req.IgnoreMXCheck = &ignoreMXCheck
	}

	domain, err := apiClient.Domains.UpdateDomain(ctx, args[0], req)
	if err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}

	cmd.Printf("Domain '%s' updated successfully\n", domain.Name)

	return formatOutput(domain, viper.GetString("output"), func(format output.Format) (interface{}, error) {
		if format == output.FormatTable || format == output.FormatCSV {
			return output.FormatDomainDetails(domain, format)
		}
		return domain, nil
	})
}

func runDomainDelete(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	if !force {
		cmd.Printf("Are you sure you want to delete domain '%s'? This action cannot be undone.\n", args[0])
		cmd.Print("Type 'yes' to confirm: ")
		reader := bufio.NewReader(cmd.InOrStdin())
		line, _ := reader.ReadString('\n')
		confirmation := strings.TrimSpace(line)
		if !strings.EqualFold(confirmation, "yes") {
			cmd.Println("Domain deletion canceled")
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	if err := apiClient.Domains.DeleteDomain(ctx, args[0]); err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	cmd.Printf("Domain '%s' deleted successfully\n", args[0])
	return nil
}

func runDomainVerify(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	// Check if --smtp flag is set
	smtpFlag, _ := cmd.Flags().GetBool("smtp")

	var verification *api.DomainVerification
	var verifyType string

	if smtpFlag {
		verification, err = apiClient.Domains.VerifySMTP(ctx, args[0])
		verifyType = "SMTP"
	} else {
		verification, err = apiClient.Domains.VerifyDomain(ctx, args[0])
		verifyType = "DNS"
	}

	if err != nil {
		return fmt.Errorf("failed to verify domain: %w", err)
	}

	if verification.IsVerified {
		fmt.Printf("Domain '%s' %s is verified ✓\n", args[0], verifyType)
	} else {
		fmt.Printf("Domain '%s' %s verification failed ✗\n", args[0], verifyType)
	}

	return formatOutput(verification, viper.GetString("output"), func(format output.Format) (interface{}, error) {
		if format == output.FormatTable || format == output.FormatCSV {
			return output.FormatDomainVerification(verification, format)
		}
		return verification, nil
	})
}

func runDomainDNS(_ *cobra.Command, args []string) error {
	return domainOperationRunner(
		args,
		func(ctx context.Context, domains *api.DomainService, domainID string) ([]api.DNSRecord, error) {
			return domains.GetDomainDNSRecords(ctx, domainID)
		},
		"failed to get DNS records",
		func(records []api.DNSRecord, format output.Format) (interface{}, error) {
			if format == output.FormatTable || format == output.FormatCSV {
				return output.FormatDNSRecords(records, format)
			}
			return records, nil
		},
	)
}

func runDomainQuota(_ *cobra.Command, args []string) error {
	return domainOperationRunner(
		args,
		func(ctx context.Context, domains *api.DomainService, domainID string) (*api.DomainQuota, error) {
			return domains.GetDomainQuota(ctx, domainID)
		},
		"failed to get domain quota",
		func(quota *api.DomainQuota, format output.Format) (interface{}, error) {
			if format == output.FormatTable || format == output.FormatCSV {
				return output.FormatDomainQuota(quota, format)
			}
			return quota, nil
		},
	)
}

func runDomainStats(_ *cobra.Command, args []string) error {
	return domainOperationRunner(
		args,
		func(ctx context.Context, domains *api.DomainService, domainID string) (*api.DomainStats, error) {
			return domains.GetDomainStats(ctx, domainID)
		},
		"failed to get domain stats",
		func(stats *api.DomainStats, format output.Format) (interface{}, error) {
			if format == output.FormatTable || format == output.FormatCSV {
				return output.FormatDomainStats(stats, format)
			}
			return stats, nil
		},
	)
}

func runDomainMembersList(_ *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	domain, err := apiClient.Domains.GetDomain(ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	return formatOutput(domain.Members, viper.GetString("output"), func(format output.Format) (interface{}, error) {
		if format == output.FormatTable || format == output.FormatCSV {
			return output.FormatDomainMembers(domain.Members, format)
		}
		return domain.Members, nil
	})
}

func runDomainMembersAdd(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	group := "user"
	if len(args) >= 3 {
		group = args[2]
	} else if cmd.Flags().Changed("group") {
		group, _ = cmd.Flags().GetString("group")
	}

	member, err := apiClient.Domains.AddDomainMember(ctx, args[0], args[1], group)
	if err != nil {
		return fmt.Errorf("failed to add domain member: %w", err)
	}

	fmt.Printf("Member '%s' added to domain '%s' with group '%s'\n", args[1], args[0], group)

	return formatOutput(member, viper.GetString("output"), func(_ output.Format) (interface{}, error) {
		return member, nil
	})
}

func runDomainMembersRemove(_ *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	if err := apiClient.Domains.RemoveDomainMember(ctx, args[0], args[1]); err != nil {
		return fmt.Errorf("failed to remove domain member: %w", err)
	}

	fmt.Printf("Member '%s' removed from domain '%s'\n", args[1], args[0])
	return nil
}

// Helper functions

func formatOutput(data interface{}, format string, tableFormatter func(output.Format) (interface{}, error)) error {
	outputFormat, err := output.ParseFormat(format)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	formatter := output.NewFormatter(outputFormat, nil)

	if outputFormat == output.FormatTable || outputFormat == output.FormatCSV {
		tableData, err := tableFormatter(outputFormat)
		if err != nil {
			return err
		}
		return formatter.Format(tableData)
	}

	return formatter.Format(data)
}
