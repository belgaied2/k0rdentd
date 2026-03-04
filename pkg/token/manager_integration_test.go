//go:build integration

package token

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
)

// TestCreateToken_RealK0s tests token creation with a real k0s binary.
// This test requires a running k0s cluster.
// Run with: go test -tags=integration ./pkg/token/...
func TestCreateToken_RealK0s(t *testing.T) {
	g := gomega.NewWithT(t)

	manager := NewManager("k0s", false)

	token, err := manager.CreateToken("controller", 1*time.Hour)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(token).ToNot(gomega.BeEmpty())
	 // Token should look like a base64-encoded string
	g.Expect(token).To(gomega.HaveLen(gomega.BeNumericallyGreaterThanOrEqualTo(50)))
}

// TestCreateControllerToken_RealK0s tests controller token creation with a real k0s binary.
// This test requires a running k0s cluster.
// Run with: go test -tags=integration ./pkg/token/...
func TestCreateControllerToken_RealK0s(t *testing.T) {
	g := gomega.NewWithT(t)

	manager := NewManager("k0s", false)

	token, err := manager.CreateControllerToken(30 * time.Minute)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(token).ToNot(gomega.BeEmpty())
}

// TestCreateWorkerToken_RealK0s tests worker token creation with a real k0s binary.
// This test requires a running k0s cluster.
// Run with: go test -tags=integration ./pkg/token/...
func TestCreateWorkerToken_RealK0s(t *testing.T) {
	g := gomega.NewWithT(t)

	manager := NewManager("k0s", false)

	token, err := manager.CreateWorkerToken(30 * time.Minute)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(token).ToNot(gomega.BeEmpty())
}
