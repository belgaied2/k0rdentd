package token

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/onsi/gomega"
)

func TestNewManager(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("creates manager with provided path", func(t *testing.T) {
		manager := NewManager("/usr/local/bin/k0s", true)
		g.Expect(manager).ToNot(gomega.BeNil())
		g.Expect(manager.k0sBinaryPath).To(gomega.Equal("/usr/local/bin/k0s"))
		g.Expect(manager.debug).To(gomega.BeTrue())
	})

	t.Run("creates manager with empty path", func(t *testing.T) {
		manager := NewManager("", false)
		g.Expect(manager).ToNot(gomega.BeNil())
		g.Expect(manager.k0sBinaryPath).To(gomega.Equal(""))
		g.Expect(manager.debug).To(gomega.BeFalse())
	})
}

func TestCreateToken_InvalidRole(t *testing.T) {
	g := gomega.NewWithT(t)

	manager := NewManager("/usr/local/bin/k0s", false)

	tests := []struct {
		name string
		role string
	}{
		{"empty role", ""},
		{"invalid role", "invalid"},
		{"uppercase controller", "CONTROLLER"},
		{"uppercase worker", "WORKER"},
		{"plural controllers", "controllers"},
		{"plural workers", "workers"},
		{"random string", "foobar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.CreateToken(tt.role, 1*time.Hour)
			g.Expect(err).To(gomega.HaveOccurred())
			g.Expect(err.Error()).To(gomega.ContainSubstring("invalid role"))
			g.Expect(token).To(gomega.BeEmpty())
		})
	}
}

func TestCreateControllerToken_CallsCreateToken(t *testing.T) {
	g := gomega.NewWithT(t)

	// Create a mock k0s binary
	tmpDir := t.TempDir()
	mockK0s := filepath.Join(tmpDir, "k0s")
	if runtime.GOOS == "windows" {
		mockK0s += ".exe"
	}

	// Create a simple script that outputs a token
	scriptContent := "#!/bin/sh\necho 'mock-controller-token-123'"
	if runtime.GOOS == "windows" {
		scriptContent = "@echo mock-controller-token-123"
	}

	err := os.WriteFile(mockK0s, []byte(scriptContent), 0755)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	manager := NewManager(mockK0s, false)
	token, err := manager.CreateControllerToken(1 * time.Hour)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(token).To(gomega.Equal("mock-controller-token-123\n"))
}

func TestCreateWorkerToken_CallsCreateToken(t *testing.T) {
	g := gomega.NewWithT(t)

	// Create a mock k0s binary
	tmpDir := t.TempDir()
	mockK0s := filepath.Join(tmpDir, "k0s")
	if runtime.GOOS == "windows" {
		mockK0s += ".exe"
	}

	// Create a simple script that outputs a token
	scriptContent := "#!/bin/sh\necho 'mock-worker-token-456'"
	if runtime.GOOS == "windows" {
		scriptContent = "@echo mock-worker-token-456"
	}

	err := os.WriteFile(mockK0s, []byte(scriptContent), 0755)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	manager := NewManager(mockK0s, false)
	token, err := manager.CreateWorkerToken(30 * time.Minute)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(token).To(gomega.Equal("mock-worker-token-456\n"))
}

func TestCreateToken_MockBinary(t *testing.T) {
	g := gomega.NewWithT(t)

	// Create a mock k0s binary
	tmpDir := t.TempDir()
	mockK0s := filepath.Join(tmpDir, "k0s")

	// Create a script that validates arguments and outputs a token
	scriptContent := `#!/bin/sh
# Validate that correct arguments are passed
if [ "$1" != "token" ] || [ "$2" != "create" ]; then
    echo "ERROR: expected 'token create'" >&2
    exit 1
fi
if [ "$3" != "--role" ]; then
    echo "ERROR: expected '--role'" >&2
    exit 1
fi
if [ "$5" != "--expiry" ]; then
    echo "ERROR: expected '--expiry'" >&2
    exit 1
fi
# Output a mock token based on role
echo "mock-${4}-token-abc123"
`
	err := os.WriteFile(mockK0s, []byte(scriptContent), 0755)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	manager := NewManager(mockK0s, true)

	t.Run("creates controller token with correct args", func(t *testing.T) {
		token, err := manager.CreateToken("controller", 24*time.Hour)
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(token).To(gomega.ContainSubstring("controller"))
	})

	t.Run("creates worker token with correct args", func(t *testing.T) {
		token, err := manager.CreateToken("worker", 1*time.Hour)
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(token).To(gomega.ContainSubstring("worker"))
	})
}

func TestCreateToken_MockBinaryError(t *testing.T) {
	g := gomega.NewWithT(t)

	// Create a mock k0s binary that fails
	tmpDir := t.TempDir()
	mockK0s := filepath.Join(tmpDir, "k0s")

	scriptContent := `#!/bin/sh
echo "Error: k0s cluster is not running" >&2
exit 1
`
	err := os.WriteFile(mockK0s, []byte(scriptContent), 0755)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	manager := NewManager(mockK0s, false)

	token, err := manager.CreateToken("controller", 1*time.Hour)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("failed to create"))
	g.Expect(err.Error()).To(gomega.ContainSubstring("k0s cluster is not running"))
	g.Expect(token).To(gomega.BeEmpty())
}

func TestCreateToken_MockBinaryEmptyOutput(t *testing.T) {
	g := gomega.NewWithT(t)

	// Create a mock k0s binary that outputs nothing
	tmpDir := t.TempDir()
	mockK0s := filepath.Join(tmpDir, "k0s")

	scriptContent := `#!/bin/sh
# Output nothing
exit 0
`
	err := os.WriteFile(mockK0s, []byte(scriptContent), 0755)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	manager := NewManager(mockK0s, false)

	token, err := manager.CreateToken("worker", 1*time.Hour)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("empty token"))
	g.Expect(token).To(gomega.BeEmpty())
}

func TestCreateToken_NonExistentBinary(t *testing.T) {
	g := gomega.NewWithT(t)

	manager := NewManager("/non/existent/path/k0s", false)

	token, err := manager.CreateToken("controller", 1*time.Hour)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(token).To(gomega.BeEmpty())
}
