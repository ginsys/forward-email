package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ginsys/forward-email/internal/cmd"
)

func main() {
	// Setup graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Execute root command
	if err := cmd.Execute(ctx); err != nil {
		cancel() // Ensure cleanup happens before exit
		os.Exit(1)
	}
}
