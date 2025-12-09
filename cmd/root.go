package cmd

import (
	"context"

	"github.com/bab-sh/bab/internal/runner"
	"github.com/bab-sh/bab/internal/version"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

type CLI struct {
	ctx        context.Context
	verbose    bool
	dryRun     bool
	listTasks  bool
	validate   bool
	completion string
	babfile    string
}

func ExecuteContext(ctx context.Context) error {
	return newCLI().execute(ctx)
}

func newCLI() *CLI {
	return &CLI{}
}

func (c *CLI) execute(ctx context.Context) error {
	c.ctx = ctx
	return c.buildCommand().Execute()
}

func (c *CLI) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "bab [task]",
		Short:             "Custom commands for every project",
		Version:           version.Version,
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: completeTaskNames,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if c.verbose {
				log.SetLevel(log.DebugLevel)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run(cmd, args)
		},
		Args: cobra.ArbitraryArgs,
	}

	cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.PersistentFlags().BoolVarP(&c.verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVarP(&c.dryRun, "dry-run", "n", false, "Show commands without executing")
	cmd.PersistentFlags().StringVarP(&c.babfile, "babfile", "b", "", "Path to Babfile")
	cmd.Flags().BoolVarP(&c.listTasks, "list", "l", false, "List all available tasks")
	cmd.Flags().BoolVar(&c.validate, "validate", false, "Validate the Babfile without executing tasks")
	cmd.Flags().StringVarP(&c.completion, "completion", "c", "", "Generate completion script (bash|zsh|fish|powershell)")

	return cmd
}

func (c *CLI) run(cmd *cobra.Command, args []string) error {
	if c.completion != "" {
		return c.runCompletion(cmd)
	}
	if c.validate {
		return c.runValidate()
	}
	if c.listTasks {
		return c.runList()
	}
	if len(args) > 0 {
		return c.runTask(args[0])
	}
	return c.runInteractive()
}

func (c *CLI) runTask(taskName string) error {
	r := runner.New(c.dryRun, c.babfile)
	return r.Run(c.ctx, taskName)
}
