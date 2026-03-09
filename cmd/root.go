package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/runner"
	"github.com/bab-sh/bab/internal/update"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var argNameRegex = regexp.MustCompile(babfile.VarNamePattern)

var (
	versionString = "dev"
	versionShort  = "dev"
)

func SetVersionInfo(version, commit, date string) {
	versionShort = version
	versionString = version
	if commit != "none" {
		versionString = fmt.Sprintf("%s\n  commit: %s\n  built:  %s", version, commit, date)
	}
}

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

	update.StartBackgroundRefresh(ctx, versionShort)

	err := c.buildCommand().Execute()

	if info := update.CheckCached(versionShort); info != nil {
		log.Warn("A new version of bab is available",
			"latest", info.LatestVersion,
			"current", info.CurrentVersion)
	}

	return err
}

func (c *CLI) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "bab [task]",
		Short:             "Clean commands for any project.",
		Version:           versionString,
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
		return c.runTask(args[0], args[1:])
	}
	return c.runInteractive()
}

func (c *CLI) runTask(taskName string, rawArgs []string) error {
	r := runner.New(c.dryRun, c.babfile)
	if len(rawArgs) > 0 {
		cliArgs, err := parseKVArgs(rawArgs)
		if err != nil {
			return err
		}
		r.CLIArgs = cliArgs
	}
	return r.Run(c.ctx, taskName)
}

func parseKVArgs(args []string) (map[string]string, error) {
	result := make(map[string]string, len(args))
	for _, arg := range args {
		idx := strings.Index(arg, "=")
		if idx <= 0 {
			return nil, fmt.Errorf("invalid argument %q: expected format key=value", arg)
		}
		key := arg[:idx]
		if !argNameRegex.MatchString(key) {
			return nil, fmt.Errorf("invalid argument name %q: must match pattern %s", key, babfile.VarNamePattern)
		}
		result[key] = arg[idx+1:]
	}
	return result, nil
}
