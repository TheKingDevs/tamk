package valueobject

import "testing"

func TestNewProjectName_Valid(t *testing.T) {
	tests := []struct {
		input string
		clean string
	}{
		{"MyApp", "MyApp"},
		{"hello-world", "hello-world"},
		{"test_app", "test_app"},
		{"  spaces  ", "spaces"},
		{"-leading", "leading"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			n, err := NewProjectName(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if n.String() != tt.clean {
				t.Errorf("Clean = %q, want %q", n.String(), tt.clean)
			}
		})
	}
}

func TestNewProjectName_Invalid(t *testing.T) {
	tests := []struct {
		input  string
		reason string
	}{
		{"a", "too short"},
		{"", "empty"},
		{"android", "reserved"},
		{"COM", "reserved (case insensitive)"},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			_, err := NewProjectName(tt.input)
			if err == nil {
				t.Errorf("expected error for %q (%s)", tt.input, tt.reason)
			}
		})
	}
}

func TestNewProjectName_TooLong(t *testing.T) {
	long := ""
	for i := 0; i < 60; i++ {
		long += "a"
	}
	_, err := NewProjectName(long)
	if err == nil {
		t.Error("expected error for name > 50 chars")
	}
}
