package config_test

import (
	"testing"

	"github.com/belgaied2/k0rdentd/pkg/config"
	"github.com/onsi/gomega"
)

func TestDefaultConfig(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test DefaultConfig function
	cfg := config.DefaultConfig()

	// Verify K0rdent configuration (default values)
	g.Expect(cfg.K0rdent.Version).To(gomega.Equal("1.2.2"))
	g.Expect(cfg.K0rdent.Helm.Chart).To(gomega.Equal("oci://registry.mirantis.com/k0rdent-enterprise/charts/k0rdent-enterprise"))
	g.Expect(cfg.K0rdent.Helm.Namespace).To(gomega.Equal("kcm-system"))

	// Verify global settings
	g.Expect(cfg.Debug).To(gomega.BeFalse())
}

func TestConfigMarshaling(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test MarshalConfig function
	cfg := config.DefaultConfig()
	data, err := config.MarshalConfig(cfg)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(data).ToNot(gomega.BeEmpty())

	// Basic validation that marshaled data contains expected values
	dataStr := string(data)
	g.Expect(dataStr).To(gomega.ContainSubstring("k0rdent"))
	g.Expect(dataStr).To(gomega.ContainSubstring("kcm-system"))
}

func TestJoinConfigIsJoin(t *testing.T) {
	tests := []struct {
		name     string
		config   config.JoinConfig
		expected bool
	}{
		{
			name:     "empty config",
			config:   config.JoinConfig{},
			expected: false,
		},
		{
			name: "missing mode",
			config: config.JoinConfig{
				Server: "192.168.1.10",
				Token:  "token123",
			},
			expected: false,
		},
		{
			name: "missing server",
			config: config.JoinConfig{
				Mode:  "worker",
				Token: "token123",
			},
			expected: false,
		},
		{
			name: "missing token",
			config: config.JoinConfig{
				Mode:   "worker",
				Server: "192.168.1.10",
			},
			expected: false,
		},
		{
			name: "valid worker config",
			config: config.JoinConfig{
				Mode:   "worker",
				Server: "192.168.1.10",
				Token:  "token123",
			},
			expected: true,
		},
		{
			name: "valid controller config",
			config: config.JoinConfig{
				Mode:   "controller",
				Server: "192.168.1.10",
				Token:  "token123",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			g.Expect(tt.config.IsJoin()).To(gomega.Equal(tt.expected))
		})
	}
}

func TestJoinConfigIsValid(t *testing.T) {
	tests := []struct {
		name     string
		config   config.JoinConfig
		expected bool
	}{
		{
			name:     "empty config",
			config:   config.JoinConfig{},
			expected: false,
		},
		{
			name: "invalid mode",
			config: config.JoinConfig{
				Mode:   "invalid",
				Server: "192.168.1.10",
				Token:  "token123",
			},
			expected: false,
		},
		{
			name: "valid worker config",
			config: config.JoinConfig{
				Mode:   "worker",
				Server: "192.168.1.10",
				Token:  "token123",
			},
			expected: true,
		},
		{
			name: "valid controller config",
			config: config.JoinConfig{
				Mode:   "controller",
				Server: "192.168.1.10",
				Token:  "token123",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			g.Expect(tt.config.IsValid()).To(gomega.Equal(tt.expected))
		})
	}
}