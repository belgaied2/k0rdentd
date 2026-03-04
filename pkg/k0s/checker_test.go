package k0s

import (
	"os"
	"strings"
	"testing"

	"github.com/onsi/gomega"
)

func TestCheckK0s_NotInstalled(t *testing.T) {
	g := gomega.NewWithT(t)

	// Temporarily modify PATH to exclude k0s
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Create a new PATH without k0s
	newPath := ""
	for _, dir := range strings.Split(originalPath, ":") {
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
