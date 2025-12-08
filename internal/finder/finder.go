package finder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

var ErrBabfileNotFound = errors.New("no Babfile found")

var babfileNames = []string{
	"Babfile",
	"Babfile.yaml",
	"Babfile.yml",
	"babfile.yaml",
	"babfile.yml",
}

func FindBabfile() (string, error) {
	log.Debug("Starting Babfile search")

	cwd, err := os.Getwd()
	if err != nil {
		log.Debug("Failed to get working directory", "error", err)
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	log.Debug("Searching for Babfile", "directory", cwd)

	currentDir := cwd
	depth := 0

	for {
		log.Debug("Searching directory", "path", currentDir, "depth", depth)

		for _, filename := range babfileNames {
			path := filepath.Join(currentDir, filename)
			log.Debug("Checking for Babfile", "path", path)
			if _, err := os.Stat(path); err == nil {
				log.Debug("Found Babfile", "path", path)
				return path, nil
			}
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			log.Debug("Reached filesystem root", "root", currentDir, "searched_depth", depth+1)
			break
		}

		currentDir = parentDir
		depth++
	}

	log.Debug("No Babfile found", "cwd", cwd, "searched_depth", depth+1, "tried_names", babfileNames)
	return "", fmt.Errorf("%w in current directory or any parent directories (searched: %s)", ErrBabfileNotFound, cwd)
}
