package credentials

import (
	"context"
	"testing"

	"github.com/belgaied2/k0rdentd/pkg/config"
	"github.com/belgaied2/k0rdentd/pkg/k8sclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

// MockK8sClient is a mock implementation of the k8sclient for testing
type MockK8sClient struct {
	mock.Mock
}

func (m *MockK8sClient) CreateSecret(ctx context.Context, secret *corev1.Secret) error {
	args := m.Called(ctx, secret)
	return args.Error(0)
}

func (m *MockK8sClient) CreateAWSClusterStaticIdentity(ctx context.Context, name, secretRef, namespace string) error {
	args := m.Called(ctx, name, secretRef, namespace)
	return args.Error(0)
}

func (m *MockK8sClient) CreateAzureClusterIdentity(ctx context.Context, name, clientID, tenantID, secretName, namespace string) error {
	args := m.Called(ctx, name, clientID, tenantID, secretName, namespace)
	return args.Error(0)
}

func (m *MockK8sClient) CreateCredential(ctx context.Context, name, description, identityKind, identityName, identityAPIVersion, namespace string) error {
	args := m.Called(ctx, name, description, identityKind, identityName, identityAPIVersion, namespace)
	return args.Error(0)
}

func (m *MockK8sClient) CreateOpenStackCredential(ctx context.Context, name, description, secretName, namespace string) error {
	args := m.Called(ctx, name, description, secretName, namespace)
	return args.Error(0)
}

func TestNewManager(t *testing.T) {
	client := &k8sclient.Client{}
	manager := NewManager(client)
	assert.NotNil(t, manager)
	assert.Equal(t, client, manager.client)
}

func TestCreateAWSCredentials(t *testing.T) {
	ctx := context.Background()

	// Create fake clients
	fakeClient := fake.NewSimpleClientset()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

	client := k8sclient.NewFromClientsetAndDynamic(fakeClient, fakeDynamicClient)
	manager := NewManager(client)

	awsCred := config.AWSCredential{
		Name:            "test-aws-cred",
		Region:          "us-west-2",
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	err := manager.createAWSCredentials(ctx, awsCred)
	assert.NoError(t, err)

	// Verify secret was created
	secret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "test-aws-cred-secret", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	// When using StringData, it's stored in Data after creation
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", secret.StringData["AccessKeyID"])
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", secret.StringData["SecretAccessKey"])
}

func TestCreateAWSCredentialsWithSessionToken(t *testing.T) {
	ctx := context.Background()

	fakeClient := fake.NewSimpleClientset()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

	client := k8sclient.NewFromClientsetAndDynamic(fakeClient, fakeDynamicClient)
	manager := NewManager(client)

	awsCred := config.AWSCredential{
		Name:            "test-aws-cred",
		Region:          "us-west-2",
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		SessionToken:    "FwoGZXIvYXdzEBYaDG...",
	}

	err := manager.createAWSCredentials(ctx, awsCred)
	assert.NoError(t, err)

	// Verify secret was created with session token
	secret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "test-aws-cred-secret", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, "FwoGZXIvYXdzEBYaDG...", secret.StringData["SessionToken"])
}

func TestCreateAzureCredentials(t *testing.T) {
	ctx := context.Background()

	fakeClient := fake.NewSimpleClientset()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

	client := k8sclient.NewFromClientsetAndDynamic(fakeClient, fakeDynamicClient)
	manager := NewManager(client)

	azureCred := config.AzureCredential{
		Name:           "test-azure-cred",
		SubscriptionID: "12345678-1234-1234-1234-123456789012",
		ClientID:       "87654321-4321-4321-4321-210987654321",
		ClientSecret:   "my-client-secret",
		TenantID:       "11111111-1111-1111-1111-111111111111",
	}

	err := manager.createAzureCredentials(ctx, azureCred)
	assert.NoError(t, err)

	// Verify secret was created
	secret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "test-azure-cred-secret", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, "my-client-secret", secret.StringData["clientSecret"])
}

func TestCreateOpenStackCredentialsWithAppCreds(t *testing.T) {
	ctx := context.Background()

	fakeClient := fake.NewSimpleClientset()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

	client := k8sclient.NewFromClientsetAndDynamic(fakeClient, fakeDynamicClient)
	manager := NewManager(client)

	openstackCred := config.OpenStackCredential{
		Name:                        "test-openstack-cred",
		AuthURL:                     "https://openstack.example.com:5000/v3",
		Region:                      "RegionOne",
		ApplicationCredentialID:     "app-cred-id",
		ApplicationCredentialSecret: "app-cred-secret",
	}

	err := manager.createOpenStackCredentials(ctx, openstackCred)
	assert.NoError(t, err)

	// Verify secret was created
	secret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "test-openstack-cred-config", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)

	// Verify clouds.yaml content
	cloudsYAML := secret.StringData["clouds.yaml"]
	assert.Contains(t, cloudsYAML, "auth_url: https://openstack.example.com:5000/v3")
	assert.Contains(t, cloudsYAML, "application_credential_id: app-cred-id")
	assert.Contains(t, cloudsYAML, "application_credential_secret: app-cred-secret")
	assert.Contains(t, cloudsYAML, "region_name: RegionOne")
}

func TestCreateOpenStackCredentialsWithPassword(t *testing.T) {
	ctx := context.Background()

	fakeClient := fake.NewSimpleClientset()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

	client := k8sclient.NewFromClientsetAndDynamic(fakeClient, fakeDynamicClient)
	manager := NewManager(client)

	openstackCred := config.OpenStackCredential{
		Name:        "test-openstack-cred",
		AuthURL:     "https://openstack.example.com:5000/v3",
		Region:      "RegionOne",
		Username:    "admin",
		Password:    "secretpassword",
		ProjectName: "my-project",
		DomainName:  "Default",
	}

	err := manager.createOpenStackCredentials(ctx, openstackCred)
	assert.NoError(t, err)

	// Verify secret was created
	secret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "test-openstack-cred-config", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)

	// Verify clouds.yaml content
	cloudsYAML := secret.StringData["clouds.yaml"]
	assert.Contains(t, cloudsYAML, "username: admin")
	assert.Contains(t, cloudsYAML, "password: secretpassword")
	assert.Contains(t, cloudsYAML, "project_name: my-project")
	assert.Contains(t, cloudsYAML, "domain_name: Default")
}

func TestCreateAll(t *testing.T) {
	ctx := context.Background()

	fakeClient := fake.NewSimpleClientset()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

	client := k8sclient.NewFromClientsetAndDynamic(fakeClient, fakeDynamicClient)
	manager := NewManager(client)

	cfg := config.CredentialsConfig{
		AWS: []config.AWSCredential{
			{
				Name:            "aws-cred-1",
				Region:          "us-west-2",
				AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
				SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
		},
		Azure: []config.AzureCredential{
			{
				Name:           "azure-cred-1",
				SubscriptionID: "12345678-1234-1234-1234-123456789012",
				ClientID:       "87654321-4321-4321-4321-210987654321",
				ClientSecret:   "my-client-secret",
				TenantID:       "11111111-1111-1111-1111-111111111111",
			},
		},
		OpenStack: []config.OpenStackCredential{
			{
				Name:                        "openstack-cred-1",
				AuthURL:                     "https://openstack.example.com:5000/v3",
				Region:                      "RegionOne",
				ApplicationCredentialID:     "app-cred-id",
				ApplicationCredentialSecret: "app-cred-secret",
			},
		},
	}

	err := manager.CreateAll(ctx, cfg)
	assert.NoError(t, err)

	// Verify AWS secret was created
	awsSecret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "aws-cred-1-secret", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, awsSecret)

	// Verify Azure secret was created
	azureSecret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "azure-cred-1-secret", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, azureSecret)

	// Verify OpenStack secret was created
	openstackSecret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "openstack-cred-1-config", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, openstackSecret)
}

func TestCreateAllWithEmptyConfig(t *testing.T) {
	ctx := context.Background()

	fakeClient := fake.NewSimpleClientset()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

	client := k8sclient.NewFromClientsetAndDynamic(fakeClient, fakeDynamicClient)
	manager := NewManager(client)

	cfg := config.CredentialsConfig{}

	err := manager.CreateAll(ctx, cfg)
	assert.NoError(t, err)

	// Verify no secrets were created
	secrets, err := fakeClient.CoreV1().Secrets(KCMNamespace).List(ctx, metav1.ListOptions{})
	assert.NoError(t, err)
	assert.Empty(t, secrets.Items)
}

func TestHasCredentials(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.CredentialsConfig
		expected bool
	}{
		{
			name:     "empty config",
			cfg:      config.CredentialsConfig{},
			expected: false,
		},
		{
			name: "with AWS credentials",
			cfg: config.CredentialsConfig{
				AWS: []config.AWSCredential{{Name: "test"}},
			},
			expected: true,
		},
		{
			name: "with Azure credentials",
			cfg: config.CredentialsConfig{
				Azure: []config.AzureCredential{{Name: "test"}},
			},
			expected: true,
		},
		{
			name: "with OpenStack credentials",
			cfg: config.CredentialsConfig{
				OpenStack: []config.OpenStackCredential{{Name: "test"}},
			},
			expected: true,
		},
		{
			name: "with multiple credentials",
			cfg: config.CredentialsConfig{
				AWS:       []config.AWSCredential{{Name: "test"}},
				Azure:     []config.AzureCredential{{Name: "test"}},
				OpenStack: []config.OpenStackCredential{{Name: "test"}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.HasCredentials()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCreateAWSCredentialsIdempotent tests that AWS credential creation is idempotent
func TestCreateAWSCredentialsIdempotent(t *testing.T) {
	ctx := context.Background()

	fakeClient := fake.NewSimpleClientset()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

	client := k8sclient.NewFromClientsetAndDynamic(fakeClient, fakeDynamicClient)
	manager := NewManager(client)

	awsCred := config.AWSCredential{
		Name:            "test-aws-cred",
		Region:          "us-west-2",
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	// First creation - should create all resources
	err := manager.createAWSCredentials(ctx, awsCred)
	assert.NoError(t, err)

	// Verify secret was created
	secret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "test-aws-cred-secret", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)

	// Second creation - should skip creation (resources already exist)
	// The function should not return an error and should not recreate resources
	err = manager.createAWSCredentials(ctx, awsCred)
	assert.NoError(t, err)

	// Verify the secret still exists and wasn't replaced
	secret, err = fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "test-aws-cred-secret", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)
}

// TestCreateAzureCredentialsIdempotent tests that Azure credential creation is idempotent
func TestCreateAzureCredentialsIdempotent(t *testing.T) {
	ctx := context.Background()

	fakeClient := fake.NewSimpleClientset()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

	client := k8sclient.NewFromClientsetAndDynamic(fakeClient, fakeDynamicClient)
	manager := NewManager(client)

	azureCred := config.AzureCredential{
		Name:           "test-azure-cred",
		SubscriptionID: "12345678-1234-1234-1234-123456789012",
		ClientID:       "87654321-4321-4321-4321-210987654321",
		ClientSecret:   "my-client-secret",
		TenantID:       "11111111-1111-1111-1111-111111111111",
	}

	// First creation - should create all resources
	err := manager.createAzureCredentials(ctx, azureCred)
	assert.NoError(t, err)

	// Verify secret was created
	secret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "test-azure-cred-secret", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)

	// Second creation - should skip creation (resources already exist)
	err = manager.createAzureCredentials(ctx, azureCred)
	assert.NoError(t, err)

	// Verify the secret still exists and wasn't replaced
	secret, err = fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "test-azure-cred-secret", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)
}

// TestCreateOpenStackCredentialsIdempotent tests that OpenStack credential creation is idempotent
func TestCreateOpenStackCredentialsIdempotent(t *testing.T) {
	ctx := context.Background()

	fakeClient := fake.NewSimpleClientset()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

	client := k8sclient.NewFromClientsetAndDynamic(fakeClient, fakeDynamicClient)
	manager := NewManager(client)

	openstackCred := config.OpenStackCredential{
		Name:                        "test-openstack-cred",
		AuthURL:                     "https://openstack.example.com:5000/v3",
		Region:                      "RegionOne",
		ApplicationCredentialID:     "app-cred-id",
		ApplicationCredentialSecret: "app-cred-secret",
	}

	// First creation - should create all resources
	err := manager.createOpenStackCredentials(ctx, openstackCred)
	assert.NoError(t, err)

	// Verify secret was created
	secret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "test-openstack-cred-config", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)

	// Second creation - should skip creation (resources already exist)
	err = manager.createOpenStackCredentials(ctx, openstackCred)
	assert.NoError(t, err)

	// Verify the secret still exists and wasn't replaced
	secret, err = fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "test-openstack-cred-config", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)
}

// TestCreateAllIdempotent tests that CreateAll is idempotent
func TestCreateAllIdempotent(t *testing.T) {
	ctx := context.Background()

	fakeClient := fake.NewSimpleClientset()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

	client := k8sclient.NewFromClientsetAndDynamic(fakeClient, fakeDynamicClient)
	manager := NewManager(client)

	cfg := config.CredentialsConfig{
		AWS: []config.AWSCredential{
			{
				Name:            "aws-cred-1",
				Region:          "us-west-2",
				AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
				SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
		},
	}

	// First creation
	err := manager.CreateAll(ctx, cfg)
	assert.NoError(t, err)

	// Verify secret was created
	awsSecret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "aws-cred-1-secret", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, awsSecret)

	// Second creation - should not fail or recreate
	err = manager.CreateAll(ctx, cfg)
	assert.NoError(t, err)

	// Verify the secret still exists
	awsSecret, err = fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "aws-cred-1-secret", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, awsSecret)
}

// TestCreateAWSCredentialsPartialFailure tests "best effort" error handling for AWS credentials
func TestCreateAWSCredentialsPartialFailure(t *testing.T) {
	ctx := context.Background()

	fakeClient := fake.NewSimpleClientset()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

	client := k8sclient.NewFromClientsetAndDynamic(fakeClient, fakeDynamicClient)
	manager := NewManager(client)

	awsCred := config.AWSCredential{
		Name:            "test-aws-cred-partial",
		Region:          "us-west-2",
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	// Create the secret first (to simulate partial state)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-aws-cred-partial-secret",
			Namespace: KCMNamespace,
		},
		StringData: map[string]string{
			"AccessKeyID":     "AKIAIOSFODNN7EXAMPLE",
			"SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}
	_, err := fakeClient.CoreV1().Secrets(KCMNamespace).Create(ctx, secret, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Now try to create the full credential set - should skip the secret and try to create the rest
	err = manager.createAWSCredentials(ctx, awsCred)
	// Should succeed since Secret exists (best effort for other resources)
	assert.NoError(t, err)

	// Verify the original secret still exists
	existingSecret, err := fakeClient.CoreV1().Secrets(KCMNamespace).Get(ctx, "test-aws-cred-partial-secret", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, existingSecret)
}

