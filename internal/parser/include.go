package parser

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/charmbracelet/log"
)

func parseIncludes(rootMap map[string]interface{}, basePath string) (babfile.IncludeMap, error) {
	includesRaw, exists := rootMap[keyIncludes]
	if !exists {
		return nil, nil
	}

	includesMap, ok := safeMapCast(includesRaw)
	if !ok {
		return nil, fmt.Errorf("'includes' must be a map, got %T", includesRaw)
	}

	includes := make(babfile.IncludeMap)
	baseDir := filepath.Dir(basePath)

	for namespace, config := range includesMap {
		configMap, ok := safeMapCast(config)
		if !ok {
			return nil, fmt.Errorf("include %q must be a map with 'babfile' key", namespace)
		}

		babfileRaw, exists := configMap[keyBabfile]
		if !exists {
			return nil, fmt.Errorf("include %q missing required 'babfile' key", namespace)
		}

		babfilePath, err := safeStringCast(babfileRaw)
		if err != nil || babfilePath == "" {
			return nil, fmt.Errorf("include %q has invalid 'babfile' value", namespace)
		}

		if filepath.IsAbs(babfilePath) {
			includes[namespace] = filepath.Clean(babfilePath)
		} else {
			includes[namespace] = filepath.Clean(filepath.Join(baseDir, babfilePath))
		}
	}

	return includes, nil
}

func resolveInclude(namespace, path string, tasks babfile.TaskMap, ctx *babfile.ParseContext) error {
	if ctx.Visited[path] {
		return fmt.Errorf("circular include detected: %s -> %s\nInclude chain: %s",
			ctx.Stack[len(ctx.Stack)-1], path,
			strings.Join(append(ctx.Stack, path), " -> "))
	}

	ctx.Visited[path] = true
	ctx.Stack = append(ctx.Stack, path)
	defer func() { ctx.Stack = ctx.Stack[:len(ctx.Stack)-1] }()

	log.Debug("Resolving include", "namespace", namespace, "path", path)

	includedTasks, err := parseWithContext(path, ctx)
	if err != nil {
		return fmt.Errorf("failed to parse included babfile %q (namespace %q): %w", path, namespace, err)
	}

	return mergeTasks(tasks, includedTasks, namespace)
}

func mergeTasks(parent, included babfile.TaskMap, namespace string) error {
	for name, task := range included {
		prefixedName := namespace + ":" + name

		if _, exists := parent[prefixedName]; exists {
			return fmt.Errorf("task name collision: %q", prefixedName)
		}

		parent[prefixedName] = &babfile.Task{
			Name:         prefixedName,
			Description:  task.Description,
			Commands:     task.Commands,
			Dependencies: task.Dependencies,
		}
	}

	return nil
}
