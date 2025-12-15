package parser

import (
	"errors"
	"fmt"

	"github.com/bab-sh/bab/internal/babfile"
	"gopkg.in/yaml.v3"
)

const (
	keyCmd       = "cmd"
	keyDeps      = "deps"
	keyDesc      = "desc"
	keyEnv       = "env"
	keyIncludes  = "includes"
	keyLevel     = "level"
	keyLog       = "log"
	keyPlatforms = "platforms"
	keyRun       = "run"
	keyTask      = "task"
	keyTasks     = "tasks"
)

func parseEnvMap(path string, node *yaml.Node, env *map[string]string) error {
	if node.Kind != yaml.MappingNode {
		return &ParseError{Path: path, Line: node.Line, Message: "env must be a mapping"}
	}

	*env = make(map[string]string, len(node.Content)/2)

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]

		if valNode.Kind != yaml.ScalarNode {
			return &ParseError{
				Path:    path,
				Line:    valNode.Line,
				Message: fmt.Sprintf("env value for %q must be a string", keyNode.Value),
			}
		}

		(*env)[keyNode.Value] = valNode.Value
	}

	return nil
}

func unmarshalBabfile(path string, data []byte) (*babfile.Schema, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return nil, &ParseError{Path: path, Message: "invalid document"}
	}

	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return nil, &ParseError{Path: path, Line: doc.Line, Message: "expected mapping at root"}
	}

	schema := &babfile.Schema{
		Env:      make(map[string]string),
		Tasks:    make(map[string]babfile.Task),
		Includes: make(map[string]babfile.Include),
	}

	for i := 0; i < len(doc.Content); i += 2 {
		key := doc.Content[i]
		val := doc.Content[i+1]

		switch key.Value {
		case keyEnv:
			if err := parseEnvMap(path, val, &schema.Env); err != nil {
				return nil, err
			}
		case keyTasks:
			if err := parseTasks(path, val, schema); err != nil {
				return nil, err
			}
		case keyIncludes:
			if err := parseIncludes(path, val, schema); err != nil {
				return nil, err
			}
		}
	}

	return schema, nil
}

func parseTasks(path string, node *yaml.Node, schema *babfile.Schema) error {
	if node.Kind != yaml.MappingNode {
		return &ParseError{Path: path, Line: node.Line, Message: "tasks must be a mapping"}
	}

	taskLines := make(map[string]int)

	for i := 0; i < len(node.Content); i += 2 {
		nameNode := node.Content[i]
		taskNode := node.Content[i+1]
		name := nameNode.Value

		if origLine, exists := taskLines[name]; exists {
			return &DuplicateError{
				Line:         nameNode.Line,
				TaskName:     name,
				OriginalLine: origLine,
			}
		}
		taskLines[name] = nameNode.Line

		task, err := parseTask(path, taskNode)
		if err != nil {
			var parseErr *ParseError
			if errors.As(err, &parseErr) {
				parseErr.Message = fmt.Sprintf("task %q: %s", name, parseErr.Message)
				return parseErr
			}
			return &ParseError{Path: path, Line: nameNode.Line, Message: fmt.Sprintf("task %q", name), Cause: err}
		}
		task.Line = nameNode.Line
		schema.Tasks[name] = task
	}

	return nil
}

func parseTask(path string, node *yaml.Node) (babfile.Task, error) {
	if node.Kind != yaml.MappingNode {
		return babfile.Task{}, &ParseError{Path: path, Line: node.Line, Message: "task must be a mapping"}
	}

	task := babfile.Task{}

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]

		switch key.Value {
		case keyDesc:
			task.Desc = val.Value
		case keyEnv:
			if err := parseEnvMap(path, val, &task.Env); err != nil {
				return babfile.Task{}, err
			}
		case keyDeps:
			task.DepsLine = key.Line
			if err := val.Decode(&task.Deps); err != nil {
				return babfile.Task{}, &ParseError{Path: path, Line: key.Line, Message: "invalid deps", Cause: err}
			}
		case keyRun:
			runItems, err := parseRunItems(path, val)
			if err != nil {
				return babfile.Task{}, err
			}
			task.Run = runItems
		}
	}

	return task, nil
}

func parseRunItems(path string, node *yaml.Node) ([]babfile.RunItem, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, &ParseError{Path: path, Line: node.Line, Message: "run must be a sequence"}
	}

	items := make([]babfile.RunItem, 0, len(node.Content))
	for i, itemNode := range node.Content {
		item, err := parseRunItem(path, itemNode)
		if err != nil {
			var parseErr *ParseError
			if errors.As(err, &parseErr) {
				parseErr.Message = fmt.Sprintf("run[%d]: %s", i, parseErr.Message)
				return nil, parseErr
			}
			return nil, &ParseError{Path: path, Line: itemNode.Line, Message: fmt.Sprintf("run[%d]", i), Cause: err}
		}
		items = append(items, item)
	}

	return items, nil
}

func parseRunItem(path string, node *yaml.Node) (babfile.RunItem, error) {
	if node.Kind != yaml.MappingNode {
		return nil, &ParseError{Path: path, Line: node.Line, Message: "run item must be a mapping"}
	}

	var cmd, task, log string
	var env map[string]string
	var platforms []babfile.Platform
	var level babfile.LogLevel
	line := node.Line

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]

		switch key.Value {
		case keyCmd:
			cmd = val.Value
		case keyTask:
			task = val.Value
		case keyLog:
			log = val.Value
		case keyLevel:
			level = babfile.LogLevel(val.Value)
		case keyEnv:
			if err := parseEnvMap(path, val, &env); err != nil {
				return nil, err
			}
		case keyPlatforms:
			if err := val.Decode(&platforms); err != nil {
				return nil, &ParseError{Path: path, Line: key.Line, Message: "invalid platforms", Cause: err}
			}
		}
	}

	hasCmd := cmd != ""
	hasTask := task != ""
	hasLog := log != ""

	count := 0
	if hasCmd {
		count++
	}
	if hasTask {
		count++
	}
	if hasLog {
		count++
	}

	switch {
	case count > 1:
		return nil, &ParseError{Path: path, Line: node.Line, Message: "run item can only have one of 'cmd', 'task', or 'log'"}
	case hasCmd:
		return babfile.CommandRun{Line: line, Cmd: cmd, Env: env, Platforms: platforms}, nil
	case hasTask:
		return babfile.TaskRun{Line: line, Task: task, Platforms: platforms}, nil
	case hasLog:
		if level == "" {
			level = babfile.LogLevelInfo
		}
		if !level.Valid() {
			return nil, &ParseError{
				Path:    path,
				Line:    node.Line,
				Message: fmt.Sprintf("invalid log level %q, must be one of: debug, info, warn, error", level),
			}
		}
		return babfile.LogRun{Line: line, Log: log, Level: level, Platforms: platforms}, nil
	default:
		return nil, &ParseError{Path: path, Line: node.Line, Message: "run item must have 'cmd', 'task', or 'log'"}
	}
}

func parseIncludes(path string, node *yaml.Node, schema *babfile.Schema) error {
	if node.Kind != yaml.MappingNode {
		return &ParseError{Path: path, Line: node.Line, Message: "includes must be a mapping"}
	}

	for i := 0; i < len(node.Content); i += 2 {
		nameNode := node.Content[i]
		incNode := node.Content[i+1]

		var inc babfile.Include
		if err := incNode.Decode(&inc); err != nil {
			return &ParseError{Path: path, Line: nameNode.Line, Message: fmt.Sprintf("include %q", nameNode.Value), Cause: err}
		}
		schema.Includes[nameNode.Value] = inc
	}

	return nil
}
