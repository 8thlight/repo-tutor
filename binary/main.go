package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"repo-tutor/internal/terminal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := terminal.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "repo-tutor:", err)
		os.Exit(1)
	}
}
