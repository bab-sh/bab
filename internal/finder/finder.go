package finder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

var babfileNames = []string{
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

	for _, filename := range babfileNames {
		path := filepath.Join(cwd, filename)
		log.Debug("Checking for Babfile", "path", path)
		if _, err := os.Stat(path); err == nil {
			log.Debug("Found Babfile", "path", path)
			return path, nil
		}
	}

	log.Debug("No Babfile found", "directory", cwd, "tried", babfileNames)
	return "", fmt.Errorf("no Babfile found in current directory (%s)", cwd)
}
