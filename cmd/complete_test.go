package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCompleteTaskNames(t *testing.T) {
	tests := []struct {
		name        string
		babfileYAML string
		args        []string
		toComplete  string
		wantTasks   []string
		wantEmpty   bool
		createFile  bool
	}{
		{
			name: "complete all tasks",
			babfileYAML: `tasks:
  build:
    desc: Build the project
    run:
      - cmd: go build
  test:
    desc: Run tests
    run:
      - cmd: go test`,
			args:       []string{},
			toComplete: "",
			wantTasks:  []string{"build", "test"},
			createFile: true,
		},
		{
			name: "complete with prefix",
			babfileYAML: `tasks:
  build:
    run:
      - cmd: echo build
  test:
    run:
      - cmd: echo test
  testrace:
    run:
      - cmd: echo test race`,
			args:       []string{},
			toComplete: "te",
			wantTasks:  []string{"test", "testrace"},
			createFile: true,
		},
		{
			name: "no completions when arg already provided",
			babfileYAML: `tasks:
  build:
    run:
      - cmd: echo build`,
			args:       []string{"build"},
			toComplete: "",
			wantEmpty:  true,
			createFile: true,
		},
		{
			name:        "no babfile returns empty",
			babfileYAML: "",
			args:        []string{},
			toComplete:  "",
			wantEmpty:   true,
			createFile:  false,
		},
		{
			name: "includes description in completion",
			babfileYAML: `tasks:
  deploy:
    desc: Deploy to production
    run:
      - cmd: echo deploy`,
			args:       []string{},
			toComplete: "",
			wantTasks:  []string{"deploy\tDeploy to production"},
			createFile: true,
		},
		{
			name: "hierarchical task names",
			babfileYAML: `tasks:
  test:
    unit:
      desc: Run unit tests
      run:
        - cmd: go test
    integration:
      desc: Run integration tests
      run:
        - cmd: go test -tags=integration`,
			args:       []string{},
			toComplete: "test:",
			wantTasks:  []string{"test:unit", "test:integration"},
			createFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.createFile {
				babfilePath := filepath.Join(tmpDir, "Babfile.yml")
				if err := os.WriteFile(babfilePath, []byte(tt.babfileYAML), 0600); err != nil {
					t.Fatalf("failed to create test Babfile: %v", err)
				}
			}

			oldDir, _ := os.Getwd()
			defer func() { _ = os.Chdir(oldDir) }()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change directory: %v", err)
			}

			cmd := &cobra.Command{}
			completions, directive := completeTaskNames(cmd, tt.args, tt.toComplete)

			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
			}

			if tt.wantEmpty {
				if len(completions) != 0 {
					t.Errorf("expected no completions, got %v", completions)
				}
				return
			}

			for _, want := range tt.wantTasks {
				found := false
				wantName := strings.Split(want, "\t")[0]
				for _, got := range completions {
					gotName := strings.Split(got, "\t")[0]
					if gotName == wantName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected completion %q not found in %v", wantName, completions)
				}
			}
		})
	}
}

func TestCompleteTaskNames_AlwaysReturnsNoFileComp(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	_ = os.Chdir(tmpDir)

	cmd := &cobra.Command{}
	_, directive := completeTaskNames(cmd, []string{}, "")

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
}

func TestCompleteTaskNames_CompletionsAreSorted(t *testing.T) {
	tmpDir := t.TempDir()
	babfilePath := filepath.Join(tmpDir, "Babfile.yml")
	babfileContent := `tasks:
  zebra:
    run:
      - cmd: echo zebra
  alpha:
    run:
      - cmd: echo alpha
  middle:
    run:
      - cmd: echo middle`

	if err := os.WriteFile(babfilePath, []byte(babfileContent), 0600); err != nil {
		t.Fatalf("failed to create test Babfile: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := &cobra.Command{}
	completions, _ := completeTaskNames(cmd, []string{}, "")

	if len(completions) != 3 {
		t.Fatalf("expected 3 completions, got %d", len(completions))
	}

	for i := 1; i < len(completions); i++ {
		prev := strings.Split(completions[i-1], "\t")[0]
		curr := strings.Split(completions[i], "\t")[0]
		if prev > curr {
			t.Errorf("completions not sorted: %q should come before %q", curr, prev)
		}
	}
}
