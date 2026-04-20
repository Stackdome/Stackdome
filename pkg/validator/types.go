package validator

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type SecretValidator interface {
	ValidateSecretData(secret *models.Secret) *errors.ServiceError
}

type InterpolationValidation interface {
	ValidateStackInterpolations(in *models.Stack) error
}

type StackValidator interface {
	ValidateForCreate(ctx context.Context, spec *models.Stack) *errors.ServiceError
	ValidateForUpdate(ctx context.Context, existing *models.Stack, spec *models.Stack) *errors.ServiceError
}

type PostgresAddonValidator interface {
	ValidateForCreate(ctx context.Context, spec *models.PostgresAddon) *errors.ServiceError
	ValidateForUpdate(ctx context.Context, existing *models.PostgresAddon, spec *models.PostgresAddon) *errors.ServiceError
}

type ObjectStoreValidator interface {
	ValidateForCreate(ctx context.Context, spec *models.ObjectStore) *errors.ServiceError
	ValidateForUpdate(ctx context.Context, existing *models.ObjectStore, spec *models.ObjectStore) *errors.ServiceError
}
