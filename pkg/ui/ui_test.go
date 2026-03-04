package ui

import (
	"testing"

	"github.com/onsi/gomega"
	networkingv1 "k8s.io/api/networking/v1"
)

func TestCloudProviderConstants(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("cloud provider constants", func(t *testing.T) {
		g.Expect(string(CloudProviderAWS)).To(gomega.Equal("aws"))
		g.Expect(string(CloudProviderGCP)).To(gomega.Equal("gcp"))
        g.Expect(string(CloudProviderAzure)).To(gomega.Equal("azure"))
        g.Expect(string(CloudProviderNone)).To(gomega.Equal("none"))
    })
}

func TestK0rdentUIConstants(t *testing.T) {
    g := gomega.NewWithT(t)

    t.Run("k0rdent UI constants", func(t *testing.T) {
        g.Expect(k0rdentUIDeploymentName).To(gomega.Equal("kcm-k0rdent-ui"))
        g.Expect(k0rdentUIServiceName).To(gomega.Equal("kcm-k0rdent-ui"))
        g.Expect(k0rdentUINamespace).To(gomega.Equal("kcm-system"))
        g.Expect(k0rdentUIIngressName).To(gomega.Equal("k0rdent-ui"))
        g.Expect(k0rdentUIIngressPath).To(gomega.Equal("/k0rdent-ui"))
    })
}

func TestRemoveDuplicateIPs(t *testing.T) {
    g := gomega.NewWithT(t)

    tests := []struct {
        name     string
        input    []string
        expected []string
    }{
        {
            name:     "single IP",
            input:    []string{"192.168.1.1"},
            expected: []string{"192.168.1.1"},
        },
        {
            name:     "no duplicates",
            input:    []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"},
            expected: []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"},
        },
        {
            name:     "all duplicates",
            input:    []string{"192.168.1.1", "192.168.1.1", "192.168.1.1"},
            expected: []string{"192.168.1.1"},
        },
        {
            name:     "some duplicates",
            input:    []string{"192.168.1.1", "10.0.0.1", "192.168.1.1", "172.16.0.1", "10.0.0.1"},
            expected: []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := removeDuplicateIPs(tt.input)
            g.Expect(result).To(gomega.Equal(tt.expected))
        })
    }

    t.Run("empty slice returns nil or empty", func(t *testing.T) {
        result := removeDuplicateIPs([]string{})
        // Function returns nil for empty input
        g.Expect(result).To(gomega.BeNil())
    })
}

func TestBuildIngressObject(t *testing.T) {
    g := gomega.NewWithT(t)

    tests := []struct {
        name        string
        ips         []string
        expectName  string
        expectNS    string
        expectPath  string
        expectRules int
    }{
        {
            name:        "single IP",
            ips:         []string{"192.168.1.1"},
            expectName:  "k0rdent-ui",
            expectNS:    "kcm-system",
            expectPath:  "/k0rdent-ui",
            expectRules: 1,
        },
        {
            name:        "multiple IPs",
            ips:         []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"},
            expectName:  "k0rdent-ui",
            expectNS:    "kcm-system",
            expectPath:  "/k0rdent-ui",
            expectRules: 1,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := buildIngressObject(tt.ips)

            g.Expect(result).ToNot(gomega.BeNil())
            g.Expect(result.Name).To(gomega.Equal(tt.expectName))
            g.Expect(result.Namespace).To(gomega.Equal(tt.expectNS))
            g.Expect(result.Spec.IngressClassName).ToNot(gomega.BeNil())
            g.Expect(*result.Spec.IngressClassName).To(gomega.Equal("nginx"))
            g.Expect(result.Spec.Rules).To(gomega.HaveLen(tt.expectRules))
            g.Expect(result.Spec.Rules[0].HTTP.Paths[0].Path).To(gomega.Equal(tt.expectPath))
            g.Expect(result.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(gomega.Equal("kcm-k0rdent-ui"))
            g.Expect(result.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number).To(gomega.Equal(int32(80)))

            // Check annotation
            g.Expect(result.Annotations).ToNot(gomega.BeNil())
            g.Expect(result.Annotations["nginx.ingress.kubernetes.io/rewrite-target"]).To(gomega.Equal("/"))
        })
    }
}

func TestBuildIngressObject_PathType(t *testing.T) {
    g := gomega.NewWithT(t)

    result := buildIngressObject([]string{"192.168.1.1"})
    g.Expect(result.Spec.Rules[0].HTTP.Paths[0].PathType).ToNot(gomega.BeNil())
    g.Expect(*result.Spec.Rules[0].HTTP.Paths[0].PathType).To(gomega.Equal(networkingv1.PathTypePrefix))
}

func TestBuildIngressObject_ServiceBackend(t *testing.T) {
    g := gomega.NewWithT(t)

    result := buildIngressObject([]string{"192.168.1.1"})

    // Verify the service backend configuration matches architecture expectations
    // These ingress should point to kcm-k0rdent-ui service on port 80
    backend := result.Spec.Rules[0].HTTP.Paths[0].Backend
    g.Expect(backend.Service).ToNot(gomega.BeNil())
    g.Expect(backend.Service.Name).To(gomega.Equal("kcm-k0rdent-ui"))
    g.Expect(backend.Service.Port.Number).To(gomega.Equal(int32(80)))
}
