package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adrg/xdg"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name    string
		latest  string
		current string
		want    bool
	}{
		{"newer major", "v2.0.0", "v1.0.0", true},
		{"newer minor", "v1.1.0", "v1.0.0", true},
		{"newer patch", "v1.0.1", "v1.0.0", true},
		{"same version", "v1.0.0", "v1.0.0", false},
		{"older version", "v1.0.0", "v2.0.0", false},
		{"without v prefix", "1.0.1", "1.0.0", true},
		{"mixed prefixes", "v1.0.1", "1.0.0", true},
		{"prerelease current", "v1.0.0", "v1.0.0-beta", true},
		{"invalid latest", "invalid", "v1.0.0", false},
		{"invalid current", "v1.0.0", "invalid", false},
		{"empty latest", "", "v1.0.0", false},
		{"empty current", "v1.0.0", "", false},
		{"both empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNewer(tt.latest, tt.current)
			if got != tt.want {
				t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
			}
		})
	}
}

func TestShouldSkip(t *testing.T) {
	t.Run("skips dev version", func(t *testing.T) {
		if !shouldSkip("dev") {
			t.Error("expected true for dev version")
		}
	})

	t.Run("skips empty version", func(t *testing.T) {
		if !shouldSkip("") {
			t.Error("expected true for empty version")
		}
	})

	t.Run("skips with BAB_NO_UPDATE_CHECK", func(t *testing.T) {
		t.Setenv("BAB_NO_UPDATE_CHECK", "1")
		if !shouldSkip("v1.0.0") {
			t.Error("expected true when BAB_NO_UPDATE_CHECK=1")
		}
	})

	t.Run("skips in CI", func(t *testing.T) {
		t.Setenv("CI", "true")
		if !shouldSkip("v1.0.0") {
			t.Error("expected true when CI=true")
		}
	})

	t.Run("does not skip normal version", func(t *testing.T) {
		t.Setenv("CI", "")
		t.Setenv("BAB_NO_UPDATE_CHECK", "")
		if shouldSkip("v1.0.0") {
			t.Error("expected false for normal version")
		}
	})
}

func TestCacheOperations(t *testing.T) {
	t.Run("save and load", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", tmpDir)
		xdg.Reload()

		c := &cache{
			CheckedAt:     time.Now(),
			LatestVersion: "v1.2.3",
			ETag:          "abc123",
		}
		saveCache(c)

		loaded := loadCache()
		if loaded == nil {
			t.Fatal("loadCache() returned nil")
		}
		if loaded.LatestVersion != c.LatestVersion {
			t.Errorf("version = %q, want %q", loaded.LatestVersion, c.LatestVersion)
		}
	})

	t.Run("nonexistent returns nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", tmpDir)
		xdg.Reload()

		if loadCache() != nil {
			t.Error("expected nil for nonexistent cache")
		}
	})

	t.Run("corrupted returns nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", tmpDir)
		xdg.Reload()

		dir := filepath.Join(tmpDir, "bab")
		_ = os.MkdirAll(dir, 0o700)
		_ = os.WriteFile(filepath.Join(dir, "update-check.json"), []byte("invalid"), 0o600)

		if loadCache() != nil {
			t.Error("expected nil for corrupted cache")
		}
	})

	t.Run("empty version returns nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", tmpDir)
		xdg.Reload()

		dir := filepath.Join(tmpDir, "bab")
		_ = os.MkdirAll(dir, 0o700)
		_ = os.WriteFile(filepath.Join(dir, "update-check.json"), []byte(`{"latest_version":""}`), 0o600)

		if loadCache() != nil {
			t.Error("expected nil for empty version")
		}
	})

	t.Run("future timestamp returns nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", tmpDir)
		xdg.Reload()

		dir := filepath.Join(tmpDir, "bab")
		_ = os.MkdirAll(dir, 0o700)
		future := time.Now().Add(48 * time.Hour).Format(time.RFC3339)
		_ = os.WriteFile(filepath.Join(dir, "update-check.json"), []byte(`{"checked_at":"`+future+`","latest_version":"v1.0.0"}`), 0o600)

		if loadCache() != nil {
			t.Error("expected nil for future timestamp")
		}
	})

	t.Run("stores UTC time", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", tmpDir)
		xdg.Reload()

		saveCache(&cache{CheckedAt: time.Now(), LatestVersion: "v1.0.0"})

		data, _ := os.ReadFile(filepath.Join(tmpDir, "bab", "update-check.json"))
		if !strings.Contains(string(data), "Z") && !strings.Contains(string(data), "+00:00") {
			t.Errorf("expected UTC time, got: %s", data)
		}
	})
}

func TestCacheIsValid(t *testing.T) {
	tests := []struct {
		name      string
		checkedAt time.Time
		want      bool
	}{
		{"recent", time.Now().Add(-1 * time.Hour), true},
		{"old", time.Now().Add(-25 * time.Hour), false},
		{"exactly 24h", time.Now().Add(-24 * time.Hour), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cache{CheckedAt: tt.checkedAt, LatestVersion: "v1.0.0"}
			if got := c.isValid(); got != tt.want {
				t.Errorf("isValid() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("nil is invalid", func(t *testing.T) {
		var c *cache
		if c.isValid() {
			t.Error("nil cache should be invalid")
		}
	})
}

func TestCleanupTempFiles(t *testing.T) {
	tmpDir := t.TempDir()

	stale := filepath.Join(tmpDir, "update-check-123.tmp")
	_ = os.WriteFile(stale, []byte("stale"), 0o600)
	_ = os.Chtimes(stale, time.Now().Add(-2*time.Hour), time.Now().Add(-2*time.Hour))

	recent := filepath.Join(tmpDir, "update-check-456.tmp")
	_ = os.WriteFile(recent, []byte("recent"), 0o600)

	cleanupTempFiles(tmpDir)

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("stale file should be removed")
	}
	if _, err := os.Stat(recent); err != nil {
		t.Error("recent file should remain")
	}
}

func TestCheckCached(t *testing.T) {
	t.Run("disabled returns nil", func(t *testing.T) {
		t.Setenv("BAB_NO_UPDATE_CHECK", "1")
		if CheckCached("v1.0.0") != nil {
			t.Error("expected nil when disabled")
		}
	})

	t.Run("dev returns nil", func(t *testing.T) {
		if CheckCached("dev") != nil {
			t.Error("expected nil for dev version")
		}
	})

	t.Run("no cache returns nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", tmpDir)
		t.Setenv("CI", "")
		t.Setenv("BAB_NO_UPDATE_CHECK", "")
		xdg.Reload()

		if CheckCached("v1.0.0") != nil {
			t.Error("expected nil when no cache")
		}
	})

	t.Run("newer version returns info", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", tmpDir)
		t.Setenv("CI", "")
		t.Setenv("BAB_NO_UPDATE_CHECK", "")
		xdg.Reload()

		saveCache(&cache{CheckedAt: time.Now(), LatestVersion: "v2.0.0"})

		result := CheckCached("v1.0.0")
		if result == nil {
			t.Fatal("expected UpdateInfo")
		}
		if result.LatestVersion != "v2.0.0" {
			t.Errorf("LatestVersion = %q, want v2.0.0", result.LatestVersion)
		}
	})

	t.Run("same version returns nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", tmpDir)
		t.Setenv("CI", "")
		t.Setenv("BAB_NO_UPDATE_CHECK", "")
		xdg.Reload()

		saveCache(&cache{CheckedAt: time.Now(), LatestVersion: "v1.0.0"})

		if CheckCached("v1.0.0") != nil {
			t.Error("expected nil for same version")
		}
	})
}
