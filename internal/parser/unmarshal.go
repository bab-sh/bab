package parser

import (
	"errors"
	"fmt"

	"github.com/bab-sh/bab/internal/babfile"
	"gopkg.in/yaml.v3"
)

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
		Tasks:    make(map[string]babfile.Task),
		Includes: make(map[string]babfile.Include),
	}

	for i := 0; i < len(doc.Content); i += 2 {
		key := doc.Content[i]
		val := doc.Content[i+1]

		switch key.Value {
		case "tasks":
			if err := parseTasks(path, val, schema); err != nil {
				return nil, err
			}
		case "includes":
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
		case "desc":
			task.Desc = val.Value
		case "deps":
			task.DepsLine = key.Line
			if err := val.Decode(&task.Deps); err != nil {
				return babfile.Task{}, &ParseError{Path: path, Line: key.Line, Message: "invalid deps", Cause: err}
			}
		case "run":
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

	var cmd, task string
	var platforms []babfile.Platform
	line := node.Line

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]

		switch key.Value {
		case "cmd":
			cmd = val.Value
		case "task":
			task = val.Value
		case "platforms":
			if err := val.Decode(&platforms); err != nil {
				return nil, &ParseError{Path: path, Line: key.Line, Message: "invalid platforms", Cause: err}
			}
		}
	}

	hasCmd := cmd != ""
	hasTask := task != ""

	switch {
	case hasCmd && hasTask:
		return nil, &ParseError{Path: path, Line: node.Line, Message: "cannot have both 'cmd' and 'task'"}
	case hasCmd:
		return babfile.CommandRun{Line: line, Cmd: cmd, Platforms: platforms}, nil
	case hasTask:
		return babfile.TaskRun{Line: line, Task: task, Platforms: platforms}, nil
	default:
		return nil, &ParseError{Path: path, Line: node.Line, Message: "must have either 'cmd' or 'task'"}
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
