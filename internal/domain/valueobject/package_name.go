package valueobject

import (
	"fmt"
	"regexp"
)

var packagePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)+$`)

type PackageName struct {
	Value string
}

func NewPackageName(name string) (PackageName, error) {
	if !packagePattern.MatchString(name) {
		return PackageName{}, fmt.Errorf("invalid package name %s: %w", name, ErrInvalidPackageName)
	}
	return PackageName{Value: name}, nil
}

func (p PackageName) String() string {
	return p.Value
}
