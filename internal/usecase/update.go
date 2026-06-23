package usecase

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TheKingDevs/tamk/internal/config"
	"github.com/TheKingDevs/tamk/internal/domain/entity"
	"github.com/TheKingDevs/tamk/internal/domain/repository"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

type UpdateUseCase struct {
	cfg        *config.Config
	updateRepo repository.UpdateRepository
}

func NewUpdateUseCase(
	cfg *config.Config,
	updateRepo repository.UpdateRepository,
) *UpdateUseCase {
	return &UpdateUseCase{
		cfg:        cfg,
		updateRepo: updateRepo,
	}
}

func (uc *UpdateUseCase) Check(ctx context.Context) (*entity.UpdateInfo, error) {
	logger.Info("Checking for updates...")
	info, err := uc.updateRepo.CheckForUpdates(ctx, uc.cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("update check failed: %w", err)
	}
	return info, nil
}

func (uc *UpdateUseCase) ShouldAutoInstall(info *entity.UpdateInfo) bool {
	return info.Level == entity.UpdateLevelCritical || info.Level == entity.UpdateLevelPatch
}

func (uc *UpdateUseCase) ShouldPrompt(info *entity.UpdateInfo) bool {
	return info.Level == entity.UpdateLevelMajor || info.Level == entity.UpdateLevelMinor
}

func (uc *UpdateUseCase) PrintUpdateInfo(info *entity.UpdateInfo) {
	levelName := map[entity.UpdateLevel]string{
		entity.UpdateLevelCritical: "CRITICAL",
		entity.UpdateLevelMajor:    "MAJOR",
		entity.UpdateLevelMinor:    "MINOR",
		entity.UpdateLevelPatch:    "PATCH",
		entity.UpdateLevelOptional: "OPTIONAL",
	}

	logger.Info("Update available",
		"current", info.CurrentVersion,
		"latest", info.LatestVersion,
		"level", levelName[info.Level],
	)

	if info.ReleaseNotes != "" {
		logger.Info("Release notes", "notes", formatReleaseNotes(info.ReleaseNotes))
	}
}

func formatReleaseNotes(notes string) string {
	if len(notes) > 500 {
		return notes[:500] + "..."
	}
	return notes
}

func (uc *UpdateUseCase) Install(ctx context.Context, info *entity.UpdateInfo) error {
	if info == nil {
		return fmt.Errorf("no update info provided")
	}

	logger.Step("Installing update", "from", info.CurrentVersion, "to", info.LatestVersion)

	method := uc.detectUpdateMethod()

	switch method {
	case "git":
		return uc.installViaGit(ctx)
	case "go":
		return uc.installViaGoInstall(ctx)
	default:
		logger.Warn("No update method available. Install manually.",
			"download", info.DownloadURL)
		return nil
	}
}

func (uc *UpdateUseCase) detectUpdateMethod() string {
	tamkHome := uc.cfg.TAMKHome
	gitDir := filepath.Join(tamkHome, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return "git"
	}

	if _, err := exec.LookPath("go"); err == nil {
		return "go"
	}

	return ""
}

func (uc *UpdateUseCase) installViaGit(ctx context.Context) error {
	tamkHome := uc.cfg.TAMKHome
	cmd := exec.CommandContext(ctx, "git", "pull", "--rebase", "--autostash")
	cmd.Dir = tamkHome
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %s: %w", string(out), err)
	}
	logger.Success("Updated via git", "output", strings.TrimSpace(string(out)))
	return nil
}

func (uc *UpdateUseCase) installViaGoInstall(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "go", "install", "github.com/TheKingDevs/tamk/cmd/tamk@latest")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go install failed: %s: %w", string(out), err)
	}
	logger.Success("Updated via go install")
	return nil
}
