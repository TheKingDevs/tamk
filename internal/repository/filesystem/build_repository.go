package filesystem

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TheKingDevs/tamk/internal/domain/entity"
	"github.com/TheKingDevs/tamk/pkg/errors"
)

type BuildRepository struct{}

func NewBuildRepository() *BuildRepository {
	return &BuildRepository{}
}

type cacheEntry struct {
	Hash      string `json:"hash"`
	Timestamp string `json:"timestamp"`
}

func (r *BuildRepository) CalculateProjectHash(ctx context.Context, projectPath string) (string, error) {
	hasher := sha256.New()
	sourceFiles := r.collectSourceFiles(projectPath)

	for _, file := range sourceFiles {
		relPath, _ := filepath.Rel(projectPath, file)
		hasher.Write([]byte(relPath))
		hasher.Write([]byte{0})

		f, err := os.Open(file)
		if err != nil {
			continue
		}
		io.Copy(hasher, f)
		f.Close()

		hasher.Write([]byte{0})
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

var ignoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"secret":       true,
	".idea":        true,
	"build":        true,
	".gradle":      true,
	".tamk-run":    true,
	"cache":        true,
}

func (r *BuildRepository) collectSourceFiles(projectPath string) []string {
	var files []string
	watchDirs := []string{"src", "res"}
	watchFiles := []string{"AndroidManifest.xml", "tamk.config"}

	for _, dir := range watchDirs {
		dirPath := filepath.Join(projectPath, dir)
		filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				if info != nil && info.IsDir() && ignoredDirs[info.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			files = append(files, path)
			return nil
		})
	}

	for _, f := range watchFiles {
		path := filepath.Join(projectPath, f)
		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		}
	}

	sort.Strings(files)
	return files
}

func (r *BuildRepository) LoadCache(ctx context.Context, projectPath string) (*entity.BuildCache, error) {
	path := filepath.Join(projectPath, ".build_cache")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &entity.BuildCache{
		Hash: entry.Hash,
	}, nil
}

func (r *BuildRepository) SaveCache(ctx context.Context, projectPath string, cache *entity.BuildCache) error {
	entry := cacheEntry{Hash: cache.Hash}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(projectPath, ".build_cache"), data, 0o644)
}

func (r *BuildRepository) MustRecompile(ctx context.Context, projectPath string, currentHash string) (bool, error) {
	cache, err := r.LoadCache(ctx, projectPath)
	if err != nil {
		return true, nil
	}
	if cache.Hash == currentHash {
		return false, nil
	}
	return true, nil
}

func (r *BuildRepository) CleanCache(ctx context.Context, projectPath string) error {
	cachePath := filepath.Join(projectPath, ".build_cache")
	if _, err := os.Stat(cachePath); err == nil {
		return os.Remove(cachePath)
	}
	return nil
}

func SanitizePath(base, path string) (string, error) {
	if strings.ContainsAny(path, ";&|`$") {
		return "", fmt.Errorf("invalid characters in path")
	}

	fullPath := filepath.Join(base, filepath.Clean(path))
	if !strings.HasPrefix(fullPath, filepath.Clean(base)) {
		return "", fmt.Errorf("path traversal detected: %w", errors.ErrPathTraversal)
	}

	return fullPath, nil
}
