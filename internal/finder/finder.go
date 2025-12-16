package finder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bab-sh/bab/internal/errs"
	"github.com/charmbracelet/log"
)

var babfileNames = []string{
	"Babfile",
	"Babfile.yaml",
	"Babfile.yml",
	"babfile.yaml",
	"babfile.yml",
}

func FindBabfile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	currentDir := cwd
	for {
		for _, filename := range babfileNames {
			path := filepath.Join(currentDir, filename)
			if _, err := os.Stat(path); err == nil {
				log.Debug("Found Babfile", "path", path)
				return path, nil
			}
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}

	log.Debug("Babfile not found", "searched", cwd)
	return "", fmt.Errorf("%w in current directory or any parent directories", errs.ErrBabfileNotFound)
}
