// Package main implements the Forward Email CLI application entry point.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ginsys/forward-email/internal/cmd"
)

// main is the entry point for the Forward Email CLI application.
// It sets up graceful shutdown handling for SIGINT and SIGTERM signals,
// executes the root command with proper context propagation, and ensures
// clean program termination on both success and error conditions.
func main() {
	// Setup graceful shutdown context that responds to SIGINT (Ctrl+C) and SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// Execute the CLI command tree with cancellation support
	err := cmd.Execute(ctx)
	cancel() // Ensure cleanup always happens
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
