package errors

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrInvalidURL", ErrInvalidURL},
		{"ErrProjectNotFound", ErrProjectNotFound},
		{"ErrSDKNotFound", ErrSDKNotFound},
		{"ErrKeystoreNotFound", ErrKeystoreNotFound},
		{"ErrKeystoreInvalidPass", ErrKeystoreInvalidPass},
		{"ErrBuildFailed", ErrBuildFailed},
		{"ErrTemplateNotFound", ErrTemplateNotFound},
		{"ErrPathTraversal", ErrPathTraversal},
		{"ErrNotAWebAppProject", ErrNotAWebAppProject},
		{"ErrAPKBaseNotFound", ErrAPKBaseNotFound},
		{"ErrUpdateCheckFailed", ErrUpdateCheckFailed},
		{"ErrDeviceNotConnected", ErrDeviceNotConnected},
		{"ErrWatchdogNotAvailable", ErrWatchdogNotAvailable},
		{"ErrWebSocketNotAvailable", ErrWebSocketNotAvailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s should have a non-empty message", tt.name)
			}
		})
	}
}

func TestSentinelErrors_AreDistinct(t *testing.T) {
	allErrors := []error{
		ErrInvalidURL,
		ErrProjectNotFound,
		ErrSDKNotFound,
		ErrKeystoreNotFound,
		ErrKeystoreInvalidPass,
		ErrBuildFailed,
		ErrTemplateNotFound,
		ErrPathTraversal,
		ErrNotAWebAppProject,
		ErrAPKBaseNotFound,
		ErrUpdateCheckFailed,
		ErrDeviceNotConnected,
		ErrWatchdogNotAvailable,
		ErrWebSocketNotAvailable,
	}

	seen := make(map[string]bool)
	for _, e := range allErrors {
		msg := e.Error()
		if seen[msg] {
			t.Errorf("duplicate error message: %q", msg)
		}
		seen[msg] = true
	}
}

func TestBuildError(t *testing.T) {
	inner := errors.New("kotlinc failed")
	be := &BuildError{
		Phase: "kotlin_compile",
		Err:   inner,
	}

	if be.Error() != "build failed at phase kotlin_compile: kotlinc failed" {
		t.Errorf("BuildError.Error() = %q", be.Error())
	}

	if !errors.Is(be, inner) {
		t.Error("BuildError should unwrap to inner error")
	}

	if be.Phase != "kotlin_compile" {
		t.Errorf("Phase = %q, want %q", be.Phase, "kotlin_compile")
	}
}

func TestBuildError_Unwrap(t *testing.T) {
	inner := errors.New("d8 conversion failed")
	be := &BuildError{Phase: "d8", Err: inner}

	if !errors.Is(be, inner) {
		t.Error("errors.Is should find inner error through BuildError")
	}
}
