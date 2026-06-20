package entity

import (
	"testing"
)

func TestBuildPhase_Values(t *testing.T) {
	phases := []BuildPhase{
		BuildPhaseAAPT2Compile,
		BuildPhaseAAPT2Link,
		BuildPhaseKotlinCompile,
		BuildPhaseD8,
		BuildPhasePackageDEX,
		BuildPhaseZipalign,
		BuildPhaseApkSign,
	}

	for _, p := range phases {
		if p == "" {
			t.Error("BuildPhase should not be empty")
		}
	}
}

func TestBuildResult_Success(t *testing.T) {
	r := &BuildResult{
		Success: true,
		APKPath: "/tmp/test/myapp-1.0.0-release.apk",
	}

	if !r.Success {
		t.Error("expected success")
	}
	if r.APKPath != "/tmp/test/myapp-1.0.0-release.apk" {
		t.Errorf("APKPath = %q, want %q", r.APKPath, "/tmp/test/myapp-1.0.0-release.apk")
	}
}

func TestBuildResult_Failure(t *testing.T) {
	r := &BuildResult{
		Success:  false,
		Phase:    BuildPhaseD8,
		ErrorMsg: "compilation error",
	}

	if r.Success {
		t.Error("expected failure")
	}
	if r.Phase != BuildPhaseD8 {
		t.Errorf("Phase = %q, want %q", r.Phase, BuildPhaseD8)
	}
}
