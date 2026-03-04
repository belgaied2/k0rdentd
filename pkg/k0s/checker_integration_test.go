//go:build integration

package k0s

import (
	"testing"

	"github.com/onsi/gomega"
)

// TestCheckK0s_Installed tests k0s checking when k0s is installed.
// This test requires a real k0s binary to be present in the system.
// Run with: go test -tags=integration ./pkg/k0s/...
func TestCheckK0s_Installed(t *testing.T) {
	g := gomega.NewWithT(t)

	// This test assumes k0s is installed in the test environment
	result, err := CheckK0s()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result).ToNot(gomega.BeNil())
	g.Expect(result.Exists).To(gomega.BeTrue())
	g.Expect(result.Installed).To(gomega.BeTrue())
	g.Expect(result.Version).To(gomega.Not(gomega.BeEmpty()))
}

// TestGetK0sVersion tests getting the k0s version.
// This test requires a real k0s binary to be present in the system.
// Run with: go test -tags=integration ./pkg/k0s/...
func TestGetK0sVersion(t *testing.T) {
	g := gomega.NewWithT(t)

	version, err := GetK0sVersion()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(version).To(gomega.Not(gomega.BeEmpty()))
}
