// Package stackdeploy converts a draft stack (user-authored state from the DB)
// into an effective stack: connections, self outputs, and volume relationships
// resolved into the concrete values the cluster CR builder consumes.
package stackdeploy

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

//go:generate mockgen -source=resolver.go -destination=mock_resolver_dependencies_test.go -package=stackdeploy

type SecretService interface {
	InternalGetByID(ctx context.Context, id string) (*models.Secret, *errors.ServiceError)
}

type PostgresAddonService interface {
	InternalGetPostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError)
	InternalGetCredentials(ctx context.Context, addonID string, database string, superuser bool) (*models.PostgresCredentials, *errors.ServiceError)
}

type VolumeService interface {
	ListVolumesUsedByStack(ctx context.Context, stackID string) ([]*models.Volume, *errors.ServiceError)
}

// DependencyNotReadyError signals that a backing resource is not available yet
// and resolution should be retried.
type DependencyNotReadyError struct {
	Message string
}

func (e DependencyNotReadyError) Error() string {
	return e.Message
}

type ResolverSpec struct {
	VolumeService        VolumeService
	PostgresAddonService PostgresAddonService
	SecretService        SecretService
}

type Resolver struct {
	volumeService        VolumeService
	postgresAddonService PostgresAddonService
	secretService        SecretService
}

func NewResolver(spec ResolverSpec) *Resolver {
	return &Resolver{
		volumeService:        spec.VolumeService,
		postgresAddonService: spec.PostgresAddonService,
		secretService:        spec.SecretService,
	}
}

// Resolve returns an effective copy of the stack with volume connections,
// self outputs, and env connections resolved. The input stack is not mutated.
func (r *Resolver) Resolve(ctx context.Context, stack *models.Stack) (*models.Stack, error) {
	effective, err := CloneStack(stack)
	if err != nil {
		return nil, err
	}
	if err := r.resolveVolumeConnections(ctx, effective); err != nil {
		return nil, err
	}
	if err := resolveSelfOutputEnvVars(effective); err != nil {
		return nil, err
	}
	if err := r.resolveEnvConnections(ctx, effective); err != nil {
		return nil, err
	}
	return effective, nil
}
