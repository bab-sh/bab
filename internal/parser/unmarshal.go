package parser

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/errs"
	"gopkg.in/yaml.v3"
)

const (
	keyCmd         = "cmd"
	keyDeps        = "deps"
	keyDesc        = "desc"
	keyDir         = "dir"
	keyEnv         = "env"
	keyIncludes    = "includes"
	keyLevel       = "level"
	keyLog         = "log"
	keyOutput      = "output"
	keyPlatforms   = "platforms"
	keyRun         = "run"
	keySilent      = "silent"
	keyTask        = "task"
	keyTasks       = "tasks"
	keyVars        = "vars"
	keyPrompt      = "prompt"
	keyType        = "type"
	keyMessage     = "message"
	keyDefault     = "default"
	keyDefaults    = "defaults"
	keyOptions     = "options"
	keyPlaceholder = "placeholder"
	keyValidate    = "validate"
	keyMin         = "min"
	keyMax         = "max"
	keyConfirm     = "confirm"
)

type promptFields struct {
	name        string
	promptType  string
	message     string
	dflt        string
	placeholder string
	validate    string
	options     []string
	defaults    []string
	min         *int
	max         *int
	confirm     *bool
}

var varNameRegex = regexp.MustCompile(babfile.VarNamePattern)

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

func parseBool(path string, node *yaml.Node, target **bool, verrs *errs.ValidationErrors) bool {
	if node.Kind != yaml.ScalarNode {
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: "expected boolean value"})
		return false
	}

	var val bool
	if err := node.Decode(&val); err != nil {
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: "expected boolean value", Cause: err})
		return false
	}

	*target = &val
	return true
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
		case keySilent:
			parseBool(path, val, &schema.Silent, verrs)
		case keyOutput:
			parseBool(path, val, &schema.Output, verrs)
		case keyDir:
			if val.Kind == yaml.ScalarNode {
				schema.Dir = val.Value
			}
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
		case keySilent:
			if !parseBool(path, val, &task.Silent, verrs) {
				hasErrors = true
			}
		case keyOutput:
			if !parseBool(path, val, &task.Output, verrs) {
				hasErrors = true
			}
		case keyDir:
			task.Dir = val.Value
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

func parsePromptKey(key, val *yaml.Node, pf *promptFields, path, taskName string, index int, verrs *errs.ValidationErrors) bool {
	switch key.Value {
	case keyPrompt:
		pf.name = val.Value
	case keyType:
		pf.promptType = val.Value
	case keyMessage:
		pf.message = val.Value
	case keyDefault:
		pf.dflt = val.Value
	case keyPlaceholder:
		pf.placeholder = val.Value
	case keyValidate:
		pf.validate = val.Value
	case keyOptions:
		if err := val.Decode(&pf.options); err != nil {
			verrs.Add(&errs.ParseError{Path: path, Line: key.Line, Message: fmt.Sprintf("task %q: run[%d]: invalid options", taskName, index), Cause: err})
			return false
		}
	case keyDefaults:
		if err := val.Decode(&pf.defaults); err != nil {
			verrs.Add(&errs.ParseError{Path: path, Line: key.Line, Message: fmt.Sprintf("task %q: run[%d]: invalid defaults", taskName, index), Cause: err})
			return false
		}
	case keyMin:
		var v int
		if err := val.Decode(&v); err != nil {
			verrs.Add(&errs.ParseError{Path: path, Line: key.Line, Message: fmt.Sprintf("task %q: run[%d]: invalid min value", taskName, index), Cause: err})
			return false
		}
		pf.min = &v
	case keyMax:
		var v int
		if err := val.Decode(&v); err != nil {
			verrs.Add(&errs.ParseError{Path: path, Line: key.Line, Message: fmt.Sprintf("task %q: run[%d]: invalid max value", taskName, index), Cause: err})
			return false
		}
		pf.max = &v
	case keyConfirm:
		if !parseBool(path, val, &pf.confirm, verrs) {
			return false
		}
	}
	return true
}

func countRunTypes(cmd, task, log, prompt string) int {
	count := 0
	if cmd != "" {
		count++
	}
	if task != "" {
		count++
	}
	if log != "" {
		count++
	}
	if prompt != "" {
		count++
	}
	return count
}

func parseRunItem(path string, node *yaml.Node, taskName string, index int, verrs *errs.ValidationErrors) (babfile.RunItem, bool) {
	if node.Kind != yaml.MappingNode {
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: fmt.Sprintf("task %q: run[%d]: run item must be a mapping", taskName, index)})
		return nil, false
	}

	var cmd, task, log string
	var dir string
	var env map[string]string
	var platforms []babfile.Platform
	var level babfile.LogLevel
	var silent *bool
	var output *bool
	line := node.Line
	hasErrors := false

	var pf promptFields

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]

		switch key.Value {
		case keyCmd:
			cmd = val.Value
		case keyDir:
			dir = val.Value
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
		case keySilent:
			if !parseBool(path, val, &silent, verrs) {
				hasErrors = true
			}
		case keyOutput:
			if !parseBool(path, val, &output, verrs) {
				hasErrors = true
			}
		case keyPlatforms:
			if err := val.Decode(&platforms); err != nil {
				verrs.Add(&errs.ParseError{Path: path, Line: key.Line, Message: fmt.Sprintf("task %q: run[%d]: invalid platforms", taskName, index), Cause: err})
				hasErrors = true
			}
		default:
			if !parsePromptKey(key, val, &pf, path, taskName, index, verrs) {
				hasErrors = true
			}
		}
	}

	count := countRunTypes(cmd, task, log, pf.name)

	switch {
	case count > 1:
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: fmt.Sprintf("task %q: run[%d]: run item can only have one of 'cmd', 'task', 'log', or 'prompt'", taskName, index)})
		return nil, false
	case count == 0:
		verrs.Add(&errs.ParseError{Path: path, Line: node.Line, Message: fmt.Sprintf("task %q: run[%d]: run item must have 'cmd', 'task', 'log', or 'prompt'", taskName, index)})
		return nil, false
	case hasErrors:
		return nil, false
	case cmd != "":
		return babfile.CommandRun{Line: line, Cmd: cmd, Dir: dir, Env: env, Silent: silent, Output: output, Platforms: platforms}, true
	case task != "":
		return babfile.TaskRun{Line: line, Task: task, Silent: silent, Output: output, Platforms: platforms}, true
	case log != "":
		return buildLogRun(path, line, taskName, index, log, level, platforms, verrs)
	case pf.name != "":
		return validatePromptRun(path, line, taskName, index, pf, platforms, verrs)
	default:
		return nil, false
	}
}

func buildLogRun(path string, line int, taskName string, index int, log string, level babfile.LogLevel, platforms []babfile.Platform, verrs *errs.ValidationErrors) (babfile.RunItem, bool) {
	if level == "" {
		level = babfile.LogLevelInfo
	}
	if !level.Valid() {
		verrs.Add(&errs.ParseError{
			Path:    path,
			Line:    line,
			Message: fmt.Sprintf("task %q: run[%d]: invalid log level %q, must be one of: debug, info, warn, error", taskName, index, level),
		})
		return nil, false
	}
	return babfile.LogRun{Line: line, Log: log, Level: level, Platforms: platforms}, true
}

type promptValidationContext struct {
	path   string
	line   int
	prefix string
	pf     promptFields
	pType  babfile.PromptType
}

func (ctx *promptValidationContext) addError(verrs *errs.ValidationErrors, msg string) {
	verrs.Add(&errs.ParseError{Path: ctx.path, Line: ctx.line, Message: msg})
}

func validatePromptBasics(ctx *promptValidationContext, verrs *errs.ValidationErrors) bool {
	hasErrors := false

	if !varNameRegex.MatchString(ctx.pf.name) {
		ctx.addError(verrs, fmt.Sprintf("%s: invalid prompt variable name %q, must match pattern %s", ctx.prefix, ctx.pf.name, babfile.VarNamePattern))
		hasErrors = true
	}

	if ctx.pf.promptType == "" {
		ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: 'type' is required", ctx.prefix, ctx.pf.name))
		hasErrors = true
	} else if !ctx.pType.Valid() {
		ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: invalid type %q, must be one of: confirm, input, select, multiselect, password, number", ctx.prefix, ctx.pf.name, ctx.pf.promptType))
		hasErrors = true
	}

	if ctx.pf.message == "" {
		ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: 'message' is required", ctx.prefix, ctx.pf.name))
		hasErrors = true
	}

	return !hasErrors
}

func validatePromptOptions(ctx *promptValidationContext, verrs *errs.ValidationErrors) bool {
	hasErrors := false

	if ctx.pType.RequiresOptions() && len(ctx.pf.options) == 0 {
		ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: 'options' is required for type %q", ctx.prefix, ctx.pf.name, ctx.pf.promptType))
		hasErrors = true
	}

	if ctx.pType == babfile.PromptTypeSelect && ctx.pf.dflt != "" && len(ctx.pf.options) > 0 {
		if !stringInSlice(ctx.pf.dflt, ctx.pf.options) {
			ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: default value %q is not in options", ctx.prefix, ctx.pf.name, ctx.pf.dflt))
			hasErrors = true
		}
	}

	if ctx.pType == babfile.PromptTypeMultiselect && len(ctx.pf.defaults) > 0 && len(ctx.pf.options) > 0 {
		optionSet := make(map[string]bool, len(ctx.pf.options))
		for _, opt := range ctx.pf.options {
			optionSet[opt] = true
		}
		for _, def := range ctx.pf.defaults {
			if !optionSet[def] {
				ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: default value %q is not in options", ctx.prefix, ctx.pf.name, def))
				hasErrors = true
			}
		}
	}

	return !hasErrors
}

func validatePromptConstraints(ctx *promptValidationContext, verrs *errs.ValidationErrors) bool {
	hasErrors := false

	if ctx.pf.min != nil && ctx.pf.max != nil && *ctx.pf.min > *ctx.pf.max {
		ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: min (%d) cannot be greater than max (%d)", ctx.prefix, ctx.pf.name, *ctx.pf.min, *ctx.pf.max))
		hasErrors = true
	}

	if ctx.pf.validate != "" {
		if _, err := regexp.Compile(ctx.pf.validate); err != nil {
			ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: invalid validate regex: %v", ctx.prefix, ctx.pf.name, err))
			hasErrors = true
		}
	}

	return !hasErrors
}

func validatePromptTypeSpecific(ctx *promptValidationContext, verrs *errs.ValidationErrors) bool {
	if !ctx.pType.Valid() {
		return true
	}

	hasErrors := false

	if len(ctx.pf.options) > 0 && !ctx.pType.RequiresOptions() {
		ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: 'options' is only valid for select/multiselect types", ctx.prefix, ctx.pf.name))
		hasErrors = true
	}

	if len(ctx.pf.defaults) > 0 && ctx.pType != babfile.PromptTypeMultiselect {
		ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: 'defaults' is only valid for multiselect type", ctx.prefix, ctx.pf.name))
		hasErrors = true
	}

	if ctx.pf.validate != "" && ctx.pType != babfile.PromptTypeInput {
		ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: 'validate' is only valid for input type", ctx.prefix, ctx.pf.name))
		hasErrors = true
	}

	if ctx.pf.confirm != nil && ctx.pType != babfile.PromptTypePassword {
		ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: 'confirm' is only valid for password type", ctx.prefix, ctx.pf.name))
		hasErrors = true
	}

	if ctx.pType == babfile.PromptTypeNumber && ctx.pf.dflt != "" {
		if _, err := strconv.Atoi(ctx.pf.dflt); err != nil {
			ctx.addError(verrs, fmt.Sprintf("%s: prompt %q: default value %q must be a valid integer for number type", ctx.prefix, ctx.pf.name, ctx.pf.dflt))
			hasErrors = true
		}
	}

	return !hasErrors
}

func stringInSlice(s string, slice []string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func validatePromptRun(path string, line int, taskName string, index int, pf promptFields, platforms []babfile.Platform, verrs *errs.ValidationErrors) (babfile.RunItem, bool) {
	ctx := &promptValidationContext{
		path:   path,
		line:   line,
		prefix: fmt.Sprintf("task %q: run[%d]", taskName, index),
		pf:     pf,
		pType:  babfile.PromptType(pf.promptType),
	}

	valid := validatePromptBasics(ctx, verrs)
	valid = validatePromptOptions(ctx, verrs) && valid
	valid = validatePromptConstraints(ctx, verrs) && valid
	valid = validatePromptTypeSpecific(ctx, verrs) && valid

	if !valid {
		return nil, false
	}

	return babfile.PromptRun{
		Line:        line,
		Prompt:      pf.name,
		Type:        ctx.pType,
		Message:     pf.message,
		Platforms:   platforms,
		Default:     pf.dflt,
		Defaults:    pf.defaults,
		Options:     pf.options,
		Placeholder: pf.placeholder,
		Validate:    pf.validate,
		Min:         pf.min,
		Max:         pf.max,
		Confirm:     pf.confirm,
	}, true
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
