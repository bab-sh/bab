// Package compiler provides functionality for compiling Babfiles to standalone shell scripts.
package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/bab-sh/bab/internal/parser"
	"github.com/bab-sh/bab/internal/registry"
	"github.com/bab-sh/bab/internal/templates"
	"github.com/charmbracelet/log"
)

// Compiler compiles Babfiles to standalone shell scripts.
type Compiler struct {
	babfilePath string
	outputDir   string
	verbose     bool
	noColor     bool
}

// TemplateTask represents a task for template rendering.
type TemplateTask struct {
	Name        string
	SafeName    string
	Description string
	Commands    []string
}

// TemplateData holds all data needed for template execution.
type TemplateData struct {
	Tasks           []TemplateTask
	RootTasks       []TemplateTask
	GroupedTasks    map[string][]TemplateTask
	RootMaxNameLen  int
	GroupMaxNameLen map[string]int
	NoColor         bool
}

// New creates a new Compiler instance.
func New(babfilePath string, options ...Option) *Compiler {
	c := &Compiler{
		babfilePath: babfilePath,
		outputDir:   ".",
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

// Option is a functional option for configuring the Compiler.
type Option func(*Compiler)

// WithOutputDir sets the output directory for generated scripts.
func WithOutputDir(dir string) Option {
	return func(c *Compiler) {
		c.outputDir = dir
	}
}

// WithVerbose enables verbose output.
func WithVerbose(verbose bool) Option {
	return func(c *Compiler) {
		c.verbose = verbose
	}
}

// WithNoColor disables color output in generated scripts.
func WithNoColor(noColor bool) Option {
	return func(c *Compiler) {
		c.noColor = noColor
	}
}

// Compile compiles the Babfile to standalone shell scripts.
func (c *Compiler) Compile() error {
	reg := registry.New()
	p := parser.New(reg)

	if err := p.ParseFile(c.babfilePath); err != nil {
		return fmt.Errorf("failed to parse Babfile: %w", err)
	}

	data := c.prepareTemplateData(reg)

	if err := c.generateShellScript(data); err != nil {
		return fmt.Errorf("failed to generate shell script: %w", err)
	}

	if err := c.generateBatchFile(data); err != nil {
		return fmt.Errorf("failed to generate batch file: %w", err)
	}

	log.Info("Successfully compiled Babfile to scripts!")
	log.Info("Generated script", "path", filepath.Join(c.outputDir, "bab.sh"))
	log.Info("Generated script", "path", filepath.Join(c.outputDir, "bab.bat"))

	return nil
}

func (c *Compiler) prepareTemplateData(reg registry.Registry) TemplateData {
	tasks := reg.List()
	tree := reg.Tree()

	templateTasks := make([]TemplateTask, 0, len(tasks))
	for _, task := range tasks {
		templateTasks = append(templateTasks, TemplateTask{
			Name:        task.Name,
			SafeName:    sanitizeName(task.Name),
			Description: task.Description,
			Commands:    task.Commands,
		})
	}

	rootTasks := make([]TemplateTask, 0)
	groupedTasks := make(map[string][]TemplateTask)
	groupMaxNameLen := make(map[string]int)

	// Recursively collect tasks from the tree
	c.collectTasksFromNode(tree, "", &rootTasks, groupedTasks, groupMaxNameLen)

	rootMaxNameLen := 0
	for _, task := range rootTasks {
		if len(task.Name) > rootMaxNameLen {
			rootMaxNameLen = len(task.Name)
		}
	}

	return TemplateData{
		Tasks:           templateTasks,
		RootTasks:       rootTasks,
		GroupedTasks:    groupedTasks,
		RootMaxNameLen:  rootMaxNameLen,
		GroupMaxNameLen: groupMaxNameLen,
		NoColor:         c.noColor,
	}
}

// collectTasksFromNode recursively collects tasks from a tree node.
func (c *Compiler) collectTasksFromNode(
	node *registry.TreeNode,
	currentGroup string,
	rootTasks *[]TemplateTask,
	groupedTasks map[string][]TemplateTask,
	groupMaxNameLen map[string]int,
) {
	for _, child := range node.Children {
		if child.IsTask() {
			task := child.Task
			templateTask := TemplateTask{
				Name:        task.Name,
				SafeName:    sanitizeName(task.Name),
				Description: task.Description,
				Commands:    task.Commands,
			}

			if currentGroup == "" {
				// Root level task
				*rootTasks = append(*rootTasks, templateTask)
			} else {
				// Grouped task
				groupedTasks[currentGroup] = append(groupedTasks[currentGroup], templateTask)

				// Update max name length for this group
				shortName := strings.TrimPrefix(task.Name, currentGroup+":")
				if len(shortName) > groupMaxNameLen[currentGroup] {
					groupMaxNameLen[currentGroup] = len(shortName)
				}
			}
		} else {
			// This is a group node
			groupName := child.Name
			if currentGroup == "" {
				// First level group
				c.collectTasksFromNode(child, groupName, rootTasks, groupedTasks, groupMaxNameLen)
			} else {
				// Nested group and continue with the same top level group for flattening
				c.collectTasksFromNode(child, currentGroup, rootTasks, groupedTasks, groupMaxNameLen)
			}
		}
	}
}

func (c *Compiler) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"trimPrefix": strings.TrimPrefix,
		"pad": func(s string, width int) string {
			if len(s) >= width {
				return s
			}
			return s + strings.Repeat(" ", width-len(s))
		},
	}
}

func (c *Compiler) generateShellScript(data TemplateData) error {
	return c.generateScript("shell", templates.ShellTemplate, "bab.sh", data, 0700)
}

func (c *Compiler) generateBatchFile(data TemplateData) error {
	return c.generateScript("batch", templates.BatchTemplate, "bab.bat", data, 0600)
}

func (c *Compiler) generateScript(name, templateStr, filename string, data TemplateData, fileMode os.FileMode) error {
	tmpl, err := template.New(name).Funcs(c.templateFuncs()).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse %s template: %w", name, err)
	}

	// Write to temp file first
	tmpFile, err := os.CreateTemp(c.outputDir, ".bab-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on error
	defer func() {
		if tmpFile != nil {
			if err := tmpFile.Close(); err != nil {
				log.Error("Failed to close temp file", "error", err)
			}
			if err := os.Remove(tmpPath); err != nil {
				log.Error("Failed to remove temp file", "path", tmpPath, "error", err)
			}
		}
	}()

	// Write content to temp file
	if err := tmpl.Execute(tmpFile, data); err != nil {
		return fmt.Errorf("failed to execute %s template: %w", name, err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, fileMode); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Move to final location
	finalPath := filepath.Join(filepath.Clean(c.outputDir), filepath.Base(filename))
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("failed to move file to final location: %w", err)
	}

	tmpFile = nil

	if c.verbose {
		log.Debug("Generated script", "type", name, "path", finalPath)
	}

	return nil
}

func sanitizeName(name string) string {
	safe := strings.ReplaceAll(name, ":", "_")

	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	safe = reg.ReplaceAllString(safe, "_")

	if len(safe) > 0 && safe[0] >= '0' && safe[0] <= '9' {
		safe = "_" + safe
	}

	return safe
}
