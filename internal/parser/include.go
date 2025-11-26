package parser

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

func parseIncludes(rootMap map[string]interface{}, basePath string) (IncludeMap, error) {
	includesRaw, exists := rootMap[keyIncludes]
	if !exists {
		return nil, nil
	}

	includesMap, ok := safeMapCast(includesRaw)
	if !ok {
		return nil, fmt.Errorf("'includes' must be a map, got %T", includesRaw)
	}

	includes := make(IncludeMap)
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

func resolveInclude(namespace, path string, tasks TaskMap, ctx *ParseContext) error {
	if ctx.visited[path] {
		return fmt.Errorf("circular include detected: %s -> %s\nInclude chain: %s",
			ctx.stack[len(ctx.stack)-1], path,
			strings.Join(append(ctx.stack, path), " -> "))
	}

	ctx.visited[path] = true
	ctx.stack = append(ctx.stack, path)
	defer func() { ctx.stack = ctx.stack[:len(ctx.stack)-1] }()

	log.Debug("Resolving include", "namespace", namespace, "path", path)

	includedTasks, err := parseWithContext(path, ctx)
	if err != nil {
		return fmt.Errorf("failed to parse included babfile %q (namespace %q): %w", path, namespace, err)
	}

	return mergeTasks(tasks, includedTasks, namespace)
}

func mergeTasks(parent, included TaskMap, namespace string) error {
	for name, task := range included {
		prefixedName := namespace + ":" + name

		if _, exists := parent[prefixedName]; exists {
			return fmt.Errorf("task name collision: %q", prefixedName)
		}

		parent[prefixedName] = &Task{
			Name:         prefixedName,
			Description:  task.Description,
			Commands:     task.Commands,
			Dependencies: task.Dependencies,
		}
	}

	return nil
}
