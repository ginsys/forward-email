package main

import (
	"context"
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
	defer cancel()

	// Execute the CLI command tree with cancellation support
	if err := cmd.Execute(ctx); err != nil {
		cancel() // Ensure cleanup happens before exit
		os.Exit(1)
	}
}
