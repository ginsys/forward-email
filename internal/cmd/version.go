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
	)

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display CLI version, build metadata, and optional license.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := buildversion.Get()

			switch {
			case jsonOut:
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			case verbose:
				fmt.Fprintln(cmd.OutOrStdout(), info.String())
			default:
				fmt.Fprintln(cmd.OutOrStdout(), info.Version)
			}

			if showLicense {
				// Best-effort: read local LICENSE if available
				if data, err := os.ReadFile("LICENSE"); err == nil {
					fmt.Fprintln(cmd.OutOrStdout())
					fmt.Fprintln(cmd.OutOrStdout(), string(data))
				} else {
					fmt.Fprintln(cmd.OutOrStdout())
					fmt.Fprintln(cmd.OutOrStdout(), "License: MIT (see repository LICENSE file)")
				}
			}
			return nil
		},
		Example: `  forward-email version
  forward-email version --verbose
  forward-email version --json
  forward-email version --license`,
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output version info as JSON")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed version info")
	cmd.Flags().BoolVar(&showLicense, "license", false, "Print license after version info")

	return cmd
}

func init() {
	rootCmd.AddCommand(newVersionCmd())
}
