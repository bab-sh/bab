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
		"index": func(m map[string]int, key string) int {
			return m[key]
		},
	}
}

func (c *Compiler) generateShellScript(data TemplateData) error {
	tmpl, err := template.New("shell").Funcs(c.templateFuncs()).Parse(templates.ShellTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse shell template: %w", err)
	}

	outputPath := filepath.Join(c.outputDir, "bab.sh")
	file, err := os.Create(outputPath) //nolint:gosec // outputPath is user-controlled via --output flag
	if err != nil {
		return fmt.Errorf("failed to create shell script: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Error("Failed to close shell script file", "error", closeErr)
		}
	}()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute shell template: %w", err)
	}

	if err := os.Chmod(outputPath, 0755); err != nil { //nolint:gosec // executable permission needed for shell scripts
		return fmt.Errorf("failed to make shell script executable: %w", err)
	}

	if c.verbose {
		log.Debug("Generated shell script", "path", outputPath)
	}

	return nil
}

func (c *Compiler) generateBatchFile(data TemplateData) error {
	tmpl, err := template.New("batch").Funcs(c.templateFuncs()).Parse(templates.BatchTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse batch template: %w", err)
	}

	outputPath := filepath.Join(c.outputDir, "bab.bat")
	file, err := os.Create(outputPath) //nolint:gosec // outputPath is user-controlled via --output flag
	if err != nil {
		return fmt.Errorf("failed to create batch file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Error("Failed to close batch file", "error", closeErr)
		}
	}()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute batch template: %w", err)
	}

	if c.verbose {
		log.Debug("Generated batch file", "path", outputPath)
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
