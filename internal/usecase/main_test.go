package usecase

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	root := findProjectRoot()
	if root != "" {
		os.Setenv("TAMK_HOME", root)
	}
	os.Exit(m.Run())
}

func findProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
