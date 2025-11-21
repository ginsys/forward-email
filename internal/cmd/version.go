package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	buildversion "github.com/ginsys/forward-email/internal/version"
	"github.com/spf13/cobra"
)

// newVersionCmd creates the `version` subcommand with output modes.
func newVersionCmd() *cobra.Command {
	var (
		jsonOut     bool
		verbose     bool
		showLicense bool
		checkUpdate bool
	)

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display CLI version, build metadata, and optional license.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			info := buildversion.Get()

			switch {
			case jsonOut:
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			case verbose:
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), info.String())
			default:
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), info.Version)
			}

			if checkUpdate {
				// Offline-friendly hint. Actual remote check handled by CI/package managers.
				_, _ = fmt.Fprintln(cmd.OutOrStdout())
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "For updates, see: https://github.com/ginsys/forward-email/releases")
			}

			if showLicense {
				// Best-effort: read local LICENSE if available
				if data, err := os.ReadFile("LICENSE"); err == nil {
					_, _ = fmt.Fprintln(cmd.OutOrStdout())
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout())
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "License: MIT (see repository LICENSE file)")
				}
			}
			return nil
		},
		Example: `  forward-email version
  forward-email version --verbose
  forward-email version --json
  forward-email version --license
  forward-email version --check-update`,
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output version info as JSON")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed version info")
	cmd.Flags().BoolVar(&showLicense, "license", false, "Print license after version info")
	cmd.Flags().BoolVar(&checkUpdate, "check-update", false, "Show update link for latest releases")

	return cmd
}

func init() {
	rootCmd.AddCommand(newVersionCmd())
}
