package installer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/belgaied2/k0rdentd/pkg/config"
	"github.com/belgaied2/k0rdentd/pkg/credentials"
	"github.com/belgaied2/k0rdentd/pkg/k8sclient"
	"github.com/belgaied2/k0rdentd/pkg/utils"
)

// Installer handles the installation and uninstallation of K0s and K0rdent
type Installer struct {
	debug     bool
	dryRun    bool
	k8sClient *k8sclient.Client
}

// NewInstaller creates a new installer instance
func NewInstaller(debug, dryRun bool) *Installer {
	return &Installer{
		debug:  debug,
		dryRun: dryRun,
	}
}

// waitForWithSpinner waits for a condition with a spinner animation
func (i *Installer) waitForWithSpinner(
	timeout time.Duration,
	message string,
	checkFunc func() (bool, error),
) error {
	tickerInterval := 100 * time.Millisecond
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	stopSpinner := make(chan bool)
	doneCh := make(chan bool)
	timeoutChan := time.After(timeout)

	go utils.RunWithSpinner(message, stopSpinner, doneCh)

	for {
		select {
		case <-timeoutChan:
			stopSpinner <- true
			<-doneCh
			return fmt.Errorf("timeout waiting for %s", message)
		default:
			ready, err := checkFunc()
			if err != nil {
				utils.GetLogger().Debugf("Check for %s failed: %v", message, err)
				continue
			}
			if ready {
				stopSpinner <- true
				<-doneCh
				return nil
			}
		}
	}
}

// Install installs K0s and K0rdent using the generated configuration
func (i *Installer) Install(k0sConfig []byte, k0rdentConfig *config.K0rdentConfig) error {
	if i.dryRun {
		utils.GetLogger().Infof("📝 Dry run mode - showing what would be done:")
		utils.GetLogger().Infof("1. Write K0s configuration to /etc/k0s/k0s.yaml")
		utils.GetLogger().Infof("2. Execute: k0s install --config /etc/k0s/k0s.yaml")
		utils.GetLogger().Infof("3. Start K0s service")
		utils.GetLogger().Infof("4. Wait for K0rdent to be installed")
		if k0rdentConfig != nil && k0rdentConfig.Credentials.HasCredentials() {
			utils.GetLogger().Infof("5. Create cloud provider credentials")
		}
		return nil
	}

	// Write K0s configuration
	if err := i.writeK0sConfig(k0sConfig); err != nil {
		return fmt.Errorf("failed to write K0s config: %w", err)
	}

	// Install K0s
	if err := i.installK0s(); err != nil {
		return fmt.Errorf("failed to install K0s: %w", err)
	}

	// Wait for k0rdent Helm chart to be installed
	if err := i.waitForK0rdentInstalled(); err != nil {
		return fmt.Errorf("k0rdent Helm chart failed to install: %w", err)
	}

	// NEW: Wait for CAPI provider Helm releases if credentials are configured
	if k0rdentConfig != nil && k0rdentConfig.Credentials.HasCredentials() {
		if err := i.waitForCAPIProviderHelmReleases(&k0rdentConfig.Credentials); err != nil {
			// Log warning but continue - credential creation will also warn
			utils.GetLogger().Warnf("⚠️ CAPI infrastructure providers failed to become ready: %v. Will attempt credential creation anyway.", err)
		}
	}

	// Create credentials if configured - log warnings on failure but don't fail the installation
	if k0rdentConfig != nil && k0rdentConfig.Credentials.HasCredentials() {
		if err := i.createCredentials(&k0rdentConfig.Credentials); err != nil {
			// Log warning but continue - user can still access K0rdent UI
			utils.GetLogger().Warnf("⚠️ Failed to create credentials: %v. You may need to create them manually through the K0rdent UI.", err)
		}
	}

	return nil
}

// createCredentials creates cloud provider credentials in the k0rdent cluster
func (i *Installer) createCredentials(credsConfig *config.CredentialsConfig) error {
	utils.GetLogger().Info("Creating cloud provider credentials...")

	credManager := credentials.NewManager(i.k8sClient)
	ctx := context.Background()

	if err := credManager.CreateAll(ctx, *credsConfig); err != nil {
		return err
	}

	utils.GetLogger().Info("✅ Cloud provider credentials created successfully")
	return nil
}

// waitForCAPIProviderHelmReleases waits for required CAPI infrastructure provider Helm releases to be deployed
func (i *Installer) waitForCAPIProviderHelmReleases(credsConfig *config.CredentialsConfig) error {
	ctx := context.Background()

	// Determine which providers are needed
	providersNeeded := i.getRequiredProviders(credsConfig)

	if len(providersNeeded) == 0 {
		utils.GetLogger().Debug("No CAPI providers needed")
		return nil
	}

	return i.waitForWithSpinner(
		15*time.Minute,
		"Waiting for CAPI infrastructure providers to be deployed",
		func() (bool, error) {
			// Check if capi-system namespace exists
			exists, err := i.k8sClient.NamespaceExists(ctx, "capi-system")
			if err != nil {
				utils.GetLogger().Debugf("capi-system namespace check failed: %v", err)
				return false, nil
			}
			if !exists {
				utils.GetLogger().Debug("capi-system namespace does not exist yet")
				return false, nil
			}

			// Check if all required provider Helm releases are deployed
			for _, provider := range providersNeeded {
				releaseName := fmt.Sprintf("%s-provider", provider)
				ready, err := i.k8sClient.IsHelmReleaseReady(ctx, "capi-system", releaseName)
				if err != nil {
					utils.GetLogger().Debugf("Provider %s Helm release check failed: %v", provider, err)
					return false, nil
				}
				if !ready {
					utils.GetLogger().Debugf("Provider %s Helm release is not ready yet", provider)
					return false, nil
				}
				utils.GetLogger().Debugf("Provider %s Helm release is deployed", provider)
			}

			utils.GetLogger().Info("All required CAPI infrastructure provider Helm releases are deployed")
			return true, nil
		},
	)
}

// getRequiredProviders returns a list of provider types that are needed based on credentials config
func (i *Installer) getRequiredProviders(credsConfig *config.CredentialsConfig) []string {
	providers := make(map[string]bool)

	if len(credsConfig.AWS) > 0 {
		providers["capa"] = true
	}
	if len(credsConfig.Azure) > 0 {
		providers["capz"] = true
	}
	if len(credsConfig.OpenStack) > 0 {
		// OpenStack doesn't need a provider for credentials, but we might want to wait for it
		// providers["capo"] = true
	}

	result := make([]string, 0, len(providers))
	for provider := range providers {
		result = append(result, provider)
	}

	return result
}

// Uninstall uninstalls K0s and K0rdent
func (i *Installer) Uninstall() error {
	if i.dryRun {
		utils.GetLogger().Infof("📝 Dry run mode - showing what would be done:")
		utils.GetLogger().Infof("1. Execute: k0s reset")
		utils.GetLogger().Infof("2. Remove K0s configuration files")
		return nil
	}

	// Stop K0s if it started
	if isK0sStarted() {
		if err := i.stopK0s(); err != nil {
			return fmt.Errorf("failed to stop K0s: %w", err)
		}
	}

	// Reset K0s
	if err := i.resetK0s(); err != nil {
		return fmt.Errorf("failed to reset K0s: %w", err)
	}

	// Remove configuration files
	if err := i.cleanupConfig(); err != nil {
		return fmt.Errorf("failed to cleanup config: %w", err)
	}

	return nil
}

// writeK0sConfig writes K0s configuration to file
func (i *Installer) writeK0sConfig(config []byte) error {
	configPath := "/etc/k0s/k0s.yaml"

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write configuration file
	if err := os.WriteFile(configPath, config, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if i.debug {
		utils.GetLogger().Debugf("📄 Wrote K0s configuration to %s", configPath)
	}

	return nil
}

// installK0s installs K0s using the generated configuration
func (i *Installer) installK0s() error {
	var stderrBuf bytes.Buffer
	cmd := exec.Command("k0s", "install", "controller", "--enable-worker", "--no-taints")
	cmd.Stderr = &stderrBuf

	if i.debug {
		utils.GetLogger().Debug("🔧 Executing: k0s install")
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Command \"k0s install\" failed: %w. stderr: %s", err, stderrBuf.String())
	}

	startCmd := exec.Command("k0s", "start")
	var startStderrBuf bytes.Buffer
	startCmd.Stderr = &startStderrBuf

	if i.debug {
		utils.GetLogger().Debug("🔧 Executing: k0s start")
		startCmd.Stdout = os.Stdout
		startCmd.Stderr = io.MultiWriter(os.Stderr, &startStderrBuf)
	}

	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("Command \"k0s start\" failed: %w. stderr: %s", err, startStderrBuf.String())
	}
	err := i.waitForK0sReady()
	if err != nil {
		return fmt.Errorf("Command \"k0s start\" was successful, but k0s never became ready")
	}

	// Initialize Kubernetes client after k0s is ready
	utils.GetLogger().Debug("Initializing Kubernetes client...")
	client, err := k8sclient.NewFromK0s()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	i.k8sClient = client
	utils.GetLogger().Debug("Kubernetes client initialized successfully")

	return nil
}

// isK0sStarted checks if K0s is started by checking its status
func isK0sStarted() bool {
	// Run k0s status command
	cmd := exec.Command("k0s", "status")
	output, err := cmd.Output()
	if err != nil {
		// Log the error but return false
		utils.GetLogger().Debugf("k0s status check failed: %v", err)
		return false
	}

	// Parse output to check if k0s is ready
	outputStr := string(output)
	return strings.Contains(outputStr, "Kube-api probing successful: true")
}

// stopK0s stops the K0s service
func (i *Installer) stopK0s() error {
	var stderrBuf bytes.Buffer
	cmd := exec.Command("k0s", "stop")
	cmd.Stderr = &stderrBuf

	if i.debug {
		utils.GetLogger().Debug("🔧 Executing: k0s stop")
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("k0s stop failed: %w. stderr: %s", err, stderrBuf.String())
	}

	if i.debug {
		utils.GetLogger().Infof("✅ K0s stopped successfully")
	}

	return nil
}

// waitForK0sReady waits for k0s to be ready by checking its status
func (i *Installer) waitForK0sReady() error {
	return i.waitForWithSpinner(
		5*time.Minute,
		"Waiting for k0s to become ready",
		func() (bool, error) {
			return isK0sStarted(), nil
		},
	)
}

// waitForK0rdentInstalled waits for the k0rdent Helm chart to be installed
func (i *Installer) waitForK0rdentInstalled() error {
	ctx := context.Background()
	return i.waitForWithSpinner(
		15*time.Minute,
		"Waiting for K0rdent to become ready",
		func() (bool, error) {
			// Check if kcm-system namespace exists
			exists, err := i.k8sClient.NamespaceExists(ctx, "kcm-system")
			if err != nil {
				utils.GetLogger().Debugf("kcm-system namespace check failed: %v", err)
				return false, nil
			}
			if !exists {
				utils.GetLogger().Debug("kcm-system namespace does not exist yet")
				return false, nil
			}

			// Check if all required deployments are ready
			allReady, err := i.areK0rdentDeploymentsReady()
			if err != nil {
				utils.GetLogger().Warnf("deployment readiness check failed: %v", err)
				return false, nil
			}

			if allReady {
				utils.GetLogger().Debug("All K0rdent deployments are ready")
			}

			return allReady, nil
		},
	)
}

// areK0rdentDeploymentsReady checks if all required K0rdent deployments are ready
func (i *Installer) areK0rdentDeploymentsReady() (bool, error) {
	ctx := context.Background()
	requiredDeployments := []string{
		"kcm-cert-manager",
		"kcm-cert-manager-cainjector",
		"kcm-cert-manager-webhook",
		"kcm-datasource-controller-manager",
		"kcm-k0rdent-enterprise-controller-manager",
		"kcm-k0rdent-ui",
		"kcm-rbac-manager",
		"kcm-regional-telemetry",
	}

	return i.k8sClient.AreAllDeploymentsReady(ctx, "kcm-system", requiredDeployments)
}

// resetK0s resets K0s installation
func (i *Installer) resetK0s() error {
	var stderrBuf bytes.Buffer
	cmd := exec.Command("k0s", "reset")
	cmd.Stderr = &stderrBuf

	if i.debug {
		utils.GetLogger().Debug("🔧 Executing: k0s reset")
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("k0s reset failed: %v. stderr: %s", err, stderrBuf.String())
	}

	if i.debug {
		utils.GetLogger().Infof("✅ K0s reset successfully")
	}

	return nil
}

// cleanupConfig removes configuration files
func (i *Installer) cleanupConfig() error {
	configFiles := []string{
		"/etc/k0s/k0s.yaml",
		"/etc/k0rdentd/k0rdentd.yaml",
	}

	for _, file := range configFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", file, err)
		}
		if i.debug {
			utils.GetLogger().Debugf("🗑️  Removed %s", file)
		}
	}

	return nil
}
