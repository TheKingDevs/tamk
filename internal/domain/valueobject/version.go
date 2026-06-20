package valueobject

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var semverPattern = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-[a-zA-Z0-9.]+)?$`)

type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Original   string
}

func ParseVersion(v string) (Version, error) {
	matches := semverPattern.FindStringSubmatch(v)
	if matches == nil {
		return Version{}, fmt.Errorf("invalid version %s: %w", v, ErrInvalidVersion)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	prerelease := ""
	if matches[4] != "" {
		prerelease = strings.TrimPrefix(matches[4], "-")
	}

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: prerelease,
		Original:   v,
	}, nil
}

func (v Version) String() string {
	if v.Prerelease != "" {
		return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Prerelease)
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) GreaterThan(other Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	if v.Patch != other.Patch {
		return v.Patch > other.Patch
	}
	return false
}

func (v Version) IsValid() bool {
	return v.Major > 0 || v.Minor > 0 || v.Patch > 0
}
