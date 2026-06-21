package entity

import (
	"testing"
)

func TestSecurityConfigForLevel(t *testing.T) {
	tests := []struct {
		name           string
		level          SecurityLevel
		wantEnabled    bool
		wantAntiDebug  bool
		wantAntiRoot   bool
		wantAntiFrida  bool
		wantIntegrity  bool
		wantRASP       bool
	}{
		{
			name:          "none",
			level:         SecurityLevelNone,
			wantEnabled:   false,
			wantAntiDebug: false,
		},
		{
			name:          "basic",
			level:         SecurityLevelBasic,
			wantEnabled:   true,
			wantAntiDebug: true,
			wantAntiRoot:  false,
		},
		{
			name:          "standard",
			level:         SecurityLevelStandard,
			wantEnabled:   true,
			wantAntiDebug: true,
			wantAntiRoot:  true,
			wantAntiFrida: true,
			wantIntegrity: true,
		},
		{
			name:          "maximum",
			level:         SecurityLevelMaximum,
			wantEnabled:   true,
			wantAntiDebug: true,
			wantAntiRoot:  true,
			wantAntiFrida: true,
			wantIntegrity: true,
			wantRASP:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := SecurityConfigForLevel(tt.level)
			if cfg.Enabled != tt.wantEnabled {
				t.Errorf("Enabled = %v, want %v", cfg.Enabled, tt.wantEnabled)
			}
			if cfg.Guardian.AntiDebug != tt.wantAntiDebug {
				t.Errorf("AntiDebug = %v, want %v", cfg.Guardian.AntiDebug, tt.wantAntiDebug)
			}
			if cfg.Guardian.AntiRoot != tt.wantAntiRoot {
				t.Errorf("AntiRoot = %v, want %v", cfg.Guardian.AntiRoot, tt.wantAntiRoot)
			}
			if cfg.Guardian.AntiFrida != tt.wantAntiFrida {
				t.Errorf("AntiFrida = %v, want %v", cfg.Guardian.AntiFrida, tt.wantAntiFrida)
			}
			if cfg.Guardian.Integrity != tt.wantIntegrity {
				t.Errorf("Integrity = %v, want %v", cfg.Guardian.Integrity, tt.wantIntegrity)
			}
			if cfg.Guardian.RASP != tt.wantRASP {
				t.Errorf("RASP = %v, want %v", cfg.Guardian.RASP, tt.wantRASP)
			}
		})
	}
}

func TestDefaultSecurityConfig(t *testing.T) {
	cfg := DefaultSecurityConfig()
	if cfg.Enabled {
		t.Error("Default config should not be enabled")
	}
	if cfg.Level != SecurityLevelNone {
		t.Errorf("Default level = %v, want %v", cfg.Level, SecurityLevelNone)
	}
}
