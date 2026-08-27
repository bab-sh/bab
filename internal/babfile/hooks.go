package babfile

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

var ValidHookNames = map[string]bool{
	"applypatch-msg":     true,
	"pre-applypatch":     true,
	"post-applypatch":    true,
	"pre-commit":         true,
	"prepare-commit-msg": true,
	"commit-msg":         true,
	"post-commit":        true,
	"pre-rebase":         true,
	"post-checkout":      true,
	"post-merge":         true,
	"pre-push":           true,
	"pre-receive":        true,
	"update":             true,
	"post-receive":       true,
	"post-update":        true,
	"pre-auto-gc":        true,
	"post-rewrite":       true,
}

type HookArgDef struct {
	Name     string
	Position int
	ReadFile bool
}

var HookArgs = map[string][]HookArgDef{
	"commit-msg":     {{Name: "msg", Position: 1, ReadFile: true}},
	"applypatch-msg": {{Name: "msg", Position: 1, ReadFile: true}},
	"prepare-commit-msg": {
		{Name: "msg_file", Position: 1},
		{Name: "source", Position: 2},
		{Name: "sha", Position: 3},
	},
	"pre-push": {
		{Name: "remote", Position: 1},
		{Name: "url", Position: 2},
	},
	"pre-rebase": {
		{Name: "upstream", Position: 1},
		{Name: "branch", Position: 2},
	},
	"post-checkout": {
		{Name: "prev_ref", Position: 1},
		{Name: "new_ref", Position: 2},
		{Name: "is_branch", Position: 3},
	},
	"update": {
		{Name: "ref", Position: 1},
		{Name: "old_sha", Position: 2},
		{Name: "new_sha", Position: 3},
	},
	"post-rewrite": {{Name: "cause", Position: 1}},
}

func MaxHookArgPosition(hookName string) int {
	defs := HookArgs[hookName]
	max := 0
	for _, d := range defs {
		if d.Position > max {
			max = d.Position
		}
	}
	return max
}

type HookDef struct {
	Line int       `json:"-" yaml:"-"`
	Name string    `json:"-" yaml:"-"`
	Run  []RunItem `json:"-" yaml:"-"`
}

type HookMap map[string]*HookDef

func IsValidHookName(name string) bool {
	return ValidHookNames[name]
}

func HooksSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "object",
		Description: "Git hook definitions",
		PropertyNames: &jsonschema.Schema{
			Enum: validHookNamesEnum(),
		},
		AdditionalProperties: &jsonschema.Schema{Ref: "#/$defs/HookDef"},
	}
}

func HookDefSchema() *jsonschema.Schema {
	minRunItems := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("run", &jsonschema.Schema{
		Type:        "array",
		Description: "Commands or tasks to execute",
		MinItems:    &minRunItems,
		Items:       &jsonschema.Schema{Ref: "#/$defs/RunItem"},
	})

	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "Hook definition",
		Required:             []string{"run"},
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           props,
	}
}

func validHookNamesEnum() []any {
	names := make([]any, 0, len(ValidHookNames))
	for name := range ValidHookNames {
		names = append(names, name)
	}
	return names
}
