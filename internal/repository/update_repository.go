package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheKingDevs/tamk/internal/config"
	"github.com/TheKingDevs/tamk/internal/domain/entity"
)

const (
	githubAPI     = "https://api.github.com/repos/TheKingDevs/tamk/releases"
	cacheDuration = 6 * time.Hour
)

type githubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	PublishedAt string `json:"published_at"`
	Prerelease  bool   `json:"prerelease"`
	Draft       bool   `json:"draft"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type UpdateRepository struct {
	client *http.Client
	cache  string
}

func NewUpdateRepository() *UpdateRepository {
	cacheDir := filepath.Join(os.TempDir(), "tamk-update")
	os.MkdirAll(cacheDir, 0o755)

	return &UpdateRepository{
		client: &http.Client{Timeout: 30 * time.Second},
		cache:  filepath.Join(cacheDir, "update_cache.json"),
	}
}

func (r *UpdateRepository) CheckForUpdates(ctx context.Context, currentVersion string) (*entity.UpdateInfo, error) {
	if info, err := r.loadCache(); err == nil && info != nil {
		return info, nil
	}

	req, err := http.NewRequest("GET", githubAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "tamk-updater/"+config.Version)
	req.Header.Set("Accept", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check updates: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var releases []githubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}

	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}
		if release.TagName == "" {
			continue
		}

		latestTag := strings.TrimPrefix(release.TagName, "v")
		if latestTag == strings.TrimPrefix(currentVersion, "v") {
			continue
		}

		downloadURL := ""
		if len(release.Assets) > 0 {
			downloadURL = release.Assets[0].BrowserDownloadURL
		}

		info := &entity.UpdateInfo{
			CurrentVersion: currentVersion,
			LatestVersion:  latestTag,
			Level:          r.DetectUpdateLevel(latestTag, currentVersion, release.Body),
			ReleaseNotes:   release.Body,
			DownloadURL:    downloadURL,
			ReleaseDate:    release.PublishedAt,
		}

		r.saveCache(info)
		return info, nil
	}

	return nil, nil
}

func (r *UpdateRepository) DetectUpdateLevel(latestVersion, currentVersion, releaseNotes string) entity.UpdateLevel {
	notes := strings.ToLower(releaseNotes)

	if strings.Contains(notes, "critical") ||
		strings.Contains(notes, "security") ||
		strings.Contains(notes, "urgent") {
		return entity.UpdateLevelCritical
	}

	if strings.Contains(notes, "patch") ||
		strings.Contains(notes, "bugfix") ||
		strings.Contains(notes, "fix") {
		return entity.UpdateLevelPatch
	}

	latest := strings.TrimPrefix(latestVersion, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")

	if len(latestParts) > 0 && len(currentParts) > 0 {
		if latestParts[0] > currentParts[0] {
			return entity.UpdateLevelMajor
		}
		if len(latestParts) > 1 && len(currentParts) > 1 {
			if latestParts[1] > currentParts[1] {
				return entity.UpdateLevelMinor
			}
		}
	}

	return entity.UpdateLevelOptional
}

func (r *UpdateRepository) DownloadUpdate(ctx context.Context, url, destPath string) error {
	resp, err := r.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (r *UpdateRepository) GetCachePath() string {
	return r.cache
}

func (r *UpdateRepository) ClearCache() error {
	if _, err := os.Stat(r.cache); err == nil {
		return os.Remove(r.cache)
	}
	return nil
}

type cachedInfo struct {
	Info      entity.UpdateInfo `json:"info"`
	Timestamp time.Time         `json:"timestamp"`
}

func (r *UpdateRepository) loadCache() (*entity.UpdateInfo, error) {
	data, err := os.ReadFile(r.cache)
	if err != nil {
		return nil, err
	}

	var cached cachedInfo
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}

	if time.Since(cached.Timestamp) > cacheDuration {
		return nil, fmt.Errorf("cache expired")
	}

	return &cached.Info, nil
}

func (r *UpdateRepository) saveCache(info *entity.UpdateInfo) {
	cached := cachedInfo{
		Info:      *info,
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(cached)
	os.WriteFile(r.cache, data, 0o644)
}
