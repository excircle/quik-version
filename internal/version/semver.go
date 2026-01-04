package version

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse parses a version string (with or without 'v' prefix) into major, minor, patch
func Parse(v string) (major, minor, patch int, err error) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")

	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", v)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return major, minor, patch, nil
}

// Format formats major, minor, patch into a version string (without 'v' prefix)
func Format(major, minor, patch int) string {
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

// IncrementMajor increments the major version and resets minor and patch
func IncrementMajor(v string) string {
	major, _, _, err := Parse(v)
	if err != nil {
		return "1.0.0"
	}
	return Format(major+1, 0, 0)
}

// IncrementMinor increments the minor version and resets patch
func IncrementMinor(v string) string {
	major, minor, _, err := Parse(v)
	if err != nil {
		return "0.1.0"
	}
	return Format(major, minor+1, 0)
}

// IncrementPatch increments the patch version
func IncrementPatch(v string) string {
	major, minor, patch, err := Parse(v)
	if err != nil {
		return "0.0.1"
	}
	return Format(major, minor, patch+1)
}

// Initial returns the initial version for a given increment type
func Initial(incrementType string) string {
	switch incrementType {
	case "major":
		return "1.0.0"
	case "patch":
		return "0.0.1"
	default:
		return "0.1.0"
	}
}
