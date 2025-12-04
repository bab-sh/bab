package validation

import (
	"fmt"
	"strings"

	"github.com/bab-sh/bab/internal/babfile"
)

func ValidateString(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be only whitespace", fieldName)
	}
	return nil
}

func ValidateCommand(command string) error {
	return ValidateString(command, "command")
}

func ValidatePath(path string) error {
	return ValidateString(path, "path")
}

func ValidateNonEmptySlice[T any](slice []T, fieldName string) error {
	if len(slice) == 0 {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	return nil
}

func ValidateDependencyName(dep string, index int, taskName string) error {
	if dep == "" {
		return fmt.Errorf("task %q has empty dependency at index %d", taskName, index)
	}
	return nil
}

// ValidPlatforms is kept for backwards compatibility but delegates to babfile package.
var ValidPlatforms = babfile.ValidPlatformStrings()

func ValidatePlatform(platform string) error {
	if babfile.IsValidPlatform(platform) {
		return nil
	}
	return fmt.Errorf("invalid platform %q (valid: %s)", platform, strings.Join(ValidPlatforms, ", "))
}
