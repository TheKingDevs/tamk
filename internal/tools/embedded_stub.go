//go:build !windows

package tools

import (
	"context"
	"fmt"
	"runtime"
)

// newEmbeddedManager returns a stub for non-Windows platforms.
// Embedded tools are only available on Windows.
func newEmbeddedManager(cfg Config) ToolManager {
	return &stubEmbeddedManager{cfg: cfg}
}

type stubEmbeddedManager struct {
	cfg Config
}

func (m *stubEmbeddedManager) AAPT2(_ context.Context) (string, error) {
	return "", fmt.Errorf("embedded tools not available on %s; use system tools", runtime.GOOS)
}

func (m *stubEmbeddedManager) ApkSigner(_ context.Context) (string, error) {
	return "", fmt.Errorf("embedded tools not available on %s; use system tools", runtime.GOOS)
}

func (m *stubEmbeddedManager) Zipalign(_ context.Context) (string, error) {
	return "", fmt.Errorf("embedded tools not available on %s; use system tools", runtime.GOOS)
}

func (m *stubEmbeddedManager) D8(_ context.Context) (string, error) {
	return "", fmt.Errorf("embedded tools not available on %s; use system tools", runtime.GOOS)
}

func (m *stubEmbeddedManager) KotlinCompiler(_ context.Context) (string, error) {
	return "", fmt.Errorf("embedded tools not available on %s; use system tools", runtime.GOOS)
}

func (m *stubEmbeddedManager) SDKJar() (string, error) {
	return "", fmt.Errorf("embedded tools not available on %s; use system tools", runtime.GOOS)
}

func (m *stubEmbeddedManager) Cleanup() error {
	return nil
}
