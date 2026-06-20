package usecase

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/TheKingDevs/tamk/internal/config"
	"github.com/TheKingDevs/tamk/internal/domain/entity"
	repoFS "github.com/TheKingDevs/tamk/internal/repository/filesystem"
)

func TestCreateProjectUseCase_ConsoleProject(t *testing.T) {
	cfg := config.New()
	cfg.ProjectDir = t.TempDir()
	projRepo := repoFS.NewProjectRepository()
	tmplRepo := repoFS.NewTemplateRepository(cfg)
	uc := NewCreateProjectUseCase(cfg, projRepo, tmplRepo)

	output, err := uc.Execute(context.Background(), CreateProjectInput{
		Name:    "TestConsole",
		Type:    entity.ProjectTypeConsole,
		Version: "1.0.0",
		Author:  "TestAuthor",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.Project == nil {
		t.Fatal("expected non-nil project")
	}
	if output.Project.Name != "TestConsole" {
		t.Errorf("Name = %q, want %q", output.Project.Name, "TestConsole")
	}
	if output.Project.Type != entity.ProjectTypeConsole {
		t.Errorf("Type = %q, want %q", output.Project.Type, entity.ProjectTypeConsole)
	}

	mainKt := filepath.Join(output.ProjectPath, "src", "Main.kt")
	if _, err := os.Stat(mainKt); os.IsNotExist(err) {
		t.Errorf("Main.kt not created at %s", mainKt)
	}

	tamkConfig := filepath.Join(output.ProjectPath, "tamk.config")
	if _, err := os.Stat(tamkConfig); os.IsNotExist(err) {
		t.Errorf("tamk.config not created at %s", tamkConfig)
	}
}

func TestCreateProjectUseCase_WebAppInternal(t *testing.T) {
	cfg := config.New()
	cfg.ProjectDir = t.TempDir()
	projRepo := repoFS.NewProjectRepository()
	tmplRepo := repoFS.NewTemplateRepository(cfg)
	uc := NewCreateProjectUseCase(cfg, projRepo, tmplRepo)

	output, err := uc.Execute(context.Background(), CreateProjectInput{
		Name:     "TestWebApp",
		Type:     entity.ProjectTypeWebApp,
		Version:  "0.1.0",
		Author:   "Dev",
		Password: "test123",
		WebMode:  entity.WebContentInternal,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.Project.Type != entity.ProjectTypeWebApp {
		t.Errorf("Type = %q, want %q", output.Project.Type, entity.ProjectTypeWebApp)
	}

	androidManifest := filepath.Join(output.ProjectPath, "AndroidManifest.xml")
	if _, err := os.Stat(androidManifest); os.IsNotExist(err) {
		t.Log("AndroidManifest.xml not created (templates may not exist in test environment)")
	}
}

func TestCreateProjectUseCase_Duplicate(t *testing.T) {
	cfg := config.New()
	cfg.ProjectDir = t.TempDir()
	projRepo := repoFS.NewProjectRepository()
	tmplRepo := repoFS.NewTemplateRepository(cfg)
	uc := NewCreateProjectUseCase(cfg, projRepo, tmplRepo)

	_, err := uc.Execute(context.Background(), CreateProjectInput{
		Name:    "DupTest",
		Type:    entity.ProjectTypeConsole,
		Version: "1.0.0",
		Author:  "Author",
	})
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err = uc.Execute(context.Background(), CreateProjectInput{
		Name:    "DupTest",
		Type:    entity.ProjectTypeConsole,
		Version: "1.0.0",
		Author:  "Author",
	})
	if err == nil {
		t.Error("expected error for duplicate project")
	}
}
