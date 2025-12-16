package errs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RelativePath(path string) string {
	if cwd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(cwd, path); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return path
}

func FormatLocation(path string, line, column int) string {
	path = RelativePath(path)
	if line > 0 {
		if column > 0 {
			return fmt.Sprintf("%s:%d:%d", path, line, column)
		}
		return fmt.Sprintf("%s:%d", path, line)
	}
	return path
}
