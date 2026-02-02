// Package k8sclient provides a Go-native way to interact with the Kubernetes cluster
// using the official client-go library, replacing kubectl exec calls.
package k8sclient

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client wraps a Kubernetes clientset and provides helper methods
type Client struct {
	clientset kubernetes.Interface
	config    *rest.Config
}

// New creates a new Client from a REST config
func New(config *rest.Config) (*Client, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}

// NewFromClientset creates a new Client from an existing clientset (useful for testing)
func NewFromClientset(clientset kubernetes.Interface) *Client {
	return &Client{
		clientset: clientset,
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
