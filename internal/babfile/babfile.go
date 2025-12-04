package babfile

type Schema struct {
	Includes map[string]Include `json:"includes,omitempty" yaml:"includes,omitempty" jsonschema:"description=External babfile imports with namespace prefixes"`
	Tasks    map[string]Task    `json:"tasks" yaml:"tasks" jsonschema:"description=Task definitions"`
}
