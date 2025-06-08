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
