package generator

import (
	"strings"
	"testing"

	"github.com/belgaied2/k0rdentd/pkg/config"
	"github.com/onsi/gomega"
)

func TestGenerateK0sConfig(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("empty k0s config fields should not generate empty api, network, or storage", func(t *testing.T) {
		cfg := &config.K0rdentdConfig{
			K0rdent: config.K0rdentConfig{
				Version: "1.2.2",
				Helm: config.K0rdentHelmConfig{
					Chart:     "k0rdent/kcm",
					Namespace: "kcm-system",
				},
			},
		}

		result, err := GenerateK0sConfig(cfg)
		g.Expect(err).ToNot(gomega.HaveOccurred())

		resultStr := string(result)

		// Should contain extensions (k0rdent installation)
		g.Expect(resultStr).To(gomega.ContainSubstring("extensions:"))
		g.Expect(resultStr).To(gomega.ContainSubstring("k0rdent"))

		// Should NOT contain empty api, network, or storage fields at spec level
		lines := strings.Split(resultStr, "\n")
		inSpec := false
		inExtensions := false
		foundAPI := false
		foundNetwork := false
		foundStorage := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "spec:" {
				inSpec = true
				continue
			}
			if inSpec && !inExtensions {
				// Check if we're still at the spec level (indented but not double-indented)
				if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
					if trimmed == "api:" {
						foundAPI = true
					}
					if trimmed == "network:" {
						foundNetwork = true
					}
					if trimmed == "storage:" {
						foundStorage = true
					}
					if trimmed == "extensions:" {
						inExtensions = true
					}
				}
			}
		}

		g.Expect(foundAPI).To(gomega.BeFalse(), "api should not be present when empty")
		g.Expect(foundNetwork).To(gomega.BeFalse(), "network should not be present when empty")
		g.Expect(foundStorage).To(gomega.BeFalse(), "storage should not be present when empty")
	})

	t.Run("partially set config should only generate fields with values", func(t *testing.T) {
		cfg := &config.K0rdentdConfig{
			K0s: config.K0sConfig{
				API: config.APIConfig{
					Address: "192.168.1.1",
					Port:    6443,
				},
				// Network and Storage are empty
			},
			K0rdent: config.K0rdentConfig{
				Version: "1.2.2",
				Helm: config.K0rdentHelmConfig{
					Chart:     "k0rdent/kcm",
					Namespace: "kcm-system",
				},
			},
		}

		result, err := GenerateK0sConfig(cfg)
		g.Expect(err).ToNot(gomega.HaveOccurred())

		resultStr := string(result)

		// Should contain API section
		g.Expect(resultStr).To(gomega.ContainSubstring("api:"))
		g.Expect(resultStr).To(gomega.ContainSubstring("address: 192.168.1.1"))
		g.Expect(resultStr).To(gomega.ContainSubstring("port: 6443"))

		// Should NOT contain network or storage at spec level
		lines := strings.Split(resultStr, "\n")
		inSpec := false
		inExtensions := false
		foundNetwork := false
		foundStorage := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "spec:" {
				inSpec = true
				continue
			}
			if inSpec && !inExtensions {
				if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
					if trimmed == "network:" {
						foundNetwork = true
					}
					if trimmed == "storage:" {
						foundStorage = true
					}
					if trimmed == "extensions:" {
						inExtensions = true
					}
				}
			}
		}

		g.Expect(foundNetwork).To(gomega.BeFalse(), "network should not be present when empty")
		g.Expect(foundStorage).To(gomega.BeFalse(), "storage should not be present when empty")
	})

	t.Run("fully set config should generate all fields", func(t *testing.T) {
		cfg := &config.K0rdentdConfig{
			K0s: config.K0sConfig{
				API: config.APIConfig{
					Address: "192.168.1.1",
					Port:    6443,
				},
				Network: config.NetworkConfig{
					Provider:    "calico",
					PodCIDR:     "10.244.0.0/16",
					ServiceCIDR: "10.96.0.0/12",
				},
				Storage: config.StorageConfig{
					Type: "etcd",
					Etcd: config.EtcdConfig{
						PeerAddress: "127.0.0.1",
					},
				},
			},
			K0rdent: config.K0rdentConfig{
				Version: "1.2.2",
				Helm: config.K0rdentHelmConfig{
					Chart:     "k0rdent/kcm",
					Namespace: "kcm-system",
					Values: map[string]interface{}{
						"replicas": 1,
					},
				},
			},
		}

		result, err := GenerateK0sConfig(cfg)
		g.Expect(err).ToNot(gomega.HaveOccurred())

		resultStr := string(result)

		// Should contain all sections
		g.Expect(resultStr).To(gomega.ContainSubstring("api:"))
		g.Expect(resultStr).To(gomega.ContainSubstring("address: 192.168.1.1"))
		g.Expect(resultStr).To(gomega.ContainSubstring("port: 6443"))

		g.Expect(resultStr).To(gomega.ContainSubstring("network:"))
		g.Expect(resultStr).To(gomega.ContainSubstring("provider: calico"))
		g.Expect(resultStr).To(gomega.ContainSubstring("podCIDR: 10.244.0.0/16"))
		g.Expect(resultStr).To(gomega.ContainSubstring("serviceCIDR: 10.96.0.0/12"))

		g.Expect(resultStr).To(gomega.ContainSubstring("storage:"))
		g.Expect(resultStr).To(gomega.ContainSubstring("type: etcd"))
		g.Expect(resultStr).To(gomega.ContainSubstring("peerAddress: 127.0.0.1"))

		g.Expect(resultStr).To(gomega.ContainSubstring("extensions:"))
		g.Expect(resultStr).To(gomega.ContainSubstring("k0rdent"))
	})

	t.Run("only network set should only generate network field", func(t *testing.T) {
		cfg := &config.K0rdentdConfig{
			K0s: config.K0sConfig{
				Network: config.NetworkConfig{
					Provider: "calico",
				},
			},
			K0rdent: config.K0rdentConfig{
				Version: "1.2.2",
				Helm: config.K0rdentHelmConfig{
					Chart:     "k0rdent/kcm",
					Namespace: "kcm-system",
				},
			},
		}

		result, err := GenerateK0sConfig(cfg)
		g.Expect(err).ToNot(gomega.HaveOccurred())

		resultStr := string(result)

		// Should contain network
		g.Expect(resultStr).To(gomega.ContainSubstring("network:"))
		g.Expect(resultStr).To(gomega.ContainSubstring("provider: calico"))

		// Should NOT contain api or storage at spec level
		lines := strings.Split(resultStr, "\n")
		inSpec := false
		inExtensions := false
		foundAPI := false
		foundStorage := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "spec:" {
				inSpec = true
				continue
			}
			if inSpec && !inExtensions {
				if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
					if trimmed == "api:" {
						foundAPI = true
					}
					if trimmed == "storage:" {
						foundStorage = true
					}
					if trimmed == "extensions:" {
						inExtensions = true
					}
				}
			}
		}

		g.Expect(foundAPI).To(gomega.BeFalse(), "api should not be present when empty")
		g.Expect(foundStorage).To(gomega.BeFalse(), "storage should not be present when empty")
	})

	t.Run("storage type without etcd peerAddress should not include etcd section", func(t *testing.T) {
		cfg := &config.K0rdentdConfig{
			K0s: config.K0sConfig{
				Storage: config.StorageConfig{
					Type: "etcd",
					// Etcd.PeerAddress is empty
				},
			},
			K0rdent: config.K0rdentConfig{
				Version: "1.2.2",
				Helm: config.K0rdentHelmConfig{
					Chart:     "k0rdent/kcm",
					Namespace: "kcm-system",
				},
			},
		}

		result, err := GenerateK0sConfig(cfg)
		g.Expect(err).ToNot(gomega.HaveOccurred())

		resultStr := string(result)

		// Should contain storage with type
		g.Expect(resultStr).To(gomega.ContainSubstring("storage:"))
		g.Expect(resultStr).To(gomega.ContainSubstring("type: etcd"))

		// Should NOT contain etcd section (since peerAddress is empty)
		g.Expect(resultStr).ToNot(gomega.ContainSubstring("peerAddress:"))
	})
}
