package usecase

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindAPKInDir_ReleaseAPK(t *testing.T) {
	tmpDir := t.TempDir()
	apk := filepath.Join(tmpDir, "myapp-1.0.0-release.apk")
	os.WriteFile(apk, []byte("fake"), 0o644)

	got, err := FindAPKInDir(tmpDir)
	if err != nil {
		t.Fatalf("FindAPKInDir() error = %v", err)
	}
	if got != apk {
		t.Errorf("FindAPKInDir() = %q, want %q", got, apk)
	}
}

func TestFindAPKInDir_DevAPK(t *testing.T) {
	tmpDir := t.TempDir()
	apk := filepath.Join(tmpDir, "myapp-1.0.0-dev.apk")
	os.WriteFile(apk, []byte("fake"), 0o644)

	got, err := FindAPKInDir(tmpDir)
	if err != nil {
		t.Fatalf("FindAPKInDir() error = %v", err)
	}
	if got != apk {
		t.Errorf("FindAPKInDir() = %q, want %q", got, apk)
	}
}

func TestFindAPKInDir_PrefersRelease(t *testing.T) {
	tmpDir := t.TempDir()
	release := filepath.Join(tmpDir, "app-1.0.0-release.apk")
	dev := filepath.Join(tmpDir, "app-1.0.0-dev.apk")
	os.WriteFile(release, []byte("release"), 0o644)
	os.WriteFile(dev, []byte("dev"), 0o644)

	got, err := FindAPKInDir(tmpDir)
	if err != nil {
		t.Fatalf("FindAPKInDir() error = %v", err)
	}
	if got != release {
		t.Errorf("FindAPKInDir() = %q, want %q (should prefer release)", got, release)
	}
}

func TestFindAPKInDir_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := FindAPKInDir(tmpDir)
	if err == nil {
		t.Error("FindAPKInDir() should return error when no APK found")
	}
}

func TestFindAllAPKs_Multiple(t *testing.T) {
	tmpDir := t.TempDir()
	apks := []string{"app-1.0.0-release.apk", "app-1.0.0-dev.apk", "other.apk"}
	for _, name := range apks {
		os.WriteFile(filepath.Join(tmpDir, name), []byte("fake"), 0o644)
	}

	got := FindAllAPKs(tmpDir)
	if len(got) != 3 {
		t.Errorf("FindAllAPKs() returned %d APKs, want 3", len(got))
	}
}

func TestFindAllAPKs_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	got := FindAllAPKs(tmpDir)
	if len(got) != 0 {
		t.Errorf("FindAllAPKs() returned %d APKs, want 0", len(got))
	}
}

func TestFindAllAPKs_IgnoresNonAPK(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "app.txt"), []byte("text"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "app.apk"), []byte("apk"), 0o644)

	got := FindAllAPKs(tmpDir)
	if len(got) != 1 {
		t.Errorf("FindAllAPKs() returned %d APKs, want 1", len(got))
	}
}
