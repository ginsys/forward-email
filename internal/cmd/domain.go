package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ginsys/forwardemail-cli/internal/client"
	"github.com/ginsys/forwardemail-cli/pkg/api"
	"github.com/ginsys/forwardemail-cli/pkg/output"
)

var (
	domainOutputFormat string
	domainPage         int
	domainLimit        int
	domainSort         string
	domainOrder        string
	domainSearch       string
	domainVerified     string
	domainPlan         string
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
	Long:  `Update settings for an existing domain.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDomainUpdate,
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
	Long:  `Verify that the DNS records for a domain are correctly configured.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDomainVerify,
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

	// Global flags for all domain commands
	domainCmd.PersistentFlags().StringVarP(&domainOutputFormat, "output", "o", "table", "Output format (table, json, yaml, csv)")

	// List command specific flags
	domainListCmd.Flags().IntVar(&domainPage, "page", 1, "Page number")
	domainListCmd.Flags().IntVar(&domainLimit, "limit", 25, "Number of results per page")
	domainListCmd.Flags().StringVar(&domainSort, "sort", "name", "Sort field (name, created_at, updated_at, is_verified, plan)")
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

	// Delete command flags
	domainDeleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")

	// Members add command flags
	domainMembersAddCmd.Flags().String("group", "user", "Member group (admin, user)")
}

func runDomainList(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	// Parse verified filter
	var verified *bool
	if domainVerified != "" {
		v, err := strconv.ParseBool(domainVerified)
		if err != nil {
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

	return formatOutput(response.Domains, domainOutputFormat, func(format output.Format) (interface{}, error) {
		if format == output.FormatTable || format == output.FormatCSV {
			return output.FormatDomainList(response.Domains, format)
		}
		return response.Domains, nil
	})
}

func runDomainGet(cmd *cobra.Command, args []string) error {
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

	return formatOutput(domain, domainOutputFormat, func(format output.Format) (interface{}, error) {
		if format == output.FormatTable || format == output.FormatCSV {
			return output.FormatDomainDetails(domain, format)
		}
		return domain, nil
	})
}

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

	fmt.Printf("Domain '%s' created successfully\n", domain.Name)

	return formatOutput(domain, domainOutputFormat, func(format output.Format) (interface{}, error) {
		if format == output.FormatTable || format == output.FormatCSV {
			return output.FormatDomainDetails(domain, format)
		}
		return domain, nil
	})
}

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

	domain, err := apiClient.Domains.UpdateDomain(ctx, args[0], req)
	if err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}

	fmt.Printf("Domain '%s' updated successfully\n", domain.Name)

	return formatOutput(domain, domainOutputFormat, func(format output.Format) (interface{}, error) {
		if format == output.FormatTable || format == output.FormatCSV {
			return output.FormatDomainDetails(domain, format)
		}
		return domain, nil
	})
}

func runDomainDelete(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	if !force {
		fmt.Printf("Are you sure you want to delete domain '%s'? This action cannot be undone.\n", args[0])
		fmt.Print("Type 'yes' to confirm: ")
		var confirmation string
		fmt.Scanln(&confirmation)
		if strings.ToLower(confirmation) != "yes" {
			fmt.Println("Domain deletion cancelled")
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

	fmt.Printf("Domain '%s' deleted successfully\n", args[0])
	return nil
}

func runDomainVerify(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	verification, err := apiClient.Domains.VerifyDomain(ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to verify domain: %w", err)
	}

	if verification.IsVerified {
		fmt.Printf("Domain '%s' is verified ✓\n", args[0])
	} else {
		fmt.Printf("Domain '%s' verification failed ✗\n", args[0])
	}

	return formatOutput(verification, domainOutputFormat, func(format output.Format) (interface{}, error) {
		if format == output.FormatTable || format == output.FormatCSV {
			return output.FormatDomainVerification(verification, format)
		}
		return verification, nil
	})
}

func runDomainDNS(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	records, err := apiClient.Domains.GetDomainDNSRecords(ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to get DNS records: %w", err)
	}

	return formatOutput(records, domainOutputFormat, func(format output.Format) (interface{}, error) {
		if format == output.FormatTable || format == output.FormatCSV {
			return output.FormatDNSRecords(records, format)
		}
		return records, nil
	})
}

func runDomainQuota(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	quota, err := apiClient.Domains.GetDomainQuota(ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to get domain quota: %w", err)
	}

	return formatOutput(quota, domainOutputFormat, func(format output.Format) (interface{}, error) {
		if format == output.FormatTable || format == output.FormatCSV {
			return output.FormatDomainQuota(quota, format)
		}
		return quota, nil
	})
}

func runDomainStats(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	stats, err := apiClient.Domains.GetDomainStats(ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to get domain stats: %w", err)
	}

	return formatOutput(stats, domainOutputFormat, func(format output.Format) (interface{}, error) {
		if format == output.FormatTable || format == output.FormatCSV {
			return output.FormatDomainStats(stats, format)
		}
		return stats, nil
	})
}

func runDomainMembersList(cmd *cobra.Command, args []string) error {
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

	return formatOutput(domain.Members, domainOutputFormat, func(format output.Format) (interface{}, error) {
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

	return formatOutput(member, domainOutputFormat, func(format output.Format) (interface{}, error) {
		return member, nil
	})
}

func runDomainMembersRemove(cmd *cobra.Command, args []string) error {
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
