//go:build windows

package tools

import (
	"archive/zip"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

//go:embed binaries/windows/aapt2.exe
var aapt2Binary []byte

//go:embed binaries/windows/apksigner.jar
var apksignerJar []byte

//go:embed binaries/windows/zipalign.exe
var zipalignBinary []byte

//go:embed binaries/windows/d8.jar
var d8Jar []byte

//go:embed binaries/windows/kotlinc.zip
var kotlinZip []byte

// embeddedToolManager extracts and caches embedded tools.
type embeddedToolManager struct {
	cfg        Config
	cache      *cachedTool
	extractDir string
	mu         sync.Once
}

func newEmbeddedManager(cfg Config) *embeddedToolManager {
	return &embeddedToolManager{
		cfg:   cfg,
		cache: newCachedTool(),
	}
}

func (m *embeddedToolManager) ensureExtracted() error {
	var firstErr error
	m.mu.Do(func() {
		dir := filepath.Join(m.cfg.CacheDir, "extracted")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			firstErr = fmt.Errorf("create cache dir: %w", err)
			return
		}
		m.extractDir = dir
	})
	return firstErr
}

func (m *embeddedToolManager) extractFile(data []byte, name string) (string, error) {
	if err := m.ensureExtracted(); err != nil {
		return "", err
	}

	path := filepath.Join(m.extractDir, name)

	// Check if already extracted
	if info, err := os.Stat(path); err == nil && info.Size() == int64(len(data)) {
		return path, nil
	}

	if err := os.WriteFile(path, data, 0o755); err != nil {
		return "", fmt.Errorf("extract %s: %w", name, err)
	}

	return path, nil
}

func (m *embeddedToolManager) extractZip(data []byte, targetDir string) (string, error) {
	if err := m.ensureExtracted(); err != nil {
		return "", err
	}

	dir := filepath.Join(m.extractDir, targetDir)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir, nil
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}

	for _, f := range zipReader.File {
		fpath := filepath.Join(dir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0o755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
			return "", err
		}
		outFile, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return "", err
		}
		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return "", err
		}
	}

	return dir, nil
}

func (m *embeddedToolManager) AAPT2(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("aapt2"); ok {
		return p, nil
	}

	p, err := m.extractFile(aapt2Binary, "aapt2.exe")
	if err != nil {
		return "", fmt.Errorf("extract aapt2: %w", err)
	}

	m.cache.Set("aapt2", p)
	return p, nil
}

func (m *embeddedToolManager) ApkSigner(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("apksigner"); ok {
		return p, nil
	}

	jarPath, err := m.extractFile(apksignerJar, "apksigner.jar")
	if err != nil {
		return "", fmt.Errorf("extract apksigner: %w", err)
	}

	// On Windows, we need to find java
	javaPath, err := findTool("java.exe")
	if err != nil {
		return "", fmt.Errorf("apksigner requires Java: %w", err)
	}

	p := fmt.Sprintf("%s -jar %s", javaPath, jarPath)
	m.cache.Set("apksigner", p)
	return p, nil
}

func (m *embeddedToolManager) Zipalign(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("zipalign"); ok {
		return p, nil
	}

	p, err := m.extractFile(zipalignBinary, "zipalign.exe")
	if err != nil {
		return "", fmt.Errorf("extract zipalign: %w", err)
	}

	m.cache.Set("zipalign", p)
	return p, nil
}

func (m *embeddedToolManager) D8(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("d8"); ok {
		return p, nil
	}

	jarPath, err := m.extractFile(d8Jar, "d8.jar")
	if err != nil {
		return "", fmt.Errorf("extract d8: %w", err)
	}

	javaPath, err := findTool("java.exe")
	if err != nil {
		return "", fmt.Errorf("d8 requires Java: %w", err)
	}

	p := fmt.Sprintf("%s -jar %s", javaPath, jarPath)
	m.cache.Set("d8", p)
	return p, nil
}

func (m *embeddedToolManager) KotlinCompiler(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("kotlinc"); ok {
		return p, nil
	}

	kotlinDir, err := m.extractZip(kotlinZip, "kotlinc")
	if err != nil {
		return "", fmt.Errorf("extract kotlinc: %w", err)
	}

	// Find kotlinc.bat or kotlinc.exe
	candidates := []string{
		filepath.Join(kotlinDir, "bin", "kotlinc.bat"),
		filepath.Join(kotlinDir, "bin", "kotlinc.exe"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			m.cache.Set("kotlinc", c)
			return c, nil
		}
	}

	return "", fmt.Errorf("kotlinc not found in extracted archive")
}

func (m *embeddedToolManager) SDKJar() (string, error) {
	if m.cfg.SDKPath == "" {
		return "", fmt.Errorf("SDK path not configured")
	}
	return m.cfg.SDKPath, nil
}

func (m *embeddedToolManager) Cleanup() error {
	if m.extractDir != "" {
		return os.RemoveAll(m.extractDir)
	}
	return nil
}
