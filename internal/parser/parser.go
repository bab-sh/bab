package parser

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/errs"
	"github.com/charmbracelet/log"
)

type ParseResult struct {
	Path         string
	GlobalVars   babfile.VarMap
	GlobalEnv    map[string]string
	GlobalSilent *bool
	GlobalOutput *bool
	GlobalDir    string
	Tasks        babfile.TaskMap
}

func Parse(path string) (*ParseResult, error) {
	if strings.TrimSpace(path) == "" {
		return nil, &errs.ParseError{Path: path, Message: "path cannot be empty", Cause: errs.ErrPathEmpty}
	}

	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return nil, &errs.ParseError{Path: path, Message: "invalid path", Cause: err}
	}

	visited := make(map[string]bool)
	result, err := parseFile(absPath, visited)
	if err != nil {
		return nil, err
	}

	if err := validateAll(absPath, result.Tasks); err != nil {
		return nil, err
	}

	return result, nil
}

func parseFile(absPath string, visited map[string]bool) (*ParseResult, error) {
	log.Debug("Parsing babfile", "path", absPath)

	if visited[absPath] {
		chain := chainFromVisited(visited, absPath)
		return nil, &errs.CircularDepError{Path: absPath, Type: "include", Chain: chain}
	}
	visited[absPath] = true
	defer delete(visited, absPath)

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, &errs.ParseError{Path: absPath, Message: "file not found", Cause: err}
	}

	bf, err := unmarshalBabfile(absPath, data)
	if err != nil {
		var verrs *errs.ValidationErrors
		if errors.As(err, &verrs) {
			return nil, verrs
		}

		var parseErr *errs.ParseError
		if errors.As(err, &parseErr) {
			return nil, parseErr
		}

		line := errs.ExtractYAMLLocation(err)
		cleanMsg := errs.CleanYAMLError(err)
		if cleanMsg == "" {
			cleanMsg = err.Error()
		}
		return nil, &errs.ParseError{
			Path:    absPath,
			Line:    line,
			Message: "invalid YAML syntax",
			Cause:   errors.New(cleanMsg),
		}
	}

	tasks := make(babfile.TaskMap, len(bf.Tasks))
	for name, task := range bf.Tasks {
		task.Name = name
		task.SourcePath = absPath
		tasks[name] = &task
	}

	baseDir := filepath.Dir(absPath)
	for namespace, inc := range bf.Includes {
		if err := resolveInclude(namespace, inc.Babfile, baseDir, tasks, visited); err != nil {
			return nil, &errs.ParseError{Path: absPath, Message: "include " + namespace + " failed", Cause: err}
		}
	}

	log.Debug("Parsed babfile", "path", absPath, "tasks", len(tasks))
	return &ParseResult{
		Path:         absPath,
		GlobalVars:   bf.Vars,
		GlobalEnv:    bf.Env,
		GlobalSilent: bf.Silent,
		GlobalOutput: bf.Output,
		GlobalDir:    bf.Dir,
		Tasks:        tasks,
	}, nil
}

func chainFromVisited(visited map[string]bool, current string) []string {
	chain := make([]string, 0, len(visited)+1)
	for path := range visited {
		chain = append(chain, filepath.Base(path))
	}
	chain = append(chain, filepath.Base(current))
	return chain
}
