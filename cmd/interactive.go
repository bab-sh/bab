package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bab-sh/bab/internal/runner"
	"github.com/bab-sh/bab/internal/tui"
	"github.com/charmbracelet/log"
)

func (c *CLI) runInteractive() error {
	result, err := runner.LoadTasks(c.babfile)
	if err != nil {
		return err
	}

	selected, err := tui.PickTask(c.ctx, result.Tasks)
	if err != nil {
		return err
	}

	if selected == nil {
		log.Debug("No task selected")
		return nil
	}

	if selected.Args != nil {
		var required []string
		for argName, def := range selected.Args {
			if def.Default == nil {
				required = append(required, argName)
			}
		}
		if len(required) > 0 {
			sort.Strings(required)
			kvPairs := make([]string, len(required))
			for i, name := range required {
				kvPairs[i] = name + "=<value>"
			}
			return fmt.Errorf("task %q requires arguments: %s — run with: bab %s %s",
				selected.Name, strings.Join(required, ", "), selected.Name, strings.Join(kvPairs, " "))
		}
	}

	log.Debug("Running task", "name", selected.Name)
	return c.runTask(selected.Name, nil)
}
