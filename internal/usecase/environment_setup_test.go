package usecase

import (
	"testing"

	"github.com/TheKingDevs/tamk/internal/config"
)

func TestSetupEnvironmentUseCase_Verify(t *testing.T) {
	cfg := config.New()
	uc := NewSetupEnvironmentUseCase(cfg)

	result := uc.VerifyEnvironment()
	t.Logf("Environment verification result: %v", result)
	t.Logf("SDK path: %s", cfg.SDKPath)
	t.Logf("Keystore path: %s", cfg.Keystore)
}
