package entity

import (
	"testing"
)

func TestProjectType_Values(t *testing.T) {
	if ProjectTypeWebApp != "webapp" {
		t.Errorf("ProjectTypeWebApp = %q, want %q", ProjectTypeWebApp, "webapp")
	}
	if ProjectTypeUIAPK != "ui_apk" {
		t.Errorf("ProjectTypeUIAPK = %q, want %q", ProjectTypeUIAPK, "ui_apk")
	}
	if ProjectTypeConsole != "console" {
		t.Errorf("ProjectTypeConsole = %q, want %q", ProjectTypeConsole, "console")
	}
}

func TestWebContentMode_Values(t *testing.T) {
	if WebContentInternal != "internal" {
		t.Errorf("WebContentInternal = %q, want %q", WebContentInternal, "internal")
	}
	if WebContentExternal != "external" {
		t.Errorf("WebContentExternal = %q, want %q", WebContentExternal, "external")
	}
}

func TestProject_Creation(t *testing.T) {
	p := &Project{
		Name:        "TestApp",
		Type:        ProjectTypeWebApp,
		Version:     "1.0.0",
		Author:      "TestAuthor",
		PackageName: "com.testauthor.testapp",
		MinSDK:      21,
		TargetSDK:   30,
	}

	if p.Name != "TestApp" {
		t.Errorf("Name = %q, want %q", p.Name, "TestApp")
	}
	if p.MinSDK != 21 {
		t.Errorf("MinSDK = %d, want %d", p.MinSDK, 21)
	}
	if p.TargetSDK != 30 {
		t.Errorf("TargetSDK = %d, want %d", p.TargetSDK, 30)
	}
}
