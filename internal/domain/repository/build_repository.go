package repository

import (
	"context"

	"github.com/Shadw-Developer/tamk/internal/domain/entity"
)

type BuildRepository interface {
	CalculateProjectHash(ctx context.Context, projectPath string) (string, error)
	LoadCache(ctx context.Context, projectPath string) (*entity.BuildCache, error)
	SaveCache(ctx context.Context, projectPath string, cache *entity.BuildCache) error
	MustRecompile(ctx context.Context, projectPath string, currentHash string) (bool, error)
	CleanCache(ctx context.Context, projectPath string) error
}
