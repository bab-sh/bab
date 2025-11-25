package finder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func normalizePath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}

func TestFindBabfileVariants(t *testing.T) {
	variants := []string{"Babfile", "Babfile.yaml", "Babfile.yml", "babfile.yaml", "babfile.yml"}

	for _, variant := range variants {
		t.Run("finds "+variant+" in current directory", func(t *testing.T) {
			tmpDir := t.TempDir()
			babfilePath := filepath.Join(tmpDir, variant)

			if err := os.WriteFile(babfilePath, []byte("test: {run: echo test}"), 0600); err != nil {
				t.Fatalf("failed to create test %s: %v", variant, err)
			}

			oldDir, _ := os.Getwd()
			defer func() { _ = os.Chdir(oldDir) }()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change directory: %v", err)
			}

			found, err := FindBabfile()
			if err != nil {
				t.Errorf("FindBabfile() unexpected error: %v", err)
			}

			foundNorm := normalizePath(t, found)
			wantNorm := normalizePath(t, babfilePath)
			if filepath.Dir(foundNorm) != filepath.Dir(wantNorm) ||
				!strings.EqualFold(filepath.Base(foundNorm), filepath.Base(wantNorm)) {
				t.Errorf("FindBabfile() = %q, want %q", found, babfilePath)
			}
		})
	}
}

func TestFindBabfileInParentDirectories(t *testing.T) {
	t.Run("finds Babfile in parent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "subdir")
		babfilePath := filepath.Join(tmpDir, "Babfile")

		if err := os.Mkdir(subDir, 0750); err != nil {
			t.Fatalf("failed to create subdirectory: %v", err)
		}

		if err := os.WriteFile(babfilePath, []byte("test: {run: echo test}"), 0600); err != nil {
			t.Fatalf("failed to create test Babfile: %v", err)
		}

		oldDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(subDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		found, err := FindBabfile()
		if err != nil {
			t.Errorf("FindBabfile() unexpected error: %v", err)
		}

		if normalizePath(t, found) != normalizePath(t, babfilePath) {
			t.Errorf("FindBabfile() = %q, want %q", found, babfilePath)
		}
	})

	t.Run("finds Babfile in grandparent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "subdir")
		subSubDir := filepath.Join(subDir, "subsubdir")
		babfilePath := filepath.Join(tmpDir, "Babfile")

		if err := os.MkdirAll(subSubDir, 0750); err != nil {
			t.Fatalf("failed to create subdirectories: %v", err)
		}

		if err := os.WriteFile(babfilePath, []byte("test: {run: echo test}"), 0600); err != nil {
			t.Fatalf("failed to create test Babfile: %v", err)
		}

		oldDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(subSubDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		found, err := FindBabfile()
		if err != nil {
			t.Errorf("FindBabfile() unexpected error: %v", err)
		}

		if normalizePath(t, found) != normalizePath(t, babfilePath) {
			t.Errorf("FindBabfile() = %q, want %q", found, babfilePath)
		}
	})
}

func TestFindBabfileErrors(t *testing.T) {
	t.Run("returns error when no Babfile found", func(t *testing.T) {
		tmpDir := t.TempDir()

		oldDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		_, err := FindBabfile()
		if err == nil {
			t.Error("FindBabfile() expected error when no Babfile exists, got nil")
		}

		if err != nil && !strings.Contains(err.Error(), "no Babfile found") {
			t.Errorf("FindBabfile() error = %q, want error containing 'no Babfile found'", err.Error())
		}
	})
}

func TestFindBabfilePriority(t *testing.T) {
	t.Run("prefers Babfile over other variants", func(t *testing.T) {
		tmpDir := t.TempDir()

		babfilePath := filepath.Join(tmpDir, "Babfile")
		babfileYamlPath := filepath.Join(tmpDir, "Babfile.yaml")

		if err := os.WriteFile(babfilePath, []byte("test: {run: echo test}"), 0600); err != nil {
			t.Fatalf("failed to create test Babfile: %v", err)
		}
		if err := os.WriteFile(babfileYamlPath, []byte("other: {run: echo other}"), 0600); err != nil {
			t.Fatalf("failed to create test Babfile.yaml: %v", err)
		}

		oldDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		found, err := FindBabfile()
		if err != nil {
			t.Errorf("FindBabfile() unexpected error: %v", err)
		}

		if normalizePath(t, found) != normalizePath(t, babfilePath) {
			t.Errorf("FindBabfile() = %q, want %q", found, babfilePath)
		}
	})
}
