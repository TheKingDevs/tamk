package usecase

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/TheKingDevs/tamk/internal/config"
	"github.com/TheKingDevs/tamk/internal/tools"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

type SetupEnvironmentUseCase struct {
	cfg *config.Config
}

func NewSetupEnvironmentUseCase(cfg *config.Config) *SetupEnvironmentUseCase {
	return &SetupEnvironmentUseCase{cfg: cfg}
}

const sdkURL = "https://dl.google.com/android/repository/platform-30_r03.zip"
const sdkZipName = "platform-30_r03.zip"

func (uc *SetupEnvironmentUseCase) Execute(ctx context.Context) error {
	logger.Info("Setting up TAMK environment...")
	logger.Info("Detected environment", "env", uc.cfg.Env)

	if err := uc.ensureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Extract embedded tools
	logger.Step("Extracting embedded tools...")
	toolMgr := tools.New(tools.Config{
		DevDir:  uc.cfg.DevDir,
		SDKPath: uc.cfg.SDKPath,
	})
	logger.Info("Tools dir", "path", toolMgr.ToolsDir())
	if err := toolMgr.Setup(); err != nil {
		logger.Warn("Failed to extract embedded tools", "error", err)
	} else {
		logger.Success("Embedded tools extracted", "dir", toolMgr.ToolsDir())
	}

	if err := uc.downloadSDK(ctx); err != nil {
		return fmt.Errorf("failed to download SDK: %w", err)
	}

	if err := uc.generateDebugKeystore(ctx); err != nil {
		return fmt.Errorf("failed to generate debug keystore: %w", err)
	}

	logger.Success("Environment setup complete")
	logger.Info("SDK path", "path", uc.cfg.SDKPath)
	logger.Info("Keystore path", "path", uc.cfg.Keystore)
	logger.Info("Tools dir", "path", toolMgr.ToolsDir())

	return nil
}

func (uc *SetupEnvironmentUseCase) VerifyEnvironment() bool {
	allGood := true

	if _, err := os.Stat(uc.cfg.SDKPath); os.IsNotExist(err) {
		logger.Warn("SDK not found", "path", uc.cfg.SDKPath)
		allGood = false
	} else {
		logger.Info("SDK found")
	}

	if _, err := os.Stat(uc.cfg.Keystore); os.IsNotExist(err) {
		logger.Warn("Keystore not found", "path", uc.cfg.Keystore)
		allGood = false
	} else {
		logger.Info("Keystore found")
	}

	tools := []string{"aapt2", "kotlinc", "d8", "apksigner", "zipalign", "keytool"}
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			logger.Warn("Tool not found", "tool", tool)
			allGood = false
		} else {
			logger.Info("Tool found", "tool", tool)
		}
	}

	return allGood
}

func (uc *SetupEnvironmentUseCase) ensureDirectories() error {
	dirs := []string{
		filepath.Dir(uc.cfg.SDKPath),
		filepath.Dir(uc.cfg.Keystore),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (uc *SetupEnvironmentUseCase) downloadSDK(ctx context.Context) error {
	if _, err := os.Stat(uc.cfg.SDKPath); err == nil {
		logger.Info("SDK already exists, skipping download")
		return nil
	}

	logger.Info("Downloading Android SDK...")

	zipPath := filepath.Join(filepath.Dir(uc.cfg.SDKPath), sdkZipName)

	req, err := http.NewRequestWithContext(ctx, "GET", sdkURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download SDK: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write SDK zip: %w", err)
	}
	out.Close()

	logger.Info("Extracting SDK...")
	extractDir := filepath.Dir(uc.cfg.SDKPath)
	unzipCmd := exec.CommandContext(ctx, "unzip", "-o", zipPath, "-d", extractDir)
	if outBytes, err := unzipCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to extract SDK: %s: %w", string(outBytes), err)
	}

	// Find and move android.jar from extracted subdirectory
	entries, _ := os.ReadDir(extractDir)
	for _, entry := range entries {
		if entry.IsDir() {
			jarPath := filepath.Join(extractDir, entry.Name(), "android.jar")
			if _, err := os.Stat(jarPath); err == nil {
				os.Rename(jarPath, uc.cfg.SDKPath)
				os.RemoveAll(filepath.Join(extractDir, entry.Name()))
				break
			}
		}
	}

	os.Remove(zipPath)
	logger.Success("SDK downloaded and extracted")

	return nil
}

func (uc *SetupEnvironmentUseCase) generateDebugKeystore(ctx context.Context) error {
	if _, err := os.Stat(uc.cfg.Keystore); err == nil {
		logger.Info("Debug keystore already exists")
		return nil
	}

	logger.Info("Generating debug keystore...")

	ksDir := filepath.Dir(uc.cfg.Keystore)
	os.MkdirAll(ksDir, 0o755)

	args := []string{
		"-genkey", "-v",
		"-keystore", uc.cfg.Keystore,
		"-alias", "androiddebugkey",
		"-keyalg", "RSA",
		"-keysize", "2048",
		"-validity", "10000",
		"-storepass", "android",
		"-keypass", "android",
		"-dname", "CN=Android Debug, OU=Android, O=Android, L=Unknown, ST=Unknown, C=US",
	}

	cmd := exec.CommandContext(ctx, "keytool", args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate keystore: %w", err)
	}

	if err := os.Chmod(uc.cfg.Keystore, 0o644); err != nil {
		return fmt.Errorf("failed to set keystore permissions: %w", err)
	}

	logger.Success("Debug keystore generated")
	return nil
}
