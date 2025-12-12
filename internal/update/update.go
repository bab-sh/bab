package update

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	goversion "github.com/hashicorp/go-version"
)

const (
	releaseURL      = "https://api.github.com/repos/bab-sh/bab/releases/latest"
	timeout         = 5 * time.Second
	maxResponseSize = 1 << 20
)

type Info struct {
	CurrentVersion string
	LatestVersion  string
}

func StartBackgroundRefresh(currentVersion string) {
	if shouldSkip(currentVersion) {
		return
	}
	cached := loadCache()
	if cached == nil || !cached.isValid() {
		go refreshCache(cached)
	}
}

func CheckCached(currentVersion string) *Info {
	if shouldSkip(currentVersion) {
		return nil
	}
	cached := loadCache()
	if cached != nil && isNewer(cached.LatestVersion, currentVersion) {
		return &Info{
			CurrentVersion: currentVersion,
			LatestVersion:  cached.LatestVersion,
		}
	}
	return nil
}

func shouldSkip(currentVersion string) bool {
	if os.Getenv("BAB_NO_UPDATE_CHECK") == "1" {
		return true
	}

	if ci := os.Getenv("CI"); ci == "true" || ci == "1" {
		return true
	}

	if currentVersion == "dev" || currentVersion == "" {
		return true
	}

	return false
}

func refreshCache(cached *cache) {
	latest, etag, err := fetchLatestVersion(cached)
	if err != nil {
		return
	}

	if latest == "" {
		return
	}

	saveCache(&cache{
		CheckedAt:     time.Now(),
		LatestVersion: latest,
		ETag:          etag,
	})
}

func fetchLatestVersion(cached *cache) (version, etag string, err error) {
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(http.MethodGet, releaseURL, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("User-Agent", "bab")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	if cached != nil && cached.ETag != "" {
		req.Header.Set("If-None-Match", cached.ETag)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotModified && cached != nil {
		return cached.LatestVersion, cached.ETag, nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", errors.New("unexpected status code")
	}

	limitedReader := io.LimitReader(resp.Body, maxResponseSize)

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(limitedReader).Decode(&release); err != nil {
		return "", "", err
	}

	if release.TagName == "" {
		return "", "", errors.New("empty tag_name in response")
	}

	return release.TagName, resp.Header.Get("ETag"), nil
}

func isNewer(latest, current string) bool {
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	if latest == "" || current == "" {
		return false
	}

	latestV, err := goversion.NewVersion(latest)
	if err != nil {
		return false
	}

	currentV, err := goversion.NewVersion(current)
	if err != nil {
		return false
	}

	return latestV.GreaterThan(currentV)
}
