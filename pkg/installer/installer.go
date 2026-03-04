package installer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/belgaied2/k0rdentd/internal/airgap"
	"github.com/belgaied2/k0rdentd/pkg/config"
	"github.com/belgaied2/k0rdentd/pkg/credentials"
	"github.com/belgaied2/k0rdentd/pkg/k0s"
	"github.com/belgaied2/k0rdentd/pkg/k8sclient"
	"github.com/belgaied2/k0rdentd/pkg/utils"
)

// Installer handles the installation and uninstallation of K0s and K0rdent
type Installer struct {
	debug      bool
	dryRun     bool
	k8sClient  *k8sclient.Client
	config     *config.K0rdentdConfig // Store full config for airgap support
	airgapped  bool
	replaceK0s bool // Replace existing k0s binary without prompting
}

// NewInstaller creates a new installer instance
func NewInstaller(debug, dryRun bool) *Installer {
	return &Installer{
		debug:      debug,
		dryRun:     dryRun,
		airgapped:  false,
		replaceK0s: false,
	}
}

// SetReplaceK0s sets whether to replace existing k0s binary without prompting
func (i *Installer) SetReplaceK0s(replace bool) {
	i.replaceK0s = replace
}

// SetConfig sets the full configuration (needed for airgap support)
func (i *Installer) SetConfig(cfg *config.K0rdentdConfig) {
	i.config = cfg
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
	// Check if running in airgap mode
	if airgap.IsAirGap() {
		return i.installAirgap(k0rdentConfig)
	}

	// Online installation
	if i.dryRun {
		utils.GetLogger().Infof("📝 Dry run mode - showing what would be done:")
		utils.GetLogger().Infof("1. Check k0s version conflicts")
		utils.GetLogger().Infof("2. Write K0s configuration to /etc/k0s/k0s.yaml")
		utils.GetLogger().Infof("3. Execute: k0s install --config /etc/k0s/k0s.yaml")
		utils.GetLogger().Infof("4. Start K0s service")
		utils.GetLogger().Infof("5. Wait for K0rdent to be installed")
		if k0rdentConfig != nil && k0rdentConfig.Credentials.HasCredentials() {
			utils.GetLogger().Infof("6. Create cloud provider credentials")
		}
		return nil
	}

	// Check for k0s version conflicts (online mode only)
	if i.config != nil {
		conflict, err := i.CheckK0sVersionConflict(i.config.K0s.Version)
		if err != nil {
			return fmt.Errorf("failed to check k0s version: %w", err)
		}
		if err := i.HandleVersionConflict(conflict); err != nil {
			return fmt.Errorf("k0s version conflict: %w", err)
		}
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

// InstallJoin installs K0s as a joining node (controller or worker)
func (i *Installer) InstallJoin(joinConfig *config.JoinConfig) error {
	logger := utils.GetLogger()

	if joinConfig == nil || !joinConfig.IsJoin() {
		return fmt.Errorf("invalid join configuration")
	}

	if !joinConfig.IsValid() {
		return fmt.Errorf("join configuration is incomplete or invalid")
	}

	logger.Infof("Joining cluster as %s node...", joinConfig.Mode)
	logger.Infof("Controller server: %s", joinConfig.Server)

	if i.dryRun {
		logger.Info("📝 Dry run mode - join installation steps:")
		logger.Infof("1. Configure containerd mirrors (if airgap)")
		logger.Infof("2. Write K0s configuration")
		logger.Infof("3. Execute: k0s install %s --token-file <token>", joinConfig.Mode)
		logger.Infof("4. Start K0s service")
		logger.Infof("5. Wait for node to be ready")
		return nil
	}

	// Configure containerd mirrors for airgap mode only
	if airgap.IsAirGap() && i.config != nil && i.config.Airgap.Registry.Address != "" {
		logger.Info("Configuring containerd registry mirrors...")
		if err := i.configureContainerdMirrors(i.config.Airgap.Registry.Address); err != nil {
			return fmt.Errorf("failed to configure containerd mirrors: %w", err)
		}
	}

	// Write k0s config for join mode
	if err := i.writeK0sJoinConfig(); err != nil {
		return fmt.Errorf("failed to write k0s config: %w", err)
	}

	// Create token file
	tokenFile := "/etc/k0s/join-token"
	if err := os.WriteFile(tokenFile, []byte(joinConfig.Token), 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}
	defer os.Remove(tokenFile) // Clean up token file after installation

	// Install k0s in join mode
	// Note: k0s token includes server information, no --server flag needed
	var stderrBuf bytes.Buffer
	installArgs := []string{}
	if joinConfig.Mode == "controller" {
		installArgs = []string{
			"install", joinConfig.Mode,
			"--enable-worker",
			"--token-file", tokenFile,
		}
	} else {
		installArgs = []string{
			"install", joinConfig.Mode,
			"--token-file", tokenFile,
		}
	}

	cmd := exec.Command("k0s", installArgs...)
	cmd.Stderr = &stderrBuf

	if i.debug {
		logger.Debugf("🔧 Executing: k0s %v", installArgs)
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("k0s install failed: %w. stderr: %s", err, stderrBuf.String())
	}

	// Start k0s service
	startCmd := exec.Command("k0s", "start")
	var startStderrBuf bytes.Buffer
	startCmd.Stderr = &startStderrBuf

	if i.debug {
		logger.Debug("🔧 Executing: k0s start")
		startCmd.Stdout = os.Stdout
		startCmd.Stderr = io.MultiWriter(os.Stderr, &startStderrBuf)
	}

	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("k0s start failed: %w. stderr: %s", err, startStderrBuf.String())
	}

	// Wait for k0s to be ready
	if err := i.waitForK0sReady(); err != nil {
		return fmt.Errorf("k0s did not become ready: %w", err)
	}

	logger.Infof("✅ Successfully joined cluster as %s", joinConfig.Mode)
	return nil
}

// writeK0sJoinConfig writes a minimal k0s configuration for join mode
func (i *Installer) writeK0sJoinConfig() error {
	configPath := "/etc/k0s/k0s.yaml"

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Minimal k0s config for join mode
	// k0s will get most config from the controller via the join token
	minimalConfig := `apiVersion: k0s.k0sproject.io/v1beta1
kind: ClusterConfig
metadata:
  name: k0s
spec:
  # Configuration will be merged with controller settings
`

	if err := os.WriteFile(configPath, []byte(minimalConfig), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if i.debug {
		utils.GetLogger().Debugf("📄 Wrote K0s join configuration to %s", configPath)
	}

	return nil
}

// configureContainerdMirrors configures containerd to use the specified registry as a mirror
func (i *Installer) configureContainerdMirrors(registryAddress string) error {
	logger := utils.GetLogger()

	// Create containerd config directory
	certsDir := "/etc/k0s/containerd.d/certs.d"
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		return fmt.Errorf("failed to create containerd certs directory: %w", err)
	}

	// List of registries to mirror
	registries := []string{
		"registry.k8s.io",
		"quay.io",
		"ghcr.io",
		"gcr.io",
		"docker.io",
	}

	// Configure mirror for each registry
	for _, registry := range registries {
		registryDir := filepath.Join(certsDir, registry)
		if err := os.MkdirAll(registryDir, 0755); err != nil {
			return fmt.Errorf("failed to create registry directory for %s: %w", registry, err)
		}

		hostsFile := filepath.Join(registryDir, "hosts.toml")
		content := fmt.Sprintf(`server = "https://%s"

[host."http://%s"]
  capabilities = ["pull", "resolve"]
  skip_verify = true
`, registry, registryAddress)

		if err := os.WriteFile(hostsFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write hosts.toml for %s: %w", registry, err)
		}

		logger.Debugf("Configured mirror for %s -> %s", registry, registryAddress)
	}

	logger.Infof("✅ Configured containerd mirrors for %d registries", len(registries))
	return nil
}

// installAirgap performs air-gapped installation
func (i *Installer) installAirgap(k0rdentConfig *config.K0rdentConfig) error {
	logger := utils.GetLogger()
	i.airgapped = true

	metadata := airgap.GetBuildMetadata()
	logger.Infof("Air-gapped installation (K0s: %s)",
		metadata.K0sVersion)

	// Check for version mismatch between config and bundled version (airgap mode)
	if i.config != nil {
		CheckAirgapVersionMismatch(i.config.K0s.Version, metadata.K0sVersion)
	}

	if i.dryRun {
		logger.Infof("📝 Dry run mode - airgap installation steps:")
		logger.Infof("1. Check k0s version mismatch (config vs bundled)")
		logger.Infof("2. Extract k0rdent version from bundle")
		logger.Infof("3. Extract k0s binary from embedded assets")
		logger.Infof("4. Generate k0s configuration for airgap mode")
		logger.Infof("5. Install k0s with embedded binary")
		logger.Infof("6. k0s will automatically install k0rdent from local registry via helm operator")
		if k0rdentConfig != nil && k0rdentConfig.Credentials.HasCredentials() {
			logger.Infof("7. Create cloud provider credentials")
		}
		return nil
	}

	// Ensure we have the full config for airgap
	if i.config == nil {
		return fmt.Errorf("airgap installation requires full configuration, but config is not set")
	}

	// Create airgap installer
	agInstaller := airgap.NewInstaller(i.config, i.debug)

	// Perform airgap-specific preparation (extract k0s, generate config)
	ctx := context.Background()
	if err := agInstaller.Install(ctx); err != nil {
		return fmt.Errorf("airgap preparation failed: %w", err)
	}

	// Now proceed with standard k0s installation
	// The k0s binary is now at /usr/local/bin/k0s
	// The k0s config is at /etc/k0s/k0s.yaml with airgap settings
	logger.Info("")
	logger.Info("Installing k0s...")
	if err := i.installK0s(); err != nil {
		return fmt.Errorf("failed to install k0s: %w", err)
	}

	// Wait for k0rdent to be installed via k0s helm operator
	if err := i.waitForK0rdentInstalled(); err != nil {
		return fmt.Errorf("k0rdent installation failed: %w", err)
	}

	// Wait for CAPI provider Helm releases if credentials are configured
	if k0rdentConfig != nil && k0rdentConfig.Credentials.HasCredentials() {
		if err := i.waitForCAPIProviderHelmReleases(&k0rdentConfig.Credentials); err != nil {
			logger.Warnf("⚠️ CAPI infrastructure providers failed to become ready: %v. Will attempt credential creation anyway.", err)
		}
	}

	// Create credentials if configured
	if k0rdentConfig != nil && k0rdentConfig.Credentials.HasCredentials() {
		if err := i.createCredentials(&k0rdentConfig.Credentials); err != nil {
			logger.Warnf("⚠️ Failed to create credentials: %v. You may need to create them manually through the K0rdent UI.", err)
		}
	}

	logger.Info("✅ Airgap installation completed successfully")
	return nil
}

// createCredentials creates cloud provider credentials in the k0rdent cluster
func (i *Installer) createCredentials(credsConfig *config.CredentialsConfig) error {
	utils.GetLogger().Debugf("Creating cloud provider credentials...")

	credManager := credentials.NewManager(i.k8sClient)
	ctx := context.Background()

	// CreateAll is idempotent - it will skip existing resources
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

			// Check if all required provider Helm releases are deployed
			for _, provider := range providersNeeded {
				releaseName := fmt.Sprintf("cluster-api-provider-%s", provider)
				ready, err := i.k8sClient.IsHelmReleaseReady(ctx, "kcm-system", releaseName)
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
		providers["aws"] = true
	}
	if len(credsConfig.Azure) > 0 {
		providers["azure"] = true
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
	// Check if k0s is already installed and running
	if isK0sInstalled() && isK0sRunning() {
		utils.GetLogger().Info("✅ K0s is already installed and running, skipping installation")

		// Initialize Kubernetes client
		utils.GetLogger().Debug("Initializing Kubernetes client...")
		client, err := k8sclient.NewFromK0s()
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes client: %w", err)
		}
		i.k8sClient = client
		utils.GetLogger().Debug("Kubernetes client initialized successfully")
		return nil
	}

	// Check if k0s is installed but not running
	if isK0sInstalled() && !isK0sRunning() {
		utils.GetLogger().Info("K0s is installed but not running, starting K0s...")
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

	// K0s is not installed, proceed with installation
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

// isK0sInstalled checks if k0s is installed by checking if the config file exists
func isK0sInstalled() bool {

	cmd := exec.Command("systemctl", "is-enabled", "k0scontroller.service")
	_, err := cmd.Output()
	if err != nil {
		return false
	}

	return true
}

// isK0sRunning checks if k0s is running by checking its status
func isK0sRunning() bool {
	return k0s.IsK0sRunning()
}

// isK0sStarted checks if K0s is started by checking its status
// Deprecated: Use isK0sRunning instead
func isK0sStarted() bool {
	return isK0sRunning()
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

	// First check if K0rdent is already ready
	exists, err := i.k8sClient.NamespaceExists(ctx, "kcm-system")
	if err == nil && exists {
		allReady, err := i.areK0rdentDeploymentsReady()
		if err == nil && allReady {
			utils.GetLogger().Info("✅ K0rdent is already installed and ready, skipping wait")
			return nil
		}
	}

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
		"kcm-k0rdent-enterprise-controller-manager", //TODO: Check that it works for k0rdent OSS as well.
		"kcm-k0rdent-ui",
		"kcm-rbac-manager",
	}
	if !i.airgapped {
		requiredDeployments = append(requiredDeployments, "kcm-regional-telemetry")
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

// CheckK0sVersionConflict checks for k0s version conflicts between installed and configured versions
// Returns a VersionConflict if there's a conflict, nil otherwise
func (i *Installer) CheckK0sVersionConflict(configVersion string) (*k0s.VersionConflict, error) {
	logger := utils.GetLogger()

	// Check if k0s binary exists
	k0sCheck, err := k0s.CheckK0s()
	if err != nil {
		return nil, fmt.Errorf("failed to check k0s: %w", err)
	}

	// No conflict if k0s is not installed
	if !k0sCheck.Exists {
		logger.Debug("k0s not installed, no version conflict")
		return nil, nil
	}

	// No conflict if no config version specified
	if configVersion == "" {
		logger.Debug("No k0s version specified in config, using installed version")
		return nil, nil
	}

	// Check if versions match
	equal, err := k0s.VersionsEqual(k0sCheck.Version, configVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to compare k0s versions: %w", err)
	}

	// No conflict if versions match
	if equal {
		logger.Debugf("k0s version %s matches config, no conflict", k0sCheck.Version)
		return nil, nil
	}

	// There's a version conflict
	conflict := &k0s.VersionConflict{
		InstalledVersion: k0sCheck.Version,
		ConfigVersion:    configVersion,
		IsRunning:        k0s.IsK0sRunning(),
	}

	logger.Debugf("k0s version conflict detected: %s", conflict.String())
	return conflict, nil
}

// HandleVersionConflict handles a k0s version conflict
// Returns an error if the conflict cannot be resolved
func (i *Installer) HandleVersionConflict(conflict *k0s.VersionConflict) error {
	logger := utils.GetLogger()

	if conflict == nil {
		return nil
	}

	// If k0s is running with a different version, we cannot proceed automatically
	if conflict.RequiresManualIntervention() {
		logger.Errorf("❌ %s", k0s.FormatRunningConflictError(conflict.InstalledVersion, conflict.ConfigVersion))
		return fmt.Errorf("%s", k0s.FormatRunningConflictError(conflict.InstalledVersion, conflict.ConfigVersion))
	}

	// If we can auto-replace (k0s not running but different version)
	if conflict.CanAutoReplace() {
		if i.replaceK0s {
			logger.Infof("Replacing k0s %s with %s (--replace-k0s flag set)",
				conflict.InstalledVersion, conflict.ConfigVersion)
			return i.replaceK0sBinary(conflict.ConfigVersion)
		}

		// No auto-replace flag, prompt or fail
		logger.Warnf("⚠️  %s", k0s.FormatConflictMessage(conflict.InstalledVersion, conflict.ConfigVersion))
		return fmt.Errorf("k0s version conflict: installed %s, config specifies %s. Use --replace-k0s to replace the existing k0s binary",
			conflict.InstalledVersion, conflict.ConfigVersion)
	}

	return nil
}

// replaceK0sBinary replaces the k0s binary with a specific version
func (i *Installer) replaceK0sBinary(version string) error {
	logger := utils.GetLogger()

	logger.Infof("Downloading and installing k0s version %s...", version)
	if err := k0s.InstallK0sVersion(version); err != nil {
		return fmt.Errorf("failed to replace k0s: %w", err)
	}

	logger.Infof("✅ k0s replaced with version %s", version)
	return nil
}

// CheckAirgapVersionMismatch checks for version mismatch in airgap mode
// Returns true if there's a mismatch (logs a warning), false otherwise
func CheckAirgapVersionMismatch(configVersion, bundledVersion string) bool {
	logger := utils.GetLogger()

	// No mismatch if config version is not specified
	if configVersion == "" {
		return false
	}

	// Check if versions match
	equal, err := k0s.VersionsEqual(configVersion, bundledVersion)
	if err != nil {
		logger.Warnf("Could not compare k0s versions: %v", err)
		return false
	}

	if !equal {
		logger.Warnf("⚠️  %s", k0s.FormatWarningMessage(configVersion, bundledVersion))
		return true
	}

	return false
}
