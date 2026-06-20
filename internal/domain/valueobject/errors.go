package valueobject

import "errors"

var (
	ErrInvalidProjectName = errors.New("invalid project name")
	ErrInvalidVersion     = errors.New("invalid version (expected SEMVER)")
	ErrInvalidPackageName = errors.New("invalid Android package name")
	ErrReservedName       = errors.New("name is reserved")
)
