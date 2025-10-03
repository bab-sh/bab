package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bab-sh/bab/internal/registry"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		babfilePath string
		options     []Option
		check       func(*testing.T, *Compiler)
	}{
		{
			name:        "default compiler",
			babfilePath: "Babfile",
			options:     nil,
			check: func(t *testing.T, c *Compiler) {
				if c.babfilePath != "Babfile" {
					t.Errorf("babfilePath = %q, want %q", c.babfilePath, "Babfile")
				}
				if c.outputDir != "." {
					t.Errorf("outputDir = %q, want %q", c.outputDir, ".")
				}
				if c.verbose {
					t.Error("verbose should be false by default")
				}
				if c.noColor {
					t.Error("noColor should be false by default")
				}
			},
		},
		{
			name:        "with output dir",
			babfilePath: "Babfile",
			options:     []Option{WithOutputDir("dist")},
			check: func(t *testing.T, c *Compiler) {
				if c.outputDir != "dist" {
					t.Errorf("outputDir = %q, want %q", c.outputDir, "dist")
				}
			},
		},
		{
			name:        "with verbose",
			babfilePath: "Babfile",
			options:     []Option{WithVerbose(true)},
			check: func(t *testing.T, c *Compiler) {
				if !c.verbose {
					t.Error("verbose should be true")
				}
			},
		},
		{
			name:        "with no color",
			babfilePath: "Babfile",
			options:     []Option{WithNoColor(true)},
			check: func(t *testing.T, c *Compiler) {
				if !c.noColor {
					t.Error("noColor should be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := New(tt.babfilePath, tt.options...)
			if compiler == nil {
				t.Fatal("New() returned nil")
			}
			if tt.check != nil {
				tt.check(t, compiler)
			}
		})
	}
}

func TestCompiler_Compile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test Babfile
	babfileContent := `
build:
  desc: Build the project
  run: go build

test:
  desc: Run tests
  run: go test

dev:
  start:
    desc: Start dev server
    run: npm run dev
`
	babfilePath := filepath.Join(tmpDir, "Babfile")
	if err := os.WriteFile(babfilePath, []byte(babfileContent), 0600); err != nil {
		t.Fatalf("failed to create test Babfile: %v", err)
	}

	outputDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		t.Fatalf("failed to create output directory: %v", err)
	}

	compiler := New(babfilePath, WithOutputDir(outputDir))
	err := compiler.Compile()
	if err != nil {
		t.Fatalf("Compile() failed: %v", err)
	}

	// Verify shell script was created
	shellPath := filepath.Join(outputDir, "bab.sh")
	if _, err := os.Stat(shellPath); os.IsNotExist(err) {
		t.Error("Compile() did not create bab.sh")
	}

	// Verify batch file was created
	batchPath := filepath.Join(outputDir, "bab.bat")
	if _, err := os.Stat(batchPath); os.IsNotExist(err) {
		t.Error("Compile() did not create bab.bat")
	}

	// Verify shell script is executable
	info, err := os.Stat(shellPath)
	if err != nil {
		t.Fatalf("failed to stat shell script: %v", err)
	}
	if info.Mode().Perm()&0111 == 0 {
		t.Error("Compile() shell script is not executable")
	}

	// Verify shell script contains tasks
	cleanPath := filepath.Clean(shellPath)
	if !filepath.IsAbs(cleanPath) && !filepath.IsLocal(cleanPath) {
		t.Fatal("shell path is not a valid local file")
	}
	shellContent, err := os.ReadFile(cleanPath)
	if err != nil {
		t.Fatalf("failed to read shell script: %v", err)
	}
	if !strings.Contains(string(shellContent), "build") {
		t.Error("Compile() shell script missing 'build' task")
	}
	if !strings.Contains(string(shellContent), "test") {
		t.Error("Compile() shell script missing 'test' task")
	}
	if !strings.Contains(string(shellContent), "dev:start") {
		t.Error("Compile() shell script missing 'dev:start' task")
	}
}

func TestCompiler_CompileInvalidBabfile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an invalid Babfile
	babfileContent := `invalid: [yaml`
	babfilePath := filepath.Join(tmpDir, "Babfile")
	if err := os.WriteFile(babfilePath, []byte(babfileContent), 0600); err != nil {
		t.Fatalf("failed to create test Babfile: %v", err)
	}

	compiler := New(babfilePath)
	err := compiler.Compile()
	if err == nil {
		t.Error("Compile() expected error for invalid Babfile, got nil")
	}
}

func TestCompiler_prepareTemplateData(t *testing.T) {
	reg := registry.New()

	tasks := []*registry.Task{
		{Name: "build", Description: "Build", Commands: []string{"go build"}},
		{Name: "test", Description: "Test", Commands: []string{"go test"}},
		{Name: "dev:start", Description: "Start dev", Commands: []string{"npm run dev"}},
		{Name: "dev:watch", Description: "Watch", Commands: []string{"npm run watch"}},
	}

	for _, task := range tasks {
		if err := reg.Register(task); err != nil {
			t.Fatalf("failed to register task: %v", err)
		}
	}

	compiler := &Compiler{noColor: false}
	data := compiler.prepareTemplateData(reg)

	// Check total tasks
	if len(data.Tasks) != 4 {
		t.Errorf("prepareTemplateData() got %d tasks, want 4", len(data.Tasks))
	}

	// Check root tasks
	if len(data.RootTasks) != 2 {
		t.Errorf("prepareTemplateData() got %d root tasks, want 2", len(data.RootTasks))
	}

	// Check grouped tasks
	if len(data.GroupedTasks) != 1 {
		t.Errorf("prepareTemplateData() got %d groups, want 1", len(data.GroupedTasks))
	}

	devTasks, exists := data.GroupedTasks["dev"]
	if !exists {
		t.Fatal("prepareTemplateData() missing 'dev' group")
	}
	if len(devTasks) != 2 {
		t.Errorf("prepareTemplateData() got %d dev tasks, want 2", len(devTasks))
	}

	// Check max name length calculations
	if data.RootMaxNameLen < 4 { // "test" is 4 chars
		t.Errorf("prepareTemplateData() RootMaxNameLen = %d, want >= 4", data.RootMaxNameLen)
	}

	devMaxLen, exists := data.GroupMaxNameLen["dev"]
	if !exists {
		t.Fatal("prepareTemplateData() missing dev group max length")
	}
	if devMaxLen < 5 { // "start" or "watch" are 5+ chars
		t.Errorf("prepareTemplateData() dev max length = %d, want >= 5", devMaxLen)
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple name",
			input: "build",
			want:  "build",
		},
		{
			name:  "name with colon",
			input: "dev:start",
			want:  "dev_start",
		},
		{
			name:  "name with spaces",
			input: "my task",
			want:  "my_task",
		},
		{
			name:  "name with special chars",
			input: "deploy-prod!",
			want:  "deploy_prod_",
		},
		{
			name:  "name starting with number",
			input: "123build",
			want:  "_123build",
		},
		{
			name:  "deeply nested name",
			input: "dev:server:start",
			want:  "dev_server_start",
		},
		{
			name:  "name with multiple special chars",
			input: "test@email.com",
			want:  "test_email_com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWithOutputDir(t *testing.T) {
	c := &Compiler{}
	opt := WithOutputDir("/tmp/output")
	opt(c)

	if c.outputDir != "/tmp/output" {
		t.Errorf("WithOutputDir() set outputDir = %q, want %q", c.outputDir, "/tmp/output")
	}
}

func TestWithVerbose_Compiler(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{"enable verbose", true},
		{"disable verbose", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Compiler{}
			opt := WithVerbose(tt.verbose)
			opt(c)

			if c.verbose != tt.verbose {
				t.Errorf("WithVerbose(%v) set verbose = %v", tt.verbose, c.verbose)
			}
		})
	}
}

func TestWithNoColor(t *testing.T) {
	tests := []struct {
		name    string
		noColor bool
	}{
		{"enable no color", true},
		{"disable no color", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Compiler{}
			opt := WithNoColor(tt.noColor)
			opt(c)

			if c.noColor != tt.noColor {
				t.Errorf("WithNoColor(%v) set noColor = %v", tt.noColor, c.noColor)
			}
		})
	}
}
