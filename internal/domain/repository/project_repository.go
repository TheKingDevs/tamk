package repository

import (
	"context"

	"github.com/Shadw-Developer/tamk/internal/domain/entity"
)

type ProjectRepository interface {
	Save(ctx context.Context, project *entity.Project, path string) error
	Load(ctx context.Context, path string) (*entity.Project, error)
	Exists(ctx context.Context, path string) bool
	Delete(ctx context.Context, path string) error
}
