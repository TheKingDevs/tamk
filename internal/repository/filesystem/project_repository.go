package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Shadw-Developer/tamk/internal/domain/entity"
)

type ProjectRepository struct{}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{}
}

type projectConfig struct {
	Type        entity.ProjectType `json:"type"`
	Name        string             `json:"name"`
	Version     string             `json:"version"`
	Author      string             `json:"author"`
	PackageName string             `json:"package"`
	WebURL      string             `json:"web_url,omitempty"`
	WebMode     string             `json:"web_mode,omitempty"`
	MinSDK      int                `json:"min_sdk,omitempty"`
	TargetSDK   int                `json:"target_sdk,omitempty"`
}

func (r *ProjectRepository) Save(ctx context.Context, project *entity.Project, path string) error {
	cfg := projectConfig{
		Type:        project.Type,
		Name:        project.Name,
		Version:     project.Version,
		Author:      project.Author,
		PackageName: project.PackageName,
		WebURL:      project.WebURL,
		WebMode:     string(project.WebMode),
		MinSDK:      project.MinSDK,
		TargetSDK:   project.TargetSDK,
	}

	tamkConfigPath := filepath.Join(path, "tamk.config")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(tamkConfigPath, data, 0o644)
}

func (r *ProjectRepository) Load(ctx context.Context, path string) (*entity.Project, error) {
	tamkConfigPath := filepath.Join(path, "tamk.config")
	data, err := os.ReadFile(tamkConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", tamkConfigPath, err)
	}

	var cfg projectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &entity.Project{
		Name:        cfg.Name,
		Type:        cfg.Type,
		Version:     cfg.Version,
		Author:      cfg.Author,
		PackageName: cfg.PackageName,
		WebURL:      cfg.WebURL,
		WebMode:     entity.WebContentMode(cfg.WebMode),
		MinSDK:      cfg.MinSDK,
		TargetSDK:   cfg.TargetSDK,
	}, nil
}

func (r *ProjectRepository) Exists(ctx context.Context, path string) bool {
	tamkConfigPath := filepath.Join(path, "tamk.config")
	_, err := os.Stat(tamkConfigPath)
	return err == nil
}

func (r *ProjectRepository) Delete(ctx context.Context, path string) error {
	return os.RemoveAll(path)
}
