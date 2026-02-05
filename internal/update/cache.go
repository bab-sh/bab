package update

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bab-sh/bab/internal/paths"
)

const (
	cacheFileName = "update-check.json"
	cacheDuration = 24 * time.Hour
)

type cache struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
	ETag          string    `json:"etag"`
}

func loadCache() *cache {
	path, _ := paths.SearchCacheFile(cacheFileName)
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var c cache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil
	}

	if c.LatestVersion == "" {
		return nil
	}

	if c.CheckedAt.After(time.Now().Add(time.Hour)) {
		return nil
	}

	return &c
}

func saveCache(c *cache) {
	path, err := paths.CacheFile(cacheFileName)
	if err != nil {
		return
	}

	dir := filepath.Dir(path)
	cleanupTempFiles(dir)

	c.CheckedAt = c.CheckedAt.UTC()

	data, err := json.Marshal(c)
	if err != nil {
		return
	}

	tmp, err := os.CreateTemp(dir, "update-check-*.tmp")
	if err != nil {
		return
	}
	tmpName := tmp.Name()

	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return
	}

	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return
	}

	if err := tmp.Close(); err != nil {
		return
	}

	if runtime.GOOS == "windows" {
		_ = os.Remove(path)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return
	}

	success = true
}

func cleanupTempFiles(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "update-check-") && strings.HasSuffix(name, ".tmp") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if time.Since(info.ModTime()) > time.Hour {
				_ = os.Remove(filepath.Join(dir, name))
			}
		}
	}
}

func (c *cache) isValid() bool {
	if c == nil {
		return false
	}
	return time.Since(c.CheckedAt) < cacheDuration
}
