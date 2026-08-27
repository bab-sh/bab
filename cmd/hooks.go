package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/hooks"
	"github.com/bab-sh/bab/internal/parser"
	"github.com/bab-sh/bab/internal/runner"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

type hookContext struct {
	result   *parser.ParseResult
	hooksDir string
}

func (c *CLI) loadHookContext() (*hookContext, error) {
	result, err := runner.LoadTasks(c.babfile)
	if err != nil {
		return nil, err
	}

	if len(result.Hooks) == 0 {
		return nil, nil
	}

	hooksDir, err := hooks.FindHooksDir()
	if err != nil {
		return nil, err
	}

	return &hookContext{result: result, hooksDir: hooksDir}, nil
}

func (c *CLI) hooksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Manage git hooks",
		Long:  "Install, uninstall, and inspect git hooks defined in the Babfile.",
	}

	cmd.AddCommand(
		c.hooksInstallCommand(),
		c.hooksUninstallCommand(),
		c.hooksStatusCommand(),
		c.hooksRunCommand(),
	)

	return cmd
}

func (c *CLI) hooksInstallCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install git hooks defined in the Babfile",
		Args:  cobra.NoArgs,
		Example: strings.Join([]string{
			"  bab hooks install",
			"  bab hooks install --force",
			"  bab hooks install --dry-run",
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			hctx, err := c.loadHookContext()
			if err != nil {
				return err
			}
			if hctx == nil {
				log.Info("No hooks defined in Babfile")
				return nil
			}

			if c.dryRun {
				names := sortedHookNames(hctx.result.Hooks)
				for _, name := range names {
					fmt.Printf("--- %s ---\n%s\n", name, hooks.GenerateScript(name, hctx.result.Hooks[name]))
				}
				return nil
			}

			if err := hooks.Install(hctx.result.Hooks, hctx.hooksDir, force); err != nil {
				return err
			}

			for _, name := range sortedHookNames(hctx.result.Hooks) {
				log.Info("Installed hook", "name", name)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing non-bab hooks (creates .bak backup)")

	return cmd
}

func (c *CLI) hooksUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove bab-managed git hooks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			hctx, err := c.loadHookContext()
			if err != nil {
				return err
			}
			if hctx == nil {
				log.Info("No hooks defined in Babfile")
				return nil
			}

			if err := hooks.Uninstall(hctx.result.Hooks, hctx.hooksDir); err != nil {
				return err
			}

			for _, name := range sortedHookNames(hctx.result.Hooks) {
				log.Info("Uninstalled hook", "name", name)
			}
			return nil
		},
	}
}

func (c *CLI) hooksStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show status of git hooks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			hctx, err := c.loadHookContext()
			if err != nil {
				return err
			}
			if hctx == nil {
				log.Info("No hooks defined in Babfile")
				return nil
			}

			statuses := hooks.Status(hctx.result.Hooks, hctx.hooksDir)
			sort.Slice(statuses, func(i, j int) bool {
				return statuses[i].Name < statuses[j].Name
			})

			for _, s := range statuses {
				switch {
				case s.Managed:
					log.Info("Hook installed", "name", s.Name)
				case s.Conflict:
					log.Warn("Hook exists (not managed by bab)", "name", s.Name)
				default:
					log.Warn("Hook not installed", "name", s.Name)
				}
			}
			return nil
		},
	}
}

func (c *CLI) hooksRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "run <hook> [args...]",
		Short:  "Execute a hook (called by generated scripts)",
		Args:   cobra.MinimumNArgs(1),
		Hidden: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return c.completeHookNames(toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			hookName := args[0]
			positionalArgs := args[1:]

			r := runner.New(c.dryRun, c.babfile)
			return r.RunHook(c.ctx, hookName, positionalArgs)
		},
	}
}

func (c *CLI) completeHookNames(toComplete string) ([]string, cobra.ShellCompDirective) {
	result, err := runner.LoadTasks(c.babfile)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for name := range result.Hooks {
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}
	sort.Strings(completions)
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func sortedHookNames(hm babfile.HookMap) []string {
	names := make([]string, 0, len(hm))
	for name := range hm {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
