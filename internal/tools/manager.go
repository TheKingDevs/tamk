// Package tools provides a platform-agnostic tool manager for build dependencies.
// It implements a "Setup Once, Run Forever" pattern:
// - First run: Extract embedded tools to persistent directory
// - Subsequent runs: Use extracted tools directly (instant)
// - Version tracking: Re-extract only when TAMK version changes
package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/TheKingDevs/tamk/internal/config"
)

// ToolManager resolves paths to build tools (aapt2, kotlinc, d8, etc.).
type ToolManager interface {
	// Setup extracts tools to persistent directory if needed.
	Setup() error

	// IsSetup returns true if tools are already extracted and valid.
	IsSetup() bool

	// AAPT2 returns path to the aapt2 binary.
	AAPT2(ctx context.Context) (string, error)

	// ApkSigner returns path to the apksigner script or binary.
	ApkSigner(ctx context.Context) (string, error)

	// Zipalign returns path to the zipalign binary.
	Zipalign(ctx context.Context) (string, error)

	// D8 returns path to the d8.jar or d8 binary.
	D8(ctx context.Context) (string, error)

	// KotlinCompiler returns path to the kotlinc script or binary.
	KotlinCompiler(ctx context.Context) (string, error)

	// SDKJar returns path to android.jar.
	SDKJar() (string, error)

	// ToolsDir returns the path to the extracted tools directory.
	ToolsDir() string

	// Cleanup removes extracted files (for reinstall).
	Cleanup() error
}

// Config holds tool manager configuration.
type Config struct {
	// DevDir is the development directory (e.g., $TAMK_HOME/development).
	DevDir string

	// SDKPath is the path to android.jar.
	SDKPath string

	// ToolsDir is the persistent directory for extracted tools.
	// Defaults to $TAMK_HOME/tools/{os}
	ToolsDir string

	// UseEmbedded forces use of embedded tools even if system tools exist.
	UseEmbedded bool
}

// New creates a new ToolManager for the current platform.
func New(cfg Config) ToolManager {
	if cfg.ToolsDir == "" {
		home := config.HomeDir()
		tamkHome := os.Getenv("TAMK_HOME")
		if tamkHome == "" {
			tamkHome = filepath.Join(home, ".tamk")
		}
		cfg.ToolsDir = filepath.Join(tamkHome, "tools", runtime.GOOS)
	}

	if runtime.GOOS == "windows" || cfg.UseEmbedded {
		return newEmbeddedManager(cfg)
	}
	return newSystemManager(cfg)
}

// NewForced creates a ToolManager that always uses embedded tools.
func NewForced(cfg Config) ToolManager {
	cfg.UseEmbedded = true
	return newEmbeddedManager(cfg)
}

// findTool searches for a tool in PATH and common locations.
func findTool(name string) (string, error) {
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}

	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".local", "bin", name),
		filepath.Join("/usr", "local", "bin", name),
		filepath.Join("/opt", "tamk", "bin", name),
	}

	if runtime.GOOS == "android" {
		prefix := os.Getenv("PREFIX")
		if prefix == "" {
			prefix = "/data/data/com.termux/files/usr"
		}
		candidates = append(candidates,
			filepath.Join(prefix, "bin", name),
			filepath.Join(prefix, "opt", "tamk", "bin", name),
		)
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	return "", fmt.Errorf("%s not found in PATH or common locations", name)
}

// versionMarker returns the path to the version marker file.
func versionMarker(toolsDir string) string {
	return filepath.Join(toolsDir, ".version")
}

// readVersion reads the version from the marker file.
func readVersion(toolsDir string) string {
	data, err := os.ReadFile(versionMarker(toolsDir))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// writeVersion writes the version to the marker file.
func writeVersion(toolsDir, version string) error {
	return os.WriteFile(versionMarker(toolsDir), []byte(version), 0o644)
}

// setExecutable sets the executable permission on a file.
func setExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode()&0o111 != 0 {
		return nil
	}
	return os.Chmod(path, info.Mode()|0o111)
}

// cachedTool provides thread-safe caching for resolved tool paths.
type cachedTool struct {
	mu    sync.RWMutex
	paths map[string]string
}

func newCachedTool() *cachedTool {
	return &cachedTool{paths: make(map[string]string)}
}

func (c *cachedTool) Get(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	p, ok := c.paths[name]
	return p, ok
}

func (c *cachedTool) Set(name, path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paths[name] = path
}
