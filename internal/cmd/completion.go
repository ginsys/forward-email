package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// completionCmd provides shell completion scripts for supported shells.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for bash, zsh, fish, and PowerShell.

To enable completions in your shell, run one of the examples below, or
refer to your OS/shell documentation.`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		w := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(w)
		case "zsh":
			return rootCmd.GenZshCompletion(w)
		case "fish":
			return rootCmd.GenFishCompletion(w, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(w)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
	Example: `  forward-email completion bash > /usr/local/etc/bash_completion.d/forward-email
  forward-email completion zsh > /usr/local/share/zsh/site-functions/_forward-email
  forward-email completion fish > ~/.config/fish/completions/forward-email.fish
  forward-email completion powershell > forward-email.ps1`,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
