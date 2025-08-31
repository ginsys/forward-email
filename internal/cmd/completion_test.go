package cmd

import (
	"bytes"
	"testing"
)

func TestCompletion_GeneratesScripts(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, sh := range shells {
		t.Run(sh, func(t *testing.T) {
			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs([]string{"completion", sh})
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("completion %s failed: %v", sh, err)
			}
			if out.Len() == 0 {
				t.Fatalf("expected non-empty completion output for %s", sh)
			}
		})
	}
}
