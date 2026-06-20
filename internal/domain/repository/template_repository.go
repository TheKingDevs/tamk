package repository

import (
	"context"

	"github.com/Shadw-Developer/tamk/internal/domain/entity"
)

type TemplateRepository interface {
	LoadTemplate(ctx context.Context, name string) (string, error)
	LoadTemplateForType(ctx context.Context, projectType, name string) (string, error)
	LoadTemplateWithFallback(ctx context.Context, name, fallbackType string) (string, error)
	ApplyPlaceholders(content string, placeholders map[string]string) string
	WriteTemplate(ctx context.Context, projectPath, destPath, content string) error
	GetTemplateDir(ctx context.Context, projectType string) (string, error)
	GetMappings(ctx context.Context, projectType string, isInternal bool) ([]entity.TemplateMapping, error)
}
