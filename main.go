package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/bab-sh/bab/cmd"
	"github.com/charmbracelet/log"
)

func main() {
	os.Exit(run())
}

func run() int {
	log.SetDefault(log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    false,
		ReportTimestamp: false,
		Level:           log.InfoLevel,
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Received interrupt signal, shutting down gracefully...")
		cancel()
	}()

	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Error("Execution failed", "error", err)
		return 1
	}

	return 0
}
