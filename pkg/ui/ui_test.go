package ui

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestShouldIgnoreInterface(t *testing.T) {
	g := gomega.NewWithT(t)

	tests := []struct {
		name     string
		expected bool
	}{
		{"cali123", true},
		{"cali0", true},
		{"califl", true},
		{"vxlan.calico", true},
		{"vxlan.calico0", true},
		{"tunl0", true},
		{"tunl123", true},
		{"wg0", true},
		{"eth0", false},
		{"enp0s0", false},
		{"lo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIgnoreInterface(tt.name)
			g.Expect(result).To(gomega.Equal(tt.expected))
		})
	}
}

func TestCloudProviderConstants(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("cloud provider constants", func(t *testing.T) {
		g.Expect(CloudProviderAWS).To(gomega.Equal("aws"))
		g.Expect(CloudProviderGCP).To(gomega.Equal("gcp"))
		g.Expect(CloudProviderAzure).To(gomega.Equal("azure"))
		g.Expect(CloudProviderNone).To(gomega.Equal("none"))
	})
}

func TestK0rdentUIConstants(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("k0rdent UI constants", func(t *testing.T) {
		g.Expect(k0rdentUIDeploymentName).To(gomega.Equal("k0rdent-k0rdent-ui"))
		g.Expect(k0rdentUIServiceName).To(gomega.Equal("k0rdent-k0rdent-ui"))
		g.Expect(k0rdentUINamespace).To(gomega.Equal("kcm-system"))
		g.Expect(k0rdentUIIngressName).To(gomega.Equal("k0rdent-ui"))
		g.Expect(k0rdentUIIngressPath).To(gomega.Equal("/k0rdent-ui"))
	})
}

// TestGetBasicAuthPassword tests the GetBasicAuthPassword function
func TestGetBasicAuthPassword(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("GetBasicAuthPassword function", func(t *testing.T) {
		// This test verifies that the function can be called without panicking
		// In a real environment with a k0s cluster, this would return the actual password
		password, err := GetBasicAuthPassword()

		// We expect this to fail in a test environment without a real cluster
		// but we want to ensure proper error handling
		if err != nil {
			// This is expected in a test environment without a real k0s cluster
			g.Expect(err.Error()).To(gomega.ContainSubstring("failed to get Basic Auth password"))
		} else {
			// If we somehow get a password, it should not be empty
			g.Expect(password).ToNot(gomega.BeEmpty())
		}
	})
}
