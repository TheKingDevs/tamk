// Package tools provides a platform-agnostic tool manager for build dependencies.
// It implements a "Lazy Extraction" pattern:
// - Unix-like systems: Use system-installed tools (lightweight)
// - Windows: Auto-extract embedded tools on first use (zero config)
package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

// ToolManager resolves paths to build tools (aapt2, kotlinc, d8, etc.).
type ToolManager interface {
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

	// Cleanup removes extracted temporary files.
	Cleanup() error
}

// Config holds tool manager configuration.
type Config struct {
	// DevDir is the development directory (e.g., $TAMK_HOME/development).
	DevDir string

	// SDKPath is the path to android.jar.
	SDKPath string

	// UseEmbedded forces use of embedded tools even if system tools exist.
	UseEmbedded bool

	// CacheDir is the directory for caching extracted tools.
	// Defaults to os.TempDir()/tamk-tools.
	CacheDir string
}

// New creates a new ToolManager for the current platform.
func New(cfg Config) ToolManager {
	if cfg.CacheDir == "" {
		cfg.CacheDir = filepath.Join(os.TempDir(), "tamk-tools")
	}

	if runtime.GOOS == "windows" || cfg.UseEmbedded {
		return newEmbeddedManager(cfg)
	}
	return newSystemManager(cfg)
}

// NewForced creates a ToolManager that always uses embedded tools.
// Useful for testing or when system tools are incompatible.
func NewForced(cfg Config) ToolManager {
	cfg.UseEmbedded = true
	return newEmbeddedManager(cfg)
}

// findTool searches for a tool in PATH and common locations.
func findTool(name string) (string, error) {
	// Check PATH first
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}

	// Check common locations
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".local", "bin", name),
		filepath.Join("/usr", "local", "bin", name),
		filepath.Join("/opt", "tamk", "bin", name),
	}

	// Termux paths
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
