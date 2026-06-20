package valueobject

import (
	"fmt"
	"regexp"
	"strings"
)

var namePattern = regexp.MustCompile(`[^a-zA-Z0-9\-_]`)

var reservedNames = map[string]bool{
	"android": true, "com": true, "org": true, "java": true,
	"kotlin": true, "assets": true, "res": true, "src": true,
	"build": true, "test": true, "main": true, "secret": true, "tamk": true,
}

type ProjectName struct {
	Original string
	Clean    string
}

func NewProjectName(name string) (ProjectName, error) {
	if len(name) < 2 {
		return ProjectName{}, fmt.Errorf("project name too short: %w", ErrInvalidProjectName)
	}
	if len(name) > 50 {
		return ProjectName{}, fmt.Errorf("name must be at most 50 characters")
	}

	clean := namePattern.ReplaceAllString(name, "")
	clean = strings.TrimLeft(clean, "-_")

	if reservedNames[strings.ToLower(clean)] {
		return ProjectName{}, fmt.Errorf("name '%s' is reserved: %w", clean, ErrReservedName)
	}

	return ProjectName{
		Original: name,
		Clean:    clean,
	}, nil
}

func (n ProjectName) String() string {
	return n.Clean
}
