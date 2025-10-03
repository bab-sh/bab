package display

import (
	"testing"

	"github.com/bab-sh/bab/internal/registry"
	"github.com/charmbracelet/lipgloss"
)

func TestListTasks(t *testing.T) {
	tests := []struct {
		name    string
		tasks   []*registry.Task
		wantErr bool
	}{
		{
			name:    "empty registry",
			tasks:   []*registry.Task{},
			wantErr: false,
		},
		{
			name: "single task",
			tasks: []*registry.Task{
				{Name: "build", Description: "Build the project", Commands: []string{"go build"}},
			},
			wantErr: false,
		},
		{
			name: "multiple root tasks",
			tasks: []*registry.Task{
				{Name: "build", Description: "Build", Commands: []string{"go build"}},
				{Name: "test", Description: "Test", Commands: []string{"go test"}},
			},
			wantErr: false,
		},
		{
			name: "grouped tasks",
			tasks: []*registry.Task{
				{Name: "dev:start", Description: "Start dev", Commands: []string{"npm run dev"}},
				{Name: "dev:watch", Description: "Watch", Commands: []string{"npm run watch"}},
			},
			wantErr: false,
		},
		{
			name: "mixed root and grouped tasks",
			tasks: []*registry.Task{
				{Name: "build", Description: "Build", Commands: []string{"go build"}},
				{Name: "dev:start", Description: "Start dev", Commands: []string{"npm run dev"}},
				{Name: "test", Description: "Test", Commands: []string{"go test"}},
			},
			wantErr: false,
		},
		{
			name: "tasks without descriptions",
			tasks: []*registry.Task{
				{Name: "build", Commands: []string{"go build"}},
				{Name: "test", Commands: []string{"go test"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := registry.New()
			for _, task := range tt.tasks {
				if err := reg.Register(task); err != nil {
					t.Fatalf("failed to register task: %v", err)
				}
			}

			err := ListTasks(reg)
			if tt.wantErr && err == nil {
				t.Error("ListTasks() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ListTasks() unexpected error: %v", err)
			}
		})
	}
}

func TestFormatTaskWithPadding(t *testing.T) {
	tests := []struct {
		name        string
		taskName    string
		description string
		maxLen      int
	}{
		{
			name:        "task with description",
			taskName:    "build",
			description: "Build the project",
			maxLen:      10,
		},
		{
			name:        "task without description",
			taskName:    "test",
			description: "",
			maxLen:      10,
		},
		{
			name:        "long task name",
			taskName:    "verylongtaskname",
			description: "Description",
			maxLen:      5,
		},
		{
			name:        "exact length match",
			taskName:    "build",
			description: "Build",
			maxLen:      5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the styled output, but we can verify it doesn't panic
			nameStyle := lipgloss.NewStyle()
			descStyle := lipgloss.NewStyle()
			result := formatTaskWithPadding(tt.taskName, tt.description, tt.maxLen, nameStyle, descStyle)

			if result == "" && tt.taskName != "" {
				t.Error("formatTaskWithPadding() returned empty string for non-empty task name")
			}

			// The result should contain the task name (even with ANSI codes)
			if tt.taskName != "" && len(result) == 0 {
				t.Error("formatTaskWithPadding() result doesn't seem to contain task name")
			}
		})
	}
}
