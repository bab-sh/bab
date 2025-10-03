package cli

import (
	"fmt"
	"os"

	"github.com/bab/bab/internal/display"
	"github.com/bab/bab/internal/executor"
	"github.com/bab/bab/internal/parser"
	"github.com/bab/bab/internal/registry"
	"github.com/bab/bab/pkg/version"
	"github.com/spf13/cobra"
)

var (
	babfile string
	dryRun  bool
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "bab [task]",
	Short: "Simple task runner for your project",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRoot,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&babfile, "file", "f", "", "Path to Babfile (default: ./Babfile)")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be executed without running")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output")
	rootCmd.Version = version.GetVersion()
}

func runRoot(cmd *cobra.Command, args []string) error {
	reg, err := loadRegistry()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return display.ListTasks(reg)
	}

	taskName := args[0]
	task, err := reg.Get(taskName)
	if err != nil {
		fmt.Printf("Task '%s' not found.\n\nRun 'bab' to see available tasks.\n", taskName)
		return err
	}

	exec := executor.New(
		executor.WithDryRun(dryRun),
		executor.WithVerbose(verbose),
	)

	return exec.Execute(task)
}

func loadRegistry() (registry.Registry, error) {
	if babfile == "" {
		babfile = findBabfile()
	}

	if babfile == "" {
		return nil, fmt.Errorf("no Babfile found")
	}

	reg := registry.New()
	p := parser.New(reg)

	if err := p.ParseFile(babfile); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", babfile, err)
	}

	return reg, nil
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
