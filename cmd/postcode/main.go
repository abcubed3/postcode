package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	cmd "github.com/abcubed3/postcode/cmd/postcode/internal/commands"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cmd.NewRootCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
