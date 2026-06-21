package usecase

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/TheKingDevs/tamk/pkg/logger"
)

// ProGuardObfuscator handles code obfuscation via ProGuard.
type ProGuardObfuscator struct {
	proguardPath string
}

// NewProGuardObfuscator creates a new obfuscator instance.
func NewProGuardObfuscator() *ProGuardObfuscator {
	path := findProGuard()
	return &ProGuardObfuscator{proguardPath: path}
}

func findProGuard() string {
	// Check common locations
	candidates := []string{
		"proguard",
		"/usr/bin/proguard",
		"/usr/local/bin/proguard",
		filepath.Join(os.Getenv("HOME"), ".tamk/tools/proguard"),
	}

	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			return c
		}
	}

	return ""
}

// IsAvailable checks if ProGuard is installed.
func (p *ProGuardObfuscator) IsAvailable() bool {
	return p.proguardPath != ""
}

// Obfuscate applies ProGuard obfuscation to compiled classes.
func (p *ProGuardObfuscator) Obfuscate(ctx context.Context, objDir, sdkPath, configPath string) error {
	if !p.IsAvailable() {
		logger.Warn("ProGuard not available, skipping obfuscation")
		return nil
	}

	logger.Step("Obfuscating code with ProGuard...")

	obfuscatedDir := filepath.Join(filepath.Dir(objDir), "obfuscated")
	os.MkdirAll(obfuscatedDir, 0o755)

	// Build ProGuard command
	args := []string{
		"@proguard-android.pro",
		"-injars", objDir,
		"-outjars", obfuscatedDir,
		"-libraryjars", sdkPath,
	}

	if configPath != "" {
		args = []string{"@" + configPath}
	}

	cmd := exec.CommandContext(ctx, p.proguardPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("ProGuard output", "output", string(output))
		return fmt.Errorf("proguard failed: %w", err)
	}

	// Copy obfuscated classes back
	if err := copyDir(obfuscatedDir, objDir); err != nil {
		return fmt.Errorf("failed to copy obfuscated classes: %w", err)
	}

	logger.Success("Code obfuscated")
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}
