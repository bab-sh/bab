package parser

import (
	"fmt"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/validation"
	"github.com/charmbracelet/log"
)

func processTaskNode(taskMap map[string]interface{}, taskName string, tasks babfile.TaskMap) error {
	runCmd, hasRun := taskMap[keyRun]

	if hasRun {
		task, err := buildTask(taskName, taskMap, runCmd)
		if err != nil {
			return err
		}
		tasks[taskName] = task
		log.Debug("Registered task", "name", taskName, "commands", len(task.Commands))
	}

	nestedKeys := getNestedKeys(taskMap)
	if len(nestedKeys) > 0 {
		if hasRun {
			nonDefaultKeys := make([]string, 0, len(nestedKeys))
			for _, k := range nestedKeys {
				if k != keyDefault {
					nonDefaultKeys = append(nonDefaultKeys, k)
				}
			}
			if len(nonDefaultKeys) > 0 {
				return fmt.Errorf("task %q cannot have both 'run' block and nested subtasks %v; use a 'default' subtask instead", taskName, nonDefaultKeys)
			}
		}
		log.Debug("Node has nested tasks, recursing", "name", taskName)
		for _, key := range nestedKeys {
			nestedMap := map[string]interface{}{key: taskMap[key]}
			if err := flatten(nestedMap, taskName, tasks); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildTask(name string, taskMap map[string]interface{}, runCmd interface{}) (*babfile.Task, error) {
	task := &babfile.Task{Name: name}

	if desc, ok := taskMap[keyDesc]; ok {
		descStr, err := safeStringCast(desc)
		if err != nil {
			return nil, fmt.Errorf("invalid description for task %q: %w", name, err)
		}
		task.Description = descStr
		log.Debug("Task has description", "name", name, "desc", task.Description)
	}

	if deps, ok := taskMap[keyDeps]; ok {
		dependencies, err := parseDependencies(name, deps)
		if err != nil {
			return nil, err
		}
		task.Dependencies = dependencies
		log.Debug("Task has dependencies", "name", name, "deps", dependencies)
	}

	commands, err := parseCommands(name, runCmd)
	if err != nil {
		return nil, err
	}
	task.Commands = commands

	log.Debug("Found executable task", "name", name, "commands", len(commands))
	return task, nil
}

func parseCommands(taskName string, runCmd interface{}) ([]babfile.Command, error) {
	if runCmd == nil {
		return nil, fmt.Errorf("task %q has nil 'run' field", taskName)
	}

	cmdSlice, ok := safeSliceCast(runCmd)
	if !ok {
		return nil, fmt.Errorf("task %q 'run' must be a list of commands", taskName)
	}

	if err := validation.ValidateNonEmptySlice(cmdSlice, fmt.Sprintf("task %q 'run' field", taskName)); err != nil {
		return nil, err
	}

	commands := make([]babfile.Command, 0, len(cmdSlice))
	for i, item := range cmdSlice {
		cmd, err := parseCommandObject(taskName, i, item)
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}

	log.Debug("Parsed commands", "task", taskName, "count", len(commands))
	return commands, nil
}

func parseCommandObject(taskName string, index int, item interface{}) (babfile.Command, error) {
	cmdMap, ok := safeMapCast(item)
	if !ok {
		return babfile.Command{}, fmt.Errorf("task %q command at index %d must be an object with 'cmd' field", taskName, index)
	}

	cmd := babfile.Command{}

	cmdRaw, hasCmd := cmdMap[keyCmd]
	if !hasCmd {
		return babfile.Command{}, fmt.Errorf("task %q command at index %d missing required 'cmd' field", taskName, index)
	}

	cmdStr, err := safeStringCast(cmdRaw)
	if err != nil {
		return babfile.Command{}, fmt.Errorf("task %q command at index %d has invalid 'cmd': %w", taskName, index, err)
	}

	if err := validation.ValidateCommand(cmdStr); err != nil {
		return babfile.Command{}, fmt.Errorf("task %q command at index %d has invalid 'cmd': %w", taskName, index, err)
	}
	cmd.Cmd = cmdStr

	if platformsRaw, hasPlatforms := cmdMap[keyPlatforms]; hasPlatforms {
		platforms, err := parsePlatforms(taskName, index, platformsRaw)
		if err != nil {
			return babfile.Command{}, err
		}
		cmd.Platforms = platforms
	}

	return cmd, nil
}

func parsePlatforms(taskName string, index int, raw interface{}) ([]babfile.Platform, error) {
	platformStrings, ok := safeStringSliceCast(raw)
	if !ok {
		return nil, fmt.Errorf("task %q command at index %d 'platforms' must be a list of strings", taskName, index)
	}

	platforms := make([]babfile.Platform, 0, len(platformStrings))
	for _, p := range platformStrings {
		if err := validation.ValidatePlatform(p); err != nil {
			return nil, fmt.Errorf("task %q command at index %d: %w", taskName, index, err)
		}
		platforms = append(platforms, babfile.Platform(p))
	}

	return platforms, nil
}

func parseDependencies(taskName string, depsValue interface{}) ([]string, error) {
	if depsValue == nil {
		return nil, fmt.Errorf("task %q has nil 'deps' value", taskName)
	}

	if depStr, ok := depsValue.(string); ok {
		if err := validation.ValidateDependencyName(depStr, 0, taskName); err != nil {
			return nil, err
		}
		log.Debug("Task has single dependency", "name", taskName, "dep", depStr)
		return []string{depStr}, nil
	}

	if depSlice, ok := safeSliceCast(depsValue); ok {
		if err := validation.ValidateNonEmptySlice(depSlice, fmt.Sprintf("task %q 'deps' list", taskName)); err != nil {
			return nil, err
		}
		dependencies := make([]string, 0, len(depSlice))
		for i, dep := range depSlice {
			depStr, err := safeStringCast(dep)
			if err != nil {
				return nil, fmt.Errorf("task %q has invalid dependency at index %d: %w", taskName, i, err)
			}
			if err := validation.ValidateDependencyName(depStr, i, taskName); err != nil {
				return nil, err
			}
			dependencies = append(dependencies, depStr)
		}
		log.Debug("Task has multiple dependencies", "name", taskName, "count", len(depSlice))
		return dependencies, nil
	}

	depStr, err := safeStringCast(depsValue)
	if err != nil {
		return nil, fmt.Errorf("task %q has invalid 'deps' value: %w", taskName, err)
	}
	if err := validation.ValidateDependencyName(depStr, 0, taskName); err != nil {
		return nil, err
	}
	log.Debug("Task has dependency of unknown type, converted to string", "name", taskName)
	return []string{depStr}, nil
}
