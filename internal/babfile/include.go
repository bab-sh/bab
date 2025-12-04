package babfile

type Include struct {
	Babfile string `json:"babfile" yaml:"babfile" jsonschema:"description=Relative or absolute path to the babfile to include"`
}
