package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (c *CLI) runCompletion(cmd *cobra.Command) error {
	switch c.completion {
	case "bash":
		return cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	default:
		return fmt.Errorf("invalid shell %q: must be one of: bash, zsh, fish, powershell", c.completion)
	}
}
