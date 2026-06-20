package usecase

import (
	"testing"

	"github.com/Shadw-Developer/tamk/internal/config"
	"github.com/Shadw-Developer/tamk/internal/domain/entity"
	repo "github.com/Shadw-Developer/tamk/internal/repository"
)

func TestUpdateUseCase_DetectLevel(t *testing.T) {
	repo := repo.NewUpdateRepository()

	tests := []struct {
		latest  string
		current string
		notes   string
		level   entity.UpdateLevel
	}{
		{"1.1.0", "1.0.0", "critical security fix", entity.UpdateLevelCritical},
		{"1.1.0", "1.0.0", "bugfix release", entity.UpdateLevelPatch},
		{"2.0.0", "1.0.0", "new features", entity.UpdateLevelMajor},
		{"1.1.0", "1.0.0", "minor improvements", entity.UpdateLevelMinor},
		{"1.0.5", "1.0.0", "docs update", entity.UpdateLevelOptional},
	}

	for _, tt := range tests {
		name := tt.notes
		if len(name) > 20 {
			name = name[:20]
		}
		t.Run(name, func(t *testing.T) {
			level := repo.DetectUpdateLevel(tt.latest, tt.current, tt.notes)
			if level != tt.level {
				t.Errorf("DetectUpdateLevel() = %d, want %d", level, tt.level)
			}
		})
	}
}

func TestUpdateUseCase_AutoInstall(t *testing.T) {
	cfg := config.New()
	updateRepo := repo.NewUpdateRepository()
	uc := NewUpdateUseCase(cfg, updateRepo)

	critical := &entity.UpdateInfo{Level: entity.UpdateLevelCritical}
	if !uc.ShouldAutoInstall(critical) {
		t.Error("CRITICAL should auto-install")
	}

	patch := &entity.UpdateInfo{Level: entity.UpdateLevelPatch}
	if !uc.ShouldAutoInstall(patch) {
		t.Error("PATCH should auto-install")
	}

	minor := &entity.UpdateInfo{Level: entity.UpdateLevelMinor}
	if uc.ShouldAutoInstall(minor) {
		t.Error("MINOR should NOT auto-install")
	}
}
