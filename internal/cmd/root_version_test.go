package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	buildversion "github.com/ginsys/forward-email/internal/version"
	"github.com/spf13/cobra"
)

// Flags these tests may set on the real command tree. rootCmd is a package-level global, so a
// value parsed in one test would otherwise leak into every later rootCmd.Execute.
var (
	rootFlagsTouched    = []string{"help", "version", "verbose", "profile", "output", "debug", "timeout"}
	versionFlagsTouched = []string{"help", "json", "verbose", "license", "check-update"}
)

// resetChangedFlags restores the named flags of cmd to their defaults when a test changed them.
func resetChangedFlags(t *testing.T, cmd *cobra.Command, names []string) {
	t.Helper()
	for _, name := range names {
		f := cmd.Flags().Lookup(name)
		if f == nil || !f.Changed {
			continue
		}
		if err := f.Value.Set(f.DefValue); err != nil {
			t.Errorf("reset flag %q on %q: %v", name, cmd.Name(), err)
		}
		f.Changed = false
	}
}

// runRealRootCmd executes the package-level rootCmd (not a stand-in) with args, using disposable
// application configuration, and restores its args, output and flag state afterwards.
func runRealRootCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	versionCmd, _, err := rootCmd.Find([]string{"version"})
	if err != nil || versionCmd == rootCmd {
		t.Fatalf("version subcommand not registered on rootCmd: %v", err)
	}
	resetFlags := func() {
		resetChangedFlags(t, rootCmd, rootFlagsTouched)
		resetChangedFlags(t, versionCmd, versionFlagsTouched)
	}
	// Not every test in this package that executes rootCmd restores it (a leftover --help would
	// win over --version), so start from defaults as well as cleaning up afterwards.
	resetFlags()
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		resetFlags()
	})

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	return out.String(), err
}

func TestRootCmd_VersionFlag_PrintsVersionTemplate(t *testing.T) {
	out, err := runRealRootCmd(t, "--version")
	if err != nil {
		t.Fatalf("forward-email --version failed: %v\noutput: %s", err, out)
	}

	v := buildversion.Get()
	want := fmt.Sprintf("forward-email version %s\ncommit: %s\nbuilt: %s\n", v.Version, v.Commit, v.Date)
	if out != want {
		t.Errorf("forward-email --version output = %q, want %q", out, want)
	}

	f := rootCmd.Flags().Lookup("version")
	if f == nil {
		t.Fatal("rootCmd has no --version flag")
	}
	if f.Shorthand != "" {
		t.Errorf("--version shorthand = %q, want none (-v belongs to --verbose)", f.Shorthand)
	}
}

func TestRootCmd_VerboseShorthand_IsNotVersion(t *testing.T) {
	out, err := runRealRootCmd(t, "-v")
	if err != nil {
		t.Fatalf("forward-email -v failed: %v\noutput: %s", err, out)
	}

	for _, unwanted := range []string{"forward-email version", "commit:", "built:"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("forward-email -v printed version output (%q):\n%s", unwanted, out)
		}
	}
	// The root command has no Run, so cobra shows help once flags parse cleanly.
	if !strings.Contains(out, "Usage:") {
		t.Errorf("forward-email -v: expected root help output, got:\n%s", out)
	}

	verbose, err := rootCmd.PersistentFlags().GetBool("verbose")
	if err != nil {
		t.Fatalf("lookup verbose flag: %v", err)
	}
	if !verbose {
		t.Error("forward-email -v did not set --verbose")
	}
	if f := rootCmd.Flags().ShorthandLookup("v"); f == nil || f.Name != "verbose" {
		t.Errorf("-v resolves to %v, want the verbose flag", f)
	}
	if f := rootCmd.Flags().Lookup("version"); f != nil && f.Changed {
		t.Error("forward-email -v set the version flag")
	}
}

func TestRootCmd_VersionSubcommand_Unchanged(t *testing.T) {
	v := buildversion.Get()

	t.Run("plain", func(t *testing.T) {
		out, err := runRealRootCmd(t, "version")
		if err != nil {
			t.Fatalf("forward-email version failed: %v\noutput: %s", err, out)
		}
		if out != v.Version+"\n" {
			t.Errorf("forward-email version output = %q, want %q", out, v.Version+"\n")
		}
	})

	for _, flag := range []string{"-v", "--verbose"} {
		t.Run("verbose "+flag, func(t *testing.T) {
			out, err := runRealRootCmd(t, "version", flag)
			if err != nil {
				t.Fatalf("forward-email version %s failed: %v\noutput: %s", flag, err, out)
			}
			if out != v.String()+"\n" {
				t.Errorf("forward-email version %s output = %q, want %q", flag, out, v.String()+"\n")
			}
		})
	}

	t.Run("json", func(t *testing.T) {
		out, err := runRealRootCmd(t, "version", "--json")
		if err != nil {
			t.Fatalf("forward-email version --json failed: %v\noutput: %s", err, out)
		}
		var got buildversion.Info
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("invalid JSON output: %v\n%s", err, out)
		}
		if got != v {
			t.Errorf("forward-email version --json = %+v, want %+v", got, v)
		}
	})

	t.Run("check-update", func(t *testing.T) {
		out, err := runRealRootCmd(t, "version", "--check-update")
		if err != nil {
			t.Fatalf("forward-email version --check-update failed: %v\noutput: %s", err, out)
		}
		if !strings.HasPrefix(out, v.Version+"\n") || !strings.Contains(out, "/releases") {
			t.Errorf("forward-email version --check-update output = %q", out)
		}
	})

	t.Run("license", func(t *testing.T) {
		out, err := runRealRootCmd(t, "version", "--license")
		if err != nil {
			t.Fatalf("forward-email version --license failed: %v\noutput: %s", err, out)
		}
		if !strings.HasPrefix(out, v.Version+"\n") || !strings.Contains(out, "License") {
			t.Errorf("forward-email version --license output = %q", out)
		}
	})

	t.Run("flag set", func(t *testing.T) {
		versionCmd, _, err := rootCmd.Find([]string{"version"})
		if err != nil {
			t.Fatalf("find version subcommand: %v", err)
		}
		want := map[string]string{"json": "", "verbose": "v", "license": "", "check-update": ""}
		for name, shorthand := range want {
			f := versionCmd.LocalNonPersistentFlags().Lookup(name)
			if f == nil {
				t.Errorf("version subcommand lost its --%s flag", name)
				continue
			}
			if f.Shorthand != shorthand {
				t.Errorf("version --%s shorthand = %q, want %q", name, f.Shorthand, shorthand)
			}
		}
	})
}
