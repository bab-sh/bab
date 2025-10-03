// Package templates contains embedded template files for generating shell scripts.
package templates

import (
	_ "embed"
)

// ShellTemplate contains the embedded shell script template.
//
//go:embed shell.tmpl
var ShellTemplate string

// BatchTemplate contains the embedded batch file template.
//
//go:embed batch.tmpl
var BatchTemplate string
