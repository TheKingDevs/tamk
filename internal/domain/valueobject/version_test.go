package valueobject

import (
	"testing"
)

func TestParseVersion_Valid(t *testing.T) {
	tests := []struct {
		input   string
		major   int
		minor   int
		patch   int
		wantErr bool
	}{
		{"1.0.0", 1, 0, 0, false},
		{"2026.3.0", 2026, 3, 0, false},
		{"0.0.1", 0, 0, 1, false},
		{"2.1.3-beta", 2, 1, 3, false},
		{"1.0.0-alpha.1", 1, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if v.Major != tt.major || v.Minor != tt.minor || v.Patch != tt.patch {
					t.Errorf("ParseVersion() = %d.%d.%d, want %d.%d.%d",
						v.Major, v.Minor, v.Patch, tt.major, tt.minor, tt.patch)
				}
			}
		})
	}
}

func TestParseVersion_Invalid(t *testing.T) {
	invalid := []string{
		"", "1", "1.0", "abc", "1.0.0.0", "v1.0.0", "01.0.0",
	}

	for _, v := range invalid {
		t.Run(v, func(t *testing.T) {
			_, err := ParseVersion(v)
			if err == nil {
				t.Errorf("expected error for input %q", v)
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	v, _ := ParseVersion("1.2.3")
	if v.String() != "1.2.3" {
		t.Errorf("String() = %s, want 1.2.3", v.String())
	}

	v, _ = ParseVersion("1.2.3-rc1")
	if v.String() != "1.2.3-rc1" {
		t.Errorf("String() = %s, want 1.2.3-rc1", v.String())
	}
}

func TestVersion_GreaterThan(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"2.0.0", "1.0.0", true},
		{"1.1.0", "1.0.0", true},
		{"1.0.1", "1.0.0", true},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "2.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.a+">"+tt.b, func(t *testing.T) {
			a, _ := ParseVersion(tt.a)
			b, _ := ParseVersion(tt.b)
			if got := a.GreaterThan(b); got != tt.want {
				t.Errorf("GreaterThan() = %v, want %v", got, tt.want)
			}
		})
	}
}
