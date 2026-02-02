package k8sclient

import (
	"fmt"
	"os/exec"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetK0sKubeconfig retrieves the admin kubeconfig from k0s
func GetK0sKubeconfig() ([]byte, error) {
	cmd := exec.Command("k0s", "kubeconfig", "admin")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("failed to get k0s kubeconfig: %w. stderr: %s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to get k0s kubeconfig: %w", err)
	}
	return output, nil
}

// NewFromK0s creates a new Client by retrieving kubeconfig from k0s
func NewFromK0s() (*Client, error) {
	kubeconfig, err := GetK0sKubeconfig()
	if err != nil {
		return nil, err
	}

	return NewFromKubeconfig(kubeconfig)
}

// NewFromKubeconfig creates a new Client from kubeconfig bytes
func NewFromKubeconfig(kubeconfig []byte) (*Client, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST config from kubeconfig: %w", err)
	}

	return New(config)
}

// NewInClusterClient creates a new Client using in-cluster configuration
// This is useful when running inside a Kubernetes pod
func NewInClusterClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	return New(config)
}
