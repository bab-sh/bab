package main

import (
	"os"

	"github.com/bab/bab/internal/cli"
	"github.com/fatih/color"
)

func main() {
	if err := cli.Execute(); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}

func init() {
	if os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
}