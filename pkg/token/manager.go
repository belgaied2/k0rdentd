package token

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/belgaied2/k0rdentd/pkg/utils"
)

// Manager handles k0s token operations
type Manager struct {
	k0sBinaryPath string
	debug         bool
}

// NewManager creates a new token manager
func NewManager(k0sBinaryPath string, debug bool) *Manager {
	return &Manager{
		k0sBinaryPath: k0sBinaryPath,
		debug:         debug,
	}
}

// CreateToken creates a join token for the specified role
func (m *Manager) CreateToken(role string, expiry time.Duration) (string, error) {
	logger := utils.GetLogger()

	if role != "controller" && role != "worker" {
		return "", fmt.Errorf("invalid role '%s': must be 'controller' or 'worker'", role)
	}

	logger.Debugf("Creating %s join token with expiry %s", role, expiry)

	args := []string{
		"token", "create",
		"--role", role,
		"--expiry", expiry.String(),
	}

	cmd := exec.Command(m.k0sBinaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if m.debug {
		logger.Debugf("Executing: %s %v", m.k0sBinaryPath, args)
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create %s token: %w. stderr: %s", role, err, stderr.String())
	}

	token := stdout.String()
	if token == "" {
		return "", fmt.Errorf("k0s token create returned empty token")
	}

	logger.Debugf("Successfully created %s token", role)
	return token, nil
}

// CreateControllerToken creates a join token for controller nodes
func (m *Manager) CreateControllerToken(expiry time.Duration) (string, error) {
	return m.CreateToken("controller", expiry)
}

// CreateWorkerToken creates a join token for worker nodes
func (m *Manager) CreateWorkerToken(expiry time.Duration) (string, error) {
	return m.CreateToken("worker", expiry)
}
