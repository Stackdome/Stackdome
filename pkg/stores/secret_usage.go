package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type SecretUsageStore interface {
	Create(ctx context.Context, secretUsage *models.SecretUsage) (*models.SecretUsage, *errors.ServiceError)
	GetBySecretIDAndStackID(ctx context.Context, secretID, stackID string) (*models.SecretUsage, *errors.ServiceError)
	GetBySecretID(ctx context.Context, secretID string) ([]*models.SecretUsage, *errors.ServiceError)
	GetByStackID(ctx context.Context, stackID string) ([]*models.SecretUsage, *errors.ServiceError)
	Delete(ctx context.Context, secretID, stackID string) *errors.ServiceError
	DeleteBySecretID(ctx context.Context, secretID string) *errors.ServiceError
	DeleteByStackID(ctx context.Context, stackID string) *errors.ServiceError
}
