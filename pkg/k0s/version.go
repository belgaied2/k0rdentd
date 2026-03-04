package k0s

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// VersionPattern matches k0s version format: v{KUBERNETES_VERSION}+k0s.{K0S_PATCH}
// Examples: v1.32.4+k0s.0, v1.31.0+k0s.0, v1.30.0+k0s.0
var VersionPattern = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)\+k0s\.(\d+)$`)

// VersionComponents represents parsed version components
type VersionComponents struct {
	Major      int
	Minor      int
	Patch      int
	K0sPatch   int
	Original   string
}

// ValidateVersion validates that a version string matches the k0s version format
func ValidateVersion(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	if !VersionPattern.MatchString(version) {
		return fmt.Errorf("invalid k0s version format: %s (expected: v1.32.4+k0s.0)", version)
	}

	return nil
}

// ParseVersion parses a k0s version string into its components
func ParseVersion(version string) (*VersionComponents, error) {
	if err := ValidateVersion(version); err != nil {
		return nil, err
	}

	matches := VersionPattern.FindStringSubmatch(version)
	if len(matches) != 5 {
		return nil, fmt.Errorf("failed to parse version: %s", version)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	k0sPatch, _ := strconv.Atoi(matches[4])

	return &VersionComponents{
		Major:    major,
		Minor:    minor,
		Patch:    patch,
		K0sPatch: k0sPatch,
		Original: version,
	}, nil
}

// CompareVersions compares two k0s version strings
// Returns:
//   -1 if v1 < v2
//    0 if v1 == v2
//    1 if v1 > v2
func CompareVersions(v1, v2 string) (int, error) {
	// Handle empty versions
	if v1 == "" && v2 == "" {
		return 0, nil
	}
	if v1 == "" {
		return -1, nil
	}
	if v2 == "" {
		return 1, nil
	}

	// Parse versions
	parsed1, err := ParseVersion(v1)
	if err != nil {
		return 0, fmt.Errorf("failed to parse first version: %w", err)
	}

	parsed2, err := ParseVersion(v2)
	if err != nil {
		return 0, fmt.Errorf("failed to parse second version: %w", err)
	}

	// Compare major version
	if parsed1.Major < parsed2.Major {
		return -1, nil
	}
	if parsed1.Major > parsed2.Major {
		return 1, nil
	}

	// Compare minor version
	if parsed1.Minor < parsed2.Minor {
		return -1, nil
	}
	if parsed1.Minor > parsed2.Minor {
		return 1, nil
	}

	// Compare patch version
	if parsed1.Patch < parsed2.Patch {
		return -1, nil
	}
	if parsed1.Patch > parsed2.Patch {
		return 1, nil
	}

	// Compare k0s patch version
	if parsed1.K0sPatch < parsed2.K0sPatch {
		return -1, nil
	}
	if parsed1.K0sPatch > parsed2.K0sPatch {
		return 1, nil
	}

	return 0, nil
}

// VersionsEqual checks if two version strings are equal
func VersionsEqual(v1, v2 string) (bool, error) {
	result, err := CompareVersions(v1, v2)
	if err != nil {
		return false, err
	}
	return result == 0, nil
}

// VersionConflict represents a detected version conflict
type VersionConflict struct {
	InstalledVersion string
	ConfigVersion    string
	IsRunning        bool
}

// String returns a human-readable description of the conflict
func (c *VersionConflict) String() string {
	return fmt.Sprintf("installed: %s, config: %s, running: %t",
		c.InstalledVersion, c.ConfigVersion, c.IsRunning)
}

// NeedsResolution returns true if the conflict needs user intervention
func (c *VersionConflict) NeedsResolution() bool {
	if c.InstalledVersion == "" || c.ConfigVersion == "" {
		return false
	}

	equal, _ := VersionsEqual(c.InstalledVersion, c.ConfigVersion)
	return !equal
}

// RequiresManualIntervention returns true if the conflict cannot be resolved automatically
func (c *VersionConflict) RequiresManualIntervention() bool {
	return c.NeedsResolution() && c.IsRunning
}

// CanAutoReplace returns true if k0s can be automatically replaced
func (c *VersionConflict) CanAutoReplace() bool {
	return c.NeedsResolution() && !c.IsRunning
}

// FormatWarningMessage formats a warning message for airgap version mismatch
func FormatWarningMessage(configVersion, bundledVersion string) string {
	return fmt.Sprintf("Config specifies k0s version %s, but bundled version is %s.\n   Using bundled version for airgap installation.",
		configVersion, bundledVersion)
}

// FormatConflictMessage formats a conflict message for user interaction
func FormatConflictMessage(installedVersion, configVersion string) string {
	var sb strings.Builder
	sb.WriteString("k0s version conflict detected!\n")
	sb.WriteString(fmt.Sprintf("   Installed: %s\n", installedVersion))
	sb.WriteString(fmt.Sprintf("   Config:    %s\n", configVersion))
	return sb.String()
}

// FormatRunningConflictError formats an error message for running k0s with version conflict
func FormatRunningConflictError(installedVersion, configVersion string) string {
	var sb strings.Builder
	sb.WriteString("Cannot replace k0s while it's running!\n")
	sb.WriteString(fmt.Sprintf("   The installed k0s (%s) is currently running as a service.\n", installedVersion))
	sb.WriteString(fmt.Sprintf("   Config specifies: %s\n", configVersion))
	sb.WriteString("\n")
	sb.WriteString("   To proceed, you must manually stop and reset k0s:\n")
	sb.WriteString("     sudo k0s stop\n")
	sb.WriteString("\n")
	sb.WriteString("   Then run k0rdentd install again.")
	return sb.String()
}
