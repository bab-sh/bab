package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/bab-sh/bab/cmd"
	"github.com/bab-sh/bab/internal/errs"
	"github.com/charmbracelet/log"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func isCancellation(err error) bool {
	return errors.Is(err, context.Canceled)
}

func main() {
	os.Exit(run())
}

func run() int {
	log.SetDefault(log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    false,
		ReportTimestamp: false,
		Level:           log.InfoLevel,
	}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd.SetVersionInfo(version, commit, date)
	if err := cmd.ExecuteContext(ctx); err != nil {
		if isCancellation(err) {
			return 0
		}
		handleError(err)
		return 1
	}

	return 0
}

func handleError(err error) {
	var verrs *errs.ValidationErrors
	if errors.As(err, &verrs) {
		for _, e := range verrs.Errors {
			log.Error(e.Error())
		}
		return
	}

	if msg := err.Error(); msg != "" {
		log.Error(msg)
	}
}
