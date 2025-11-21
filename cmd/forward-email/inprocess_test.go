package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	appcmd "github.com/ginsys/forward-email/internal/cmd"
	"github.com/stretchr/testify/assert"
)

// captureOutput redirects stdout/stderr during fn and returns their combined output.
func captureOutput(fn func()) (stdout, stderr string) {
	// Capture stdout
	oldOut := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut
	// Capture stderr
	oldErr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	// Execute function
	fn()

	// Restore
	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr

	var bufOut, bufErr bytes.Buffer
	_, _ = io.Copy(&bufOut, rOut)
	_, _ = io.Copy(&bufErr, rErr)
	_ = rOut.Close()
	_ = rErr.Close()

	return bufOut.String(), bufErr.String()
}

// TestInProcess_BasicFlows executes the CLI in-process so coverage is attributed
// to invoked packages, improving module-level coverage beyond external exec tests.
func TestInProcess_BasicFlows(t *testing.T) {
	t.Setenv("FORWARDEMAIL_KEYRING_BACKEND", "none")
	t.Setenv("FORWARDEMAIL_NO_COLOR", "1")

	// Isolate config to a temporary directory
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)

	cases := []struct {
		name     string
		args     []string
		wantErr  bool
		wantOut  string
		wantErrS string
	}{
		{name: "help", args: []string{"forward-email", "--help"}, wantOut: "Forward Email CLI"},
		{name: "version", args: []string{"forward-email", "version", "--verbose"}, wantOut: "forward-email version"},
		{name: "invalid", args: []string{"forward-email", "does-not-exist"}, wantErr: true, wantErrS: "unknown command"},
		{name: "auth status", args: []string{"forward-email", "auth", "status"}, wantOut: "Authentication Status"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Set process args for Cobra
			oldArgs := os.Args
			os.Args = tc.args
			defer func() { os.Args = oldArgs }()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			out, errOut := captureOutput(func() {
				_ = appcmd.Execute(ctx) // We assert using outputs instead of error type
			})

			if tc.wantErr {
				assert.Contains(t, errOut, tc.wantErrS, "stderr should contain expected error text")
			} else {
				assert.Empty(t, errOut, fmt.Sprintf("unexpected stderr: %s", errOut))
			}
			if tc.wantOut != "" {
				assert.Contains(t, out, tc.wantOut, "stdout should contain expected text")
			}
		})
	}
}
