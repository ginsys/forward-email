package cmd

import (
	"bufio"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ginsys/forward-email/internal/client"
	"github.com/ginsys/forward-email/pkg/api"
	"github.com/ginsys/forward-email/pkg/output"
)

// Global variables for alias command flags.
// These store parsed command-line arguments for alias operations, filtering, and configuration.
var (
	// List/filtering flags
	aliasPage       int    // Page number for pagination (default: 1)
	aliasLimit      int    // Number of results per page (default: 25)
	aliasSort       string // Sort field: name, created_at, updated_at
	aliasOrder      string // Sort order: asc or desc
	aliasSearch     string // Search filter for alias names
	aliasEnabled    string // Enabled status filter: "true" or "false"
	aliasLabels     string // Labels filter (comma-separated)
	aliasHasIMAP    string // IMAP enabled filter: "true" or "false"
	aliasDomain     string // Domain filter for aliases
	aliasAllDomains bool   // Include aliases from all domains
	aliasColumns    string // Custom column selection for output
	aliasOrderBy    string // Alternative sort field specification

	// Create/Update operation flags
	aliasRecipients  []string // Recipient email addresses or webhooks
	aliasLabelsFlag  []string // Labels to assign to the alias
	aliasDescription string   // Alias description text
	aliasEnableFlag  bool     // Enable the alias for forwarding
	aliasDisableFlag bool     // Disable the alias forwarding
	aliasIMAPFlag    bool     // Enable IMAP access for the alias
	aliasPGPFlag     bool     // Enable PGP encryption for the alias
	aliasPublicKey   string   // PGP public key for encryption
)

// aliasCmd represents the alias command
var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage Forward Email aliases",
	Long: `Manage Forward Email aliases including creating, listing, updating, 
and configuring alias settings and recipients.`,
}

// aliasListCmd represents the alias list command
var aliasListCmd = &cobra.Command{
	Use:   "list [domain...]",
	Short: "List aliases",
	Long: `List aliases for one or more domains, or all available domains.

Examples:
  forward-email alias list example.com              # Single domain
  forward-email alias list example.com,test.com     # Multiple domains (comma-separated)
  forward-email alias list --all-domains            # All available domains
  forward-email alias list --domain example.com     # Using flag
  forward-email alias list --columns name,domain,recipients  # Custom columns
  forward-email alias list --order-by domain,name   # Sort by domain, then name
  forward-email alias list --order-by enabled:desc,created:asc  # Sort with direction`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAliasList,
}

// aliasGetCmd represents the alias get command
var aliasGetCmd = &cobra.Command{
	Use:   "get [domain] <alias-id>",
	Short: "Get alias details",
	Long: `Get detailed information about a specific alias.
	
You can specify the domain either as a positional argument or using the --domain flag:
  forward-email alias get example.com alias123
  forward-email alias get alias123 --domain example.com`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runAliasGet,
}

// aliasCreateCmd represents the alias create command
var aliasCreateCmd = &cobra.Command{
	Use:   "create [domain] <name> --recipients <email1,email2>",
	Short: "Create a new alias",
	Long: `Create a new alias for a domain with specified recipients.
	
You can specify the domain either as a positional argument or using the --domain flag:
  forward-email alias create example.com sales --recipients sales@company.com
  forward-email alias create sales --domain example.com --recipients sales@company.com`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runAliasCreate,
}

// aliasUpdateCmd represents the alias update command
var aliasUpdateCmd = &cobra.Command{
	Use:   "update [domain] <alias-id>",
	Short: "Update alias settings",
	Long: `Update settings for an existing alias.
	
You can specify the domain either as a positional argument or using the --domain flag:
  forward-email alias update example.com alias123 --enable
  forward-email alias update alias123 --domain example.com --enable`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runAliasUpdate,
}

// aliasDeleteCmd represents the alias delete command
var aliasDeleteCmd = &cobra.Command{
	Use:   "delete [domain] <alias-id>",
	Short: "Delete an alias",
	Long: `Delete an alias from a domain.
	
You can specify the domain either as a positional argument or using the --domain flag:
  forward-email alias delete example.com alias123
  forward-email alias delete alias123 --domain example.com`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runAliasDelete,
}

// aliasEnableCmd represents the alias enable command
var aliasEnableCmd = &cobra.Command{
	Use:   "enable [domain] <alias-id>",
	Short: "Enable an alias",
	Long: `Enable an alias to start receiving emails.
	
You can specify the domain either as a positional argument or using the --domain flag:
  forward-email alias enable example.com alias123
  forward-email alias enable alias123 --domain example.com`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runAliasEnable,
}

// aliasDisableCmd represents the alias disable command
var aliasDisableCmd = &cobra.Command{
	Use:   "disable [domain] <alias-id>",
	Short: "Disable an alias",
	Long: `Disable an alias to stop receiving emails.
	
You can specify the domain either as a positional argument or using the --domain flag:
  forward-email alias disable example.com alias123
  forward-email alias disable alias123 --domain example.com`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runAliasDisable,
}

// aliasRecipientsCmd represents the alias recipients command
var aliasRecipientsCmd = &cobra.Command{
	Use:   "recipients [domain] <alias-id> --recipients <email1,email2>",
	Short: "Update alias recipients",
	Long: `Update the list of recipients for an alias.
	
You can specify the domain either as a positional argument or using the --domain flag:
  forward-email alias recipients example.com alias123 --recipients new@email.com
  forward-email alias recipients alias123 --domain example.com --recipients new@email.com`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runAliasRecipients,
}

// aliasPasswordCmd represents the alias password command
var aliasPasswordCmd = &cobra.Command{
	Use:   "password [domain] <alias-id>",
	Short: "Generate IMAP password",
	Long: `Generate a new IMAP password for an alias.
	
You can specify the domain either as a positional argument or using the --domain flag:
  forward-email alias password example.com alias123
  forward-email alias password alias123 --domain example.com`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runAliasPassword,
}

// aliasQuotaCmd represents the alias quota command
var aliasQuotaCmd = &cobra.Command{
	Use:   "quota [domain] <alias-id>",
	Short: "Show alias quota",
	Long: `Show storage and email quota information for an alias.
	
You can specify the domain either as a positional argument or using the --domain flag:
  forward-email alias quota example.com alias123
  forward-email alias quota alias123 --domain example.com`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runAliasQuota,
}

// aliasStatsCmd represents the alias stats command
var aliasStatsCmd = &cobra.Command{
	Use:   "stats [domain] <alias-id>",
	Short: "Show alias statistics",
	Long: `Show usage statistics for an alias.
	
You can specify the domain either as a positional argument or using the --domain flag:
  forward-email alias stats example.com alias123
  forward-email alias stats alias123 --domain example.com`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runAliasStats,
}

func init() {
	// Register with root command
	rootCmd.AddCommand(aliasCmd)

	// Add subcommands
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasGetCmd)
	aliasCmd.AddCommand(aliasCreateCmd)
	aliasCmd.AddCommand(aliasUpdateCmd)
	aliasCmd.AddCommand(aliasDeleteCmd)
	aliasCmd.AddCommand(aliasEnableCmd)
	aliasCmd.AddCommand(aliasDisableCmd)
	aliasCmd.AddCommand(aliasRecipientsCmd)
	aliasCmd.AddCommand(aliasPasswordCmd)
	aliasCmd.AddCommand(aliasQuotaCmd)
	aliasCmd.AddCommand(aliasStatsCmd)

	// Global flags (output inherited from root command)
	aliasCmd.PersistentFlags().StringVarP(&aliasDomain, "domain", "d", "", "Domain name or ID (required unless using --all-domains)")

	// List command flags
	aliasListCmd.Flags().IntVar(&aliasPage, "page", 1, "Page number")
	aliasListCmd.Flags().IntVar(&aliasLimit, "limit", 25, "Number of aliases per page")
	aliasListCmd.Flags().StringVar(&aliasSort, "sort", "name", "Sort by (name, created, updated)")
	aliasListCmd.Flags().StringVar(&aliasOrder, "order", "asc", "Sort order (asc, desc)")
	aliasListCmd.Flags().StringVar(&aliasSearch, "search", "", "Search alias names")
	aliasListCmd.Flags().StringVar(&aliasEnabled, "enabled", "", "Filter by enabled status (true/false)")
	aliasListCmd.Flags().StringVar(&aliasLabels, "labels", "", "Filter by labels (comma-separated)")
	aliasListCmd.Flags().StringVar(&aliasHasIMAP, "has-imap", "", "Filter by IMAP capability (true/false)")
	aliasListCmd.Flags().BoolVar(&aliasAllDomains, "all-domains", false, "List aliases from all available domains")
	aliasListCmd.Flags().StringVar(&aliasColumns, "columns", "", "Specify columns to display (name,domain,recipients,enabled,imap,labels,created)")
	aliasListCmd.Flags().StringVar(&aliasOrderBy, "order-by", "", "Sort by columns (e.g., 'domain,name' or 'enabled:desc,created:asc')")

	// Create command flags
	aliasCreateCmd.Flags().StringSliceVar(&aliasRecipients, "recipients", nil, "Recipient email addresses")
	aliasCreateCmd.Flags().StringSliceVar(&aliasLabelsFlag, "labels", nil, "Labels for the alias")
	aliasCreateCmd.Flags().StringVar(&aliasDescription, "description", "", "Description for the alias")
	aliasCreateCmd.Flags().BoolVar(&aliasEnableFlag, "enabled", true, "Enable the alias")
	aliasCreateCmd.Flags().BoolVar(&aliasIMAPFlag, "imap", false, "Enable IMAP access")
	aliasCreateCmd.Flags().BoolVar(&aliasPGPFlag, "pgp", false, "Enable PGP encryption")
	aliasCreateCmd.Flags().StringVar(&aliasPublicKey, "public-key", "", "PGP public key")
	// Validation is handled in runAliasCreate to produce clear error messages
	aliasCreateCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		if strings.Contains(err.Error(), "required flag(s) \"recipients\"") {
			return fmt.Errorf("at least one recipient is required")
		}
		return err
	})

	// Update command flags
	aliasUpdateCmd.Flags().StringSliceVar(&aliasRecipients, "recipients", nil, "Update recipient email addresses")
	aliasUpdateCmd.Flags().StringSliceVar(&aliasLabelsFlag, "labels", nil, "Update labels for the alias")
	aliasUpdateCmd.Flags().StringVar(&aliasDescription, "description", "", "Update description for the alias")
	aliasUpdateCmd.Flags().BoolVar(&aliasEnableFlag, "enable", false, "Enable the alias")
	aliasUpdateCmd.Flags().BoolVar(&aliasDisableFlag, "disable", false, "Disable the alias")
	aliasUpdateCmd.Flags().BoolVar(&aliasIMAPFlag, "imap", false, "Enable IMAP access")
	aliasUpdateCmd.Flags().BoolVar(&aliasPGPFlag, "pgp", false, "Enable PGP encryption")
	aliasUpdateCmd.Flags().StringVar(&aliasPublicKey, "public-key", "", "Update PGP public key")

	// Recipients command flags
	aliasRecipientsCmd.Flags().StringSliceVar(&aliasRecipients, "recipients", nil, "New recipient email addresses")
	// Validation is handled in runAliasRecipients to produce clear error messages
	aliasRecipientsCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		if strings.Contains(err.Error(), "required flag(s) \"recipients\"") {
			return fmt.Errorf("at least one recipient is required")
		}
		return err
	})

	// Note: Domain flag is no longer required since all commands accept domain as a positional argument
	// Users can specify domain either as first positional argument or using --domain flag
}

// formatAliasListMultiDomain formats aliases from multiple domains with proper domain resolution
func formatAliasListMultiDomain(aliases []api.Alias, format output.Format, domainMap map[string]string) (*output.TableData, error) {
	if format != output.FormatTable && format != output.FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for aliases")
	}

	headers := []string{"NAME", "DOMAIN", "RECIPIENTS", "ENABLED", "IMAP", "LABELS", "CREATED"}
	table := output.NewTableData(headers)

	for _, alias := range aliases {
		enabled := output.FormatValue(alias.IsEnabled)
		imap := output.FormatValue(alias.HasIMAP)

		var recipients, labels string

		// Build recipients/labels string (same for table and CSV)
		recipients = strings.Join(alias.Recipients, ", ")
		labels = strings.Join(alias.Labels, ", ")

		var created string
		if alias.CreatedAt.IsZero() {
			created = "-"
		} else {
			created = alias.CreatedAt.Format("2006-01-02")
		}

		// Resolve domain ID to domain name using the mapping
		domainName := alias.DomainID
		if mappedName, exists := domainMap[alias.DomainID]; exists {
			domainName = mappedName
		}

		row := []string{
			alias.Name,
			domainName,
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

// formatAliasListWithCustomColumns formats aliases with user-specified columns
func formatAliasListWithCustomColumns(aliases []api.Alias, format output.Format, domainMap map[string]string, columnsStr string) (*output.TableData, error) {
	if format != output.FormatTable && format != output.FormatCSV {
		return nil, fmt.Errorf("use direct JSON/YAML encoding for aliases")
	}

	// Parse column specification
	columnNames := strings.Split(columnsStr, ",")
	for i, col := range columnNames {
		columnNames[i] = strings.TrimSpace(strings.ToUpper(col))
	}

	// Validate columns
	validColumns := map[string]bool{
		"NAME": true, "DOMAIN": true, "RECIPIENTS": true, "ENABLED": true,
		"IMAP": true, "LABELS": true, "CREATED": true, "UPDATED": true,
		"DESCRIPTION": true, "PGP": true, "ID": true,
	}

	for _, col := range columnNames {
		if !validColumns[col] {
			return nil, fmt.Errorf("invalid column '%s'. Valid columns: %s", col, validColumnsList)
		}
	}

	table := output.NewTableData(columnNames)

	for _, alias := range aliases {
		row := make([]string, len(columnNames))

		// Resolve domain ID to domain name using the mapping
		domainName := alias.DomainID
		if mappedName, exists := domainMap[alias.DomainID]; exists {
			domainName = mappedName
		}

		for i, col := range columnNames {
			switch col {
			case "NAME":
				row[i] = alias.Name
			case "DOMAIN":
				row[i] = domainName
			case "RECIPIENTS":
				row[i] = strings.Join(alias.Recipients, ", ")
			case "ENABLED":
				row[i] = output.FormatValue(alias.IsEnabled)
			case "IMAP":
				row[i] = output.FormatValue(alias.HasIMAP)
			case "LABELS":
				row[i] = strings.Join(alias.Labels, ", ")
			case "CREATED":
				if alias.CreatedAt.IsZero() {
					row[i] = "-"
				} else {
					row[i] = alias.CreatedAt.Format("2006-01-02")
				}
			case "UPDATED":
				if alias.UpdatedAt.IsZero() {
					row[i] = "-"
				} else {
					row[i] = alias.UpdatedAt.Format("2006-01-02")
				}
			case "DESCRIPTION":
				row[i] = alias.Description
			case "PGP":
				row[i] = output.FormatValue(alias.HasPGP)
			case "ID":
				row[i] = alias.ID
			}
		}
		table.AddRow(row)
	}

	return table, nil
}

// sortAliases sorts a slice of aliases based on the order specification
func sortAliases(aliases []api.Alias, domainMap map[string]string, orderBy string) error {
	// Parse the order specification
	sortCriteria, err := parseSortCriteria(orderBy)
	if err != nil {
		return err
	}

	// Implement custom sorting
	sort.Slice(aliases, func(i, j int) bool {
		a, b := aliases[i], aliases[j]

		// Compare based on each sort criteria in order
		for _, criterion := range sortCriteria {
			cmp := compareAliases(a, b, criterion, domainMap)
			if cmp != 0 {
				if criterion.descending {
					return cmp > 0
				}
				return cmp < 0
			}
			// If equal, continue to next criterion
		}

		// If all criteria are equal, maintain stable sort by name
		return a.Name < b.Name
	})

	return nil
}

// sortCriterion represents a single sort field with direction
type sortCriterion struct {
	field      string
	descending bool
}

// parseSortCriteria parses the order-by string into sort criteria
func parseSortCriteria(orderBy string) ([]sortCriterion, error) {
	parts := strings.Split(orderBy, ",")
	criteria := make([]sortCriterion, len(parts))

	validFields := map[string]bool{
		"name": true, "domain": true, "enabled": true, "imap": true,
		"created": true, "updated": true, "recipients": true, "labels": true,
	}

	for i, part := range parts {
		part = strings.TrimSpace(part)

		// Check for direction suffix (:asc or :desc)
		fieldName := part
		descending := false

		if strings.Contains(part, ":") {
			fieldParts := strings.Split(part, ":")
			if len(fieldParts) != 2 {
				return nil, fmt.Errorf("invalid sort format '%s'. Use 'field' or 'field:asc/desc'", part)
			}
			fieldName = fieldParts[0]
			direction := strings.ToLower(fieldParts[1])
			if direction == "desc" {
				descending = true
			} else if direction != sortAsc {
				return nil, fmt.Errorf("invalid sort direction '%s'. Use 'asc' or 'desc'", direction)
			}
		}

		fieldName = strings.ToLower(fieldName)
		if !validFields[fieldName] {
			return nil, fmt.Errorf("invalid sort field '%s'. Valid fields: %s", fieldName, "name, domain, enabled, imap, created, updated, recipients, labels")
		}

		criteria[i] = sortCriterion{
			field:      fieldName,
			descending: descending,
		}
	}

	return criteria, nil
}

// compareAliases compares two aliases based on a specific criterion
// Returns: -1 if a < b, 0 if a == b, 1 if a > b
func compareAliases(a, b api.Alias, criterion sortCriterion, domainMap map[string]string) int {
	switch criterion.field {
	case sortFieldName:
		return strings.Compare(a.Name, b.Name)
	case "domain":
		domainA := a.DomainID
		if mapped, exists := domainMap[a.DomainID]; exists {
			domainA = mapped
		}
		domainB := b.DomainID
		if mapped, exists := domainMap[b.DomainID]; exists {
			domainB = mapped
		}
		return strings.Compare(domainA, domainB)
	case "enabled":
		// Convert bool to int for comparison (false=0, true=1)
		aVal, bVal := 0, 0
		if a.IsEnabled {
			aVal = 1
		}
		if b.IsEnabled {
			bVal = 1
		}
		return aVal - bVal
	case "imap":
		aVal, bVal := 0, 0
		if a.HasIMAP {
			aVal = 1
		}
		if b.HasIMAP {
			bVal = 1
		}
		return aVal - bVal
	case "created":
		if a.CreatedAt.Before(b.CreatedAt) {
			return -1
		} else if a.CreatedAt.After(b.CreatedAt) {
			return 1
		}
		return 0
	case "updated":
		if a.UpdatedAt.Before(b.UpdatedAt) {
			return -1
		} else if a.UpdatedAt.After(b.UpdatedAt) {
			return 1
		}
		return 0
	case "recipients":
		// Sort by number of recipients, then by first recipient name
		if len(a.Recipients) != len(b.Recipients) {
			return len(a.Recipients) - len(b.Recipients)
		}
		if len(a.Recipients) > 0 && len(b.Recipients) > 0 {
			return strings.Compare(a.Recipients[0], b.Recipients[0])
		}
		return 0
	case "labels":
		// Sort by number of labels, then by first label name
		if len(a.Labels) != len(b.Labels) {
			return len(a.Labels) - len(b.Labels)
		}
		if len(a.Labels) > 0 && len(b.Labels) > 0 {
			return strings.Compare(a.Labels[0], b.Labels[0])
		}
		return 0
	default:
		return 0
	}
}

func runAliasList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	// Validate flag combinations
	if aliasAllDomains && (len(args) > 0 || aliasDomain != "") {
		return fmt.Errorf("cannot use --all-domains with domain arguments or --domain flag")
	}

	// Determine which domains to query
	var domains []string
	if aliasAllDomains {
		// Get all available domains
		domainList, listErr := apiClient.Domains.ListDomains(ctx, &api.ListDomainsOptions{
			Page:  1,
			Limit: 1000, // Get all domains
		})
		if listErr != nil {
			return fmt.Errorf("failed to fetch domains: %v", listErr)
		}
		for _, domain := range domainList.Domains {
			domains = append(domains, domain.Name)
		}
		if len(domains) == 0 {
			fmt.Println("No domains found")
			return nil
		}
	} else {
		// Get domain(s) from positional argument or flag
		domain := aliasDomain
		if len(args) > 0 {
			domain = args[0]
		}
		if domain == "" {
			return fmt.Errorf("domain is required - specify as argument, use --domain flag, or use --all-domains")
		}

		// Split comma-separated domains
		domains = strings.Split(domain, ",")
		for i, d := range domains {
			domains[i] = strings.TrimSpace(d)
		}
	}

	// Parse boolean flags
	var enabled *bool
	if aliasEnabled != "" {
		if val, parseErr := strconv.ParseBool(aliasEnabled); parseErr == nil {
			enabled = &val
		}
	}

	var hasIMAP *bool
	if aliasHasIMAP != "" {
		if val, parseErr := strconv.ParseBool(aliasHasIMAP); parseErr == nil {
			hasIMAP = &val
		}
	}

	// Create a mapping from domain ID to domain name
	var domainMap map[string]string
	var allAliases []api.Alias
	var totalCount int
	var totalPages int

	// Initialize domain mapping - will be populated as we fetch aliases
	domainMap = make(map[string]string)

	for _, domain := range domains {
		opts := &api.ListAliasesOptions{
			Domain:  domain,
			Page:    aliasPage,
			Limit:   aliasLimit,
			Sort:    aliasSort,
			Order:   aliasOrder,
			Search:  aliasSearch,
			Enabled: enabled,
			Labels:  aliasLabels,
			HasIMAP: hasIMAP,
		}

		response, listErr := apiClient.Aliases.ListAliases(ctx, opts)
		if listErr != nil {
			fmt.Printf("Warning: failed to list aliases for domain %s: %v\n", domain, listErr)
			continue
		}

		// Add aliases to the list with proper domain tracking
		// WORKAROUND: Forward Email API doesn't populate domain_id field, so we set it ourselves
		for _, alias := range response.Aliases {
			// Set the DomainID to the domain we queried so mapping works
			alias.DomainID = domain
			allAliases = append(allAliases, alias)
		}

		// Ensure domain mapping exists for this domain
		domainMap[domain] = domain

		totalCount += response.TotalCount
		if response.TotalPages > totalPages {
			totalPages = response.TotalPages
		}
	}

	if len(allAliases) == 0 {
		fmt.Println("No aliases found")
		return nil
	}

	// Apply custom sorting if specified
	if aliasOrderBy != "" {
		if sortErr := sortAliases(allAliases, domainMap, aliasOrderBy); sortErr != nil {
			return fmt.Errorf("failed to sort aliases: %v", sortErr)
		}
	}

	format, err := output.ParseFormat(viper.GetString("output"))
	if err != nil {
		return fmt.Errorf("invalid output format: %v", err)
	}

	if format == output.FormatJSON || format == output.FormatYAML {
		formatter := output.NewFormatter(format, cmd.OutOrStdout())
		return formatter.Format(allAliases)
	}

	// Handle different formatting scenarios
	var tableData *output.TableData
	if aliasColumns != "" {
		// Custom columns specified - ensure domain map is populated for single domains
		if len(domains) == 1 && !aliasAllDomains && len(domainMap) == 0 {
			// For single domain with custom columns, create a simple mapping
			domainMap = make(map[string]string)
			for _, alias := range allAliases {
				domainMap[alias.DomainID] = domains[0]
			}
		}
		tableData, err = formatAliasListWithCustomColumns(allAliases, format, domainMap, aliasColumns)
	} else if len(domains) == 1 && !aliasAllDomains {
		// Single domain - use simple formatting with the domain name we queried
		tableData, err = output.FormatAliasList(allAliases, format, domains[0])
	} else {
		// Multiple domains or --all-domains - use multi-domain formatting with mapping
		tableData, err = formatAliasListMultiDomain(allAliases, format, domainMap)
	}
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	err = formatter.Format(tableData)
	if err != nil {
		return err
	}

	// Show pagination info for non-JSON/YAML formats
	if len(allAliases) > 0 {
		if len(domains) == 1 {
			fmt.Printf("\nShowing %d aliases from domain %s\n", len(allAliases), domains[0])
		} else {
			fmt.Printf("\nShowing %d aliases from %d domains\n", len(allAliases), len(domains))
		}
		if totalCount > len(allAliases) {
			fmt.Printf("Total: %d aliases (use --page to see more)\n", totalCount)
		}
	}

	return nil
}

func runAliasGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse domain and alias ID from positional arguments or flags
	domain := aliasDomain
	var aliasID string

	if len(args) == 2 {
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	} else if len(args) == 1 {
		// Only one argument - could be alias ID with --domain flag, or domain+aliasID
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	} else {
		return fmt.Errorf("alias ID is required")
	}

	if domain == "" {
		return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
	}

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	alias, err := apiClient.Aliases.GetAlias(ctx, domain, aliasID)
	if err != nil {
		return fmt.Errorf("failed to get alias: %v", err)
	}

	format, err := output.ParseFormat(viper.GetString("output"))
	if err != nil {
		return fmt.Errorf("invalid output format: %v", err)
	}

	if format == output.FormatJSON || format == output.FormatYAML {
		formatter := output.NewFormatter(format, cmd.OutOrStdout())
		return formatter.Format(alias)
	}

	// Format as table
	tableData, err := output.FormatAliasDetails(alias, format)
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	return formatter.Format(tableData)
}

func runAliasCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse domain and alias name from positional arguments or flags
	domain := aliasDomain
	var aliasName string

	if len(args) == 2 {
		// Domain and alias name provided as positional arguments
		domain = args[0]
		aliasName = args[1]
	} else if len(args) == 1 {
		// One argument provided; decide whether it's domain or alias name
		if domain != "" {
			// Domain was provided via flag; single arg must be alias name
			aliasName = args[0]
		} else if len(aliasRecipients) > 0 {
			// Creating an alias (recipients given) but no domain specified
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		} else {
			// Likely domain provided without alias name
			return fmt.Errorf("alias name is required")
		}
	} else {
		return fmt.Errorf("alias name is required")
	}

	if domain == "" {
		return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
	}

	if len(aliasRecipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	req := &api.CreateAliasRequest{
		Name:        aliasName,
		Recipients:  aliasRecipients,
		Labels:      aliasLabelsFlag,
		Description: aliasDescription,
		IsEnabled:   aliasEnableFlag,
		HasIMAP:     aliasIMAPFlag,
		HasPGP:      aliasPGPFlag,
		PublicKey:   aliasPublicKey,
	}

	alias, err := apiClient.Aliases.CreateAlias(ctx, domain, req)
	if err != nil {
		return fmt.Errorf("failed to create alias: %v", err)
	}

	cmd.Printf("✅ Alias '%s' created successfully\n", alias.Name)

	format, err := output.ParseFormat(viper.GetString("output"))
	if err != nil {
		return fmt.Errorf("invalid output format: %v", err)
	}

	if format == output.FormatJSON || format == output.FormatYAML {
		formatter := output.NewFormatter(format, cmd.OutOrStdout())
		return formatter.Format(alias)
	}

	// Format as table
	tableData, err := output.FormatAliasDetails(alias, format)
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	return formatter.Format(tableData)
}

func runAliasUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse domain and alias ID from positional arguments or flags
	domain := aliasDomain
	var aliasID string

	if len(args) == 2 {
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	} else if len(args) == 1 {
		if domain != "" {
			// Domain provided via flag; single arg must be alias ID
			aliasID = args[0]
		} else {
			// Decide based on argument shape: if it looks like a domain, alias ID is missing
			if strings.Contains(args[0], ".") {
				return fmt.Errorf("alias ID is required")
			}
			// Otherwise treat as alias ID with missing domain
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
	} else {
		return fmt.Errorf("alias ID is required")
	}

	if domain == "" {
		return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
	}

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	req := &api.UpdateAliasRequest{}

	// Check what flags were set and update accordingly
	if len(aliasRecipients) > 0 {
		req.Recipients = aliasRecipients
	}
	if len(aliasLabelsFlag) > 0 {
		req.Labels = aliasLabelsFlag
	}
	if cmd.Flags().Changed("description") {
		req.Description = &aliasDescription
	}
	if cmd.Flags().Changed("enable") && aliasEnableFlag {
		enabled := true
		req.IsEnabled = &enabled
	}
	if cmd.Flags().Changed("disable") && aliasDisableFlag {
		enabled := false
		req.IsEnabled = &enabled
	}
	if cmd.Flags().Changed("imap") {
		req.HasIMAP = &aliasIMAPFlag
	}
	if cmd.Flags().Changed("pgp") {
		req.HasPGP = &aliasPGPFlag
	}
	if cmd.Flags().Changed("public-key") {
		req.PublicKey = &aliasPublicKey
	}

	alias, err := apiClient.Aliases.UpdateAlias(ctx, domain, aliasID, req)
	if err != nil {
		return fmt.Errorf("failed to update alias: %v", err)
	}

	cmd.Printf("✅ Alias '%s' updated successfully\n", alias.Name)

	format, err := output.ParseFormat(viper.GetString("output"))
	if err != nil {
		return fmt.Errorf("invalid output format: %v", err)
	}

	if format == output.FormatJSON || format == output.FormatYAML {
		formatter := output.NewFormatter(format, cmd.OutOrStdout())
		return formatter.Format(alias)
	}

	// Format as table
	tableData, err := output.FormatAliasDetails(alias, format)
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	return formatter.Format(tableData)
}

func runAliasDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse domain and alias ID from positional arguments or flags
	domain := aliasDomain
	var aliasID string

	if len(args) == 2 {
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	} else if len(args) == 1 {
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	} else {
		return fmt.Errorf("alias ID is required")
	}

	if domain == "" {
		return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
	}

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	// Get alias info first for confirmation
	alias, err := apiClient.Aliases.GetAlias(ctx, domain, aliasID)
	if err != nil {
		return fmt.Errorf("failed to get alias: %v", err)
	}

	cmd.Printf("⚠️  Are you sure you want to delete alias '%s'? This action cannot be undone.\n", alias.Name)
	cmd.Print("Type 'yes' to confirm: ")

	reader := bufio.NewReader(cmd.InOrStdin())
	line, _ := reader.ReadString('\n')
	confirmation := strings.TrimSpace(line)

	if !strings.EqualFold(confirmation, yesStr) {
		cmd.Printf("❌ Deletion canceled\n")
		return nil
	}

	err = apiClient.Aliases.DeleteAlias(ctx, domain, aliasID)
	if err != nil {
		return fmt.Errorf("failed to delete alias: %v", err)
	}

	cmd.Printf("✅ Alias '%s' deleted successfully\n", alias.Name)
	return nil
}

func runAliasEnable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse domain and alias ID from positional arguments or flags
	domain := aliasDomain
	var aliasID string

	if len(args) == 2 {
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	} else if len(args) == 1 {
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	} else {
		return fmt.Errorf("alias ID is required")
	}

	if domain == "" {
		return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
	}

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	alias, err := apiClient.Aliases.EnableAlias(ctx, domain, aliasID)
	if err != nil {
		return fmt.Errorf("failed to enable alias: %v", err)
	}

	cmd.Printf("✅ Alias '%s' enabled successfully\n", alias.Name)
	return nil
}

func runAliasDisable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse domain and alias ID from positional arguments or flags
	domain := aliasDomain
	var aliasID string

	if len(args) == 2 {
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	} else if len(args) == 1 {
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	} else {
		return fmt.Errorf("alias ID is required")
	}

	if domain == "" {
		return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
	}

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	alias, err := apiClient.Aliases.DisableAlias(ctx, domain, aliasID)
	if err != nil {
		return fmt.Errorf("failed to disable alias: %v", err)
	}

	cmd.Printf("✅ Alias '%s' disabled successfully\n", alias.Name)
	return nil
}

func runAliasRecipients(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse domain and alias ID from positional arguments or flags
	domain := aliasDomain
	var aliasID string

	if len(args) == 2 {
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	} else if len(args) == 1 {
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	} else {
		return fmt.Errorf("alias ID is required")
	}

	if domain == "" {
		return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
	}

	if len(aliasRecipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	alias, err := apiClient.Aliases.UpdateRecipients(ctx, domain, aliasID, aliasRecipients)
	if err != nil {
		return fmt.Errorf("failed to update recipients: %v", err)
	}

	cmd.Printf("✅ Recipients updated for alias '%s'\n", alias.Name)
	cmd.Printf("New recipients: %s\n", strings.Join(alias.Recipients, ", "))
	return nil
}

func runAliasPassword(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse domain and alias ID from positional arguments or flags
	domain := aliasDomain
	var aliasID string

	if len(args) == 2 {
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	} else if len(args) == 1 {
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	} else {
		return fmt.Errorf("alias ID is required")
	}

	if domain == "" {
		return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
	}

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	response, err := apiClient.Aliases.GeneratePassword(ctx, domain, aliasID)
	if err != nil {
		return fmt.Errorf("failed to generate password: %v", err)
	}

	cmd.Printf("✅ New IMAP password generated\n")
	cmd.Printf("Password: %s\n", response.Password)
	cmd.Println("⚠️  Store this password securely - it cannot be retrieved again")
	return nil
}

func runAliasQuota(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse domain and alias ID from positional arguments or flags
	domain := aliasDomain
	var aliasID string

	if len(args) == 2 {
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	} else if len(args) == 1 {
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	} else {
		return fmt.Errorf("alias ID is required")
	}

	if domain == "" {
		return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
	}

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	quota, err := apiClient.Aliases.GetAliasQuota(ctx, domain, aliasID)
	if err != nil {
		return fmt.Errorf("failed to get alias quota: %v", err)
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
	tableData, err := output.FormatAliasQuota(quota, format)
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	return formatter.Format(tableData)
}

func runAliasStats(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse domain and alias ID from positional arguments or flags
	domain := aliasDomain
	var aliasID string

	if len(args) == 2 {
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	} else if len(args) == 1 {
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	} else {
		return fmt.Errorf("alias ID is required")
	}

	if domain == "" {
		return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
	}

	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	stats, err := apiClient.Aliases.GetAliasStats(ctx, domain, aliasID)
	if err != nil {
		return fmt.Errorf("failed to get alias stats: %v", err)
	}

	format, err := output.ParseFormat(viper.GetString("output"))
	if err != nil {
		return fmt.Errorf("invalid output format: %v", err)
	}

	if format == output.FormatJSON || format == output.FormatYAML {
		formatter := output.NewFormatter(format, cmd.OutOrStdout())
		return formatter.Format(stats)
	}

	// Format as table
	tableData, err := output.FormatAliasStats(stats, format)
	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	return formatter.Format(tableData)
}
