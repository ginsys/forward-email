package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain_Integration tests the complete CLI application by building and running it
func TestMain_Integration(t *testing.T) {
	binary := buildTestBinary(t)
	defer cleanupTestBinary(t, binary)

	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string
		wantErr  string
		timeout  time.Duration
	}{
		{
			name:     "help command",
			args:     []string{"--help"},
			wantCode: 0,
			wantOut:  "Forward Email CLI",
			timeout:  5 * time.Second,
		},
		{
			name:     "short help flag",
			args:     []string{"-h"},
			wantCode: 0,
			wantOut:  "Forward Email CLI",
			timeout:  5 * time.Second,
		},
		{
			name:     "version flag",
			args:     []string{"--version"},
			wantCode: 0,
			wantOut:  "forward-email version",
			timeout:  5 * time.Second,
		},
		{
			name:     "invalid command",
			args:     []string{"invalid-command"},
			wantCode: 1,
			wantErr:  "unknown command",
			timeout:  5 * time.Second,
		},
		{
			name:     "auth help",
			args:     []string{"auth", "--help"},
			wantCode: 0,
			wantOut:  "Manage authentication credentials",
			timeout:  5 * time.Second,
		},
		{
			name:     "email help",
			args:     []string{"email", "--help"},
			wantCode: 0,
			wantOut:  "Send emails and manage sent email history",
			timeout:  5 * time.Second,
		},
		{
			name:     "domain help",
			args:     []string{"domain", "--help"},
			wantCode: 0,
			wantOut:  "Manage Forward Email domains",
			timeout:  5 * time.Second,
		},
		{
			name:     "alias help",
			args:     []string{"alias", "--help"},
			wantCode: 0,
			wantOut:  "Manage Forward Email aliases",
			timeout:  5 * time.Second,
		},
		{
			name:     "profile help",
			args:     []string{"profile", "--help"},
			wantCode: 0,
			wantOut:  "Manage configuration profiles",
			timeout:  5 * time.Second,
		},
		{
			name:     "debug help",
			args:     []string{"debug", "--help"},
			wantCode: 0,
			wantOut:  "Debug utilities for troubleshooting",
			timeout:  5 * time.Second,
		},
		{
			name:     "auth status without authentication",
			args:     []string{"auth", "status"},
			wantCode: 0,
			wantOut:  "No API key configured",
			timeout:  10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, binary, tt.args...) //nolint:gosec // Test code
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Check exit code
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else if ctx.Err() == context.DeadlineExceeded {
					t.Fatalf("command timed out after %v", tt.timeout)
				} else {
					t.Fatalf("failed to execute command: %v", err)
				}
			}

			assert.Equal(t, tt.wantCode, exitCode, "exit code mismatch for %s", tt.name)

			// Check stdout
			if tt.wantOut != "" {
				assert.Contains(t, stdout.String(), tt.wantOut,
					"stdout should contain expected text for %s", tt.name)
			}

			// Check stderr for errors
			if tt.wantErr != "" {
				assert.Contains(t, stderr.String(), tt.wantErr,
					"stderr should contain expected error text for %s", tt.name)
			}
		})
	}
}

// TestMain_GracefulShutdown tests signal handling and graceful shutdown
func TestMain_GracefulShutdown(t *testing.T) {
	binary := buildTestBinary(t)
	defer cleanupTestBinary(t, binary)

	t.Run("SIGINT_handling", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Start a long-running command that would normally wait for input
		//nolint:gosec // Test code
		cmd := exec.CommandContext(ctx, binary, "auth", "login", "--interactive")

		err := cmd.Start()
		require.NoError(t, err, "should be able to start command")

		// Give the command time to start
		time.Sleep(100 * time.Millisecond)

		// Send SIGINT signal
		err = cmd.Process.Signal(os.Interrupt)
		require.NoError(t, err, "should be able to send SIGINT")

		// Wait for process to exit
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case err := <-done:
			// Process should exit cleanly with error code 1 (interrupted)
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					// Exit code should be 1 (error) or 130 (SIGINT)
					exitCode := exitError.ExitCode()
					assert.True(t, exitCode == 1 || exitCode == 130,
						"expected exit code 1 or 130 for interrupted command, got %d", exitCode)
				}
			}
		case <-ctx.Done():
			t.Error("command did not exit within timeout after SIGINT")
			_ = cmd.Process.Kill() // Ignore error in cleanup path
		}
	})

	t.Run("SIGTERM_handling", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Start a long-running command
		//nolint:gosec // Test code
		cmd := exec.CommandContext(ctx, binary, "auth", "login", "--interactive")

		err := cmd.Start()
		require.NoError(t, err, "should be able to start command")

		// Give the command time to start
		time.Sleep(100 * time.Millisecond)

		// Send SIGTERM signal
		err = cmd.Process.Signal(syscall.SIGTERM)
		require.NoError(t, err, "should be able to send SIGTERM")

		// Wait for process to exit
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case err := <-done:
			// Process should exit cleanly with some error code
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode := exitError.ExitCode()
					// SIGTERM typically results in exit code 1 or 143 (128 + 15)
					assert.True(t, exitCode == 1 || exitCode == 143,
						"expected exit code 1 or 143 for SIGTERM, got %d", exitCode)
				}
			}
		case <-ctx.Done():
			t.Error("command did not exit within timeout after SIGTERM")
			_ = cmd.Process.Kill() // Ignore error in cleanup path
		}
	})
}

// TestMain_ErrorHandling tests error scenarios and proper exit codes
func TestMain_ErrorHandling(t *testing.T) {
	binary := buildTestBinary(t)
	defer cleanupTestBinary(t, binary)

	tests := []struct {
		name        string
		args        []string
		wantCode    int
		description string
	}{
		{
			name:        "invalid_flag",
			args:        []string{"--invalid-flag"},
			wantCode:    1,
			description: "Should exit with code 1 for unknown flags",
		},
		{
			name:        "missing_required_arg",
			args:        []string{"domain", "get"},
			wantCode:    1,
			description: "Should exit with code 1 for missing required arguments",
		},
		{
			name:        "too_many_args",
			args:        []string{"domain", "get", "domain1", "domain2"},
			wantCode:    1,
			description: "Should exit with code 1 for too many arguments",
		},
		{
			name:        "invalid_subcommand",
			args:        []string{"email", "invalid-subcommand"},
			wantCode:    0,
			description: "Should show help for invalid subcommands (Cobra behavior)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binary, tt.args...) //nolint:gosec // Test code
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()

			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else if ctx.Err() == context.DeadlineExceeded {
					t.Fatalf("command timed out")
				} else {
					t.Fatalf("failed to execute command: %v", err)
				}
			}

			assert.Equal(t, tt.wantCode, exitCode, tt.description)
			if tt.wantCode == 1 {
				assert.NotEmpty(t, stderr.String(), "should output error message")
			}
		})
	}
}

// TestMain_OutputFormats tests different output formats
func TestMain_OutputFormats(t *testing.T) {
	binary := buildTestBinary(t)
	defer cleanupTestBinary(t, binary)

	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{
			name:     "help_output_format",
			args:     []string{"--help"},
			wantCode: 0,
		},
		{
			name:     "version_output_format",
			args:     []string{"--version"},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binary, tt.args...) //nolint:gosec // Test code
			var stdout bytes.Buffer
			cmd.Stdout = &stdout

			err := cmd.Run()

			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				}
			}

			assert.Equal(t, tt.wantCode, exitCode)
			assert.NotEmpty(t, stdout.String(), "should produce output")

			// Check that output looks properly formatted
			output := stdout.String()
			assert.False(t, strings.Contains(output, "Error:"),
				"help/version output should not contain error messages")
		})
	}
}

// TestMain_EnvironmentVariables tests that CLI respects environment variables
func TestMain_EnvironmentVariables(t *testing.T) {
	binary := buildTestBinary(t)
	defer cleanupTestBinary(t, binary)

	t.Run("config_dir_environment", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Create temporary config directory
		tmpDir := t.TempDir()

		cmd := exec.CommandContext(ctx, binary, "profile", "list") //nolint:gosec // Test code
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("XDG_CONFIG_HOME=%s", tmpDir),
			"FORWARDEMAIL_KEYRING_BACKEND=none",
		)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		// Should not crash due to environment variable
		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			}
		}

		// Command may fail (no profiles configured), but it should handle the env var
		assert.True(t, exitCode == 0 || exitCode == 1,
			"should handle custom config directory, got exit code %d", exitCode)
	})
}

// Helper functions

// buildTestBinary compiles the CLI binary for testing
func buildTestBinary(t *testing.T) string {
	t.Helper()

	// Create a temporary binary path
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "forward-email-test")

	// Build the binary
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//nolint:gosec // Test code with hardcoded "go" command
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, ".")
	cmd.Dir = "." // Build in current directory (cmd/forward-email)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "failed to build test binary: %s", stderr.String())

	// Verify the binary was created and is executable
	info, err := os.Stat(binaryPath)
	require.NoError(t, err, "test binary should exist")
	require.False(t, info.IsDir(), "test binary should not be a directory")

	return binaryPath
}

// cleanupTestBinary removes the test binary
func cleanupTestBinary(t *testing.T, binaryPath string) {
	t.Helper()
	if binaryPath != "" {
		_ = os.Remove(binaryPath) // Ignore cleanup error
	}
}
