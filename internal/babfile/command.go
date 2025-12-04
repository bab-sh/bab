package babfile

type Command struct {
	Cmd       string     `json:"cmd" yaml:"cmd" jsonschema:"description=Shell command to execute,minLength=1"`
	Platforms []Platform `json:"platforms,omitempty" yaml:"platforms,omitempty" jsonschema:"description=Platforms to run this command on (if empty runs on all platforms)"`
}
