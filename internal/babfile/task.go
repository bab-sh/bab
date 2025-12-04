package babfile

type Task struct {
	Desc string    `json:"desc,omitempty" yaml:"desc,omitempty" jsonschema:"description=Human-readable description of the task"`
	Deps []string  `json:"deps,omitempty" yaml:"deps,omitempty" jsonschema:"description=Task dependencies to run before this task"`
	Run  []Command `json:"run,omitempty" yaml:"run,omitempty" jsonschema:"description=List of commands to execute"`
}
