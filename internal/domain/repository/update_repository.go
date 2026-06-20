package repository

import (
	"context"

	"github.com/Shadw-Developer/tamk/internal/domain/entity"
)

//go:generate mockgen -source=update_repository.go -destination=mock/update_repository.go -package=mock

type UpdateRepository interface {
	CheckForUpdates(ctx context.Context, currentVersion string) (*entity.UpdateInfo, error)
	DetectUpdateLevel(latestVersion, currentVersion, releaseNotes string) entity.UpdateLevel
	DownloadUpdate(ctx context.Context, url, destPath string) error
	GetCachePath() string
	ClearCache() error
}
