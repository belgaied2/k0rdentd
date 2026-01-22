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

	// Verify K0s configuration
	g.Expect(cfg.K0s.Version).To(gomega.Equal("v1.27.4+k0s.0"))
	g.Expect(cfg.K0s.API.Address).To(gomega.Equal("0.0.0.0"))
	g.Expect(cfg.K0s.API.Port).To(gomega.Equal(6443))
	g.Expect(cfg.K0s.Network.Provider).To(gomega.Equal("calico"))
	g.Expect(cfg.K0s.Storage.Type).To(gomega.Equal("etcd"))

	// Verify K0rdent configuration
	g.Expect(cfg.K0rdent.Version).To(gomega.Equal("v0.1.0"))
	g.Expect(cfg.K0rdent.Helm.Chart).To(gomega.Equal("k0rdent/k0rdent"))
	g.Expect(cfg.K0rdent.Helm.Namespace).To(gomega.Equal("k0rdent-system"))

	// Verify global settings
	g.Expect(cfg.Debug).To(gomega.BeFalse())
	g.Expect(cfg.LogLevel).To(gomega.Equal("info"))
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
	g.Expect(dataStr).To(gomega.ContainSubstring("v1.27.4+k0s.0"))
	g.Expect(dataStr).To(gomega.ContainSubstring("k0rdent/k0rdent"))
}