package k8sclient_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/belgaied2/k0rdentd/pkg/k8sclient"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestNamespaceExists(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should return true when namespace exists", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace",
			},
		}
		_, err := fakeClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		exists, err := client.NamespaceExists(ctx, "test-namespace")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(exists).To(gomega.BeTrue())
	})

	t.Run("should return false when namespace does not exist", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		exists, err := client.NamespaceExists(ctx, "non-existent-namespace")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(exists).To(gomega.BeFalse())
	})
}

func TestGetDeploymentReadyReplicas(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should return the number of ready replicas", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		replicas := int32(3)
		readyReplicas := int32(2)

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
			Status: appsv1.DeploymentStatus{
				ReadyReplicas: readyReplicas,
			},
		}
		_, err := fakeClient.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		count, err := client.GetDeploymentReadyReplicas(ctx, "default", "test-deployment")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(count).To(gomega.Equal(readyReplicas))
	})

	t.Run("should return 0 when deployment does not exist", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		count, err := client.GetDeploymentReadyReplicas(ctx, "default", "non-existent")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(count).To(gomega.Equal(int32(0)))
	})
}

func TestGetDeploymentReplicas(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should return the total number of replicas", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		replicas := int32(5)

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
		}
		_, err := fakeClient.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		count, err := client.GetDeploymentReplicas(ctx, "default", "test-deployment")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(count).To(gomega.Equal(replicas))
	})

	t.Run("should return 0 when deployment does not exist", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		count, err := client.GetDeploymentReplicas(ctx, "default", "non-existent")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(count).To(gomega.Equal(int32(0)))
	})
}

func TestIsDeploymentReady(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should return true when all replicas are ready", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		replicas := int32(3)

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
			Status: appsv1.DeploymentStatus{
				ReadyReplicas: replicas,
			},
		}
		_, err := fakeClient.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		ready, err := client.IsDeploymentReady(ctx, "default", "test-deployment")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(ready).To(gomega.BeTrue())
	})

	t.Run("should return false when not all replicas are ready", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		replicas := int32(3)

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
			Status: appsv1.DeploymentStatus{
				ReadyReplicas: 1,
			},
		}
		_, err := fakeClient.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		ready, err := client.IsDeploymentReady(ctx, "default", "test-deployment")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(ready).To(gomega.BeFalse())
	})

	t.Run("should return false when deployment does not exist", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		ready, err := client.IsDeploymentReady(ctx, "default", "non-existent")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(ready).To(gomega.BeFalse())
	})
}

func TestGetPodPhases(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should return pod phases for matching pods", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		pod1 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-1",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		}
		pod2 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-2",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
			},
		}

		_, err := fakeClient.CoreV1().Pods("default").Create(ctx, pod1, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())
		_, err = fakeClient.CoreV1().Pods("default").Create(ctx, pod2, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		phases, err := client.GetPodPhases(ctx, "default", "app=test")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(phases).To(gomega.HaveLen(2))
		g.Expect(phases).To(gomega.ContainElements(corev1.PodRunning, corev1.PodPending))
	})

	t.Run("should return empty slice when no pods match", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		phases, err := client.GetPodPhases(ctx, "default", "app=non-existent")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(phases).To(gomega.BeEmpty())
	})
}

func TestIsAnyPodRunning(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should return true when at least one pod is running", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "running-pod",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		}
		_, err := fakeClient.CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		running, err := client.IsAnyPodRunning(ctx, "default", "app=test")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(running).To(gomega.BeTrue())
	})

	t.Run("should return false when no pods are running", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pending-pod",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
			},
		}
		_, err := fakeClient.CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		running, err := client.IsAnyPodRunning(ctx, "default", "app=test")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(running).To(gomega.BeFalse())
	})
}

func TestServiceExists(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should return true when service exists", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
		}
		_, err := fakeClient.CoreV1().Services("default").Create(ctx, svc, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		exists, err := client.ServiceExists(ctx, "default", "test-service")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(exists).To(gomega.BeTrue())
	})

	t.Run("should return false when service does not exist", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		exists, err := client.ServiceExists(ctx, "default", "non-existent")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(exists).To(gomega.BeFalse())
	})
}

func TestGetServiceNodePort(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should return the NodePort when service has one", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{
					{
						Port:     80,
						NodePort: 30080,
					},
				},
			},
		}
		_, err := fakeClient.CoreV1().Services("default").Create(ctx, svc, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		nodePort, err := client.GetServiceNodePort(ctx, "default", "test-service")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(nodePort).To(gomega.Equal(int32(30080)))
	})

	t.Run("should return error when service has no ports", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{},
			},
		}
		_, err := fakeClient.CoreV1().Services("default").Create(ctx, svc, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		_, err = client.GetServiceNodePort(ctx, "default", "test-service")
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestPatchServiceType(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should change the service type", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
			},
		}
		_, err := fakeClient.CoreV1().Services("default").Create(ctx, svc, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		err = client.PatchServiceType(ctx, "default", "test-service", corev1.ServiceTypeNodePort)
		g.Expect(err).NotTo(gomega.HaveOccurred())

		updatedSvc, err := fakeClient.CoreV1().Services("default").Get(ctx, "test-service", metav1.GetOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(updatedSvc.Spec.Type).To(gomega.Equal(corev1.ServiceTypeNodePort))
	})

	t.Run("should return error when service does not exist", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		err := client.PatchServiceType(ctx, "default", "non-existent", corev1.ServiceTypeNodePort)
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestApplyIngress(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should create a new ingress when it does not exist", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ingress",
				Namespace: "default",
			},
		}

		err := client.ApplyIngress(ctx, ingress)
		g.Expect(err).NotTo(gomega.HaveOccurred())

		createdIngress, err := fakeClient.NetworkingV1().Ingresses("default").Get(ctx, "test-ingress", metav1.GetOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(createdIngress.Name).To(gomega.Equal("test-ingress"))
	})

	t.Run("should update an existing ingress", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		existingIngress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ingress",
				Namespace: "default",
			},
		}
		_, err := fakeClient.NetworkingV1().Ingresses("default").Create(ctx, existingIngress, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		updatedIngress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ingress",
				Namespace: "default",
				Annotations: map[string]string{
					"new-annotation": "value",
				},
			},
		}

		err = client.ApplyIngress(ctx, updatedIngress)
		g.Expect(err).NotTo(gomega.HaveOccurred())

		ingress, err := fakeClient.NetworkingV1().Ingresses("default").Get(ctx, "test-ingress", metav1.GetOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(ingress.Annotations["new-annotation"]).To(gomega.Equal("value"))
	})
}

func TestGetDeploymentEnvVar(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should return the environment variable value", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		replicas := int32(1)
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "test",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "test",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-container",
								Image: "test-image",
								Env: []corev1.EnvVar{
									{
										Name:  "TEST_VAR",
										Value: "test-value",
									},
								},
							},
						},
					},
				},
			},
		}
		_, err := fakeClient.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		value, err := client.GetDeploymentEnvVar(ctx, "default", "test-deployment", "test-container", "TEST_VAR")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(value).To(gomega.Equal("test-value"))
	})

	t.Run("should return error when environment variable does not exist", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		replicas := int32(1)
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "test",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "test",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-container",
								Image: "test-image",
							},
						},
					},
				},
			},
		}
		_, err := fakeClient.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		_, err = client.GetDeploymentEnvVar(ctx, "default", "test-deployment", "test-container", "NON_EXISTENT_VAR")
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("should return error when container does not exist", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		replicas := int32(1)
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "test",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "test",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-container",
								Image: "test-image",
							},
						},
					},
				},
			},
		}
		_, err := fakeClient.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		_, err = client.GetDeploymentEnvVar(ctx, "default", "test-deployment", "non-existent-container", "TEST_VAR")
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestAreAllDeploymentsReady(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should return true when all deployments are ready", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		replicas := int32(1)

		for _, name := range []string{"deployment-1", "deployment-2"} {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
				},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas: replicas,
				},
			}
			_, err := fakeClient.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{})
			g.Expect(err).NotTo(gomega.HaveOccurred())
		}

		allReady, err := client.AreAllDeploymentsReady(ctx, "default", []string{"deployment-1", "deployment-2"})
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(allReady).To(gomega.BeTrue())
	})

	t.Run("should return false when any deployment is not ready", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		replicas := int32(2)

		// First deployment is ready
		deployment1 := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deployment-1",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
			Status: appsv1.DeploymentStatus{
				ReadyReplicas: replicas,
			},
		}
		_, err := fakeClient.AppsV1().Deployments("default").Create(ctx, deployment1, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		// Second deployment is not ready
		deployment2 := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deployment-2",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
			Status: appsv1.DeploymentStatus{
				ReadyReplicas: 0,
			},
		}
		_, err = fakeClient.AppsV1().Deployments("default").Create(ctx, deployment2, metav1.CreateOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		allReady, err := client.AreAllDeploymentsReady(ctx, "default", []string{"deployment-1", "deployment-2"})
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(allReady).To(gomega.BeFalse())
	})
}

func TestErrorHandling(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("should handle API errors gracefully", func(t *testing.T) {
		ctx := context.Background()
		fakeClient := fake.NewSimpleClientset()
		client := k8sclient.NewFromClientset(fakeClient)

		// Add a reactor that returns an error
		fakeClient.PrependReactor("get", "namespaces", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, errors.NewInternalError(fmt.Errorf("simulated API error"))
		})

		_, err := client.NamespaceExists(ctx, "test-namespace")
		g.Expect(err).To(gomega.HaveOccurred())
	})
}
