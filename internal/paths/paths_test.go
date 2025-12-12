package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

func TestCacheFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmpDir)
	xdg.Reload()

	path, err := CacheFile("test.json")
	if err != nil {
		t.Fatalf("CacheFile() error = %v", err)
	}

	want := filepath.Join(tmpDir, "bab", "test.json")
	if path != want {
		t.Errorf("CacheFile() = %q, want %q", path, want)
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("CacheFile() should create parent directory")
	}
}
