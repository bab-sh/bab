package parser

import (
	"testing"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		prefix   string
		wantErr  bool
		errMsg   string
		validate func(t *testing.T, tasks TaskMap)
	}{
		{
			name: "simple single task",
			data: map[string]interface{}{
				"hello": map[string]interface{}{
					"run": "echo hello",
				},
			},
			prefix:  "",
			wantErr: false,
			validate: func(t *testing.T, tasks TaskMap) {
				if len(tasks) != 1 {
					t.Errorf("expected 1 task, got %d", len(tasks))
				}
				if tasks["hello"] == nil {
					t.Error("task 'hello' not found")
				}
			},
		},
		{
			name: "multiple tasks at same level",
			data: map[string]interface{}{
				"build": map[string]interface{}{
					"run": "go build",
				},
				"test": map[string]interface{}{
					"run": "go test",
				},
			},
			prefix:  "",
			wantErr: false,
			validate: func(t *testing.T, tasks TaskMap) {
				if len(tasks) != 2 {
					t.Errorf("expected 2 tasks, got %d", len(tasks))
				}
				if tasks["build"] == nil {
					t.Error("task 'build' not found")
				}
				if tasks["test"] == nil {
					t.Error("task 'test' not found")
				}
			},
		},
		{
			name: "nested tasks",
			data: map[string]interface{}{
				"ci": map[string]interface{}{
					"test": map[string]interface{}{
						"run": "go test",
					},
					"lint": map[string]interface{}{
						"run": "golangci-lint run",
					},
				},
			},
			prefix:  "",
			wantErr: false,
			validate: func(t *testing.T, tasks TaskMap) {
				if len(tasks) != 2 {
					t.Errorf("expected 2 tasks, got %d", len(tasks))
				}
				if tasks["ci:test"] == nil {
					t.Error("task 'ci:test' not found")
				}
				if tasks["ci:lint"] == nil {
					t.Error("task 'ci:lint' not found")
				}
			},
		},
		{
			name: "task with run and nested tasks",
			data: map[string]interface{}{
				"ci": map[string]interface{}{
					"run": "echo starting ci",
					"test": map[string]interface{}{
						"run": "go test",
					},
				},
			},
			prefix:  "",
			wantErr: false,
			validate: func(t *testing.T, tasks TaskMap) {
				if len(tasks) != 2 {
					t.Errorf("expected 2 tasks, got %d", len(tasks))
				}
				if tasks["ci"] == nil {
					t.Error("task 'ci' not found")
				}
				if tasks["ci:test"] == nil {
					t.Error("task 'ci:test' not found")
				}
			},
		},
		{
			name: "with prefix",
			data: map[string]interface{}{
				"test": map[string]interface{}{
					"run": "go test",
				},
			},
			prefix:  "ci",
			wantErr: false,
			validate: func(t *testing.T, tasks TaskMap) {
				if len(tasks) != 1 {
					t.Errorf("expected 1 task, got %d", len(tasks))
				}
				if tasks["ci:test"] == nil {
					t.Error("task 'ci:test' not found")
				}
			},
		},
		{
			name: "deeply nested tasks",
			data: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": map[string]interface{}{
							"run": "echo abc",
						},
					},
				},
			},
			prefix:  "",
			wantErr: false,
			validate: func(t *testing.T, tasks TaskMap) {
				if len(tasks) != 1 {
					t.Errorf("expected 1 task, got %d", len(tasks))
				}
				if tasks["a:b:c"] == nil {
					t.Error("task 'a:b:c' not found")
				}
			},
		},
		{
			name: "task value is not a map",
			data: map[string]interface{}{
				"test": "not a map",
			},
			prefix:  "",
			wantErr: true,
			errMsg:  "must be a map",
		},
		{
			name: "task value is nil",
			data: map[string]interface{}{
				"test": nil,
			},
			prefix:  "",
			wantErr: true,
			errMsg:  "must be a map",
		},
		{
			name: "task value is a slice",
			data: map[string]interface{}{
				"test": []interface{}{"item1", "item2"},
			},
			prefix:  "",
			wantErr: true,
			errMsg:  "must be a map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := make(TaskMap)
			err := flatten(tt.data, tt.prefix, tasks)

			if tt.wantErr {
				if err == nil {
					t.Errorf("flatten() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("flatten() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("flatten() unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, tasks)
			}
		})
	}
}

func TestBuildTaskName(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		key    string
		want   string
	}{
		{"no prefix", "", "test", "test"},
		{"with prefix", "ci", "test", "ci:test"},
		{"nested prefix", "ci:build", "test", "ci:build:test"},
		{"empty key", "ci", "", "ci:"},
		{"both empty", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTaskName(tt.prefix, tt.key)
			if got != tt.want {
				t.Errorf("buildTaskName(%q, %q) = %q, want %q", tt.prefix, tt.key, got, tt.want)
			}
		})
	}
}

func TestGetNestedKeys(t *testing.T) {
	tests := []struct {
		name    string
		taskMap map[string]interface{}
		want    []string
	}{
		{
			name: "no nested keys",
			taskMap: map[string]interface{}{
				"run":  "echo test",
				"desc": "Test task",
				"deps": []string{"build"},
			},
			want: []string{},
		},
		{
			name: "only nested keys",
			taskMap: map[string]interface{}{
				"test": map[string]interface{}{"run": "go test"},
				"lint": map[string]interface{}{"run": "golangci-lint"},
			},
			want: []string{"test", "lint"},
		},
		{
			name: "mixed keys",
			taskMap: map[string]interface{}{
				"run":  "echo parent",
				"desc": "Parent task",
				"test": map[string]interface{}{"run": "go test"},
				"lint": map[string]interface{}{"run": "golangci-lint"},
			},
			want: []string{"test", "lint"},
		},
		{
			name:    "empty map",
			taskMap: map[string]interface{}{},
			want:    []string{},
		},
		{
			name: "single nested key",
			taskMap: map[string]interface{}{
				"run":   "echo test",
				"child": map[string]interface{}{"run": "echo child"},
			},
			want: []string{"child"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getNestedKeys(tt.taskMap)

			if len(got) != len(tt.want) {
				t.Errorf("getNestedKeys() returned %d keys, want %d", len(got), len(tt.want))
				return
			}

			wantMap := make(map[string]bool)
			for _, k := range tt.want {
				wantMap[k] = true
			}

			for _, k := range got {
				if !wantMap[k] {
					t.Errorf("getNestedKeys() returned unexpected key %q", k)
				}
			}
		})
	}
}

func TestProcessTaskNode(t *testing.T) {
	tests := []struct {
		name     string
		taskMap  map[string]interface{}
		taskName string
		wantErr  bool
		validate func(t *testing.T, tasks TaskMap)
	}{
		{
			name: "task with run command",
			taskMap: map[string]interface{}{
				"run": "echo hello",
			},
			taskName: "hello",
			wantErr:  false,
			validate: func(t *testing.T, tasks TaskMap) {
				if len(tasks) != 1 {
					t.Errorf("expected 1 task, got %d", len(tasks))
				}
				if tasks["hello"] == nil {
					t.Error("task 'hello' not found")
				}
			},
		},
		{
			name: "task without run command but with nested tasks",
			taskMap: map[string]interface{}{
				"test": map[string]interface{}{
					"run": "go test",
				},
			},
			taskName: "ci",
			wantErr:  false,
			validate: func(t *testing.T, tasks TaskMap) {
				if len(tasks) != 1 {
					t.Errorf("expected 1 task, got %d", len(tasks))
				}
				if tasks["ci:test"] == nil {
					t.Error("task 'ci:test' not found")
				}
			},
		},
		{
			name: "task with both run and nested tasks",
			taskMap: map[string]interface{}{
				"run": "echo parent",
				"test": map[string]interface{}{
					"run": "go test",
				},
			},
			taskName: "ci",
			wantErr:  false,
			validate: func(t *testing.T, tasks TaskMap) {
				if len(tasks) != 2 {
					t.Errorf("expected 2 tasks, got %d", len(tasks))
				}
				if tasks["ci"] == nil {
					t.Error("task 'ci' not found")
				}
				if tasks["ci:test"] == nil {
					t.Error("task 'ci:test' not found")
				}
			},
		},
		{
			name: "task with invalid run command",
			taskMap: map[string]interface{}{
				"run": "",
			},
			taskName: "test",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := make(TaskMap)
			err := processTaskNode(tt.taskMap, tt.taskName, tasks)

			if tt.wantErr {
				if err == nil {
					t.Error("processTaskNode() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("processTaskNode() unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, tasks)
			}
		})
	}
}
