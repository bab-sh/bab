package finder

import (
	"fmt"
	"os"
	"path/filepath"

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
	log.Debug("Starting Babfile search")

	cwd, err := os.Getwd()
	if err != nil {
		log.Debug("Failed to get working directory", "error", err)
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	log.Debug("Searching for Babfile", "directory", cwd)

	currentDir := cwd
	searchedDirs := []string{}

	for {
		log.Debug("Searching directory", "path", currentDir)
		searchedDirs = append(searchedDirs, currentDir)

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
			log.Debug("Reached filesystem root", "root", currentDir)
			break
		}

		currentDir = parentDir
	}

	log.Debug("No Babfile found", "searched", searchedDirs, "tried-names", babfileNames)
	return "", fmt.Errorf("no Babfile found in current directory or any parent directories (searched: %s)", cwd)
}
