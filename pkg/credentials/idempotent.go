// Package credentials provides idempotent resource creation utilities
package credentials

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/belgaied2/k0rdentd/pkg/utils"
)

// ResourceType identifies the type of Kubernetes resource being created
type ResourceType string

const (
	ResourceTypeSecret                   ResourceType = "Secret"
	ResourceTypeAWSClusterStaticIdentity ResourceType = "AWSClusterStaticIdentity"
	ResourceTypeAzureClusterIdentity     ResourceType = "AzureClusterIdentity"
	ResourceTypeCredential               ResourceType = "Credential"
)

// ExistsFunc is a function that checks if a resource exists
type ExistsFunc func(ctx context.Context, namespace, name string) (bool, error)

// CreateFunc is a function that creates a resource
type CreateFunc func(ctx context.Context) error

// ResourceSpec defines the specification of a resource to be created
type ResourceSpec struct {
	// Type is the type of the resource
	Type ResourceType
	// Namespace is the Kubernetes namespace
	Namespace string
	// Name is the name of the resource
	Name string
}

// createIfNotExists is a reusable function that checks for resource existence
// before attempting creation. If the resource already exists, it logs at DEBUG
// level and skips creation entirely (no API calls).
//
// If the resource does not exist, it calls the create function.
//
// This function implements a "best effort" error handling strategy:
// - If the resource exists, it returns nil (success)
// - If the check for existence fails, it returns an error
// - If creation fails, it logs a warning and returns the error
func (m *Manager) createIfNotExists(
	ctx context.Context,
	spec ResourceSpec,
	existsFn ExistsFunc,
	createFn CreateFunc,
) error {
	utils.GetLogger().Debugf("Checking if %s %s/%s exists", spec.Type, spec.Namespace, spec.Name)

	exists, err := existsFn(ctx, spec.Namespace, spec.Name)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check if %s %s/%s exists: %w", spec.Type, spec.Namespace, spec.Name, err)
	}

	if exists {
		utils.GetLogger().Debugf("Skipping %s %s/%s (already exists)", spec.Type, spec.Namespace, spec.Name)
		return nil
	}

	utils.GetLogger().Debugf("Creating %s %s/%s", spec.Type, spec.Namespace, spec.Name)

	if err := createFn(ctx); err != nil {
		return fmt.Errorf("failed to create %s %s/%s: %w", spec.Type, spec.Namespace, spec.Name, err)
	}

	utils.GetLogger().Infof("✅ Created %s %s/%s", spec.Type, spec.Namespace, spec.Name)
	return nil
}
