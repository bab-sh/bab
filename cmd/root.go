package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bab-sh/bab/internal/executor"
	"github.com/bab-sh/bab/internal/parser"
	"github.com/bab-sh/bab/internal/registry"
	"github.com/bab-sh/bab/internal/tui"
	"github.com/bab-sh/bab/internal/version"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	babfile string
	dryRun  bool
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "bab [task]",
	Short: "Interactive task runner for your project",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRoot,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&babfile, "file", "f", "", "Path to Babfile (default: ./Babfile)")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be executed without running")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output")
	rootCmd.Version = version.GetVersion()

	rootCmd.AddCommand(newListCmd())
}

func runRoot(_ *cobra.Command, args []string) error {
	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	reg, babfilePath, err := loadRegistry()
	if err != nil {
		return err
	}

	// Get project root (directory containing the Babfile)
	projectRoot := filepath.Dir(babfilePath)

	// If no task specified, launch interactive TUI
	if len(args) == 0 {
		return tui.Run(reg, projectRoot, dryRun, verbose)
	}

	// Execute the specified task directly
	taskName := args[0]
	task, err := reg.Get(taskName)
	if err != nil {
		log.Error("Task not found", "task", taskName)
		log.Info("Run 'bab' to see available tasks")
		log.Info("Run 'bab list' for a non-interactive list")
		return err
	}

	exec := executor.New(
		executor.WithDryRun(dryRun),
		executor.WithVerbose(verbose),
		executor.WithProjectRoot(projectRoot),
	)

	return exec.Execute(task)
}

func loadRegistry() (registry.Registry, string, error) {
	if babfile == "" {
		babfile = findBabfile()
	}

	if babfile == "" {
		log.Error("No Babfile found in current directory")
		log.Info("Looking for: Babfile, Babfile.yaml, Babfile.yml")
		return nil, "", fmt.Errorf("no Babfile found")
	}

	// Get absolute path to the Babfile
	absPath, err := filepath.Abs(babfile)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	reg := registry.New()
	p := parser.New(reg)

	if err := p.ParseFile(absPath); err != nil {
		log.Error("Failed to parse Babfile", "path", absPath, "error", err)
		return nil, "", err
	}

	return reg, absPath, nil
}

func findBabfile() string {
	candidates := []string{
		"Babfile",
		"Babfile.yaml",
		"Babfile.yml",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}
