package cmd

import (
	"fmt"

	"github.com/bab/bab/internal/compiler"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func newCompileCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "compile",
		Short: "Compile Babfile to standalone shell scripts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompile(outputDir)
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for generated scripts")

	return cmd
}

func runCompile(outputDir string) error {
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
	)

	return c.Compile()
}
