package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nShutting down (interrupt again to force quit)...")
		cancel()

		<-sigCh
		fmt.Fprintln(os.Stderr, "Force quit.")
		os.Exit(1)
	}()

	return ctx, func() {
		signal.Stop(sigCh)
		cancel()
	}
}
