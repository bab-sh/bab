package parser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

func Parse(path string) (TaskMap, error) {
	if strings.TrimSpace(path) == "" {
		return nil, &ParseError{Path: path, Message: "path cannot be empty"}
	}

	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return nil, &ParseError{Path: path, Message: "invalid path", Cause: err}
	}

	visited := make(map[string]bool)
	tasks, err := parseFile(absPath, visited)
	if err != nil {
		return nil, err
	}

	if err := validateDependencies(tasks); err != nil {
		return nil, err
	}

	if err := validateRunTaskRefs(tasks); err != nil {
		return nil, err
	}

	if err := validateRunCycles(tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func parseFile(absPath string, visited map[string]bool) (TaskMap, error) {
	log.Debug("Parsing babfile", "path", absPath)

	if visited[absPath] {
		chain := chainFromVisited(visited, absPath)
		return nil, &CircularError{Type: "include", Chain: chain}
	}
	visited[absPath] = true
	defer delete(visited, absPath)

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, &ParseError{Path: absPath, Message: "failed to read file", Cause: err}
	}

	bf, err := unmarshalBabfile(data)
	if err != nil {
		return nil, &ParseError{Path: absPath, Message: "invalid YAML", Cause: err}
	}

	tasks := make(TaskMap, len(bf.Tasks))
	for name, task := range bf.Tasks {
		tasks[name] = convertTask(name, task)
	}

	baseDir := filepath.Dir(absPath)
	for namespace, inc := range bf.Includes {
		if err := resolveInclude(namespace, inc.Babfile, baseDir, tasks, visited); err != nil {
			return nil, &ParseError{Path: absPath, Message: "include " + namespace + " failed", Cause: err}
		}
	}

	log.Debug("Parsed babfile", "path", absPath, "tasks", len(tasks))
	return tasks, nil
}

func chainFromVisited(visited map[string]bool, current string) []string {
	chain := make([]string, 0, len(visited)+1)
	for path := range visited {
		chain = append(chain, filepath.Base(path))
	}
	chain = append(chain, filepath.Base(current))
	return chain
}
