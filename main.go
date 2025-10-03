package main

import (
	"os"

	"github.com/bab-sh/bab/cmd"
	"github.com/charmbracelet/log"
)

func main() {
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    false,
		ReportTimestamp: false,
		Level:           log.InfoLevel,
	})

	if os.Getenv("NO_COLOR") != "" {
		styles := log.DefaultStyles()
		for i := range styles.Levels {
			styles.Levels[i] = styles.Levels[i].UnsetBackground().UnsetForeground().UnsetBold()
		}
		styles.Key = styles.Key.UnsetForeground()
		styles.Value = styles.Value.UnsetForeground()
		logger.SetStyles(styles)
	}

	log.SetDefault(logger)

	if err := cmd.Execute(); err != nil {
		log.Fatal("Error", "err", err)
	}
}
