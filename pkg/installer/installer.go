package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/belgaied2/k0rdentd/pkg/utils"
)

// Installer handles the installation and uninstallation of K0s and K0rdent
type Installer struct {
	debug  bool
	dryRun bool
}

var (
	spinner       = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	tickerSpinner = time.NewTicker(time.Second / 10)
)

// NewInstaller creates a new installer instance
func NewInstaller(debug, dryRun bool) *Installer {
	return &Installer{
		debug:  debug,
		dryRun: dryRun,
	}
}

// Install installs K0s and K0rdent using the generated configuration
func (i *Installer) Install(k0sConfig []byte) error {
	if i.dryRun {
		fmt.Printf("📝 Dry run mode - showing what would be done:")
		fmt.Printf("1. Write K0s configuration to /etc/k0s/k0s.yaml")
		fmt.Printf("2. Execute: k0s install --config /etc/k0s/k0s.yaml")
		fmt.Printf("3. Start K0s service")
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

	return nil
}

// Uninstall uninstalls K0s and K0rdent
func (i *Installer) Uninstall() error {
	if i.dryRun {
		fmt.Printf("📝 Dry run mode - showing what would be done:")
		fmt.Printf("1. Execute: k0s reset")
		fmt.Printf("2. Remove K0s configuration files")
		return nil
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
	cmd := exec.Command("k0s", "install", "controller", "--enable-worker", "--no-taints")

	if i.debug {
		utils.GetLogger().Debug("🔧 Executing: k0s install")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Command \"k0s install\" failed: %w", err)
	}

	startCmd := exec.Command("k0s", "start")
	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("Command \"k0s start\" failed: %w", err)
	}
	err := i.waitForK0sReady()
	if err != nil {
		return fmt.Errorf("Command \"k0s start\" was successful, but k0s never became ready")
	}
	fmt.Println("✅ K0s installed and started successfully")
	return nil
}

// waitForK0sReady waits for k0s to be ready by checking its status
func (i *Installer) waitForK0sReady() error {

	// Set timeout to 5 minutes
	timeout := time.After(5 * time.Minute)
	tickerK0s := time.NewTicker(5 * time.Second)

	stopSpinner := make(chan bool)
	doneCh := make(chan bool)
	go func() {
		defer tickerSpinner.Stop()
		defer close(doneCh)

		tickerCounter := 0
		for {
			select {
			case <-stopSpinner:
				fmt.Println("\r⠿ Waiting for k0s to become ready...")
				return
			case <-tickerSpinner.C:
				// fmt.Printff("\r%s Waiting for k0s to become ready...", spinner[i])
				fmt.Printf("\r%s Waiting for k0s to become ready...", spinner[tickerCounter%len(spinner)])
				tickerCounter++
			}
		}
	}()

	for {
		select {
		case <-timeout:
			stopSpinner <- true
			return fmt.Errorf("\u2718 Timed out waiting for k0s to become ready!")
		case <-tickerK0s.C:
			// Run k0s status command
			cmd := exec.Command("sudo", "k0s", "status")
			output, err := cmd.Output()
			if err != nil {
				// Log the error but continue waiting
				utils.GetLogger().Debugf("k0s status check failed: %v", err)
			}

			// Parse output to check if k0s is ready
			outputStr := string(output)
			if strings.Contains(outputStr, "Kube-api probing successful: true") {
				// fmt.Printf("\n")
				// fmt.Printf("✅ k0s is ready!")
				stopSpinner <- true
				<-doneCh

				return nil
			}
		}
	}
}

// waitForK0rdentInstalled waits for the k0rdent Helm chart to be installed
func (i *Installer) waitForK0rdentInstalled() error {

	// Set timeout to 5 minutes
	timeout := time.After(5 * time.Minute)
	tickerK0rdent := time.NewTicker(10 * time.Second)
	stopSpinner := make(chan bool)
	doneCh := make(chan bool)
	tickerSpinner.Reset(100 * time.Millisecond)
	go func() {
		defer tickerSpinner.Stop()
		defer close(doneCh)

		tickerCounter := 0
		for {
			select {
			case <-stopSpinner:
				fmt.Println("\r⠿ Waiting for k0rdent to become ready...")
				return
			case <-tickerSpinner.C:
				// fmt.Printff("\r%s Waiting for k0s to become ready...", spinner[i])
				fmt.Printf("\r%s Waiting for K0rdent to become ready...", spinner[tickerCounter%len(spinner)])
				tickerCounter++
			}
		}
	}()

	for {
		select {
		case <-timeout:
			stopSpinner <- true
			return fmt.Errorf("timeout waiting for k0rdent Helm chart to be installed")
		case <-tickerK0rdent.C:
			// Check if kcm-system namespace exists
			cmd := exec.Command("sudo", "k0s", "kubectl", "get", "namespaces", "kcm-system", "-o", "jsonpath='{.status.phase}'")
			_, err := cmd.Output()
			if err != nil {
				utils.GetLogger().Debugf("kcm-system namespace check failed: %v", err)
				continue
			}

			// Check if all required pods are running
			podsCmd := exec.Command("sudo", "k0s", "kubectl", "get", "pods", "-n", "kcm-system", "-o", "jsonpath='{.items[*].status.phase}'")
			podsOutput, podsErr := podsCmd.Output()
			if podsErr != nil {
				fmt.Printf("\n")
				utils.GetLogger().Debugf("k0rdent pods check failed: %v", podsErr)
				continue
			}

			podsStatus := string(podsOutput)
			// Check if all pods are Running
			if strings.Contains(podsStatus, "Running") && !strings.Contains(podsStatus, "Pending") && !strings.Contains(podsStatus, "Error") {
				stopSpinner <- true
				<-doneCh
				fmt.Println("✅ k0rdent Helm chart installed successfully!")
				return nil
			}
		}
	}
}

// resetK0s resets K0s installation
func (i *Installer) resetK0s() error {
	cmd := exec.Command("k0s", "reset")

	if i.debug {
		utils.GetLogger().Debug("🔧 Executing: k0s reset")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("k0s reset failed: %w", err)
	}

	if i.debug {
		fmt.Printf("✅ K0s reset successfully")
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
