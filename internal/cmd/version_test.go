package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionCommand_Basic(t *testing.T) {
	cmd := newVersionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Errorf("expected version output, got empty string")
	}
}

func TestVersionCommand_Verbose(t *testing.T) {
	cmd := newVersionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--verbose"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("version --verbose failed: %v", err)
	}
	s := out.String()
	for _, want := range []string{"forward-email version", "commit:", "os/arch:"} {
		if !strings.Contains(s, want) {
			t.Errorf("verbose output should contain %q; got: %q", want, s)
		}
	}
}

func TestVersionCommand_JSON(t *testing.T) {
	cmd := newVersionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("version --json failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(out.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out.String())
	}
	for _, key := range []string{"version", "commit", "date", "os", "arch"} {
		if _, ok := m[key]; !ok {
			t.Errorf("json output missing key %q; got: %v", key, m)
		}
	}
}

func TestVersionCommand_CheckUpdate(t *testing.T) {
	cmd := newVersionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--check-update"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("version --check-update failed: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "releases") {
		t.Errorf("expected update hint in output, got: %q", s)
	}
}
