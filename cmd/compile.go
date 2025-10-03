package cmd

import (
	"fmt"

	"github.com/bab/bab/internal/compiler"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func newCompileCmd() *cobra.Command {
	var outputDir string
	var noColor bool

	cmd := &cobra.Command{
		Use:   "compile",
		Short: "Compile Babfile to standalone shell scripts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompile(outputDir, noColor)
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for generated scripts")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colors in generated scripts")

	return cmd
}

func runCompile(outputDir string, noColor bool) error {
	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	babfilePath := findBabfile()
	if babfilePath == "" {
		return fmt.Errorf("no Babfile found")
	}

	log.Info("Using Babfile", "path", babfilePath)
	log.Info("Output directory", "path", outputDir)
	log.Info("Compiling Babfile to scripts")

	c := compiler.New(babfilePath,
		compiler.WithOutputDir(outputDir),
		compiler.WithVerbose(verbose),
		compiler.WithNoColor(noColor),
	)

	return c.Compile()
}
