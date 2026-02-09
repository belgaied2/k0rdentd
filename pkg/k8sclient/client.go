// Package k8sclient provides a Go-native way to interact with the Kubernetes cluster
// using the official client-go library, replacing kubectl exec calls.
package k8sclient

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client wraps a Kubernetes clientset and provides helper methods
type Client struct {
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
	config        *rest.Config
}

// HelmRelease represents the simplified structure of the decoded Helm secret
type HelmRelease struct {
	Info struct {
		Status string `json:"status"`
	} `json:"info"`
}

// New creates a new Client from a REST config
func New(config *rest.Config) (*Client, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		config:        config,
	}, nil
}

// NewFromClientset creates a new Client from an existing clientset (useful for testing)
func NewFromClientset(clientset kubernetes.Interface) *Client {
	return &Client{
		clientset: clientset,
	}
}

// NewFromClientsetAndDynamic creates a new Client from existing clientsets (useful for testing)
func NewFromClientsetAndDynamic(clientset kubernetes.Interface, dynamicClient dynamic.Interface) *Client {
	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
	}
}

// Clientset returns the underlying kubernetes clientset
func (c *Client) Clientset() kubernetes.Interface {
	return c.clientset
}

// NamespaceExists checks if a namespace exists
func (c *Client) NamespaceExists(ctx context.Context, name string) (bool, error) {
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get namespace %s: %w", name, err)
	}
	return true, nil
}

// GetDeploymentReadyReplicas returns the number of ready replicas for a deployment
func (c *Client) GetDeploymentReadyReplicas(ctx context.Context, namespace, name string) (int32, error) {
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get deployment %s/%s: %w", namespace, name, err)
	}

	return deployment.Status.ReadyReplicas, nil
}

// GetDeploymentReplicas returns the total number of replicas for a deployment
func (c *Client) GetDeploymentReplicas(ctx context.Context, namespace, name string) (int32, error) {
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get deployment %s/%s: %w", namespace, name, err)
	}

	if deployment.Spec.Replicas == nil {
		return 0, nil
	}
	return *deployment.Spec.Replicas, nil
}

// IsDeploymentReady checks if a deployment is fully ready (ready replicas == total replicas)
func (c *Client) IsDeploymentReady(ctx context.Context, namespace, name string) (bool, error) {
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get deployment %s/%s: %w", namespace, name, err)
	}

	if deployment.Spec.Replicas == nil {
		return false, nil
	}

	return deployment.Status.ReadyReplicas == *deployment.Spec.Replicas, nil
}

// GetPodPhases returns the phases of pods matching the label selector in the given namespace
func (c *Client) GetPodPhases(ctx context.Context, namespace, labelSelector string) ([]corev1.PodPhase, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
	}

	phases := make([]corev1.PodPhase, 0, len(pods.Items))
	for _, pod := range pods.Items {
		phases = append(phases, pod.Status.Phase)
	}

	return phases, nil
}

// IsAnyPodRunning checks if any pod matching the label selector is in Running phase
func (c *Client) IsAnyPodRunning(ctx context.Context, namespace, labelSelector string) (bool, error) {
	phases, err := c.GetPodPhases(ctx, namespace, labelSelector)
	if err != nil {
		return false, err
	}

	for _, phase := range phases {
		if phase == corev1.PodRunning {
			return true, nil
		}
	}

	return false, nil
}

// PatchServiceType patches a service to change its type
func (c *Client) PatchServiceType(ctx context.Context, namespace, name string, svcType corev1.ServiceType) error {
	service, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get service %s/%s: %w", namespace, name, err)
	}

	service.Spec.Type = svcType
	_, err = c.clientset.CoreV1().Services(namespace).Update(ctx, service, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update service %s/%s: %w", namespace, name, err)
	}

	return nil
}

// GetServiceNodePort returns the NodePort for a service
func (c *Client) GetServiceNodePort(ctx context.Context, namespace, name string) (int32, error) {
	service, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get service %s/%s: %w", namespace, name, err)
	}

	if len(service.Spec.Ports) == 0 {
		return 0, fmt.Errorf("service %s/%s has no ports defined", namespace, name)
	}

	return service.Spec.Ports[0].NodePort, nil
}

// ServiceExists checks if a service exists
func (c *Client) ServiceExists(ctx context.Context, namespace, name string) (bool, error) {
	_, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get service %s/%s: %w", namespace, name, err)
	}
	return true, nil
}

// ApplyIngress creates or updates an ingress
func (c *Client) ApplyIngress(ctx context.Context, ingress *networkingv1.Ingress) error {
	existing, err := c.clientset.NetworkingV1().Ingresses(ingress.Namespace).Get(ctx, ingress.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create new ingress
			_, err = c.clientset.NetworkingV1().Ingresses(ingress.Namespace).Create(ctx, ingress, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create ingress %s/%s: %w", ingress.Namespace, ingress.Name, err)
			}
			return nil
		}
		return fmt.Errorf("failed to get ingress %s/%s: %w", ingress.Namespace, ingress.Name, err)
	}

	// Update existing ingress
	ingress.ResourceVersion = existing.ResourceVersion
	_, err = c.clientset.NetworkingV1().Ingresses(ingress.Namespace).Update(ctx, ingress, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ingress %s/%s: %w", ingress.Namespace, ingress.Name, err)
	}

	return nil
}

// GetDeploymentEnvVar extracts an environment variable value from a specific container in a deployment
func (c *Client) GetDeploymentEnvVar(ctx context.Context, namespace, deploymentName, containerName, envVarName string) (string, error) {
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get deployment %s/%s: %w", namespace, deploymentName, err)
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			for _, env := range container.Env {
				if env.Name == envVarName {
					return env.Value, nil
				}
			}
			return "", fmt.Errorf("environment variable %s not found in container %s", envVarName, containerName)
		}
	}

	return "", fmt.Errorf("container %s not found in deployment %s/%s", containerName, namespace, deploymentName)
}

// AreAllDeploymentsReady checks if all deployments in the list are ready
func (c *Client) AreAllDeploymentsReady(ctx context.Context, namespace string, deploymentNames []string) (bool, error) {
	for _, name := range deploymentNames {
		ready, err := c.IsDeploymentReady(ctx, namespace, name)
		if err != nil {
			return false, err
		}
		if !ready {
			return false, nil
		}
	}
	return true, nil
}

// CreateSecret creates a Kubernetes Secret
func (c *Client) CreateSecret(ctx context.Context, secret *corev1.Secret) error {
	_, err := c.clientset.CoreV1().Secrets(secret.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// Secret already exists, update it
			_, err = c.clientset.CoreV1().Secrets(secret.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update secret %s/%s: %w", secret.Namespace, secret.Name, err)
			}
			return nil
		}
		return fmt.Errorf("failed to create secret %s/%s: %w", secret.Namespace, secret.Name, err)
	}
	return nil
}

// CreateAWSClusterStaticIdentity creates an AWSClusterStaticIdentity custom resource
func (c *Client) CreateAWSClusterStaticIdentity(ctx context.Context, name, secretRef, namespace string) error {
	gvr := schema.GroupVersionResource{
		Group:    "infrastructure.cluster.x-k8s.io",
		Version:  "v1beta2",
		Resource: "awsclusterstaticidentities",
	}

	identity := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "infrastructure.cluster.x-k8s.io/v1beta2",
			"kind":       "AWSClusterStaticIdentity",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"k0rdent.mirantis.com/component": "kcm",
				},
			},
			"spec": map[string]interface{}{
				"secretRef": secretRef,
				"allowedNamespaces": map[string]interface{}{
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{},
					},
				},
			},
		},
	}

	_, err := c.dynamicClient.Resource(gvr).Create(ctx, identity, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// Resource already exists, update it
			existing, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get existing AWSClusterStaticIdentity %s/%s: %w", namespace, name, err)
			}
			identity.SetResourceVersion(existing.GetResourceVersion())
			_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, identity, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update AWSClusterStaticIdentity %s/%s: %w", namespace, name, err)
			}
			return nil
		}
		return fmt.Errorf("failed to create AWSClusterStaticIdentity %s/%s: %w", namespace, name, err)
	}
	return nil
}

// CreateAzureClusterIdentity creates an AzureClusterIdentity custom resource
func (c *Client) CreateAzureClusterIdentity(ctx context.Context, name, clientID, tenantID, secretName, namespace string) error {
	gvr := schema.GroupVersionResource{
		Group:    "infrastructure.cluster.x-k8s.io",
		Version:  "v1beta1",
		Resource: "azureclusteridentities",
	}

	identity := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "infrastructure.cluster.x-k8s.io/v1beta1",
			"kind":       "AzureClusterIdentity",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"clusterctl.cluster.x-k8s.io/move-hierarchy": "true",
					"k0rdent.mirantis.com/component":             "kcm",
				},
			},
			"spec": map[string]interface{}{
				"type":              "ServicePrincipal",
				"clientID":          clientID,
				"tenantID":          tenantID,
				"allowedNamespaces": map[string]interface{}{},
				"clientSecret": map[string]interface{}{
					"name":      secretName,
					"namespace": namespace,
				},
			},
		},
	}

	_, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, identity, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// Resource already exists, update it
			existing, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get existing AzureClusterIdentity %s/%s: %w", namespace, name, err)
			}
			identity.SetResourceVersion(existing.GetResourceVersion())
			_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, identity, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update AzureClusterIdentity %s/%s: %w", namespace, name, err)
			}
			return nil
		}
		return fmt.Errorf("failed to create AzureClusterIdentity %s/%s: %w", namespace, name, err)
	}
	return nil
}

// CreateCredential creates a k0rdent Credential custom resource
func (c *Client) CreateCredential(ctx context.Context, name, description, identityKind, identityName, identityAPIVersion, namespace string) error {
	gvr := schema.GroupVersionResource{
		Group:    "k0rdent.mirantis.com",
		Version:  "v1beta1",
		Resource: "credentials",
	}

	credential := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "k0rdent.mirantis.com/v1beta1",
			"kind":       "Credential",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"k0rdent.mirantis.com/component": "kcm",
				},
			},
			"spec": map[string]interface{}{
				"description": description,
				"identityRef": map[string]interface{}{
					"apiVersion": identityAPIVersion,
					"kind":       identityKind,
					"name":       identityName,
					"namespace":  namespace,
				},
			},
		},
	}

	_, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, credential, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// Resource already exists, update it
			existing, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get existing Credential %s/%s: %w", namespace, name, err)
			}
			credential.SetResourceVersion(existing.GetResourceVersion())
			_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, credential, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update Credential %s/%s: %w", namespace, name, err)
			}
			return nil
		}
		return fmt.Errorf("failed to create Credential %s/%s: %w", namespace, name, err)
	}
	return nil
}

// HelmReleaseStatus represents the status of a Helm release
type HelmReleaseStatus string

const (
	HelmReleaseStatusDeployed HelmReleaseStatus = "deployed"
	HelmReleaseStatusFailed   HelmReleaseStatus = "failed"
	HelmReleaseStatusPending  HelmReleaseStatus = "pending"
	HelmReleaseStatusUnknown  HelmReleaseStatus = "unknown"
)

// GetHelmReleaseStatus checks the status of a Helm release in the given namespace
func (c *Client) GetHelmReleaseStatus(ctx context.Context, namespace, releaseName string) (HelmReleaseStatus, error) {
	// Helm stores release information in Secrets in the namespace where it was installed
	// The secret name follows the pattern: sh.helm.release.v1.<releaseName>
	secretName := fmt.Sprintf("sh.helm.release.v1.%s.v1", releaseName)

	secret, err := c.clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return HelmReleaseStatusUnknown, nil
		}
		return HelmReleaseStatusUnknown, fmt.Errorf("failed to get Helm release secret %s/%s: %w", namespace, secretName, err)
	}

	// The release status is stored in the secret's data
	if releaseData, ok := secret.Data["release"]; ok {

		releaseGzipped, err := base64.StdEncoding.DecodeString(string(releaseData))
		if err != nil {
			return HelmReleaseStatusUnknown, fmt.Errorf("issue getting content of release secret: %w", err)
		}

		gzipReader, err := gzip.NewReader(bytes.NewReader(releaseGzipped))
		if err != nil {
			return HelmReleaseStatusUnknown, fmt.Errorf("issue getting content of release secret: %w", err)
		}

		defer gzipReader.Close()

		releaseJson, err := io.ReadAll(gzipReader)
		if err != nil {
			return HelmReleaseStatusUnknown, fmt.Errorf("issue getting content of release secret: %w", err)
		}

		var release HelmRelease
		if err := json.Unmarshal(releaseJson, &release); err != nil {
			return HelmReleaseStatusUnknown, fmt.Errorf("unable to extract status content from release")
		}

		// Parse the status to determine if it's deployed
		statusStr := string(release.Info.Status)
		if strings.Contains(statusStr, "deployed") {
			return HelmReleaseStatusDeployed, nil
		} else if strings.Contains(statusStr, "failed") {
			return HelmReleaseStatusFailed, nil
		} else if strings.Contains(statusStr, "pending") {
			return HelmReleaseStatusPending, nil
		}
	}

	return HelmReleaseStatusUnknown, nil
}

// IsHelmReleaseReady checks if a Helm release is deployed successfully
func (c *Client) IsHelmReleaseReady(ctx context.Context, namespace, releaseName string) (bool, error) {
	status, err := c.GetHelmReleaseStatus(ctx, namespace, releaseName)
	if err != nil {
		return false, err
	}
	return status == HelmReleaseStatusDeployed, nil
}

// SecretExists checks if a secret exists
func (c *Client) SecretExists(ctx context.Context, namespace, name string) (bool, error) {
	_, err := c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get secret %s/%s: %w", namespace, name, err)
	}
	return true, nil
}

// AWSClusterStaticIdentityExists checks if an AWSClusterStaticIdentity exists
func (c *Client) AWSClusterStaticIdentityExists(ctx context.Context, namespace, name string) (bool, error) {
	gvr := schema.GroupVersionResource{
		Group:    "infrastructure.cluster.x-k8s.io",
		Version:  "v1beta2",
		Resource: "awsclusterstaticidentities",
	}
	_, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get AWSClusterStaticIdentity %s/%s: %w", namespace, name, err)
	}
	return true, nil
}

// AzureClusterIdentityExists checks if an AzureClusterIdentity exists
func (c *Client) AzureClusterIdentityExists(ctx context.Context, namespace, name string) (bool, error) {
	gvr := schema.GroupVersionResource{
		Group:    "infrastructure.cluster.x-k8s.io",
		Version:  "v1beta1",
		Resource: "azureclusteridentities",
	}
	_, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get AzureClusterIdentity %s/%s: %w", namespace, name, err)
	}
	return true, nil
}

// CredentialExists checks if a k0rdent Credential exists
func (c *Client) CredentialExists(ctx context.Context, namespace, name string) (bool, error) {
	gvr := schema.GroupVersionResource{
		Group:    "k0rdent.mirantis.com",
		Version:  "v1beta1",
		Resource: "credentials",
	}
	_, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get Credential %s/%s: %w", namespace, name, err)
	}
	return true, nil
}

// CreateOpenStackCredential creates a k0rdent Credential for OpenStack (references Secret directly)
func (c *Client) CreateOpenStackCredential(ctx context.Context, name, description, secretName, namespace string) error {
	gvr := schema.GroupVersionResource{
		Group:    "k0rdent.mirantis.com",
		Version:  "v1beta1",
		Resource: "credentials",
	}

	credential := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "k0rdent.mirantis.com/v1beta1",
			"kind":       "Credential",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
				"labels": map[string]interface{}{
					"k0rdent.mirantis.com/component": "kcm",
				},
			},
			"spec": map[string]interface{}{
				"description": description,
				"identityRef": map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Secret",
					"name":       secretName,
					"namespace":  namespace,
				},
			},
		},
	}

	_, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, credential, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// Resource already exists, update it
			existing, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get existing Credential %s/%s: %w", namespace, name, err)
			}
			credential.SetResourceVersion(existing.GetResourceVersion())
			_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, credential, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update Credential %s/%s: %w", namespace, name, err)
			}
			return nil
		}
		return fmt.Errorf("failed to create Credential %s/%s: %w", namespace, name, err)
	}
	return nil
}
