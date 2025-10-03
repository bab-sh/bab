package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/bab/bab/internal/parser"
	"github.com/bab/bab/internal/registry"
	"github.com/bab/bab/internal/templates"
	"github.com/fatih/color"
)

type Compiler struct {
	babfilePath string
	outputDir   string
	verbose     bool
}

type TemplateTask struct {
	Name        string
	SafeName    string
	Description string
	Commands    []string
}

type TemplateData struct {
	Tasks        []TemplateTask
	RootTasks    []TemplateTask
	GroupedTasks map[string][]TemplateTask
}

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

type Option func(*Compiler)

func WithOutputDir(dir string) Option {
	return func(c *Compiler) {
		c.outputDir = dir
	}
}

func WithVerbose(verbose bool) Option {
	return func(c *Compiler) {
		c.verbose = verbose
	}
}

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

	success := color.New(color.FgGreen, color.Bold)
	success.Println("\n✓ Successfully compiled Babfile to scripts!")
	info := color.New(color.FgCyan)
	info.Printf("  • %s\n", filepath.Join(c.outputDir, "bab.sh"))
	info.Printf("  • %s\n", filepath.Join(c.outputDir, "bab.bat"))

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
	if root, exists := tree[""]; exists {
		for _, task := range root {
			rootTasks = append(rootTasks, TemplateTask{
				Name:        task.Name,
				SafeName:    sanitizeName(task.Name),
				Description: task.Description,
				Commands:    task.Commands,
			})
		}
	}

	groupedTasks := make(map[string][]TemplateTask)
	for group, groupTasks := range tree {
		if group == "" {
			continue
		}
		tasks := make([]TemplateTask, 0, len(groupTasks))
		for _, task := range groupTasks {
			tasks = append(tasks, TemplateTask{
				Name:        task.Name,
				SafeName:    sanitizeName(task.Name),
				Description: task.Description,
				Commands:    task.Commands,
			})
		}
		groupedTasks[group] = tasks
	}

	return TemplateData{
		Tasks:        templateTasks,
		RootTasks:    rootTasks,
		GroupedTasks: groupedTasks,
	}
}

func (c *Compiler) generateShellScript(data TemplateData) error {
	tmpl, err := template.New("shell").Parse(templates.ShellTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse shell template: %w", err)
	}

	outputPath := filepath.Join(c.outputDir, "bab.sh")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create shell script: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute shell template: %w", err)
	}

	if err := os.Chmod(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to make shell script executable: %w", err)
	}

	if c.verbose {
		fmt.Printf("Generated: %s\n", outputPath)
	}

	return nil
}

func (c *Compiler) generateBatchFile(data TemplateData) error {
	tmpl, err := template.New("batch").Parse(templates.BatchTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse batch template: %w", err)
	}

	outputPath := filepath.Join(c.outputDir, "bab.bat")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create batch file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute batch template: %w", err)
	}

	if c.verbose {
		fmt.Printf("Generated: %s\n", outputPath)
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
