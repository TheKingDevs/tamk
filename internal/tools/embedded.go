//go:build windows

package tools

import (
	"archive/zip"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/TheKingDevs/tamk/internal/config"
)

//go:embed binaries/windows/apksigner.jar
var apksignerJar []byte

//go:embed binaries/windows/d8.jar
var d8Jar []byte

//go:embed binaries/windows/kotlinc.zip
var kotlinZip []byte

// Download URLs for tools not embedded in the binary.
const (
	aapt2DownloadURL    = "https://dl.google.com/android/maven2/com/android/tools/build/aapt2/8.4.1/aapt2-8.4.1-windows.zip"
	zipalignDownloadURL = "https://dl.google.com/android/repository/platform-tools-latest-windows.zip"
)

// embeddedToolManager extracts and caches embedded tools.
type embeddedToolManager struct {
	cfg        Config
	cache      *cachedTool
	toolsDir   string
	extractDir string
	mu         sync.Once
}

func newEmbeddedManager(cfg Config) *embeddedToolManager {
	return &embeddedToolManager{
		cfg:      cfg,
		cache:    newCachedTool(),
		toolsDir: cfg.ToolsDir,
	}
}

func (m *embeddedToolManager) Setup() error {
	if m.IsSetup() {
		return nil
	}

	if err := os.MkdirAll(m.toolsDir, 0o755); err != nil {
		return fmt.Errorf("create tools dir: %w", err)
	}

	// Download and extract aapt2.exe
	if err := m.downloadAAPT2(); err != nil {
		return fmt.Errorf("download aapt2: %w", err)
	}

	// Extract apksigner.jar
	apksignerPath := filepath.Join(m.toolsDir, "apksigner.jar")
	if err := os.WriteFile(apksignerPath, apksignerJar, 0o644); err != nil {
		return fmt.Errorf("extract apksigner: %w", err)
	}

	// Download and extract zipalign.exe
	if err := m.downloadZipalign(); err != nil {
		return fmt.Errorf("download zipalign: %w", err)
	}

	// Extract d8.jar
	d8Path := filepath.Join(m.toolsDir, "d8.jar")
	if err := os.WriteFile(d8Path, d8Jar, 0o644); err != nil {
		return fmt.Errorf("extract d8: %w", err)
	}

	// Extract kotlinc.zip
	if err := m.extractZip(kotlinZip, "kotlinc"); err != nil {
		return fmt.Errorf("extract kotlinc: %w", err)
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

func (m *embeddedToolManager) downloadAAPT2() error {
	aapt2Path := filepath.Join(m.toolsDir, "aapt2.exe")
	if _, err := os.Stat(aapt2Path); err == nil {
		return nil
	}

	data, err := downloadFile(aapt2DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download aapt2: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open aapt2 zip: %w", err)
	}

	for _, f := range zipReader.File {
		if strings.HasSuffix(f.Name, "aapt2.exe") {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.OpenFile(aapt2Path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, rc)
			return err
		}
	}

	return fmt.Errorf("aapt2.exe not found in downloaded zip")
}

func (m *embeddedToolManager) downloadZipalign() error {
	zipalignPath := filepath.Join(m.toolsDir, "zipalign.exe")
	if _, err := os.Stat(zipalignPath); err == nil {
		return nil
	}

	data, err := downloadFile(zipalignDownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download zipalign: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open platform-tools zip: %w", err)
	}

	for _, f := range zipReader.File {
		if strings.HasSuffix(f.Name, "zipalign.exe") {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.OpenFile(zipalignPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, rc)
			return err
		}
	}

	return fmt.Errorf("zipalign.exe not found in downloaded zip")
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d downloading %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}

func (m *embeddedToolManager) IsSetup() bool {
	ver := readVersion(m.toolsDir)
	return ver == config.Version
}

func (m *embeddedToolManager) ToolsDir() string {
	return m.toolsDir
}

func (m *embeddedToolManager) extractZip(data []byte, targetDir string) error {
	dir := filepath.Join(m.toolsDir, targetDir)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return nil
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}

	for _, f := range zipReader.File {
		fpath := filepath.Join(dir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0o755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *embeddedToolManager) AAPT2(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("aapt2"); ok {
		return p, nil
	}

	if !m.IsSetup() {
		if err := m.Setup(); err != nil {
			return "", err
		}
	}

	p := filepath.Join(m.toolsDir, "aapt2.exe")
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("aapt2 not found: %w", err)
	}

	m.cache.Set("aapt2", p)
	return p, nil
}

func (m *embeddedToolManager) ApkSigner(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("apksigner"); ok {
		return p, nil
	}

	if !m.IsSetup() {
		if err := m.Setup(); err != nil {
			return "", err
		}
	}

	jarPath := filepath.Join(m.toolsDir, "apksigner.jar")
	if _, err := os.Stat(jarPath); err != nil {
		return "", fmt.Errorf("apksigner not found: %w", err)
	}

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

	if !m.IsSetup() {
		if err := m.Setup(); err != nil {
			return "", err
		}
	}

	p := filepath.Join(m.toolsDir, "zipalign.exe")
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("zipalign not found: %w", err)
	}

	m.cache.Set("zipalign", p)
	return p, nil
}

func (m *embeddedToolManager) D8(_ context.Context) (string, error) {
	if p, ok := m.cache.Get("d8"); ok {
		return p, nil
	}

	if !m.IsSetup() {
		if err := m.Setup(); err != nil {
			return "", err
		}
	}

	jarPath := filepath.Join(m.toolsDir, "d8.jar")
	if _, err := os.Stat(jarPath); err != nil {
		return "", fmt.Errorf("d8 not found: %w", err)
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

	if !m.IsSetup() {
		if err := m.Setup(); err != nil {
			return "", err
		}
	}

	candidates := []string{
		filepath.Join(m.toolsDir, "kotlinc", "bin", "kotlinc.bat"),
		filepath.Join(m.toolsDir, "kotlinc", "bin", "kotlinc.exe"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			m.cache.Set("kotlinc", c)
			return c, nil
		}
	}

	return "", fmt.Errorf("kotlinc not found")
}

func (m *embeddedToolManager) SDKJar() (string, error) {
	// Check configured SDK path first
	if m.cfg.SDKPath != "" {
		if _, err := os.Stat(m.cfg.SDKPath); err == nil {
			return m.cfg.SDKPath, nil
		}
	}

	// Use extracted android.jar
	if m.IsSetup() {
		jarPath := filepath.Join(m.toolsDir, "android.jar")
		if _, err := os.Stat(jarPath); err == nil {
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

	return jarPath, nil
}

func (m *embeddedToolManager) Cleanup() error {
	return os.RemoveAll(m.toolsDir)
}
