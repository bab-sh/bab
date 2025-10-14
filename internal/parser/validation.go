package parser

import (
	"fmt"
	"strings"
)

func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path cannot be only whitespace")
	}
	return nil
}

func validateCommand(command string) error {
	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("command cannot be only whitespace")
	}
	return nil
}
