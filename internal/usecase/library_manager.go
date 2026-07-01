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
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load library registry: %w", err)
		}
		logger.Debug(fmt.Sprintf("libraries.json not found at %s, creating default registry", registryPath))
		registry := defaultLibraryRegistry()
		if err := m.persistRegistry(registryPath, registry); err != nil {
			logger.Warn(fmt.Sprintf("could not persist default library registry: %v", err))
		}
		m.registry = registry
		return nil
	}

	var registry LibraryRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("failed to parse library registry: %w", err)
	}

	m.registry = &registry
	return nil
}

func (m *LibraryManager) persistRegistry(path string, reg *LibraryRegistry) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create libs directory: %w", err)
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}
	return nil
}

func defaultLibraryRegistry() *LibraryRegistry {
	return &LibraryRegistry{
		Version: "1.0.0",
		Categories: map[string]string{
			"kotlin":      "Kotlin",
			"kotlinx":     "KotlinX",
			"networking":  "Networking",
			"json":        "JSON",
			"logging":     "Logging",
			"utils":       "Utils",
		},
		Libraries: map[string]*LibraryInfo{
			"kotlin-stdlib": {
				Name:        "Kotlin Standard Library",
				Group:       "org.jetbrains.kotlin",
				Artifact:    "kotlin-stdlib",
				Version:     "1.9.24",
				Jar:         "kotlin-stdlib-1.9.24.jar",
				URL:         "https://repo1.maven.org/maven2/org/jetbrains/kotlin/kotlin-stdlib/1.9.24/kotlin-stdlib-1.9.24.jar",
				Category:    "kotlin",
				Description: "Kotlin standard library with core utilities and collections",
			},
			"kotlinx-coroutines-core": {
				Name:         "KotlinX Coroutines Core",
				Group:        "org.jetbrains.kotlinx",
				Artifact:     "kotlinx-coroutines-core",
				Version:      "1.8.1",
				Jar:          "kotlinx-coroutines-core-1.8.1.jar",
				URL:          "https://repo1.maven.org/maven2/org/jetbrains/kotlinx/kotlinx-coroutines-core/1.8.1/kotlinx-coroutines-core-1.8.1.jar",
				Category:     "kotlinx",
				Description:  "Kotlin coroutines core library for async programming",
				Dependencies: []string{"kotlin-stdlib"},
			},
			"kotlinx-coroutines-android": {
				Name:         "KotlinX Coroutines Android",
				Group:        "org.jetbrains.kotlinx",
				Artifact:     "kotlinx-coroutines-android",
				Version:      "1.8.1",
				Jar:          "kotlinx-coroutines-android-1.8.1.jar",
				URL:          "https://repo1.maven.org/maven2/org/jetbrains/kotlinx/kotlinx-coroutines-android/1.8.1/kotlinx-coroutines-android-1.8.1.jar",
				Category:     "kotlinx",
				Description:  "Kotlin coroutines Android dispatcher (Main thread)",
				Dependencies: []string{"kotlinx-coroutines-core"},
			},
			"kotlinx-serialization-core": {
				Name:         "KotlinX Serialization Core",
				Group:        "org.jetbrains.kotlinx",
				Artifact:     "kotlinx-serialization-core",
				Version:      "1.6.3",
				Jar:          "kotlinx-serialization-core-1.6.3.jar",
				URL:          "https://repo1.maven.org/maven2/org/jetbrains/kotlinx/kotlinx-serialization-core/1.6.3/kotlinx-serialization-core-1.6.3.jar",
				Category:     "kotlinx",
				Description:  "Kotlin multiplatform serialization framework",
				Dependencies: []string{"kotlin-stdlib"},
			},
			"okhttp3": {
				Name:        "OkHttp 3",
				Group:       "com.squareup.okhttp3",
				Artifact:    "okhttp3",
				Version:     "4.12.0",
				Jar:         "okhttp-4.12.0.jar",
				URL:         "https://repo1.maven.org/maven2/com/squareup/okhttp3/okhttp/4.12.0/okhttp-4.12.0.jar",
				Category:    "networking",
				Description: "HTTP/HTTPS client for Android with connection pooling",
			},
			"okhttp3-logging": {
				Name:         "OkHttp 3 Logging Interceptor",
				Group:        "com.squareup.okhttp3",
				Artifact:     "okhttp3-logging",
				Version:      "4.12.0",
				Jar:          "logging-interceptor-4.12.0.jar",
				URL:          "https://repo1.maven.org/maven2/com/squareup/okhttp3/logging-interceptor/4.12.0/logging-interceptor-4.12.0.jar",
				Category:     "networking",
				Description:  "OkHttp logging interceptor for request/response logging",
				Dependencies: []string{"okhttp3"},
			},
			"retrofit2": {
				Name:         "Retrofit 2",
				Group:        "com.squareup.retrofit2",
				Artifact:     "retrofit2",
				Version:      "2.11.0",
				Jar:          "retrofit-2.11.0.jar",
				URL:          "https://repo1.maven.org/maven2/com/squareup/retrofit2/retrofit/2.11.0/retrofit-2.11.0.jar",
				Category:     "networking",
				Description:  "Type-safe HTTP client for Android and Java",
				Dependencies: []string{"okhttp3"},
			},
			"retrofit2-gson": {
				Name:         "Retrofit 2 Gson Converter",
				Group:        "com.squareup.retrofit2",
				Artifact:     "retrofit2-gson",
				Version:      "2.11.0",
				Jar:          "converter-gson-2.11.0.jar",
				URL:          "https://repo1.maven.org/maven2/com/squareup/retrofit2/converter-gson/2.11.0/converter-gson-2.11.0.jar",
				Category:     "networking",
				Description:  "Retrofit Gson converter for JSON serialization",
				Dependencies: []string{"retrofit2", "gson"},
			},
			"gson": {
				Name:        "Gson",
				Group:       "com.google.code.gson",
				Artifact:    "gson",
				Version:     "2.11.0",
				Jar:         "gson-2.11.0.jar",
				URL:         "https://repo1.maven.org/maven2/com/google/code/gson/gson/2.11.0/gson-2.11.0.jar",
				Category:    "json",
				Description: "Google JSON library for Java object serialization",
			},
			"jackson-databind": {
				Name:        "Jackson Databind",
				Group:       "com.fasterxml.jackson.core",
				Artifact:    "jackson-databind",
				Version:     "2.17.2",
				Jar:         "jackson-databind-2.17.2.jar",
				URL:         "https://repo1.maven.org/maven2/com/fasterxml/jackson/core/jackson-databind/2.17.2/jackson-databind-2.17.2.jar",
				Category:    "json",
				Description: "Jackson JSON data binding library",
			},
			"slf4j-api": {
				Name:        "SLF4J API",
				Group:       "org.slf4j",
				Artifact:    "slf4j-api",
				Version:     "2.0.13",
				Jar:         "slf4j-api-2.0.13.jar",
				URL:         "https://repo1.maven.org/maven2/org/slf4j/slf4j-api/2.0.13/slf4j-api-2.0.13.jar",
				Category:    "logging",
				Description: "Simple Logging Facade for Java",
			},
			"logback-classic": {
				Name:         "Logback Classic",
				Group:        "ch.qos.logback",
				Artifact:     "logback-classic",
				Version:      "1.5.6",
				Jar:          "logback-classic-1.5.6.jar",
				URL:          "https://repo1.maven.org/maven2/ch/qos/logback/logback-classic/1.5.6/logback-classic-1.5.6.jar",
				Category:     "logging",
				Description:  "Logback classic logging implementation (SLF4J binding)",
				Dependencies: []string{"slf4j-api"},
			},
			"guava": {
				Name:        "Guava",
				Group:       "com.google.guava",
				Artifact:    "guava",
				Version:     "33.2.1-jre",
				Jar:         "guava-33.2.1-jre.jar",
				URL:         "https://repo1.maven.org/maven2/com/google/guava/guava/33.2.1-jre/guava-33.2.1-jre.jar",
				Category:    "utils",
				Description: "Google core libraries for Java (collections, caching, primitives)",
			},
			"apache-commons-lang3": {
				Name:        "Apache Commons Lang 3",
				Group:       "org.apache.commons",
				Artifact:    "apache-commons-lang3",
				Version:     "3.14.0",
				Jar:         "commons-lang3-3.14.0.jar",
				URL:         "https://repo1.maven.org/maven2/org/apache/commons/commons-lang3/3.14.0/commons-lang3-3.14.0.jar",
				Category:    "utils",
				Description: "Apache Commons Lang utility library (string, math, concurrency)",
			},
		},
	}
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

	logger.Info(fmt.Sprintf("Installing library %s %s", lib.Name, lib.Version))

	for _, dep := range lib.Dependencies {
		if !m.isInstalled(dep) {
			logger.Info(fmt.Sprintf("Installing dependency %s", dep))
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
		logger.Info(fmt.Sprintf("Library already installed at %s", jarPath))
		return nil
	}

	logger.Step(fmt.Sprintf("Downloading from %s", lib.URL))
	if err := m.downloadFile(ctx, lib.URL, jarPath); err != nil {
		return fmt.Errorf("failed to download library: %w", err)
	}

	logger.Success(fmt.Sprintf("Installed %s at %s", lib.Name, jarPath))
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

	logger.Success(fmt.Sprintf("Removed library %s", lib.Name))
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
