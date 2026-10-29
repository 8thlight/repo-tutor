package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"repo-tutor/internal/terminal"
)

const signalExitBase = 128

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	received := make(chan syscall.Signal, 1)
	go func() {
		received <- (<-signals).(syscall.Signal)
		cancel()
	}()
	err := terminal.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	select {
	case stopping := <-received:
		os.Exit(signalExitBase + int(stopping))
	default:
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "repo-tutor:", err)
		os.Exit(1)
	}
}
