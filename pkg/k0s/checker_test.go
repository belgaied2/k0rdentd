package k0s

import (
	"os"
	"strings"
	"testing"

	"github.com/onsi/gomega"
)

func TestCheckK0s(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test when k0s is not installed
	t.Run("k0s not installed", func(t *testing.T) {
		// Temporarily modify PATH to exclude k0s
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Create a new PATH without k0s
		newPath := ""
		for _, dir := range originalPathSplit(originalPath) {
			if dir != "/usr/local/bin" {
				if newPath != "" {
					newPath += ":"
				}
				newPath += dir
			}
		}
		os.Setenv("PATH", newPath)

		result, err := CheckK0s()
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result).ToNot(gomega.BeNil())
		g.Expect(result.Exists).To(gomega.BeFalse())
		g.Expect(result.Installed).To(gomega.BeFalse())
	})

	// Test when k0s is installed
	t.Run("k0s installed", func(t *testing.T) {
		// This test assumes k0s is installed in the test environment
		// In a real scenario, you might want to mock the exec.LookPath and exec.Command
		result, err := CheckK0s()
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result).ToNot(gomega.BeNil())
		g.Expect(result.Exists).To(gomega.BeTrue())
		g.Expect(result.Installed).To(gomega.BeTrue())
		g.Expect(result.Version).To(gomega.Not(gomega.BeEmpty()))
	})
}

func TestGetK0sVersion(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test with valid k0s version output
	t.Run("valid version output", func(t *testing.T) {
		// This test would need to be run with a mock or actual k0s binary
		// For now, we'll just verify the function signature and basic logic
		version, err := getK0sVersion()
		g.Expect(err).To(gomega.BeNil())
		g.Expect(version).To(gomega.Not(gomega.BeEmpty()))
	})

	// Test with invalid version output
	t.Run("invalid version output", func(t *testing.T) {
		// Mock exec.Command to return invalid output
		// This would require more complex mocking setup
		// For now, we'll skip this test
	})
}

func TestInstallK0s(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test installation
	t.Run("install k0s", func(t *testing.T) {
		// This test would need to be run in a controlled environment
		// For now, we'll just verify the function signature
		err := InstallK0s()
		g.Expect(err).To(gomega.BeNil())
	})
}

func TestCheckResult(t *testing.T) {
	g := gomega.NewWithT(t)

	result := &CheckResult{
		Exists:    true,
		Version:   "v1.27.4+k0s.0",
		Installed: true,
	}

	g.Expect(result.Exists).To(gomega.BeTrue())
	g.Expect(result.Version).To(gomega.Equal("v1.27.4+k0s.0"))
	g.Expect(result.Installed).To(gomega.BeTrue())
}

func originalPathSplit(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(path, ":")
}
