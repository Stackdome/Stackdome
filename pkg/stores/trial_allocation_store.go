package stores

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=trial_allocation_store.go -destination=../mocks/mock_trial_allocation_store.go -package=mocks
type TrialAllocationStore interface {
	AcquireWithTx(ctx context.Context, organisationID string, now, expiresAt time.Time, capacity int) (*models.TrialAllocation, *errors.ServiceError)
	RevalidateWithTx(ctx context.Context, organisationID string, now time.Time) (*models.TrialAllocation, *errors.ServiceError)
	RevalidateIfExistsWithTx(ctx context.Context, organisationID string, now time.Time) (*models.TrialAllocation, *errors.ServiceError)
	GetActiveByOrganisationID(ctx context.Context, organisationID string, now time.Time) (*models.TrialAllocation, *errors.ServiceError)
	GetByOrganisationID(ctx context.Context, organisationID string) (*models.TrialAllocation, *errors.ServiceError)
}
