package network

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
