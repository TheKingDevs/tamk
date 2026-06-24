//go:build !windows

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheKingDevs/tamk/internal/config"
)

// systemToolManager uses system-installed tools with embedded fallback.
type systemToolManager struct {
	cfg      Config
	cache    *cachedTool
	toolsDir string
}

func newSystemManager(cfg Config) *systemToolManager {
	return &systemToolManager{
		cfg:      cfg,
		cache:    newCachedTool(),
		toolsDir: cfg.ToolsDir,
	}
}

func (m *systemToolManager) Setup() error {
	if m.IsSetup() {
		return nil
	}

	if err := os.MkdirAll(m.toolsDir, 0o755); err != nil {
		return fmt.Errorf("create tools dir: %w", err)
	}

	// Extract android.jar
	jarPath := filepath.Join(m.toolsDir, "android.jar")
	if err := os.WriteFile(jarPath, androidJar, 0o644); err != nil {
		return fmt.Errorf("extract android.jar: %w", err)
	}

	// Write version marker
	if err := writeVersion(m.toolsDir, config.Version); err != nil {
		return fmt.Errorf("write version: %w", err)
	}

	return nil
}

func (m *systemToolManager) IsSetup() bool {
	ver := readVersion(m.toolsDir)
	return ver == config.Version
}

func (m *systemToolManager) ToolsDir() string {
	return m.toolsDir
}

func (m *systemToolManager) AAPT2(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("aapt2"); ok {
		return p, nil
	}

	p, err := findTool("aapt2")
	if err != nil {
		return "", fmt.Errorf("aapt2 not found: install via 'pkg install aapt2' (Termux) or 'apt install aapt' (Debian): %w", err)
	}

	m.cache.Set("aapt2", p)
	return p, nil
}

func (m *systemToolManager) ApkSigner(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("apksigner"); ok {
		return p, nil
	}

	p, err := findTool("apksigner")
	if err != nil {
		jar, jarErr := findTool("apksigner.jar")
		if jarErr != nil {
			return "", fmt.Errorf("apksigner not found: install Android SDK build-tools: %w", err)
		}
		java, javaErr := findTool("java")
		if javaErr != nil {
			return "", fmt.Errorf("apksigner.jar found but java not available: %w", javaErr)
		}
		p = fmt.Sprintf("%s -jar %s", java, jar)
	}

	m.cache.Set("apksigner", p)
	return p, nil
}

func (m *systemToolManager) Zipalign(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("zipalign"); ok {
		return p, nil
	}

	p, err := findTool("zipalign")
	if err != nil {
		return "", fmt.Errorf("zipalign not found: install Android SDK build-tools: %w", err)
	}

	m.cache.Set("zipalign", p)
	return p, nil
}

func (m *systemToolManager) D8(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("d8"); ok {
		return p, nil
	}

	p, err := findTool("d8")
	if err != nil {
		jar, jarErr := findTool("d8.jar")
		if jarErr != nil {
			return "", fmt.Errorf("d8 not found: install Android SDK build-tools: %w", err)
		}
		java, javaErr := findTool("java")
		if javaErr != nil {
			return "", fmt.Errorf("d8.jar found but java not available: %w", javaErr)
		}
		p = fmt.Sprintf("%s -jar %s", java, jar)
	}

	m.cache.Set("d8", p)
	return p, nil
}

func (m *systemToolManager) KotlinCompiler(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("kotlinc"); ok {
		return p, nil
	}

	p, err := findTool("kotlinc")
	if err != nil {
		return "", fmt.Errorf("kotlinc not found: install Kotlin compiler: %w", err)
	}

	m.cache.Set("kotlinc", p)
	return p, nil
}

func (m *systemToolManager) SDKJar() (string, error) {
	if p, ok := m.cache.Get("sdk_jar"); ok {
		return p, nil
	}

	// Check configured SDK path first
	if m.cfg.SDKPath != "" {
		if _, err := os.Stat(m.cfg.SDKPath); err == nil {
			m.cache.Set("sdk_jar", m.cfg.SDKPath)
			return m.cfg.SDKPath, nil
		}
	}

	// Use extracted android.jar
	if m.IsSetup() {
		jarPath := filepath.Join(m.toolsDir, "android.jar")
		if _, err := os.Stat(jarPath); err == nil {
			m.cache.Set("sdk_jar", jarPath)
			return jarPath, nil
		}
	}

	// Extract if not setup
	if err := m.Setup(); err != nil {
		return "", err
	}

	jarPath := filepath.Join(m.toolsDir, "android.jar")
	if _, err := os.Stat(jarPath); err != nil {
		return "", fmt.Errorf("android.jar not found after setup")
	}

	m.cache.Set("sdk_jar", jarPath)
	return jarPath, nil
}

func (m *systemToolManager) Cleanup() error {
	return os.RemoveAll(m.toolsDir)
}
