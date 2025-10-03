package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindBabfile(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore current directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	tests := []struct {
		name         string
		files        []string
		expectedFile string
	}{
		{
			name:         "find Babfile",
			files:        []string{"Babfile"},
			expectedFile: "Babfile",
		},
		{
			name:         "find Babfile.yaml",
			files:        []string{"Babfile.yaml"},
			expectedFile: "Babfile.yaml",
		},
		{
			name:         "find Babfile.yml",
			files:        []string{"Babfile.yml"},
			expectedFile: "Babfile.yml",
		},
		{
			name:         "prefer Babfile over yaml",
			files:        []string{"Babfile", "Babfile.yaml", "Babfile.yml"},
			expectedFile: "Babfile",
		},
		{
			name:         "no babfile found",
			files:        []string{"README.md"},
			expectedFile: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a unique subdirectory for each test
			testDir := filepath.Join(tmpDir, tt.name)
			if err := os.MkdirAll(testDir, 0750); err != nil {
				t.Fatalf("failed to create test directory: %v", err)
			}

			// Create test files
			for _, file := range tt.files {
				filePath := filepath.Join(testDir, file)
				if err := os.WriteFile(filePath, []byte(""), 0600); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			// Change to test directory
			if err := os.Chdir(testDir); err != nil {
				t.Fatalf("failed to change directory: %v", err)
			}

			result := findBabfile()
			if result != tt.expectedFile {
				t.Errorf("findBabfile() = %q, want %q", result, tt.expectedFile)
			}
		})
	}
}

func TestLoadRegistry(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		babfile    string
		content    string
		wantErr    bool
		checkTasks []string
	}{
		{
			name:    "valid babfile",
			babfile: filepath.Join(tmpDir, "valid-Babfile"),
			content: `
build:
  desc: Build
  run: go build

test:
  desc: Test
  run: go test
`,
			wantErr:    false,
			checkTasks: []string{"build", "test"},
		},
		{
			name:    "invalid babfile",
			babfile: filepath.Join(tmpDir, "invalid-Babfile"),
			content: `invalid: [yaml`,
			wantErr: true,
		},
		{
			name:    "nonexistent babfile",
			babfile: filepath.Join(tmpDir, "nonexistent"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.content != "" {
				if err := os.WriteFile(tt.babfile, []byte(tt.content), 0600); err != nil {
					t.Fatalf("failed to create test babfile: %v", err)
				}
			}

			// Set the babfile flag
			oldBabfile := babfile
			babfile = tt.babfile
			defer func() { babfile = oldBabfile }()

			reg, err := loadRegistry()

			if tt.wantErr {
				if err == nil {
					t.Error("loadRegistry() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("loadRegistry() unexpected error: %v", err)
			}

			// Verify expected tasks are registered
			for _, taskName := range tt.checkTasks {
				if _, err := reg.Get(taskName); err != nil {
					t.Errorf("loadRegistry() task %q not found: %v", taskName, err)
				}
			}
		})
	}
}

func TestLoadRegistryNoBabfile(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore current directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Change to empty directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Reset babfile flag
	oldBabfile := babfile
	babfile = ""
	defer func() { babfile = oldBabfile }()

	_, err = loadRegistry()
	if err == nil {
		t.Error("loadRegistry() expected error when no Babfile found, got nil")
	}
}
