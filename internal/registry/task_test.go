package registry

import (
	"testing"
)

func TestNewTask(t *testing.T) {
	name := "test-task"
	task := NewTask(name)

	if task == nil {
		t.Fatal("NewTask returned nil")
	}

	if task.Name != name {
		t.Errorf("expected name %q, got %q", name, task.Name)
	}

	if task.Commands == nil {
		t.Error("Commands slice should be initialized, got nil")
	}

	if len(task.Commands) != 0 {
		t.Errorf("expected empty Commands slice, got length %d", len(task.Commands))
	}
}

func TestTask_IsGrouped(t *testing.T) {
	tests := []struct {
		name     string
		taskName string
		want     bool
	}{
		{
			name:     "simple task without colon",
			taskName: "build",
			want:     false,
		},
		{
			name:     "grouped task with single colon",
			taskName: "dev:start",
			want:     true,
		},
		{
			name:     "deeply nested task",
			taskName: "dev:server:start",
			want:     true,
		},
		{
			name:     "empty task name",
			taskName: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Name: tt.taskName}
			got := task.IsGrouped()
			if got != tt.want {
				t.Errorf("IsGrouped() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_GroupPath(t *testing.T) {
	tests := []struct {
		name     string
		taskName string
		want     []string
	}{
		{
			name:     "simple task without groups",
			taskName: "build",
			want:     []string{},
		},
		{
			name:     "single level group",
			taskName: "dev:start",
			want:     []string{"dev"},
		},
		{
			name:     "two level group",
			taskName: "build:platforms:linux",
			want:     []string{"build", "platforms"},
		},
		{
			name:     "deep nesting (4 levels)",
			taskName: "build:platforms:linux:amd64:optimized",
			want:     []string{"build", "platforms", "linux", "amd64"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Name: tt.taskName}
			got := task.GroupPath()

			if len(got) != len(tt.want) {
				t.Errorf("GroupPath() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for i, segment := range tt.want {
				if got[i] != segment {
					t.Errorf("GroupPath()[%d] = %q, want %q", i, got[i], segment)
				}
			}
		})
	}
}

func TestTask_LeafName(t *testing.T) {
	tests := []struct {
		name     string
		taskName string
		want     string
	}{
		{
			name:     "simple task",
			taskName: "build",
			want:     "build",
		},
		{
			name:     "single level group",
			taskName: "dev:start",
			want:     "start",
		},
		{
			name:     "two level group",
			taskName: "build:platforms:linux",
			want:     "linux",
		},
		{
			name:     "deep nesting",
			taskName: "build:platforms:linux:amd64:optimized",
			want:     "optimized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Name: tt.taskName}
			got := task.LeafName()

			if got != tt.want {
				t.Errorf("LeafName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    *Task
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid task",
			task: &Task{
				Name:     "build",
				Commands: []string{"go build"},
			},
			wantErr: false,
		},
		{
			name: "task with empty name",
			task: &Task{
				Name:     "",
				Commands: []string{"go build"},
			},
			wantErr: true,
			errMsg:  "task name cannot be empty",
		},
		{
			name: "task with no commands",
			task: &Task{
				Name:     "build",
				Commands: []string{},
			},
			wantErr: true,
			errMsg:  "task 'build' validation failed: commands - task has no commands",
		},
		{
			name: "task with nil commands",
			task: &Task{
				Name:     "build",
				Commands: nil,
			},
			wantErr: true,
			errMsg:  "task 'build' validation failed: commands - task has no commands",
		},
		{
			name: "task with multiple commands",
			task: &Task{
				Name:     "test",
				Commands: []string{"go test", "go vet"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error, got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Validate() error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}
