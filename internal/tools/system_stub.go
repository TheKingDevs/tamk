//go:build windows

package tools

import (
	"context"
	"fmt"
)

// newSystemManager returns a stub for Windows.
// Windows uses embedded tools instead.
func newSystemManager(cfg Config) ToolManager {
	return &stubSystemManager{cfg: cfg}
}

type stubSystemManager struct {
	cfg Config
}

func (m *stubSystemManager) Setup() error {
	return fmt.Errorf("system tools not available on Windows; use embedded tools")
}

func (m *stubSystemManager) IsSetup() bool {
	return false
}

func (m *stubSystemManager) ToolsDir() string {
	return ""
}

func (m *stubSystemManager) AAPT2(_ context.Context) (string, error) {
	return "", fmt.Errorf("system tools not available on Windows; use embedded tools")
}

func (m *stubSystemManager) ApkSigner(_ context.Context) (string, error) {
	return "", fmt.Errorf("system tools not available on Windows; use embedded tools")
}

func (m *stubSystemManager) Zipalign(_ context.Context) (string, error) {
	return "", fmt.Errorf("system tools not available on Windows; use embedded tools")
}

func (m *stubSystemManager) D8(_ context.Context) (string, error) {
	return "", fmt.Errorf("system tools not available on Windows; use embedded tools")
}

func (m *stubSystemManager) KotlinCompiler(_ context.Context) (string, error) {
	return "", fmt.Errorf("system tools not available on Windows; use embedded tools")
}

func (m *stubSystemManager) SDKJar() (string, error) {
	return "", fmt.Errorf("system tools not available on Windows; use embedded tools")
}

func (m *stubSystemManager) Cleanup() error {
	return nil
}
