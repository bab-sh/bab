package main

import (
	"os"

	"github.com/bab-sh/bab/cmd"
	"github.com/charmbracelet/log"
)

func main() {
	log.SetDefault(log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    false,
		ReportTimestamp: false,
		Level:           log.InfoLevel,
	}))

	cmd.Execute()
}
