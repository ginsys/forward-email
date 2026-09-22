package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// buildTimeoutEnv overrides the compile budget used when the integration
	// tests build the CLI binary. It accepts any Go duration (for example
	// "10m"). Slow or cold CI runners can raise it without patching the tests.
	buildTimeoutEnv = "FORWARDEMAIL_TEST_BUILD_TIMEOUT"

	// defaultBuildTimeout is the compile budget when buildTimeoutEnv is unset.
	// A cold module and build cache — the normal state on a hosted runner right
	// after a dependency change — can need well over the 30s this used to
	// allow, while still being far below this bound; a genuinely hung build
	// therefore fails in bounded time instead of hanging until the Go test
	// timeout.
	defaultBuildTimeout = 5 * time.Minute

	// waitDelayAfterKill bounds how long Wait may keep draining the build's
	// output pipes after the deadline killed it. Without it (exec.Cmd.WaitDelay
	// defaults to zero) those pipes are read until EOF, which a surviving
	// grandchild of the compiler can postpone indefinitely.
	waitDelayAfterKill = 15 * time.Second

	// buildHelperEnv selects the behaviour of the helper child process used by
	// the build-budget tests. It is never set for a normal test run.
	buildHelperEnv = "FORWARDEMAIL_TEST_BUILD_HELPER"

	// helperCompileErrorText stands in for compiler diagnostics in the
	// failed-build test, and must survive into the reported error.
	helperCompileErrorText = "simulated-compile-error: undefined: notAFunction"
)

// parseBuildBudget resolves the compile budget from the raw environment value,
// falling back to def when it is unset or empty.
func parseBuildBudget(raw string, def time.Duration) (time.Duration, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return def, nil
	}
	d, err := time.ParseDuration(trimmed)
	if err != nil {
		return 0, fmt.Errorf("%s=%q is not a valid Go duration: %w", buildTimeoutEnv, raw, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s=%q must be a positive duration", buildTimeoutEnv, raw)
	}
	return d, nil
}

// buildBudget returns the effective compile budget for this process.
func buildBudget(t *testing.T) time.Duration {
	t.Helper()

	budget, err := parseBuildBudget(os.Getenv(buildTimeoutEnv), defaultBuildTimeout)
	require.NoError(t, err, "invalid build timeout override")
	return budget
}

// lastBuildBudget records the budget the most recent buildTestBinary call ran
// under, so a test can prove the environment override actually reaches the
// build rather than merely being parseable in isolation.
var lastBuildBudget atomic.Int64

// buildCommandFunc builds the compile command, bound to the budget's context.
type buildCommandFunc func(ctx context.Context) *exec.Cmd

// runBuildWithBudget runs the command returned by newCmd under a finite budget
// and, on failure, returns an error a CI reader can act on: it separates a
// budget overrun from a failed compile, and always reports the elapsed time,
// the effective budget, the context error and whatever output was captured.
//
// The distinction matters because the context kills the child when the deadline
// fires: the run error is then only "signal: killed" and the captured output is
// usually empty, which on its own names neither the deadline nor the budget.
func runBuildWithBudget(budget time.Duration, newCmd buildCommandFunc) error {
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()

	cmd := newCmd(ctx)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if cmd.WaitDelay == 0 {
		// Killing the compiler does not close the output pipes its own children
		// inherited, and with WaitDelay unset those pipes are read until EOF, so
		// Wait can outlive the deadline indefinitely — long enough for the Go
		// test timeout to fire first and destroy the diagnostic again. Bound it.
		cmd.WaitDelay = waitDelayAfterKill
	}

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start).Round(time.Millisecond)
	if runErr == nil {
		return nil
	}

	output := buildOutputDetails(stdout.String(), stderr.String())
	if ctxErr := ctx.Err(); errors.Is(ctxErr, context.DeadlineExceeded) {
		return fmt.Errorf(
			"build hit its %s budget: elapsed %s, ctx.Err()=%v, run error=%v. "+
				"The build was killed at the deadline, so any output below may be truncated or empty; "+
				"if it does contain compiler diagnostics, the build failed on its own just as the budget ran out. "+
				"Otherwise this environment is simply slow (cold module/build cache) — raise the budget with %s=<duration>.%s",
			budget, elapsed, ctxErr, runErr, buildTimeoutEnv, output)
	}

	return fmt.Errorf(
		"build failed after %s, inside its %s budget (ctx.Err()=%v): run error=%v.%s",
		elapsed, budget, ctx.Err(), runErr, output)
}

// buildOutputDetails renders captured child output for the failure message.
func buildOutputDetails(stdout, stderr string) string {
	var b strings.Builder
	if s := strings.TrimSpace(stderr); s != "" {
		b.WriteString("\n--- build stderr ---\n")
		b.WriteString(s)
	}
	if s := strings.TrimSpace(stdout); s != "" {
		b.WriteString("\n--- build stdout ---\n")
		b.WriteString(s)
	}
	if b.Len() == 0 {
		return "\n--- no build output was captured ---"
	}
	return b.String()
}

// buildTestBinary compiles the CLI binary for testing and returns its path.
func buildTestBinary(t *testing.T) string {
	t.Helper()

	// Create a temporary binary path
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "forward-email-test")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	budget := buildBudget(t)
	lastBuildBudget.Store(int64(budget))
	err := runBuildWithBudget(budget, func(ctx context.Context) *exec.Cmd {
		cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, ".")
		cmd.Dir = "." // Build in current directory (cmd/forward-email)
		return cmd
	})
	require.NoError(t, err, "failed to build test binary")

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

// helperCommand returns a command re-running this test binary as a child in the
// given helper mode. It touches nothing outside the test process: no network, no
// configuration, no credentials.
func helperCommand(ctx context.Context, mode string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestBuildHelperProcess$")
	cmd.Env = append(os.Environ(), buildHelperEnv+"="+mode)
	return cmd
}

// TestBuildHelperProcess is not an independent test: it is the child process
// spawned by the build-budget tests, and does nothing unless buildHelperEnv is
// set by its parent.
func TestBuildHelperProcess(t *testing.T) {
	mode := os.Getenv(buildHelperEnv)
	if mode == "" {
		t.Skip("helper process for the build-budget tests; nothing to do when run directly")
	}

	switch mode {
	case "block":
		// Outlive any budget the parent gives us, so the parent's deadline is
		// what ends this process. Bounded so a failure to kill surfaces as a
		// test failure rather than a hang.
		time.Sleep(60 * time.Second)
	case "fail":
		fmt.Fprintln(os.Stderr, helperCompileErrorText)
		os.Exit(2)
	case "succeed":
		// Exit normally.
	default:
		fmt.Fprintf(os.Stderr, "unknown helper mode %q\n", mode)
		os.Exit(3)
	}
}

// TestRunBuildWithBudget_DeadlineExceeded proves the deadline branch produces a
// diagnostic naming the budget, the elapsed time and ctx.Err() — the
// information the old "failed to build test binary: " message destroyed. It uses
// a blocking child and a tiny budget, so it never waits on a real build.
func TestRunBuildWithBudget_DeadlineExceeded(t *testing.T) {
	const budget = 200 * time.Millisecond

	start := time.Now()
	err := runBuildWithBudget(budget, func(ctx context.Context) *exec.Cmd {
		return helperCommand(ctx, "block")
	})
	elapsed := time.Since(start)

	require.Error(t, err, "a build that outlives its budget must fail")
	msg := err.Error()

	assert.Contains(t, msg, "hit its 200ms budget", "must name the budget it exceeded")
	assert.Contains(t, msg, context.DeadlineExceeded.Error(), "must report ctx.Err()")
	assert.Contains(t, msg, "elapsed ", "must report how long the build actually ran")
	assert.Contains(t, msg, buildTimeoutEnv, "must point at the override for slow environments")
	assert.NotContains(t, msg, "build failed after", "must not be reported as a compile failure")

	assert.Less(t, elapsed, 30*time.Second,
		"the deadline must kill the child promptly, not wait for it to exit on its own")
}

// TestRunBuildWithBudget_CompileError proves a real compile failure is reported
// as such — with the compiler output — and is not mislabelled as a timeout.
func TestRunBuildWithBudget_CompileError(t *testing.T) {
	err := runBuildWithBudget(30*time.Second, func(ctx context.Context) *exec.Cmd {
		return helperCommand(ctx, "fail")
	})

	require.Error(t, err, "a non-zero build exit must fail")
	msg := err.Error()

	assert.Contains(t, msg, "build failed after", "must be reported as a build failure")
	assert.Contains(t, msg, helperCompileErrorText, "must surface the captured compiler output")
	assert.NotContains(t, msg, "hit its", "must not be mislabelled as a budget overrun")
	assert.NotContains(t, msg, context.DeadlineExceeded.Error(), "must not claim a deadline fired")
}

// TestRunBuildWithBudget_Success proves the happy path returns no error and does
// not report phantom output.
func TestRunBuildWithBudget_Success(t *testing.T) {
	err := runBuildWithBudget(30*time.Second, func(ctx context.Context) *exec.Cmd {
		return helperCommand(ctx, "succeed")
	})
	assert.NoError(t, err, "a command that exits zero must be reported as a successful build")
}

// TestParseBuildBudget covers the override's parsing, including the values that
// must be rejected rather than silently turned into an unbounded build.
func TestParseBuildBudget(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    time.Duration
		wantErr string
	}{
		{name: "unset uses the documented default", raw: "", want: defaultBuildTimeout},
		{name: "blank uses the documented default", raw: "   ", want: defaultBuildTimeout},
		{name: "override is honoured", raw: "12m", want: 12 * time.Minute},
		{name: "override is trimmed", raw: " 90s ", want: 90 * time.Second},
		{name: "garbage is rejected", raw: "soon", wantErr: "not a valid Go duration"},
		{name: "bare number is rejected", raw: "300", wantErr: "not a valid Go duration"},
		{name: "zero is rejected", raw: "0s", wantErr: "must be a positive duration"},
		{name: "negative is rejected", raw: "-1m", wantErr: "must be a positive duration"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseBuildBudget(tc.raw, defaultBuildTimeout)
			if tc.wantErr != "" {
				require.Error(t, err, "invalid override must not fall back to a default")
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestBuildBudget_EnvOverrideIsReachable proves the override is actually wired
// into the value buildTestBinary uses, not merely parseable in isolation: the
// second half would fail if buildTestBinary stopped consulting buildBudget.
func TestBuildBudget_EnvOverrideIsReachable(t *testing.T) {
	// Neutralise any override the surrounding environment already set, so both
	// halves of this test assert on a known starting point.
	t.Setenv(buildTimeoutEnv, "")
	assert.Equal(t, defaultBuildTimeout, buildBudget(t), "an unset override must yield the default")

	const override = 7*time.Minute + 30*time.Second
	t.Setenv(buildTimeoutEnv, "7m30s")
	assert.Equal(t, override, buildBudget(t), "the environment override must win")

	// The build itself must run under that same value. The compile is a cache
	// hit here, since other tests in this package have already built it.
	binary := buildTestBinary(t)
	defer cleanupTestBinary(t, binary)
	assert.Equal(t, int64(override), lastBuildBudget.Load(),
		"buildTestBinary must build under the overridden budget, not a hard-coded one")
}

// TestBuildTestBinary_HealthyPath proves a successful build still yields a
// usable binary: it exists outside a directory, and runs. Everything it touches
// is a disposable temporary directory — no real account, mail, credential,
// keyring or user configuration is involved.
func TestBuildTestBinary_HealthyPath(t *testing.T) {
	binary := buildTestBinary(t)
	defer cleanupTestBinary(t, binary)

	require.NotEmpty(t, binary, "a successful build must return a binary path")

	info, err := os.Stat(binary)
	require.NoError(t, err, "the returned path must exist")
	require.False(t, info.IsDir(), "the returned path must be a file")
	if runtime.GOOS != "windows" {
		assert.NotZero(t, info.Mode().Perm()&0o111, "the returned file must be executable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfgDir := t.TempDir()
	cmd := exec.CommandContext(ctx, binary, "--help")
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+cfgDir,
		"FORWARDEMAIL_KEYRING_BACKEND=none",
		"FORWARDEMAIL_NO_COLOR=1",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	require.NoError(t, cmd.Run(), "the built binary must run: %s", stderr.String())
	assert.Contains(t, stdout.String(), "Forward Email CLI", "the built binary must be this CLI")
}
