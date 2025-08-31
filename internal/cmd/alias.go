// Package cmd provides the command-line interface for Forward Email operations.
package cmd

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"os"
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
	aliasRecipients   []string // Recipient email addresses or webhooks
	aliasLabelsFlag   []string // Labels to assign to the alias
	aliasDescription  string   // Alias description text
	aliasEnableFlag   bool     // Enable the alias for forwarding
	aliasDisableFlag  bool     // Disable the alias forwarding
	aliasIMAPFlag     bool     // Enable IMAP access for the alias
	aliasPGPFlag      bool     // Enable PGP encryption for the alias
	aliasPublicKey    string   // PGP public key for encryption
	aliasImportFile   string
	aliasExportFile   string
	aliasImportDryRun bool
)

type SyncAction struct {
	typ        string
	domain     string
	aliasID    string
	name       string
	recipients []string
	enabled    *bool
	labels     []string
	desc       *string
	hasIMAP    *bool
	hasPGP     *bool
}

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

// aliasImportCmd represents importing aliases from CSV
var aliasImportCmd = &cobra.Command{
	Use:   "import <domain> --file <path>",
	Short: "Import aliases from CSV",
	Long: "Import aliases into a domain from a CSV file with columns: " +
		"Name, Recipients (comma-separated), Enabled (true/false), " +
		"Labels (comma-separated), Description.",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := strings.TrimSpace(args[0])
		if domain == "" {
			return fmt.Errorf("domain is required")
		}
		if aliasImportFile == "" {
			return fmt.Errorf("--file is required")
		}

		f, err := os.Open(aliasImportFile)
		if err != nil {
			return fmt.Errorf("failed to open CSV: %v", err)
		}
		defer func() { _ = f.Close() }()

		r := csv.NewReader(f)
		r.FieldsPerRecord = -1
		rows, err := r.ReadAll()
		if err != nil {
			return fmt.Errorf("failed to read CSV: %v", err)
		}
		if len(rows) == 0 {
			return fmt.Errorf("empty CSV")
		}
		// Map header
		header := make(map[string]int)
		for i, h := range rows[0] {
			header[strings.ToLower(strings.TrimSpace(h))] = i
		}
		required := []string{"name", "recipients"}
		for _, k := range required {
			if _, ok := header[k]; !ok {
				return fmt.Errorf("missing required column: %s", k)
			}
		}

		ctx := context.Background()
		apiClient, err := client.NewAPIClient()
		if err != nil {
			return fmt.Errorf("failed to create API client: %v", err)
		}

		// Fetch existing aliases to decide create/update
		existing, err := listAllAliases(ctx, apiClient, domain)
		if err != nil {
			return fmt.Errorf("failed to list aliases for %s: %v", domain, err)
		}
		byName := mapAliasesByName(existing)

		// Plan and process rows
		type impAction struct{ typ, name string }
		var impPlan []impAction
		for _, row := range rows[1:] {
			if len(row) == 0 {
				continue
			}
			name := strings.TrimSpace(row[header["name"]])
			if name == "" {
				continue
			}
			var recStr string
			if idx, ok := header["recipients"]; ok && idx < len(row) {
				recStr = row[idx]
			}
			recipients := splitCSVList(recStr)
			if len(recipients) == 0 {
				return fmt.Errorf("alias %s: at least one recipient required", name)
			}
			var labels []string
			if idx, ok := header["labels"]; ok && idx < len(row) {
				labels = splitCSVList(row[idx])
			}
			var enabledPtr *bool
			if idx, ok := header["enabled"]; ok && idx < len(row) {
				s := strings.TrimSpace(strings.ToLower(row[idx]))
				if s != "" {
					v := s == "true" || s == "1" || s == "yes"
					enabledPtr = &v
				}
			}
			var descPtr *string
			if idx, ok := header["description"]; ok && idx < len(row) {
				d := strings.TrimSpace(row[idx])
				if d != "" {
					descPtr = &d
				}
			}

			if ex, ok := byName[name]; ok {
				// Update
				req := &api.UpdateAliasRequest{Recipients: recipients, Labels: labels}
				if enabledPtr != nil {
					req.IsEnabled = enabledPtr
				}
				if descPtr != nil {
					req.Description = descPtr
				}
				if aliasImportDryRun {
					impPlan = append(impPlan, impAction{typ: "UPDATE", name: name})
				} else {
					if _, err := apiClient.Aliases.UpdateAlias(ctx, domain, ex.ID, req); err != nil {
						return fmt.Errorf("update %s failed: %v", name, err)
					}
				}
			} else {
				// Create
				req := &api.CreateAliasRequest{Name: name, Recipients: recipients, Labels: labels, IsEnabled: true}
				if enabledPtr != nil {
					req.IsEnabled = *enabledPtr
				}
				if descPtr != nil {
					req.Description = *descPtr
				}
				if aliasImportDryRun {
					impPlan = append(impPlan, impAction{typ: "CREATE", name: name})
				} else {
					if _, err := apiClient.Aliases.CreateAlias(ctx, domain, req); err != nil {
						return fmt.Errorf("create %s failed: %v", name, err)
					}
				}
			}
		}
		if aliasImportDryRun {
			headers := []string{"ACTION", "ALIAS"}
			tbl := output.NewTableData(headers)
			for _, a := range impPlan {
				tbl.AddRow([]string{a.typ, a.name})
			}
			formatter := output.NewFormatter(output.FormatTable, cmd.OutOrStdout())
			return formatter.Format(tbl)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Imported aliases into %s from %s\n", domain, aliasImportFile)
		return nil
	},
}

// aliasExportCmd represents exporting aliases to CSV
var aliasExportCmd = &cobra.Command{
	Use:   "export <domain> --file <path>",
	Short: "Export aliases to CSV",
	Long: "Export aliases from a domain to a CSV file with columns: " +
		"Name, Recipients (comma-separated), Enabled, Labels (comma-separated), Description.",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := strings.TrimSpace(args[0])
		if domain == "" {
			return fmt.Errorf("domain is required")
		}
		if aliasExportFile == "" {
			return fmt.Errorf("--file is required")
		}

		ctx := context.Background()
		apiClient, err := client.NewAPIClient()
		if err != nil {
			return fmt.Errorf("failed to create API client: %v", err)
		}
		aliases, err := listAllAliases(ctx, apiClient, domain)
		if err != nil {
			return fmt.Errorf("failed to list aliases for %s: %v", domain, err)
		}

		if err := writeAliasesCSV(aliasExportFile, aliases); err != nil {
			return fmt.Errorf("failed to write CSV: %v", err)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Exported %d aliases from %s to %s\n", len(aliases), domain, aliasExportFile)
		return nil
	},
}

// aliasSyncCmd represents the alias sync command (scaffold per specification)
var aliasSyncCmd = &cobra.Command{
	Use:   "sync <source-domain> <target-domain>",
	Short: "Sync aliases between domains",
	Long: `Synchronize aliases from a source domain to a target domain.

Modes:
- merge: bidirectional merge, preserving unique aliases in both domains
- replace: one-way, replace target with source
- preserve: one-way, copy from source without deleting in target

Examples:
  forward-email alias sync example.com target.com --mode merge --dry-run
  forward-email alias sync example.com target.com --mode replace
  forward-email alias sync example.com target.com --mode preserve --conflicts
`,
	Args: cobra.ExactArgs(2),
	RunE: runAliasSync,
}

var (
	aliasSyncMode     string
	aliasSyncDryRun   bool
	aliasSyncStrategy string // overwrite|skip|merge
)

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
	aliasCmd.AddCommand(aliasImportCmd)
	aliasCmd.AddCommand(aliasExportCmd)

	// Sync command flags
	aliasCmd.AddCommand(aliasSyncCmd)
	aliasSyncCmd.Flags().StringVar(&aliasSyncMode, "mode", "merge", "Sync mode: merge|replace|preserve")
	aliasSyncCmd.Flags().BoolVar(&aliasSyncDryRun, "dry-run", false, "Show planned changes without applying")
	aliasSyncCmd.Flags().StringVar(&aliasSyncStrategy, "conflicts", "", "Conflict strategy: overwrite|skip|merge")

	// CSV flags
	aliasImportCmd.Flags().StringVar(&aliasImportFile, "file", "", "Path to input CSV file")
	aliasImportCmd.Flags().BoolVar(&aliasImportDryRun, "dry-run", false, "Preview import without applying changes")
	aliasExportCmd.Flags().StringVar(&aliasExportFile, "file", "", "Path to output CSV file")

	// Global flags (output inherited from root command)
	aliasCmd.PersistentFlags().StringVarP(&aliasDomain, "domain", "d", "",
		"Domain name or ID (required unless using --all-domains)")

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
	aliasListCmd.Flags().StringVar(&aliasColumns, "columns", "",
		"Specify columns to display (name,domain,recipients,enabled,imap,labels,created)")
	aliasListCmd.Flags().StringVar(&aliasOrderBy, "order-by", "",
		"Sort by columns (e.g., 'domain,name' or 'enabled:desc,created:asc')")

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

// runAliasSync performs alias synchronization between two domains.
func runAliasSync(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: forward-email alias sync <source-domain> <target-domain> [--mode merge|replace|preserve]")
	}
	src := strings.TrimSpace(args[0])
	dst := strings.TrimSpace(args[1])
	if src == dst {
		return fmt.Errorf("source and target domains must differ")
	}

	mode := strings.ToLower(strings.TrimSpace(aliasSyncMode))
	switch mode {
	case "merge", "replace", "preserve":
	default:
		return fmt.Errorf("invalid --mode: %s (valid: merge|replace|preserve)", aliasSyncMode)
	}
	if aliasSyncStrategy != "" {
		s := strings.ToLower(aliasSyncStrategy)
		if s != "overwrite" && s != "skip" && s != "merge" {
			return fmt.Errorf("invalid --conflicts strategy: %s (valid: overwrite|skip|merge)", aliasSyncStrategy)
		}
	}

	ctx := context.Background()
	apiClient, err := client.NewAPIClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %v", err)
	}

	// Fetch aliases for both domains
	srcAliases, err := listAllAliases(ctx, apiClient, src)
	if err != nil {
		return fmt.Errorf("failed to list aliases for %s: %v", src, err)
	}
	dstAliases, err := listAllAliases(ctx, apiClient, dst)
	if err != nil {
		return fmt.Errorf("failed to list aliases for %s: %v", dst, err)
	}

	// Index by name
	srcByName := mapAliasesByName(srcAliases)
	dstByName := mapAliasesByName(dstAliases)

	// Build plan
	var plan []SyncAction

	addCreate := func(domain, name string, a api.Alias) {
		plan = append(plan, SyncAction{
			typ:        "create",
			domain:     domain,
			name:       name,
			recipients: a.Recipients,
			enabled:    &a.IsEnabled,
			labels:     a.Labels,
		})
	}
	addUpdate := func(domain string, id string, name string, desired api.Alias) {
		plan = append(plan, SyncAction{
			typ:        "update",
			domain:     domain,
			aliasID:    id,
			name:       name,
			recipients: desired.Recipients,
			enabled:    &desired.IsEnabled,
			labels:     desired.Labels,
		})
	}
	addDelete := func(domain, id, name string) {
		plan = append(plan, SyncAction{typ: "delete", domain: domain, aliasID: id, name: name})
	}

	switch mode {
	case "merge":
		// union of names
		names := make(map[string]struct{})
		for n := range srcByName {
			names[n] = struct{}{}
		}
		for n := range dstByName {
			names[n] = struct{}{}
		}
		for name := range names {
			s, sOK := srcByName[name]
			d, dOK := dstByName[name]
			switch {
			case sOK && !dOK:
				addCreate(dst, name, s)
			case !sOK && dOK:
				addCreate(src, name, d)
			case sOK && dOK:
				// conflict: compare recipients/flags
				diff := !equalStringSets(s.Recipients, d.Recipients) || s.IsEnabled != d.IsEnabled || !equalStringSets(s.Labels, d.Labels)
				if diff {
					strategy := strings.ToLower(aliasSyncStrategy)
					if strategy == "" && !aliasSyncDryRun {
						sChosen, applyAll, perr := promptConflict(cmd, name, s, d)
						if perr != nil {
							return perr
						}
						strategy = sChosen
						if applyAll {
							aliasSyncStrategy = strategy
						}
					}
					switch strategy {
					case "overwrite":
						addUpdate(dst, d.ID, d.Name, s)
						addUpdate(src, s.ID, s.Name, d) // keep symmetric? For merge, prefer source? We'll merge both ways below
					case "skip":
						// do nothing
					case "merge":
						merged := mergeRecipients(s.Recipients, d.Recipients)
						sDesired := s
						sDesired.Recipients = merged
						dDesired := d
						dDesired.Recipients = merged
						addUpdate(src, s.ID, s.Name, sDesired)
						addUpdate(dst, d.ID, d.Name, dDesired)
					default:
						// default to merge behavior for merge mode when unspecified
						merged := mergeRecipients(s.Recipients, d.Recipients)
						sDesired := s
						sDesired.Recipients = merged
						dDesired := d
						dDesired.Recipients = merged
						addUpdate(src, s.ID, s.Name, sDesired)
						addUpdate(dst, d.ID, d.Name, dDesired)
					}
				}
			}
		}
	case "replace", "preserve":
		// one-way from src to dst
		for name, s := range srcByName {
			if d, ok := dstByName[name]; !ok {
				addCreate(dst, name, s)
			} else {
				// exists: update if different per strategy
				diff := !equalStringSets(s.Recipients, d.Recipients) || s.IsEnabled != d.IsEnabled || !equalStringSets(s.Labels, d.Labels)
				if diff {
					strategy := strings.ToLower(aliasSyncStrategy)
					if strategy == "" && !aliasSyncDryRun {
						sChosen, applyAll, perr := promptConflict(cmd, name, s, d)
						if perr != nil {
							return perr
						}
						strategy = sChosen
						if applyAll {
							aliasSyncStrategy = strategy
						}
					}
					switch strategy {
					case "overwrite":
						addUpdate(dst, d.ID, d.Name, s)
					case "merge":
						desired := d
						desired.Recipients = mergeRecipients(s.Recipients, d.Recipients)
						addUpdate(dst, d.ID, d.Name, desired)
					case "skip":
						// no-op
					default:
						// default to overwrite in one-way
						addUpdate(dst, d.ID, d.Name, s)
					}
				}
			}
		}
		if mode == "replace" {
			for name, d := range dstByName {
				if _, ok := srcByName[name]; !ok {
					addDelete(dst, d.ID, d.Name)
				}
			}
		}
	}

	if aliasSyncDryRun {
		return printSyncPlan(cmd, src, dst, plan)
	}

	// Execute plan
	for _, a := range plan {
		switch a.typ {
		case "create":
			req := &api.CreateAliasRequest{Recipients: a.recipients, Labels: a.labels, Name: a.name, IsEnabled: true}
			if a.enabled != nil {
				req.IsEnabled = *a.enabled
			}
			if _, err := apiClient.Aliases.CreateAlias(ctx, a.domain, req); err != nil {
				return fmt.Errorf("create %s@%s failed: %v", a.name, a.domain, err)
			}
		case "update":
			req := &api.UpdateAliasRequest{Recipients: a.recipients}
			if a.enabled != nil {
				req.IsEnabled = a.enabled
			}
			if len(a.labels) > 0 {
				req.Labels = a.labels
			}
			if _, err := apiClient.Aliases.UpdateAlias(ctx, a.domain, a.aliasID, req); err != nil {
				return fmt.Errorf("update %s in %s failed: %v", a.aliasID, a.domain, err)
			}
		case "delete":
			if err := apiClient.Aliases.DeleteAlias(ctx, a.domain, a.aliasID); err != nil {
				return fmt.Errorf("delete %s in %s failed: %v", a.aliasID, a.domain, err)
			}
		}
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Alias sync completed: %s -> %s (mode=%s, actions=%d)\n", src, dst, mode, len(plan))
	return nil
}

func listAllAliases(ctx context.Context, c *api.Client, domain string) ([]api.Alias, error) {
	// Fetch single page large enough for most cases; could loop if needed
	resp, err := c.Aliases.ListAliases(ctx, &api.ListAliasesOptions{Domain: domain, Limit: 1000})
	if err != nil {
		return nil, err
	}
	return resp.Aliases, nil
}

func mapAliasesByName(list []api.Alias) map[string]api.Alias {
	m := make(map[string]api.Alias, len(list))
	for _, a := range list {
		m[a.Name] = a
	}
	return m
}

func equalStringSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	ma := make(map[string]int, len(a))
	for _, s := range a {
		ma[strings.ToLower(strings.TrimSpace(s))]++
	}
	for _, s := range b {
		k := strings.ToLower(strings.TrimSpace(s))
		if ma[k] == 0 {
			return false
		}
		ma[k]--
		if ma[k] < 0 {
			return false
		}
	}
	for _, v := range ma {
		if v != 0 {
			return false
		}
	}
	return true
}

func mergeRecipients(a, b []string) []string {
	m := make(map[string]struct{}, len(a)+len(b))
	for _, s := range a {
		m[strings.ToLower(strings.TrimSpace(s))] = struct{}{}
	}
	for _, s := range b {
		m[strings.ToLower(strings.TrimSpace(s))] = struct{}{}
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func printSyncPlan(cmd *cobra.Command, src, dst string, plan []SyncAction) error {
	headers := []string{"ACTION", "DOMAIN", "ALIAS", "DETAILS"}
	tbl := output.NewTableData(headers)
	for _, a := range plan {
		alias := a.name
		if alias == "" {
			alias = a.aliasID
		}
		details := ""
		switch a.typ {
		case "create":
			details = fmt.Sprintf("recipients=%v enabled=%v", a.recipients, derefBool(a.enabled))
		case "update":
			details = fmt.Sprintf("recipients=%v enabled=%v", a.recipients, derefBool(a.enabled))
		case "delete":
			details = "remove alias"
		}
		tbl.AddRow([]string{strings.ToUpper(a.typ), a.domain, alias, details})
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "DRY RUN: Alias Sync Plan (%s -> %s, actions=%d)\n", src, dst, len(plan))
	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	formatter := output.NewFormatter(output.FormatTable, cmd.OutOrStdout())
	return formatter.Format(tbl)
}

func derefBool(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func writeAliasesCSV(path string, aliases []api.Alias) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"Name", "Recipients", "Enabled", "Labels", "Description"}); err != nil {
		return err
	}
	for _, a := range aliases {
		rec := strings.Join(a.Recipients, ",")
		lab := strings.Join(a.Labels, ",")
		if err := w.Write([]string{a.Name, rec, fmt.Sprintf("%v", a.IsEnabled), lab, a.Description}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func splitCSVList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// promptConflict interactively asks the user how to resolve a conflict.
// Returns strategy (overwrite|skip|merge) and whether to apply to all.
func promptConflict(cmd *cobra.Command, alias string, src, dst api.Alias) (string, bool, error) {
	r := bufio.NewReader(cmd.InOrStdin())
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "Conflict for alias '%s':\n", alias)
	_, _ = fmt.Fprintf(out, "  source → %v\n", src.Recipients)
	_, _ = fmt.Fprintf(out, "  target → %v\n", dst.Recipients)
	_, _ = fmt.Fprintln(out, "Choose: [o] overwrite target, [s] skip, [m] merge, [a] apply to all (with last choice)")
	for {
		_, _ = fmt.Fprint(out, "Choice [o/s/m/a]: ")
		line, err := r.ReadString('\n')
		if err != nil {
			return "", false, err
		}
		c := strings.ToLower(strings.TrimSpace(line))
		switch c {
		case "o":
			return "overwrite", false, nil
		case "s":
			return "skip", false, nil
		case "m":
			return "merge", false, nil
		case "a":
			// apply-all uses last non-empty choice; default to merge
			return "merge", true, nil
		}
	}
}

// formatAliasListMultiDomain formats aliases from multiple domains with proper domain resolution
func formatAliasListMultiDomain(
	aliases []api.Alias, format output.Format, domainMap map[string]string,
) (*output.TableData, error) {
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
func formatAliasListWithCustomColumns(
	aliases []api.Alias, format output.Format, domainMap map[string]string, columnsStr string,
) (*output.TableData, error) {
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
			validFieldsList := "name, domain, enabled, imap, created, updated, recipients, labels"
			return nil, fmt.Errorf("invalid sort field '%s'. Valid fields: %s", fieldName, validFieldsList)
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
	if aliasColumns != "" { //nolint:gocritic // ifElseChain: Complex conditions not suitable for switch
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

	switch len(args) {
	case 2:
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	case 1:
		// Only one argument - could be alias ID with --domain flag, or domain+aliasID
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	default:
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

	switch len(args) {
	case 2:
		// Domain and alias name provided as positional arguments
		domain = args[0]
		aliasName = args[1]
	case 1:
		// One argument provided; decide whether it's domain or alias name
		if domain != "" { //nolint:gocritic // ifElseChain: Complex nested conditions not suitable for switch
			// Domain was provided via flag; single arg must be alias name
			aliasName = args[0]
		} else if len(aliasRecipients) > 0 {
			// Creating an alias (recipients given) but no domain specified
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		} else {
			// Likely domain provided without alias name
			return fmt.Errorf("alias name is required")
		}
	default:
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

	switch len(args) {
	case 2:
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	case 1:
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
	default:
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

	switch len(args) {
	case 2:
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	case 1:
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	default:
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

	switch len(args) {
	case 2:
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	case 1:
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	default:
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

	switch len(args) {
	case 2:
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	case 1:
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	default:
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

	switch len(args) {
	case 2:
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	case 1:
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	default:
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

	switch len(args) {
	case 2:
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	case 1:
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	default:
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

	switch len(args) {
	case 2:
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	case 1:
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	default:
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

	switch len(args) {
	case 2:
		// Domain and alias ID provided as positional arguments
		domain = args[0]
		aliasID = args[1]
	case 1:
		// Only one argument - could be alias ID with --domain flag
		if domain == "" {
			return fmt.Errorf("domain is required - specify as first argument or use --domain flag")
		}
		aliasID = args[0]
	default:
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
