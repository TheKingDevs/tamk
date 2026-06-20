package usecase

import (
	"context"
	"fmt"

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
