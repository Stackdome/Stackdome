package pgstore

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
)

type ResourceReferenceStoreSpec struct {
	SessionFactory db.SessionFactory
}

type resourceReferenceStore struct {
	sessionFactory db.SessionFactory
}

func NewResourceReferenceStore(spec ResourceReferenceStoreSpec) stores.ResourceReferenceStore {
	return &resourceReferenceStore{sessionFactory: spec.SessionFactory}
}

func (s *resourceReferenceStore) ReplaceSpecWithTx(ctx context.Context, stackID string, refs []models.ResourceReference) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	if err := tx.Where("stack_id = ? AND release_id IS NULL", stackID).
		Delete(&models.ResourceReference{}).Error; err != nil {
		return errors.GeneralError("failed to clear spec references: %s", err.Error())
	}
	if len(refs) == 0 {
		return nil
	}
	for i := range refs {
		refs[i].StackID = stackID
		refs[i].ReleaseID = nil
	}
	if err := tx.Create(&refs).Error; err != nil {
		return errors.GeneralError("failed to insert spec references: %s", err.Error())
	}
	return nil
}

func (s *resourceReferenceStore) InsertReleaseWithTx(ctx context.Context, releaseID, stackID string, refs []models.ResourceReference) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	if len(refs) == 0 {
		return nil
	}
	rid := releaseID
	for i := range refs {
		refs[i].StackID = stackID
		refs[i].ReleaseID = &rid
	}
	if err := tx.Create(&refs).Error; err != nil {
		return errors.GeneralError("failed to insert release references: %s", err.Error())
	}
	return nil
}

// ListByReferent returns every gripping reference row for a referent.
// Spec rows (release_id IS NULL) always grip — they represent the live stack spec.
// Release rows grip only if their release is NOT in a non-gripping terminal state
// (Failed/Superseded/Cancelled). This is derived at query time via a JOIN on
// stack_releases.state so it tracks release lifecycle without write-time hooks.
func (s *resourceReferenceStore) ListByReferent(ctx context.Context, referentType models.ReferentType, referentID string) ([]models.ResourceReference, *errors.ServiceError) {
	var refs []models.ResourceReference
	if err := s.sessionFactory.New(ctx).
		Model(&models.ResourceReference{}).
		Select("resource_references.*").
		Joins("LEFT JOIN stack_releases ON stack_releases.id = resource_references.release_id").
		Where("resource_references.referent_type = ? AND resource_references.referent_id = ?", referentType, referentID).
		Where("resource_references.release_id IS NULL OR stack_releases.state NOT IN ?", models.NonGrippingReleaseStates).
		Find(&refs).Error; err != nil {
		return nil, errors.GeneralError("failed to list references: %s", err.Error())
	}
	return refs, nil
}
