package validator

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

type SecretValidator interface {
	ValidateSecretData(secret *models.Secret) *errors.ServiceError
	ValidateSecretType(secretType models.SecretType, secret *models.Secret) *errors.ServiceError
}

//go:generate mockgen -destination=../mocks/mock_stack_validator.go -package=mocks github.com/Stackdome/stackdome/pkg/validator StackValidator
type StackValidator interface {
	ValidateForCreate(ctx context.Context, spec *models.Stack) *errors.ServiceError
	ValidateForUpdate(ctx context.Context, existing *models.Stack, spec *models.Stack) *errors.ServiceError
	// ValidateConnections runs only the connection-scoped rules against the
	// full stack context, skipping per-resource validation and its DB
	// referential lookups. Used for connection-only mutations so a
	// pre-existing, unrelated stack invalidity can't block a connection edit.
	ValidateConnections(ctx context.Context, spec *models.Stack) *errors.ServiceError
	// ValidateShell runs only the rules scoped to the stack's own columns
	// (currently the settings limits), skipping resource, uniqueness, and
	// connection validation entirely. Used for thin shell mutations, which
	// never carry children.
	ValidateShell(ctx context.Context, spec *models.Stack) *errors.ServiceError
}

type PostgresAddonValidator interface {
	ValidateForCreate(ctx context.Context, spec *models.PostgresAddon) *errors.ServiceError
	ValidateForUpdate(ctx context.Context, existing *models.PostgresAddon, spec *models.PostgresAddon) *errors.ServiceError
}

type ObjectStoreValidator interface {
	ValidateForCreate(ctx context.Context, spec *models.ObjectStore) *errors.ServiceError
	ValidateForUpdate(ctx context.Context, existing *models.ObjectStore, spec *models.ObjectStore) *errors.ServiceError
}
