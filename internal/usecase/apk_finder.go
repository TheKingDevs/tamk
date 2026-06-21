package usecase

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FindAPKInDir searches for APK files in the given directory.
// It looks for *-release.apk first, then *-dev.apk.
func FindAPKInDir(projectPath string) (string, error) {
	patterns := []string{"*-release.apk", "*-dev.apk"}
	for _, p := range patterns {
		matches, err := filepath.Glob(filepath.Join(projectPath, p))
		if err == nil && len(matches) > 0 {
			return matches[0], nil
		}
	}
	return "", fmt.Errorf("APK not found in %s", projectPath)
}

// FindAllAPKs returns all APK files in the given directory.
func FindAllAPKs(projectPath string) []string {
	var apks []string
	entries, err := filepath.Glob(filepath.Join(projectPath, "*.apk"))
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !strings.HasSuffix(e, ".apk") {
			continue
		}
		apks = append(apks, e)
	}
	return apks
}
