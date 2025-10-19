package finder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindBabfile(t *testing.T) {
	t.Run("finds Babfile in current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		babfilePath := filepath.Join(tmpDir, "Babfile")

		if err := os.WriteFile(babfilePath, []byte("test: {run: echo test}"), 0644); err != nil {
			t.Fatalf("failed to create test Babfile: %v", err)
		}

		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		found, err := FindBabfile()
		if err != nil {
			t.Errorf("FindBabfile() unexpected error: %v", err)
		}

		if found != babfilePath {
			t.Errorf("FindBabfile() = %q, want %q", found, babfilePath)
		}
	})

	t.Run("finds Babfile.yaml in current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		babfilePath := filepath.Join(tmpDir, "Babfile.yaml")

		if err := os.WriteFile(babfilePath, []byte("test: {run: echo test}"), 0644); err != nil {
			t.Fatalf("failed to create test Babfile.yaml: %v", err)
		}

		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		found, err := FindBabfile()
		if err != nil {
			t.Errorf("FindBabfile() unexpected error: %v", err)
		}

		if found != babfilePath {
			t.Errorf("FindBabfile() = %q, want %q", found, babfilePath)
		}
	})

	t.Run("finds Babfile.yml in current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		babfilePath := filepath.Join(tmpDir, "Babfile.yml")

		if err := os.WriteFile(babfilePath, []byte("test: {run: echo test}"), 0644); err != nil {
			t.Fatalf("failed to create test Babfile.yml: %v", err)
		}

		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		found, err := FindBabfile()
		if err != nil {
			t.Errorf("FindBabfile() unexpected error: %v", err)
		}

		if found != babfilePath {
			t.Errorf("FindBabfile() = %q, want %q", found, babfilePath)
		}
	})

	t.Run("finds babfile.yaml in current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		babfilePath := filepath.Join(tmpDir, "babfile.yaml")

		if err := os.WriteFile(babfilePath, []byte("test: {run: echo test}"), 0644); err != nil {
			t.Fatalf("failed to create test babfile.yaml: %v", err)
		}

		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		found, err := FindBabfile()
		if err != nil {
			t.Errorf("FindBabfile() unexpected error: %v", err)
		}

		if found != babfilePath {
			t.Errorf("FindBabfile() = %q, want %q", found, babfilePath)
		}
	})

	t.Run("finds babfile.yml in current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		babfilePath := filepath.Join(tmpDir, "babfile.yml")

		if err := os.WriteFile(babfilePath, []byte("test: {run: echo test}"), 0644); err != nil {
			t.Fatalf("failed to create test babfile.yml: %v", err)
		}

		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		found, err := FindBabfile()
		if err != nil {
			t.Errorf("FindBabfile() unexpected error: %v", err)
		}

		if found != babfilePath {
			t.Errorf("FindBabfile() = %q, want %q", found, babfilePath)
		}
	})

	t.Run("finds Babfile in parent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "subdir")
		babfilePath := filepath.Join(tmpDir, "Babfile")

		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatalf("failed to create subdirectory: %v", err)
		}

		if err := os.WriteFile(babfilePath, []byte("test: {run: echo test}"), 0644); err != nil {
			t.Fatalf("failed to create test Babfile: %v", err)
		}

		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		if err := os.Chdir(subDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		found, err := FindBabfile()
		if err != nil {
			t.Errorf("FindBabfile() unexpected error: %v", err)
		}

		if found != babfilePath {
			t.Errorf("FindBabfile() = %q, want %q", found, babfilePath)
		}
	})

	t.Run("finds Babfile in grandparent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "subdir")
		subSubDir := filepath.Join(subDir, "subsubdir")
		babfilePath := filepath.Join(tmpDir, "Babfile")

		if err := os.MkdirAll(subSubDir, 0755); err != nil {
			t.Fatalf("failed to create subdirectories: %v", err)
		}

		if err := os.WriteFile(babfilePath, []byte("test: {run: echo test}"), 0644); err != nil {
			t.Fatalf("failed to create test Babfile: %v", err)
		}

		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		if err := os.Chdir(subSubDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		found, err := FindBabfile()
		if err != nil {
			t.Errorf("FindBabfile() unexpected error: %v", err)
		}

		if found != babfilePath {
			t.Errorf("FindBabfile() = %q, want %q", found, babfilePath)
		}
	})

	t.Run("returns error when no Babfile found", func(t *testing.T) {
		tmpDir := t.TempDir()

		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		_, err := FindBabfile()
		if err == nil {
			t.Error("FindBabfile() expected error when no Babfile exists, got nil")
		}

		if err != nil && !contains(err.Error(), "no Babfile found") {
			t.Errorf("FindBabfile() error = %q, want error containing 'no Babfile found'", err.Error())
		}
	})

	t.Run("prefers Babfile over other variants", func(t *testing.T) {
		tmpDir := t.TempDir()

		babfilePath := filepath.Join(tmpDir, "Babfile")
		babfileYamlPath := filepath.Join(tmpDir, "Babfile.yaml")

		if err := os.WriteFile(babfilePath, []byte("test: {run: echo test}"), 0644); err != nil {
			t.Fatalf("failed to create test Babfile: %v", err)
		}
		if err := os.WriteFile(babfileYamlPath, []byte("other: {run: echo other}"), 0644); err != nil {
			t.Fatalf("failed to create test Babfile.yaml: %v", err)
		}

		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		found, err := FindBabfile()
		if err != nil {
			t.Errorf("FindBabfile() unexpected error: %v", err)
		}

		if found != babfilePath {
			t.Errorf("FindBabfile() = %q, want %q (should prefer Babfile)", found, babfilePath)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
