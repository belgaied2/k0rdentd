package k0s

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/belgaied2/k0rdentd/pkg/config"
	"github.com/belgaied2/k0rdentd/pkg/utils"
)

// CheckResult represents the result of k0s binary check
type CheckResult struct {
	Exists    bool
	Version   string
	Installed bool
}

// CheckK0s checks if k0s binary exists and returns its version
func CheckK0s(cfg *config.K0rdentdConfig) (*CheckResult, error) {
	result := &CheckResult{}

	// Check if k0s binary exists
	path, err := exec.LookPath("k0s")
	if err != nil {
		utils.GetLogger().Debug("k0s binary not found in PATH")
		result.Exists = false
		result.Installed = false
		return result, nil
	}

	result.Exists = true
	result.Installed = true
	utils.GetLogger().Debugf("Found k0s binary at: %s", path)

	// Get k0s version
	version, err := getK0sVersion()
	if err != nil {
		utils.GetLogger().Warnf("Failed to get k0s version: %v", err)
		return result, fmt.Errorf("failed to get k0s version: %w", err)
	}

	result.Version = version
	utils.GetLogger().Infof("Found K0s in version %s", version)

	return result, nil
}

// getK0sVersion executes `k0s version` and parses the output
func getK0sVersion() (string, error) {
	cmd := exec.Command("k0s", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute 'k0s version': %w", err)
	}

	// Parse the version from output
	// Expected format: "k0s version: v1.27.4+k0s.0"
	outputStr := strings.TrimSpace(string(output))
	parts := strings.Split(outputStr, ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected version output format: %s", outputStr)
	}

	version := strings.TrimSpace(parts[1])
	return version, nil
}

// InstallK0s installs k0s binary using the official install script
func InstallK0s() error {
	utils.GetLogger().Info("Installing k0s using official install script")

	// Detect architecture and map to k0s accepted values
	archCmd := exec.Command("uname", "-m")
	archOutput, err := archCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to detect architecture: %w", err)
	}
	arch := strings.TrimSpace(string(archOutput))

	// Map uname -m output to k0s accepted architecture values
	switch arch {
	case "x86_64":
		arch = "amd64"
	case "aarch64":
		arch = "arm64"
	case "armv7l", "armhf":
		arch = "arm"
	}

	// Use the official k0s install script with architecture override
	installCmd := fmt.Sprintf("curl -sSf https://get.k0s.sh | sudo K0S_ARCH=%s sh", arch)
	cmd := exec.Command("bash", "-c", installCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	utils.GetLogger().Debugf("Running: curl -sSf https://get.k0s.sh | sudo K0S_ARCH=%s sh", arch)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install k0s: %w", err)
	}

	utils.GetLogger().Info("k0s installed successfully")

	return nil
}
