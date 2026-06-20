package valueobject

import "testing"

func TestNewPackageName_Valid(t *testing.T) {
	tests := []string{
		"com.example.app",
		"com.google.android",
		"io.github.user.project",
		"com.author.myapp",
	}

	for _, pkg := range tests {
		t.Run(pkg, func(t *testing.T) {
			p, err := NewPackageName(pkg)
			if err != nil {
				t.Errorf("unexpected error for %q: %v", pkg, err)
				return
			}
			if p.String() != pkg {
				t.Errorf("String() = %q, want %q", p.String(), pkg)
			}
		})
	}
}

func TestNewPackageName_Invalid(t *testing.T) {
	tests := []string{
		"",
		"com",
		"com.example.123",
		"com.Example.App",
		"com.example.app.",
	}

	for _, pkg := range tests {
		t.Run(pkg, func(t *testing.T) {
			_, err := NewPackageName(pkg)
			if err == nil {
				t.Errorf("expected error for %q", pkg)
			}
		})
	}
}
