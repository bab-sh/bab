package executor

import (
	"fmt"
	"strings"
)

func validateCommand(command string) error {
	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("command cannot be only whitespace")
	}
	return nil
}
