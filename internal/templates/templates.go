package templates

import (
	_ "embed"
)

//go:embed shell.tmpl
var ShellTemplate string

//go:embed batch.tmpl
var BatchTemplate string
