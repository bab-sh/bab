package cli

import (
	"fmt"

	"github.com/bab/bab/internal/compiler"
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
	babfilePath := findBabfile()
	if babfilePath == "" {
		return fmt.Errorf("no Babfile found")
	}

	fmt.Printf("Using Babfile: %s\n", babfilePath)
	fmt.Printf("Output directory: %s\n", outputDir)
	fmt.Println("\nCompiling Babfile to scripts...")

	c := compiler.New(babfilePath,
		compiler.WithOutputDir(outputDir),
		compiler.WithVerbose(verbose),
	)

	return c.Compile()
}
