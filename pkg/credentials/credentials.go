// Package credentials provides functionality for creating cloud provider credentials
// in Kubernetes for use with k0rdent.
package credentials

import (
	"context"
	"fmt"

	"github.com/belgaied2/k0rdentd/pkg/config"
	"github.com/belgaied2/k0rdentd/pkg/k8sclient"
	"github.com/belgaied2/k0rdentd/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// KCMNamespace is the namespace where k0rdent credentials are created
	KCMNamespace = "kcm-system"
	// KCMComponentLabel is the label used to identify k0rdent components
	KCMComponentLabel = "k0rdent.mirantis.com/component"
	// KCMComponentValue is the value for the k0rdent component label
	KCMComponentValue = "kcm"
)

// Manager handles the creation of cloud provider credentials
type Manager struct {
	client *k8sclient.Client
}

// NewManager creates a new credentials manager
func NewManager(client *k8sclient.Client) *Manager {
	return &Manager{
		client: client,
	}
}

// CreateAll creates all configured credentials for all cloud providers.
// This function is idempotent - it can be run multiple times and will only
// create resources that don't already exist.
func (m *Manager) CreateAll(ctx context.Context, cfg config.CredentialsConfig) error {
	utils.GetLogger().Debug("Starting credentials creation")

	// Create AWS credentials
	for _, cred := range cfg.AWS {
		if err := m.createAWSCredentials(ctx, cred); err != nil {
			return fmt.Errorf("failed to create AWS credential %s: %w", cred.Name, err)
		}
	}

	// Create Azure credentials
	for _, cred := range cfg.Azure {
		if err := m.createAzureCredentials(ctx, cred); err != nil {
			return fmt.Errorf("failed to create Azure credential %s: %w", cred.Name, err)
		}
	}

	// Create OpenStack credentials
	for _, cred := range cfg.OpenStack {
		if err := m.createOpenStackCredentials(ctx, cred); err != nil {
			return fmt.Errorf("failed to create OpenStack credential %s: %w", cred.Name, err)
		}
	}

	utils.GetLogger().Info("✅ All cloud provider credentials created successfully")
	return nil
}

// createAWSCredentials creates AWS credentials (Secret + AWSClusterStaticIdentity + Credential)
// This function is idempotent - each resource is only created if it doesn't already exist.
func (m *Manager) createAWSCredentials(ctx context.Context, cred config.AWSCredential) error {
	utils.GetLogger().Debugf("Creating AWS credential: %s", cred.Name)
	utils.GetLogger().Debugf("AWS credential region: %s", cred.Region)

	secretName := fmt.Sprintf("%s-secret", cred.Name)
	identityName := fmt.Sprintf("%s-identity", cred.Name)

	// Prepare Secret with AWS credentials
	secretData := map[string]string{
		"AccessKeyID":     cred.AccessKeyID,
		"SecretAccessKey": cred.SecretAccessKey,
	}
	if cred.SessionToken != "" {
		secretData["SessionToken"] = cred.SessionToken
		utils.GetLogger().Debug("Including SessionToken in AWS secret (MFA/SSO enabled)")
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: KCMNamespace,
			Labels: map[string]string{
				KCMComponentLabel: KCMComponentValue,
			},
		},
		StringData: secretData,
	}

	// Create Secret (if not exists)
	if err := m.createIfNotExists(
		ctx,
		ResourceSpec{
			Type:      ResourceTypeSecret,
			Namespace: KCMNamespace,
			Name:      secretName,
		},
		m.client.SecretExists,
		func(ctx context.Context) error {
			return m.client.CreateSecret(ctx, secret)
		},
	); err != nil {
		utils.GetLogger().Warnf("⚠️ Failed to create AWS Secret: %v (continuing with best effort)", err)
		return fmt.Errorf("failed to create AWS secret: %w", err)
	}

	// Create AWSClusterStaticIdentity (if not exists)
	if err := m.createIfNotExists(
		ctx,
		ResourceSpec{
			Type:      ResourceTypeAWSClusterStaticIdentity,
			Namespace: KCMNamespace,
			Name:      identityName,
		},
		m.client.AWSClusterStaticIdentityExists,
		func(ctx context.Context) error {
			return m.client.CreateAWSClusterStaticIdentity(ctx, identityName, secretName, KCMNamespace)
		},
	); err != nil {
		utils.GetLogger().Warnf("⚠️ Failed to create AWSClusterStaticIdentity %s: %v (continuing with best effort)", identityName, err)
	}

	// Create k0rdent Credential (if not exists)
	description := fmt.Sprintf("AWS credentials for %s in region %s", cred.Name, cred.Region)
	if err := m.createIfNotExists(
		ctx,
		ResourceSpec{
			Type:      ResourceTypeCredential,
			Namespace: KCMNamespace,
			Name:      cred.Name,
		},
		m.client.CredentialExists,
		func(ctx context.Context) error {
			return m.client.CreateCredential(ctx, cred.Name, description, "AWSClusterStaticIdentity", identityName, "infrastructure.cluster.x-k8s.io/v1beta2", KCMNamespace)
		},
	); err != nil {
		utils.GetLogger().Warnf("⚠️ Failed to create k0rdent Credential %s: %v (continuing with best effort)", cred.Name, err)
	}

	return nil
}

// createAzureCredentials creates Azure credentials (Secret + AzureClusterIdentity + Credential)
// This function is idempotent - each resource is only created if it doesn't already exist.
func (m *Manager) createAzureCredentials(ctx context.Context, cred config.AzureCredential) error {
	utils.GetLogger().Debugf("Creating Azure credential: %s", cred.Name)

	secretName := fmt.Sprintf("%s-secret", cred.Name)
	identityName := fmt.Sprintf("%s-identity", cred.Name)

	// Prepare Secret with Azure client secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: KCMNamespace,
			Labels: map[string]string{
				KCMComponentLabel: KCMComponentValue,
			},
		},
		StringData: map[string]string{
			"clientSecret": cred.ClientSecret,
		},
	}

	// Create Secret (if not exists)
	if err := m.createIfNotExists(
		ctx,
		ResourceSpec{
			Type:      ResourceTypeSecret,
			Namespace: KCMNamespace,
			Name:      secretName,
		},
		m.client.SecretExists,
		func(ctx context.Context) error {
			return m.client.CreateSecret(ctx, secret)
		},
	); err != nil {
		utils.GetLogger().Warnf("⚠️ Failed to create Azure Secret: %v (continuing with best effort)", err)
		return fmt.Errorf("failed to create Azure secret: %w", err)
	}

	// Create AzureClusterIdentity (if not exists)
	if err := m.createIfNotExists(
		ctx,
		ResourceSpec{
			Type:      ResourceTypeAzureClusterIdentity,
			Namespace: KCMNamespace,
			Name:      identityName,
		},
		m.client.AzureClusterIdentityExists,
		func(ctx context.Context) error {
			return m.client.CreateAzureClusterIdentity(ctx, identityName, cred.ClientID, cred.TenantID, secretName, KCMNamespace)
		},
	); err != nil {
		utils.GetLogger().Warnf("⚠️ Failed to create AzureClusterIdentity %s: %v (continuing with best effort)", identityName, err)
	}

	// Create k0rdent Credential (if not exists)
	description := fmt.Sprintf("Azure credentials for %s (subscription: %s)", cred.Name, cred.SubscriptionID)
	if err := m.createIfNotExists(
		ctx,
		ResourceSpec{
			Type:      ResourceTypeCredential,
			Namespace: KCMNamespace,
			Name:      cred.Name,
		},
		m.client.CredentialExists,
		func(ctx context.Context) error {
			return m.client.CreateCredential(ctx, cred.Name, description, "AzureClusterIdentity", identityName, "infrastructure.cluster.x-k8s.io/v1beta1", KCMNamespace)
		},
	); err != nil {
		utils.GetLogger().Warnf("⚠️ Failed to create k0rdent Credential %s: %v (continuing with best effort)", cred.Name, err)
	}

	return nil
}

// createOpenStackCredentials creates OpenStack credentials (Secret + Credential)
// This function is idempotent - each resource is only created if it doesn't already exist.
func (m *Manager) createOpenStackCredentials(ctx context.Context, cred config.OpenStackCredential) error {
	utils.GetLogger().Debugf("Creating OpenStack credential: %s", cred.Name)

	secretName := fmt.Sprintf("%s-config", cred.Name)

	// Build clouds.yaml content
	cloudsYAML := m.buildOpenStackCloudsYAML(cred)

	// Prepare Secret with clouds.yaml
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: KCMNamespace,
			Labels: map[string]string{
				KCMComponentLabel: KCMComponentValue,
			},
		},
		StringData: map[string]string{
			"clouds.yaml": cloudsYAML,
		},
	}

	// Create Secret (if not exists)
	if err := m.createIfNotExists(
		ctx,
		ResourceSpec{
			Type:      ResourceTypeSecret,
			Namespace: KCMNamespace,
			Name:      secretName,
		},
		m.client.SecretExists,
		func(ctx context.Context) error {
			return m.client.CreateSecret(ctx, secret)
		},
	); err != nil {
		utils.GetLogger().Warnf("⚠️ Failed to create OpenStack Secret: %v (continuing with best effort)", err)
		return fmt.Errorf("failed to create OpenStack secret: %w", err)
	}

	// Create k0rdent Credential (if not exists, no Identity object for OpenStack)
	description := fmt.Sprintf("OpenStack credentials for %s (region: %s)", cred.Name, cred.Region)
	if err := m.createIfNotExists(
		ctx,
		ResourceSpec{
			Type:      ResourceTypeCredential,
			Namespace: KCMNamespace,
			Name:      cred.Name,
		},
		m.client.CredentialExists,
		func(ctx context.Context) error {
			return m.client.CreateOpenStackCredential(ctx, cred.Name, description, secretName, KCMNamespace)
		},
	); err != nil {
		utils.GetLogger().Warnf("⚠️ Failed to create k0rdent Credential %s: %v (continuing with best effort)", cred.Name, err)
	}

	return nil
}

// buildOpenStackCloudsYAML generates the clouds.yaml content for OpenStack credentials
func (m *Manager) buildOpenStackCloudsYAML(cred config.OpenStackCredential) string {
	var authSection string

	if cred.ApplicationCredentialID != "" && cred.ApplicationCredentialSecret != "" {
		// Use application credentials
		authSection = fmt.Sprintf(`    auth:
      auth_url: %s
      application_credential_id: %s
      application_credential_secret: %s
    auth_type: v3applicationcredential`, cred.AuthURL, cred.ApplicationCredentialID, cred.ApplicationCredentialSecret)
	} else if cred.Username != "" && cred.Password != "" {
		// Use username/password authentication
		authSection = fmt.Sprintf(`    auth:
      auth_url: %s
      username: %s
      password: %s
      project_name: %s
      domain_name: %s`, cred.AuthURL, cred.Username, cred.Password, cred.ProjectName, cred.DomainName)
	}

	return fmt.Sprintf(`clouds:
  openstack:
%s
    region_name: %s
    interface: public
    identity_api_version: 3`, authSection, cred.Region)
}
