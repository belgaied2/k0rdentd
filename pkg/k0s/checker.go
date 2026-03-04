package k0s

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/belgaied2/k0rdentd/pkg/utils"
	"github.com/sirupsen/logrus"
)

// IsK0sRunning checks if k0s is currently running as a service
func IsK0sRunning() bool {
	// Run k0s status command
	cmd := exec.Command("k0s", "status")
	output, err := cmd.Output()
	if err != nil {
		utils.GetLogger().Debugf("k0s status check failed: %v", err)
		return false
	}

	// Parse output to check if k0s is ready
	outputStr := string(output)
	return strings.Contains(outputStr, "Kube-api probing successful: true")
}

// CheckResult represents the result of k0s binary check
type CheckResult struct {
	Exists    bool
	Version   string
	Installed bool
}

// CheckK0s checks if k0s binary exists and returns its version
func CheckK0s() (*CheckResult, error) {
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
	version, err := GetK0sVersion()
	if err != nil {
		return result, fmt.Errorf("failed to get k0s version: %w", err)
	}

	result.Version = version
	utils.GetLogger().Infof("🎯 Found K0s in version %s", version)

	return result, nil
}

// GetK0sVersion executes `k0s version` and parses the output
func GetK0sVersion() (string, error) {
	cmd := exec.Command("k0s", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute 'k0s version': %w", err)
	}

	// Parse the version from output
	// Expected format: "v1.27.4+k0s.0"
	return strings.TrimSpace(string(output)), nil
}

// InstallK0s installs k0s binary using the official install script (latest version)
func InstallK0s() error {
	return InstallK0sVersion("")
}

// InstallK0sVersion installs a specific k0s version using the official install script
// If version is empty, installs the latest stable version
func InstallK0sVersion(version string) error {
	// Validate version format if specified
	if version != "" {
		if err := ValidateVersion(version); err != nil {
			return fmt.Errorf("invalid k0s version: %w", err)
		}
		utils.GetLogger().Infof("Installing k0s version %s...", version)
	} else {
		utils.GetLogger().Info("Installing k0s (latest stable version)...")
	}

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

	// Download the install script first
	utils.GetLogger().Debug("Downloading k0s install script...")
	curlCmd := exec.Command("curl", "-sSf", "https://get.k0s.sh")
	scriptContent, err := curlCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to download k0s install script: %w", err)
	}
	utils.GetLogger().Debug("Downloaded k0s install script successfully")

	// Build environment variables for the install script
	envVars := []string{"K0S_ARCH=" + arch}
	if version != "" {
		envVars = append(envVars, "K0S_VERSION="+version)
	}

	// Run the script with sudo
	utils.GetLogger().Debug("Executing k0s install script...")
	cmd := exec.Command("sudo", append(envVars, "sh")...)
	cmd.Stdin = bytes.NewReader(scriptContent)

	// Run with spinner, only show output in debug mode
	var stdoutBuf, stderrBuf bytes.Buffer
	err = runWithSpinner("Installing k0s", cmd, &stdoutBuf, &stderrBuf)
	if err != nil {
		if utils.GetLogger().GetLevel() >= logrus.TraceLevel {
			return fmt.Errorf("failed to install k0s: %w. stdout: %s, stderr: %s", err, stdoutBuf.String(), stderrBuf.String())
		}
		return fmt.Errorf("failed to install k0s: %w", err)
	}

	// Verify installation
	installedVersion, err := GetK0sVersion()
	if err != nil {
		return fmt.Errorf("k0s installed but failed to verify version: %w", err)
	}

	if version != "" {
		// Verify that the installed version matches the requested version
		equal, err := VersionsEqual(installedVersion, version)
		if err != nil {
			utils.GetLogger().Warnf("Could not verify installed version: %v", err)
		} else if !equal {
			return fmt.Errorf("k0s version mismatch: requested %s, installed %s", version, installedVersion)
		}
		utils.GetLogger().Infof("k0s version %s installed successfully", installedVersion)
	} else {
		utils.GetLogger().Infof("k0s version %s installed successfully", installedVersion)
	}

	return nil
}

// runWithSpinner runs a command with a spinner animation
func runWithSpinner(message string, cmd *exec.Cmd, stdoutBuf, stderrBuf *bytes.Buffer) error {
	stopSpinner := make(chan bool)
	doneCh := make(chan bool)

	go utils.RunWithSpinner(message, stopSpinner, doneCh)

	// Capture output - only show in debug mode
	isDebug := utils.GetLogger().GetLevel() >= logrus.DebugLevel
	if isDebug {
		cmd.Stdout = io.MultiWriter(os.Stdout, stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, stderrBuf)
	} else {
		cmd.Stdout = stdoutBuf
		cmd.Stderr = stderrBuf
	}

	err := cmd.Run()
	stopSpinner <- true
	<-doneCh

	return err
}
