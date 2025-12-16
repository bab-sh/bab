package parser

import (
	"fmt"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/errs"
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
	keyVars      = "vars"
)

func parseEnvMap(path string, node *yaml.Node, env *map[string]string, verrs *errs.ValidationErrors) bool {
	if node.Kind != yaml.MappingNode {
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: "env must be a mapping"})
		return false
	}

	*env = make(map[string]string, len(node.Content)/2)
	hasErrors := false

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]

		if valNode.Kind != yaml.ScalarNode {
			verrs.Add(&errs.ParseError{
				Path:    path,
				Line:    valNode.Line,
				Message: fmt.Sprintf("env value for %q must be a string", keyNode.Value),
			})
			hasErrors = true
			continue
		}

		(*env)[keyNode.Value] = valNode.Value
	}

	return !hasErrors
}

func parseVarMap(path string, node *yaml.Node, vars *babfile.VarMap, verrs *errs.ValidationErrors) bool {
	if node.Kind != yaml.MappingNode {
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: "vars must be a mapping"})
		return false
	}

	*vars = make(babfile.VarMap, len(node.Content)/2)
	hasErrors := false

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]

		if valNode.Kind != yaml.ScalarNode {
			verrs.Add(&errs.ParseError{
				Path:    path,
				Line:    valNode.Line,
				Message: fmt.Sprintf("vars value for %q must be a string", keyNode.Value),
			})
			hasErrors = true
			continue
		}

		(*vars)[keyNode.Value] = valNode.Value
	}

	return !hasErrors
}

func unmarshalBabfile(path string, data []byte) (*babfile.Schema, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return nil, &errs.ParseError{Path: path, Message: "invalid document"}
	}

	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return nil, &errs.ParseError{Path: path, Line: doc.Line, Message: "expected mapping at root"}
	}

	schema := &babfile.Schema{
		Vars:     make(babfile.VarMap),
		Env:      make(map[string]string),
		Tasks:    make(map[string]babfile.Task),
		Includes: make(map[string]babfile.Include),
	}

	verrs := &errs.ValidationErrors{}

	for i := 0; i < len(doc.Content); i += 2 {
		key := doc.Content[i]
		val := doc.Content[i+1]

		switch key.Value {
		case keyVars:
			parseVarMap(path, val, &schema.Vars, verrs)
		case keyEnv:
			parseEnvMap(path, val, &schema.Env, verrs)
		case keyTasks:
			parseTasks(path, val, schema, verrs)
		case keyIncludes:
			parseIncludes(path, val, schema, verrs)
		}
	}

	if verrs.HasErrors() {
		return nil, verrs
	}

	return schema, nil
}

func parseTasks(path string, node *yaml.Node, schema *babfile.Schema, verrs *errs.ValidationErrors) {
	if node.Kind != yaml.MappingNode {
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: "tasks must be a mapping"})
		return
	}

	taskLines := make(map[string]int)

	for i := 0; i < len(node.Content); i += 2 {
		nameNode := node.Content[i]
		taskNode := node.Content[i+1]
		name := nameNode.Value

		if origLine, exists := taskLines[name]; exists {
			verrs.Add(&errs.DuplicateTaskError{
				Path:         path,
				Line:         nameNode.Line,
				TaskName:     name,
				OriginalLine: origLine,
			})
			continue
		}
		taskLines[name] = nameNode.Line

		task, ok := parseTask(path, taskNode, name, verrs)
		if !ok {
			continue
		}
		task.Line = nameNode.Line
		schema.Tasks[name] = task
	}
}

func parseTask(path string, node *yaml.Node, taskName string, verrs *errs.ValidationErrors) (babfile.Task, bool) {
	if node.Kind != yaml.MappingNode {
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: fmt.Sprintf("task %q: task must be a mapping", taskName)})
		return babfile.Task{}, false
	}

	task := babfile.Task{}
	hasErrors := false

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]

		switch key.Value {
		case keyDesc:
			task.Desc = val.Value
		case keyVars:
			if !parseVarMap(path, val, &task.Vars, verrs) {
				hasErrors = true
			}
		case keyEnv:
			if !parseEnvMap(path, val, &task.Env, verrs) {
				hasErrors = true
			}
		case keyDeps:
			task.DepsLine = key.Line
			if err := val.Decode(&task.Deps); err != nil {
				verrs.Add(&errs.ParseError{Path: path, Line: key.Line, Message: fmt.Sprintf("task %q: invalid deps", taskName), Cause: err})
				hasErrors = true
			}
		case keyRun:
			runItems, ok := parseRunItems(path, val, taskName, verrs)
			if !ok {
				hasErrors = true
			}
			task.Run = runItems
		}
	}

	return task, !hasErrors
}

func parseRunItems(path string, node *yaml.Node, taskName string, verrs *errs.ValidationErrors) ([]babfile.RunItem, bool) {
	if node.Kind != yaml.SequenceNode {
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: fmt.Sprintf("task %q: run must be a sequence", taskName)})
		return nil, false
	}

	items := make([]babfile.RunItem, 0, len(node.Content))
	hasErrors := false

	for i, itemNode := range node.Content {
		item, ok := parseRunItem(path, itemNode, taskName, i, verrs)
		if !ok {
			hasErrors = true
			continue
		}
		items = append(items, item)
	}

	return items, !hasErrors
}

func parseRunItem(path string, node *yaml.Node, taskName string, index int, verrs *errs.ValidationErrors) (babfile.RunItem, bool) {
	if node.Kind != yaml.MappingNode {
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: fmt.Sprintf("task %q: run[%d]: run item must be a mapping", taskName, index)})
		return nil, false
	}

	var cmd, task, log string
	var env map[string]string
	var platforms []babfile.Platform
	var level babfile.LogLevel
	line := node.Line
	hasErrors := false

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
			if !parseEnvMap(path, val, &env, verrs) {
				hasErrors = true
			}
		case keyPlatforms:
			if err := val.Decode(&platforms); err != nil {
				verrs.Add(&errs.ParseError{Path: path, Line: key.Line, Message: fmt.Sprintf("task %q: run[%d]: invalid platforms", taskName, index), Cause: err})
				hasErrors = true
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
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: fmt.Sprintf("task %q: run[%d]: run item can only have one of 'cmd', 'task', or 'log'", taskName, index)})
		return nil, false
	case count == 0:
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: fmt.Sprintf("task %q: run[%d]: run item must have 'cmd', 'task', or 'log'", taskName, index)})
		return nil, false
	case hasErrors:
		return nil, false
	case hasCmd:
		return babfile.CommandRun{Line: line, Cmd: cmd, Env: env, Platforms: platforms}, true
	case hasTask:
		return babfile.TaskRun{Line: line, Task: task, Platforms: platforms}, true
	case hasLog:
		if level == "" {
			level = babfile.LogLevelInfo
		}
		if !level.Valid() {
			verrs.Add(&errs.ParseError{
				Path:    path,
				Line:    node.Line,
				Message: fmt.Sprintf("task %q: run[%d]: invalid log level %q, must be one of: debug, info, warn, error", taskName, index, level),
			})
			return nil, false
		}
		return babfile.LogRun{Line: line, Log: log, Level: level, Platforms: platforms}, true
	default:
		return nil, false
	}
}

func parseIncludes(path string, node *yaml.Node, schema *babfile.Schema, verrs *errs.ValidationErrors) {
	if node.Kind != yaml.MappingNode {
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: "includes must be a mapping"})
		return
	}

	for i := 0; i < len(node.Content); i += 2 {
		nameNode := node.Content[i]
		incNode := node.Content[i+1]

		var inc babfile.Include
		if err := incNode.Decode(&inc); err != nil {
			verrs.Add(&errs.ParseError{Path: path, Line: nameNode.Line, Message: fmt.Sprintf("include %q", nameNode.Value), Cause: err})
			continue
		}
		schema.Includes[nameNode.Value] = inc
	}
}
