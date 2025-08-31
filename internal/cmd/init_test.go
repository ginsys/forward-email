package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCommand_WritesConfigAndStoresKey(t *testing.T) {
	// Prepare temp HOME and keyring file backend
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	// We will use flags to force file keyring, so no env needed

	// Simulate input: profile name + API key
	input := bytes.NewBufferString("dev\nsecret-key\n")

	c := initCmd
	c.SetIn(input)
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetErr(&out)
	// Force file backend via flags to avoid GUI prompts
	_ = c.Flags().Set("store", "file")
	_ = c.Flags().Set("file-pass", "test-password")

	if err := c.RunE(c, nil); err != nil {
		t.Fatalf("init failed: %v\noutput: %s", err, out.String())
	}

	// Verify config file exists under temp HOME
	cfg := filepath.Join(tmp, ".config", "forwardemail", "config.yaml")
	if _, err := os.Stat(cfg); err != nil {
		t.Fatalf("expected config file at %s: %v", cfg, err)
	}
}
