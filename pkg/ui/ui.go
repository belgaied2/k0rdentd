package ui

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/belgaied2/k0rdentd/pkg/k8sclient"
	"github.com/belgaied2/k0rdentd/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	// k0rdentUIDeploymentName is the name of the k0rdent UI deployment
	k0rdentUIDeploymentName = "k0rdent-k0rdent-ui"
	// k0rdentUIServiceName is the name of the k0rdent UI service
	k0rdentUIServiceName = "k0rdent-k0rdent-ui"
	// k0rdentUINamespace is the namespace where k0rdent runs
	k0rdentUINamespace = "kcm-system"
	// k0rdentUIIngressName is the name of the ingress we'll create
	k0rdentUIIngressName = "k0rdent-ui"
	// k0rdentUIIngressPath is the path where k0rdent UI will be accessible
	k0rdentUIIngressPath = "/k0rdent-ui"
)

// getK8sClient creates a Kubernetes client from k0s kubeconfig
func getK8sClient() (*k8sclient.Client, error) {
	return k8sclient.NewFromK0s()
}

// DeploymentReady checks if k0rdent UI deployment is ready
func DeploymentReady() (bool, error) {
	ctx := context.Background()
	client, err := getK8sClient()
	if err != nil {
		return false, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	readyReplicas, err := client.GetDeploymentReadyReplicas(ctx, k0rdentUINamespace, k0rdentUIDeploymentName)
	if err != nil {
		return false, fmt.Errorf("failed to check k0rdent UI deployment: %w", err)
	}

	return readyReplicas > 0, nil
}

// GetBasicAuthPassword extracts the Basic Auth password from the k0rdent UI deployment
func GetBasicAuthPassword() (string, error) {
	ctx := context.Background()
	client, err := getK8sClient()
	if err != nil {
		return "", fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return client.GetDeploymentEnvVar(ctx, k0rdentUINamespace, k0rdentUIDeploymentName, "k0rdent-ui", "BASIC_AUTH_PASSWORD")
}

// ServiceExists checks if k0rdent UI service exists
func ServiceExists() (bool, error) {
	ctx := context.Background()
	client, err := getK8sClient()
	if err != nil {
		return false, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return client.ServiceExists(ctx, k0rdentUINamespace, k0rdentUIServiceName)
}

// GetNodePort extracts the NodePort from the k0rdent UI service
func GetNodePort() (int32, error) {
	ctx := context.Background()
	client, err := getK8sClient()
	if err != nil {
		return 0, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return client.GetServiceNodePort(ctx, k0rdentUINamespace, k0rdentUIServiceName)
}

// CreateIngress creates an ingress to expose k0rdent UI service
func CreateIngress(ips []string) error {
	if len(ips) == 0 {
		return fmt.Errorf("no IPs provided for ingress")
	}

	ctx := context.Background()
	client, err := getK8sClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	ingress := buildIngressObject(ips)
	return client.ApplyIngress(ctx, ingress)
}

// buildIngressObject builds a Kubernetes Ingress object
func buildIngressObject(ips []string) *networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k0rdentUIIngressName,
			Namespace: k0rdentUINamespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: ptr.To("nginx"),
			Rules: []networkingv1.IngressRule{
				{
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     k0rdentUIIngressPath,
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: k0rdentUIServiceName,
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// TestUIAccess tests if k0rdent UI is accessible on the given IP
func TestUIAccess(ip string) bool {
	url := fmt.Sprintf("http://%s%s", ip, k0rdentUIIngressPath)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		utils.GetLogger().Debugf("Failed to access k0rdent UI at %s: %v", url, err)
		return false
	}
	defer resp.Body.Close()

	// Check if response looks like HTML
	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/html") {
		utils.GetLogger().Debugf("Successfully accessed k0rdent UI at %s (HTML response)", url)
		return true
	}

	utils.GetLogger().Debugf("k0rdent UI at %s returned non-HTML content type: %s", url, contentType)
	return false
}

// ExposeUI exposes k0rdent UI by creating an ingress and printing access URLs
func ExposeUI() error {
	utils.GetLogger().Info("Checking k0rdent UI deployment status...")

	// Wait for deployment to be ready with timeout (default 5 minutes)
	timeout := 5 * time.Minute
	checkInterval := 5 * time.Second
	startTime := time.Now()

	for {
		ready, err := DeploymentReady()
		if err != nil {
			return fmt.Errorf("failed to check k0rdent UI deployment readiness: %w", err)
		}
		if ready {
			utils.GetLogger().Info("k0rdent UI deployment is ready")
			break
		}

		if time.Since(startTime) > timeout {
			return fmt.Errorf("timeout waiting for k0rdent UI deployment to be ready after %v", timeout)
		}

		utils.GetLogger().Debug("k0rdent UI deployment not ready yet, retrying...")
		time.Sleep(checkInterval)
	}

	// Check if service exists
	exists, err := ServiceExists()
	if err != nil {
		return fmt.Errorf("failed to check k0rdent UI service existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("k0rdent UI service not found")
	}

	utils.GetLogger().Info("k0rdent UI deployment and service are ready")

	// Modify the existing SVC to make it NodePort instead of ClusterIP
	client, err := getK8sClient()
	if err != nil {
		utils.GetLogger().Warnf("Failed to create Kubernetes client: %v", err)
	} else {
		err = client.PatchServiceType(context.Background(), k0rdentUINamespace, k0rdentUIServiceName, corev1.ServiceTypeNodePort)
		if err != nil {
			utils.GetLogger().Warnf("Failed to modify service type to NodePort: %v", err)
		} else {
			utils.GetLogger().Info("Modified k0rdent UI service to NodePort type")
		}
	}

	// Get NodePort for the service
	nodePort, err := GetNodePort()
	if err != nil {
		utils.GetLogger().Warnf("Failed to get NodePort: %v", err)
		nodePort = 0
	} else {
		utils.GetLogger().Infof("Detected NodePort: %d", nodePort)
	}

	// Get external/public IPs
	localIPs, err := GetLocalIPs()
	if err != nil {
		utils.GetLogger().Warnf("Failed to get external IPs: %v", err)
	} else {
		utils.GetLogger().Debugf("Found external IPs: %v", localIPs)
	}

	// Try to get external IP from cloud metadata
	externalIP := GetExternalIP()
	if externalIP != "" {
		utils.GetLogger().Infof("Detected external IP: %s", externalIP)
	}

	// Combine all IPs (external first, then local)
	var allIPs []string
	if externalIP != "" {
		allIPs = append(allIPs, externalIP)
	}
	for _, ip := range localIPs {
		allIPs = append(allIPs, ip.String())
	}

	// Remove duplicates
	uniqueIPs := removeDuplicateIPs(allIPs)

	if len(uniqueIPs) == 0 {
		utils.GetLogger().Info("⚠️  Warning: Could not detect any IP addresses")
		utils.GetLogger().Info("   You can use the following command to port-forward to k0rdent UI:")
		utils.GetLogger().Info("   k0s kubectl port-forward -n kcm-system svc/k0rdent-k0rdent-ui 8080:80")
		utils.GetLogger().Info("   Then access at: http://localhost:8080/k0rdent-ui")
		return nil
	}

	// Create ingress
	if err := CreateIngress(uniqueIPs); err != nil {
		return fmt.Errorf("failed to create ingress: %w", err)
	}

	// Test UI access on primary IP
	primaryIP := uniqueIPs[0]
	if TestUIAccess(primaryIP) {
		utils.GetLogger().Infof("✅ Successfully tested k0rdent UI access on %s\n", primaryIP)
	} else {
		utils.GetLogger().Infof("⚠️  Warning: Could not access k0rdent UI on %s\n", primaryIP)
		utils.GetLogger().Infof("   The ingress has been created, but the UI may not be ready yet\n")
	}

	// Print all possible access URLs
	utils.GetLogger().Info("\n🌐 If an Ingress Controller is installed, K0rdent UI is accessible at:")
	for _, ip := range uniqueIPs {
		url := fmt.Sprintf("http://%s%s", ip, k0rdentUIIngressPath)
		utils.GetLogger().Infof("   %s\n", url)
	}

	// Add NodePort access URLs if available
	if nodePort > 0 {
		utils.GetLogger().Info("\n🔌 NodePort access (requires firewall rules):")
		for _, ip := range uniqueIPs {
			url := fmt.Sprintf("http://%s:%d", ip, nodePort)
			utils.GetLogger().Infof("   %s\n", url)
		}
		utils.GetLogger().Info("\n⚠️  Note: Firewall rules may need to be configured to allow access to the NodePort")
	}

	// Suggest port-forwarding alternative
	utils.GetLogger().Info("\n💡 Alternatively, you can use port-forwarding:")
	utils.GetLogger().Info("   k0s kubectl port-forward -n kcm-system svc/k0rdent-k0rdent-ui 8080:3000")
	utils.GetLogger().Info("   Then access at: http://localhost:8080")

	// Get and display Basic Auth credentials
	password, err := GetBasicAuthPassword()
	if err != nil {
		utils.GetLogger().Warnf("Failed to get Basic Auth password: %v", err)
		utils.GetLogger().Info("\n🔐 Basic Auth credentials: Not available")
	} else {
		// Default username is typically "admin" for k0rdent
		username := "admin"
		utils.GetLogger().Infof("\n🔐 Basic Auth credentials:\n")
		utils.GetLogger().Infof("   Username: %s\n", username)
		utils.GetLogger().Infof("   Password: %s\n", password)
	}

	return nil
}

// removeDuplicateIPs removes duplicate IPs from a slice
func removeDuplicateIPs(ips []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, ip := range ips {
		if !seen[ip] {
			seen[ip] = true
			result = append(result, ip)
		}
	}

	return result
}
