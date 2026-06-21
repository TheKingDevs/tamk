package usecase

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/TheKingDevs/tamk/internal/domain/entity"
)

func TestApkFilename(t *testing.T) {
	tests := []struct {
		name    string
		project *entity.Project
		env     string
		want    string
	}{
		{
			name: "release build",
			project: &entity.Project{
				Name:    "My App",
				Version: "1.0.0",
			},
			env:  "release",
			want: "my-app-1.0.0-release.apk",
		},
		{
			name: "dev build",
			project: &entity.Project{
				Name:    "TestProject",
				Version: "0.1.0",
			},
			env:  "dev",
			want: "testproject-0.1.0-dev.apk",
		},
		{
			name: "single word name",
			project: &entity.Project{
				Name:    "App",
				Version: "2.0.0",
			},
			env:  "release",
			want: "app-2.0.0-release.apk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := apkFilename(tt.project, tt.env)
			if got != tt.want {
				t.Errorf("apkFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateKeystorePassword_CorrectPassword(t *testing.T) {
	tmpDir := t.TempDir()
	keystorePath := filepath.Join(tmpDir, "test.keystore")

	// Create a test keystore
	err := createTestKeystore(keystorePath, "correctpass")
	if err != nil {
		t.Skipf("keytool not available: %v", err)
	}

	if err := validateKeystorePassword(keystorePath, "correctpass"); err != nil {
		t.Errorf("validateKeystorePassword() with correct password returned error: %v", err)
	}
}

func TestValidateKeystorePassword_WrongPassword(t *testing.T) {
	tmpDir := t.TempDir()
	keystorePath := filepath.Join(tmpDir, "test.keystore")

	err := createTestKeystore(keystorePath, "correctpass")
	if err != nil {
		t.Skipf("keytool not available: %v", err)
	}

	if err := validateKeystorePassword(keystorePath, "wrongpass"); err == nil {
		t.Error("validateKeystorePassword() with wrong password should return error")
	}
}

func TestValidateKeystorePassword_EmptyPassword(t *testing.T) {
	tmpDir := t.TempDir()
	keystorePath := filepath.Join(tmpDir, "test.keystore")

	err := createTestKeystore(keystorePath, "correctpass")
	if err != nil {
		t.Skipf("keytool not available: %v", err)
	}

	if err := validateKeystorePassword(keystorePath, ""); err == nil {
		t.Error("validateKeystorePassword() with empty password should return error")
	}
}

func TestValidateKeystorePassword_FileNotFound(t *testing.T) {
	err := validateKeystorePassword("/nonexistent/path/keystore", "password")
	if err == nil {
		t.Error("validateKeystorePassword() with nonexistent file should return error")
	}
}

func TestFailedResult(t *testing.T) {
	result := failedResult(entity.BuildPhaseAAPT2Compile, "test error message")

	if result.Success {
		t.Error("failedResult() should return Success=false")
	}
	if result.Phase != entity.BuildPhaseAAPT2Compile {
		t.Errorf("Phase = %q, want %q", result.Phase, entity.BuildPhaseAAPT2Compile)
	}
	if result.ErrorMsg == "" {
		t.Error("ErrorMsg should not be empty")
	}
}

// createTestKeystore creates a test keystore using keytool.
func createTestKeystore(path, password string) error {
	cmd := exec.Command("keytool", "-genkeypair",
		"-alias", "test",
		"-keyalg", "RSA",
		"-keysize", "2048",
		"-validity", "1",
		"-keystore", path,
		"-storepass", password,
		"-dname", "CN=Test, OU=Test, O=Test, L=Test, ST=Test, C=US",
	)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
