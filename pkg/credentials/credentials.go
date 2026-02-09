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

// ExistsAll checks if all configured credentials exist in the cluster
func (m *Manager) ExistsAll(ctx context.Context, cfg config.CredentialsConfig) (bool, error) {
	// Check AWS credentials
	for _, cred := range cfg.AWS {
		exists, err := m.awsCredentialExists(ctx, cred)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}

	// Check Azure credentials
	for _, cred := range cfg.Azure {
		exists, err := m.azureCredentialExists(ctx, cred)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}

	// Check OpenStack credentials
	for _, cred := range cfg.OpenStack {
		exists, err := m.openStackCredentialExists(ctx, cred)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}

	return true, nil
}

// awsCredentialExists checks if all components of an AWS credential exist
func (m *Manager) awsCredentialExists(ctx context.Context, cred config.AWSCredential) (bool, error) {
	secretName := fmt.Sprintf("%s-secret", cred.Name)
	identityName := fmt.Sprintf("%s-identity", cred.Name)

	// Check Secret exists
	secretExists, err := m.client.SecretExists(ctx, KCMNamespace, secretName)
	if err != nil {
		return false, fmt.Errorf("failed to check AWS secret: %w", err)
	}
	if !secretExists {
		return false, nil
	}

	// Check AWSClusterStaticIdentity exists
	identityExists, err := m.client.AWSClusterStaticIdentityExists(ctx, KCMNamespace, identityName)
	if err != nil {
		return false, fmt.Errorf("failed to check AWSClusterStaticIdentity: %w", err)
	}
	if !identityExists {
		return false, nil
	}

	// Check Credential exists
	credExists, err := m.client.CredentialExists(ctx, KCMNamespace, cred.Name)
	if err != nil {
		return false, fmt.Errorf("failed to check Credential: %w", err)
	}
	return credExists, nil
}

// azureCredentialExists checks if all components of an Azure credential exist
func (m *Manager) azureCredentialExists(ctx context.Context, cred config.AzureCredential) (bool, error) {
	secretName := fmt.Sprintf("%s-secret", cred.Name)
	identityName := fmt.Sprintf("%s-identity", cred.Name)

	// Check Secret exists
	secretExists, err := m.client.SecretExists(ctx, KCMNamespace, secretName)
	if err != nil {
		return false, fmt.Errorf("failed to check Azure secret: %w", err)
	}
	if !secretExists {
		return false, nil
	}

	// Check AzureClusterIdentity exists
	identityExists, err := m.client.AzureClusterIdentityExists(ctx, KCMNamespace, identityName)
	if err != nil {
		return false, fmt.Errorf("failed to check AzureClusterIdentity: %w", err)
	}
	if !identityExists {
		return false, nil
	}

	// Check Credential exists
	credExists, err := m.client.CredentialExists(ctx, KCMNamespace, cred.Name)
	if err != nil {
		return false, fmt.Errorf("failed to check Credential: %w", err)
	}
	return credExists, nil
}

// openStackCredentialExists checks if all components of an OpenStack credential exist
func (m *Manager) openStackCredentialExists(ctx context.Context, cred config.OpenStackCredential) (bool, error) {
	secretName := fmt.Sprintf("%s-config", cred.Name)

	// Check Secret exists
	secretExists, err := m.client.SecretExists(ctx, KCMNamespace, secretName)
	if err != nil {
		return false, fmt.Errorf("failed to check OpenStack secret: %w", err)
	}
	if !secretExists {
		return false, nil
	}

	// Check Credential exists (no Identity object for OpenStack)
	credExists, err := m.client.CredentialExists(ctx, KCMNamespace, cred.Name)
	if err != nil {
		return false, fmt.Errorf("failed to check Credential: %w", err)
	}
	return credExists, nil
}

// CreateAll creates all configured credentials for all cloud providers
func (m *Manager) CreateAll(ctx context.Context, cfg config.CredentialsConfig) error {
	utils.GetLogger().Debug("Starting credentials creation")

	// Create AWS credentials
	for _, cred := range cfg.AWS {
		// Check if credential already exists
		exists, err := m.awsCredentialExists(ctx, cred)
		if err != nil {
			return fmt.Errorf("failed to check if AWS credential %s exists: %w", cred.Name, err)
		}
		if exists {
			utils.GetLogger().Infof("✓ AWS credential %s already exists, skipping creation", cred.Name)
			continue
		}

		if err := m.createAWSCredentials(ctx, cred); err != nil {
			return fmt.Errorf("failed to create AWS credential %s: %w", cred.Name, err)
		}
	}

	// Create Azure credentials
	for _, cred := range cfg.Azure {
		// Check if credential already exists
		exists, err := m.azureCredentialExists(ctx, cred)
		if err != nil {
			return fmt.Errorf("failed to check if Azure credential %s exists: %w", cred.Name, err)
		}
		if exists {
			utils.GetLogger().Infof("✓ Azure credential %s already exists, skipping creation", cred.Name)
			continue
		}

		if err := m.createAzureCredentials(ctx, cred); err != nil {
			return fmt.Errorf("failed to create Azure credential %s: %w", cred.Name, err)
		}
	}

	// Create OpenStack credentials
	for _, cred := range cfg.OpenStack {
		// Check if credential already exists
		exists, err := m.openStackCredentialExists(ctx, cred)
		if err != nil {
			return fmt.Errorf("failed to check if OpenStack credential %s exists: %w", cred.Name, err)
		}
		if exists {
			utils.GetLogger().Infof("✓ OpenStack credential %s already exists, skipping creation", cred.Name)
			continue
		}

		if err := m.createOpenStackCredentials(ctx, cred); err != nil {
			return fmt.Errorf("failed to create OpenStack credential %s: %w", cred.Name, err)
		}
	}

	utils.GetLogger().Info("✅ All cloud provider credentials created successfully")
	return nil
}

// createAWSCredentials creates AWS credentials (Secret + AWSClusterStaticIdentity + Credential)
func (m *Manager) createAWSCredentials(ctx context.Context, cred config.AWSCredential) error {
	utils.GetLogger().Infof("Creating AWS credential: %s", cred.Name)
	utils.GetLogger().Debugf("AWS credential region: %s", cred.Region)

	secretName := fmt.Sprintf("%s-secret", cred.Name)
	identityName := fmt.Sprintf("%s-identity", cred.Name)

	// Create Secret with AWS credentials
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

	if err := m.client.CreateSecret(ctx, secret); err != nil {
		return fmt.Errorf("failed to create AWS secret: %w", err)
	}
	utils.GetLogger().Infof("✅ Created AWS secret: %s", secretName)

	// Create AWSClusterStaticIdentity
	if err := m.client.CreateAWSClusterStaticIdentity(ctx, identityName, secretName, KCMNamespace); err != nil {
		utils.GetLogger().Warnf("⚠️ Failed to create AWSClusterStaticIdentity %s: %v (Secret was created but Identity creation failed)", identityName, err)
	} else {
		utils.GetLogger().Infof("✅ Created AWSClusterStaticIdentity: %s", identityName)
	}

	// Create k0rdent Credential
	description := fmt.Sprintf("AWS credentials for %s in region %s", cred.Name, cred.Region)
	if err := m.client.CreateCredential(ctx, cred.Name, description, "AWSClusterStaticIdentity", identityName, "infrastructure.cluster.x-k8s.io/v1beta2", KCMNamespace); err != nil {
		utils.GetLogger().Warnf("⚠️ Failed to create k0rdent Credential %s: %v (Identity may not exist or creation failed)", cred.Name, err)
	} else {
		utils.GetLogger().Infof("✅ Created k0rdent Credential: %s", cred.Name)
	}

	return nil
}

// createAzureCredentials creates Azure credentials (Secret + AzureClusterIdentity + Credential)
func (m *Manager) createAzureCredentials(ctx context.Context, cred config.AzureCredential) error {
	utils.GetLogger().Infof("Creating Azure credential: %s", cred.Name)

	secretName := fmt.Sprintf("%s-secret", cred.Name)
	identityName := fmt.Sprintf("%s-identity", cred.Name)

	// Create Secret with Azure client secret
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

	if err := m.client.CreateSecret(ctx, secret); err != nil {
		return fmt.Errorf("failed to create Azure secret: %w", err)
	}
	utils.GetLogger().Infof("✅ Created Azure secret: %s", secretName)

	// Create AzureClusterIdentity
	if err := m.client.CreateAzureClusterIdentity(ctx, identityName, cred.ClientID, cred.TenantID, secretName, KCMNamespace); err != nil {
		return fmt.Errorf("failed to create AzureClusterIdentity: %w", err)
	}
	utils.GetLogger().Infof("✅ Created AzureClusterIdentity: %s", identityName)

	// Create k0rdent Credential
	description := fmt.Sprintf("Azure credentials for %s (subscription: %s)", cred.Name, cred.SubscriptionID)
	if err := m.client.CreateCredential(ctx, cred.Name, description, "AzureClusterIdentity", identityName, "infrastructure.cluster.x-k8s.io/v1beta1", KCMNamespace); err != nil {
		return fmt.Errorf("failed to create k0rdent Credential: %w", err)
	}
	utils.GetLogger().Infof("✅ Created k0rdent Credential: %s", cred.Name)

	return nil
}

// createOpenStackCredentials creates OpenStack credentials (Secret + Credential)
func (m *Manager) createOpenStackCredentials(ctx context.Context, cred config.OpenStackCredential) error {
	utils.GetLogger().Infof("Creating OpenStack credential: %s", cred.Name)

	secretName := fmt.Sprintf("%s-config", cred.Name)

	// Build clouds.yaml content
	cloudsYAML := m.buildOpenStackCloudsYAML(cred)

	// Create Secret with clouds.yaml
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

	if err := m.client.CreateSecret(ctx, secret); err != nil {
		return fmt.Errorf("failed to create OpenStack secret: %w", err)
	}
	utils.GetLogger().Infof("✅ Created OpenStack secret: %s", secretName)

	// Create k0rdent Credential (no Identity object for OpenStack)
	description := fmt.Sprintf("OpenStack credentials for %s (region: %s)", cred.Name, cred.Region)
	if err := m.client.CreateOpenStackCredential(ctx, cred.Name, description, secretName, KCMNamespace); err != nil {
		return fmt.Errorf("failed to create k0rdent Credential: %w", err)
	}
	utils.GetLogger().Infof("✅ Created k0rdent Credential: %s", cred.Name)

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
