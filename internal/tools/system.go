//go:build !windows

package tools

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

// systemToolManager uses system-installed tools.
// This is the default for Linux, macOS, and Termux.
type systemToolManager struct {
	cfg     Config
	cache   *cachedTool
	sdkPath string
}

func newSystemManager(cfg Config) *systemToolManager {
	return &systemToolManager{
		cfg:     cfg,
		cache:   newCachedTool(),
		sdkPath: cfg.SDKPath,
	}
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

	// apksigner may be a script or binary
	p, err := findTool("apksigner")
	if err != nil {
		// Try apksigner.jar
		jar, jarErr := findTool("apksigner.jar")
		if jarErr != nil {
			return "", fmt.Errorf("apksigner not found: install Android SDK build-tools: %w", err)
		}
		// Wrap jar with java
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

	// Try d8 binary first, then d8.jar
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
	if m.sdkPath == "" {
		return "", fmt.Errorf("SDK path not configured")
	}

	if _, err := exec.LookPath("stat"); err == nil {
		if out, err := exec.Command("stat", "-f", "%z", m.sdkPath).Output(); err == nil && len(out) > 0 {
			return m.sdkPath, nil
		}
	}

	if info, err := filepath.Glob(m.sdkPath); err == nil && len(info) > 0 {
		return m.sdkPath, nil
	}

	return m.sdkPath, nil
}

func (m *systemToolManager) Cleanup() error {
	return nil
}
