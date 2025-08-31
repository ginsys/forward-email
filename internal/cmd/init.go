package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	kr "github.com/99designs/keyring"
	ikeyring "github.com/ginsys/forward-email/internal/keyring"
)

// initCmd provides an interactive setup wizard for first-time users.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive setup wizard",
	Long:  "Guide to configure a profile, store API key securely, and create a config file.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		in := bufio.NewReader(cmd.InOrStdin())

		// Ask for profile name
		_, _ = fmt.Fprint(cmd.OutOrStdout(), "Profile name [default]: ")
		profile, _ := in.ReadString('\n')
		profile = strings.TrimSpace(profile)
		if profile == "" {
			profile = "default"
		}

		// Ask for API key (basic masking not implemented in plain stdin)
		_, _ = fmt.Fprint(cmd.OutOrStdout(), "API key: ")
		apiKey, _ := in.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			return errors.New("API key is required")
		}

		// Store API key based on requested backend
		store := cmd.Flag("store").Value.String()
		filePass := cmd.Flag("file-pass").Value.String()
		switch store {
		case "auto", "keyring":
			kr, err := ikeyring.New(ikeyring.Config{})
			if err != nil {
				if store == "keyring" {
					return fmt.Errorf("failed to initialize system keyring: %w", err)
				}
				kr = nil
			}
			if kr != nil {
				if err := kr.SetAPIKey(profile, apiKey); err != nil {
					return fmt.Errorf("failed to store API key: %w", err)
				}
				break
			}
			fallthrough
		case "config":
			// Will be saved to config below
		case "file":
			// Persistent file keyring under config dir
			cfgDir := os.Getenv("XDG_CONFIG_HOME")
			if cfgDir == "" {
				home, herr := os.UserHomeDir()
				if herr != nil {
					return fmt.Errorf("failed to determine home dir: %w", herr)
				}
				cfgDir = filepath.Join(home, ".config")
			}
			feDir := filepath.Join(cfgDir, "forwardemail", "keyring")
			if mkErr := os.MkdirAll(feDir, 0o750); mkErr != nil {
				return fmt.Errorf("failed to create keyring dir: %w", mkErr)
			}
			if store == "file" {
				if filePass == "" {
					_, _ = fmt.Fprint(cmd.OutOrStdout(), "File keyring passphrase: ")
					passBytes, perr := in.ReadString('\n')
					if perr != nil {
						return fmt.Errorf("failed to read passphrase: %w", perr)
					}
					filePass = strings.TrimSpace(passBytes)
					if filePass == "" {
						return errors.New("file keyring passphrase cannot be empty")
					}
				}
				kr, err := ikeyring.New(ikeyring.Config{
					AllowedBackends:  []kr.BackendType{kr.FileBackend},
					FileDir:          feDir,
					FilePasswordFunc: func(string) (string, error) { return filePass, nil },
				})
				if err != nil {
					return fmt.Errorf("failed to initialize file keyring: %w", err)
				}
				if err := kr.SetAPIKey(profile, apiKey); err != nil {
					return fmt.Errorf("failed to store API key: %w", err)
				}
			}
		default:
			return fmt.Errorf("invalid --store value: %s", store)
		}

		// Prepare config directory and file
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home dir: %w", err)
		}
		cfgDir := filepath.Join(home, ".config", "forwardemail")
		if err := os.MkdirAll(cfgDir, 0o750); err != nil {
			return fmt.Errorf("failed to create config dir: %w", err)
		}
		cfgPath := filepath.Join(cfgDir, "config.yaml")

		// Minimal config: set current_profile and a default profile
		v := viper.New()
		v.SetConfigFile(cfgPath)
		v.SetConfigType("yaml")
		v.Set("current_profile", profile)
		v.Set("profiles."+profile+".base_url", "https://api.forwardemail.net")
		v.Set("profiles."+profile+".timeout", "30s")
		v.Set("profiles."+profile+".output", "table")

		if err := v.WriteConfigAs(cfgPath); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n✅ Setup complete. Config: %s (profile: %s)\n", cfgPath, profile)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Tip: run 'forward-email auth verify' to validate credentials.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("store", "auto", "Credential store: auto|keyring|file|config")
	initCmd.Flags().String("file-pass", "", "Passphrase for file keyring (used when --store=file)")
}
