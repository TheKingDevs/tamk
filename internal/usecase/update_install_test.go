package usecase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheKingDevs/tamk/internal/config"
	repo "github.com/TheKingDevs/tamk/internal/repository"
)

func TestUpdateUseCase_DetectUpdateMethod_Git(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	cfg := config.New()
	cfg.TAMKHome = tmpDir
	updateRepo := repo.NewUpdateRepository()
	uc := NewUpdateUseCase(cfg, updateRepo)

	method := uc.detectUpdateMethod()
	if method != "git" {
		t.Errorf("detectUpdateMethod() = %q, want %q", method, "git")
	}
}

func TestUpdateUseCase_DetectUpdateMethod_Go(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.New()
	cfg.TAMKHome = tmpDir
	updateRepo := repo.NewUpdateRepository()
	uc := NewUpdateUseCase(cfg, updateRepo)

	method := uc.detectUpdateMethod()
	if method != "go" {
		t.Errorf("detectUpdateMethod() = %q, want %q", method, "go")
	}
}

func TestUpdateUseCase_Install_NilInfo(t *testing.T) {
	cfg := config.New()
	updateRepo := repo.NewUpdateRepository()
	uc := NewUpdateUseCase(cfg, updateRepo)

	err := uc.Install(nil, nil)
	if err == nil {
		t.Error("Install(nil) should return error, got nil")
	}
	if err.Error() != "no update info provided" {
		t.Errorf("Install(nil) error = %q, want %q", err.Error(), "no update info provided")
	}
}
