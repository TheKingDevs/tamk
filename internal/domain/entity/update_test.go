package entity

import "testing"

func TestUpdateLevel_Order(t *testing.T) {
	if UpdateLevelOptional >= UpdateLevelPatch {
		t.Error("Optional should be less than Patch")
	}
	if UpdateLevelPatch >= UpdateLevelMinor {
		t.Error("Patch should be less than Minor")
	}
	if UpdateLevelMinor >= UpdateLevelMajor {
		t.Error("Minor should be less than Major")
	}
	if UpdateLevelMajor >= UpdateLevelCritical {
		t.Error("Major should be less than Critical")
	}
}

func TestUpdateInfo(t *testing.T) {
	info := &UpdateInfo{
		CurrentVersion: "1.0.0",
		LatestVersion:  "1.1.0",
		Level:          UpdateLevelMinor,
		ReleaseNotes:   "Bug fixes and improvements",
		DownloadURL:    "https://github.com/Shadw-Developer/tamk/releases/v1.1.0",
	}

	if info.CurrentVersion != "1.0.0" {
		t.Errorf("CurrentVersion = %q, want %q", info.CurrentVersion, "1.0.0")
	}
	if info.Level != UpdateLevelMinor {
		t.Errorf("Level = %d, want %d", info.Level, UpdateLevelMinor)
	}
}
