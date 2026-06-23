package usecase

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestZipDir_Basic(t *testing.T) {
	srcDir := t.TempDir()
	destFile := filepath.Join(t.TempDir(), "out.zip")

	os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("hello"), 0o644)
	os.WriteFile(filepath.Join(srcDir, "file2.txt"), []byte("world"), 0o644)

	if err := zipDir(srcDir, destFile); err != nil {
		t.Fatalf("zipDir() error = %v", err)
	}

	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Fatal("zip file was not created")
	}

	r, err := zip.OpenReader(destFile)
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 2 {
		t.Errorf("zip contains %d files, want 2", len(r.File))
	}
}

func TestZipDir_Empty(t *testing.T) {
	srcDir := t.TempDir()
	destFile := filepath.Join(t.TempDir(), "empty.zip")

	if err := zipDir(srcDir, destFile); err != nil {
		t.Fatalf("zipDir() error = %v", err)
	}

	r, err := zip.OpenReader(destFile)
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 0 {
		t.Errorf("zip contains %d files, want 0", len(r.File))
	}
}

func TestZipDir_WithSubdirectory(t *testing.T) {
	srcDir := t.TempDir()
	subDir := filepath.Join(srcDir, "sub")
	os.MkdirAll(subDir, 0o755)
	os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested"), 0o644)
	os.WriteFile(filepath.Join(srcDir, "root.txt"), []byte("root"), 0o644)

	destFile := filepath.Join(t.TempDir(), "nested.zip")
	if err := zipDir(srcDir, destFile); err != nil {
		t.Fatalf("zipDir() error = %v", err)
	}

	r, err := zip.OpenReader(destFile)
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}
	defer r.Close()

	if len(r.File) < 2 {
		t.Errorf("zip contains %d files, want at least 2", len(r.File))
	}

	found := false
	for _, f := range r.File {
		if f.Name == "sub/nested.txt" || f.Name == "sub\\nested.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("nested file not found in zip")
	}
}

func TestZipDir_NonexistentSource(t *testing.T) {
	destFile := filepath.Join(t.TempDir(), "out.zip")
	err := zipDir("/nonexistent/path", destFile)
	if err == nil {
		t.Error("zipDir() with nonexistent source should return error")
	}
}
