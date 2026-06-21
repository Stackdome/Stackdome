package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

//go:generate mockgen -destination=../mocks/mock_resource_reference_store.go -package=mocks github.com/ashishmax31/stackdome-api-server/pkg/stores ResourceReferenceStore
type ResourceReferenceStore interface {
	// ReplaceSpecWithTx deletes all spec-scoped rows (release_id IS NULL) for the
	// stack and reinserts refs. Must run inside a transaction.
	ReplaceSpecWithTx(ctx context.Context, stackID string, refs []models.ResourceReference) *errors.ServiceError
	// InsertReleaseWithTx inserts release-scoped rows. Must run inside a transaction.
	InsertReleaseWithTx(ctx context.Context, releaseID, stackID string, refs []models.ResourceReference) *errors.ServiceError
	// ListByReferent returns every reference row (any scope, org-wide) for a referent.
	ListByReferent(ctx context.Context, referentType models.ReferentType, referentID string) ([]models.ResourceReference, *errors.ServiceError)
}
