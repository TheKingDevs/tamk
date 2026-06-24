package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/TheKingDevs/tamk/internal/config"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

type LibraryInfo struct {
	Name         string   `json:"name"`
	Group        string   `json:"group"`
	Artifact     string   `json:"artifact"`
	Version      string   `json:"version"`
	Jar          string   `json:"jar"`
	URL          string   `json:"url"`
	Category     string   `json:"category"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
}

type LibraryRegistry struct {
	Version    string                  `json:"version"`
	Libraries  map[string]*LibraryInfo `json:"libraries"`
	Categories map[string]string       `json:"categories"`
}

type LibraryManager struct {
	cfg      *config.Config
	registry *LibraryRegistry
	client   *http.Client
	mu       sync.RWMutex
}

func NewLibraryManager(cfg *config.Config) *LibraryManager {
	return &LibraryManager{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (m *LibraryManager) loadRegistry() error {
	m.mu.RLock()
	if m.registry != nil {
		m.mu.RUnlock()
		return nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.registry != nil {
		return nil
	}

	registryPath := filepath.Join(m.cfg.TAMKHome, "libs", "libraries.json")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return fmt.Errorf("failed to load library registry: %w", err)
	}

	var registry LibraryRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("failed to parse library registry: %w", err)
	}

	m.registry = &registry
	return nil
}

func (m *LibraryManager) ListLibraries() error {
	if err := m.loadRegistry(); err != nil {
		return err
	}

	categories := make(map[string][]*LibraryInfo)
	for _, lib := range m.registry.Libraries {
		categories[lib.Category] = append(categories[lib.Category], lib)
	}

	sortedCats := make([]string, 0, len(categories))
	for cat := range categories {
		sortedCats = append(sortedCats, cat)
	}
	sort.Strings(sortedCats)

	fmt.Println()
	fmt.Println("  Available libraries:")
	fmt.Println()

	for _, cat := range sortedCats {
		catName := m.registry.Categories[cat]
		if catName == "" {
			catName = cat
		}
		fmt.Printf("  \033[1;33m%s\033[0m\n", catName)
		for _, lib := range categories[cat] {
			installed := m.isInstalled(lib.Artifact)
			status := "\033[90m[not installed]\033[0m"
			if installed {
				status = "\033[32m[installed]\033[0m"
			}
			fmt.Printf("    \033[36m%-30s\033[0m %s %s\n", lib.Artifact, lib.Version, status)
			fmt.Printf("      %s\n", lib.Description)
		}
		fmt.Println()
	}

	return nil
}

func (m *LibraryManager) isInstalled(artifact string) bool {
	libsDir := filepath.Join(m.cfg.TAMKHome, "libs")
	entries, err := os.ReadDir(libsDir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			sub, err := os.ReadDir(filepath.Join(libsDir, entry.Name()))
			if err == nil {
				for _, s := range sub {
					if s.IsDir() && s.Name() == artifact {
						return true
					}
				}
			}
		}
	}
	return false
}

func (m *LibraryManager) InstallLibrary(ctx context.Context, artifact string) error {
	if err := m.loadRegistry(); err != nil {
		return err
	}

	lib, ok := m.registry.Libraries[artifact]
	if !ok {
		return fmt.Errorf("library '%s' not found. Use 'tamk libs list' to see available libraries", artifact)
	}

	logger.Info("Installing library", "name", lib.Name, "version", lib.Version)

	for _, dep := range lib.Dependencies {
		if !m.isInstalled(dep) {
			logger.Info("Installing dependency", "name", dep)
			if err := m.InstallLibrary(ctx, dep); err != nil {
				return fmt.Errorf("failed to install dependency %s: %w", dep, err)
			}
		}
	}

	installDir := filepath.Join(m.cfg.TAMKHome, "libs", lib.Category, lib.Artifact)
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	jarPath := filepath.Join(installDir, lib.Jar)
	if _, err := os.Stat(jarPath); err == nil {
		logger.Info("Library already installed", "path", jarPath)
		return nil
	}

	logger.Step("Downloading", "url", lib.URL)
	if err := m.downloadFile(ctx, lib.URL, jarPath); err != nil {
		return fmt.Errorf("failed to download library: %w", err)
	}

	logger.Success("Installed", "library", lib.Name, "path", jarPath)
	return nil
}

func (m *LibraryManager) RemoveLibrary(artifact string) error {
	if err := m.loadRegistry(); err != nil {
		return err
	}

	lib, ok := m.registry.Libraries[artifact]
	if !ok {
		return fmt.Errorf("library '%s' not found", artifact)
	}

	installDir := filepath.Join(m.cfg.TAMKHome, "libs", lib.Category, lib.Artifact)
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return fmt.Errorf("library '%s' is not installed", artifact)
	}

	if err := os.RemoveAll(installDir); err != nil {
		return fmt.Errorf("failed to remove library: %w", err)
	}

	logger.Success("Removed", "library", lib.Name)
	return nil
}

func (m *LibraryManager) GetClasspath() (string, error) {
	libsDir := filepath.Join(m.cfg.TAMKHome, "libs")
	var jars []string

	filepath.Walk(libsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".jar") {
			jars = append(jars, path)
		}
		return nil
	})

	sort.Strings(jars)
	return strings.Join(jars, string(os.PathListSeparator)), nil
}

func (m *LibraryManager) GetInstalledLibraries() []string {
	libsDir := filepath.Join(m.cfg.TAMKHome, "libs")
	var installed []string

	filepath.Walk(libsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".jar") {
			rel, _ := filepath.Rel(libsDir, path)
			parts := strings.Split(rel, string(os.PathSeparator))
			if len(parts) >= 2 {
				installed = append(installed, parts[len(parts)-2])
			}
		}
		return nil
	})

	return installed
}

func (m *LibraryManager) downloadFile(ctx context.Context, url, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
