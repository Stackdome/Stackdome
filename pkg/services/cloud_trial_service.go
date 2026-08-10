package services

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
)

//go:generate mockgen -source=cloud_trial_service.go -destination=cloud_trial_service_mock.go -package=services -self_package=github.com/Stackdome/stackdome/pkg/services
type CloudTrialService interface {
	AcquireWithTx(ctx context.Context, organisationID string) (*models.TrialAllocation, *errors.ServiceError)
	RevalidateWithTx(ctx context.Context, organisationID string) (*models.TrialAllocation, *errors.ServiceError)
	RevalidateIfExistsWithTx(ctx context.Context, organisationID string) (*models.TrialAllocation, *errors.ServiceError)
	RequireActive(ctx context.Context, organisationID string) (*models.TrialAllocation, *errors.ServiceError)
	EnsureNoAllocation(ctx context.Context, organisationID string) *errors.ServiceError
}

type CloudTrialServiceSpec struct {
	Store    stores.TrialAllocationStore
	Capacity int
	TTL      time.Duration
	Now      func() time.Time
}

type cloudTrialService struct {
	store    stores.TrialAllocationStore
	capacity int
	ttl      time.Duration
	now      func() time.Time
}

func NewCloudTrialService(spec CloudTrialServiceSpec) CloudTrialService {
	now := spec.Now
	if now == nil {
		now = time.Now
	}
	return &cloudTrialService{store: spec.Store, capacity: spec.Capacity, ttl: spec.TTL, now: now}
}

func (s *cloudTrialService) AcquireWithTx(ctx context.Context, organisationID string) (*models.TrialAllocation, *errors.ServiceError) {
	now := s.now()
	return s.store.AcquireWithTx(ctx, organisationID, now, now.Add(s.ttl), s.capacity)
}

func (s *cloudTrialService) RevalidateWithTx(ctx context.Context, organisationID string) (*models.TrialAllocation, *errors.ServiceError) {
	return s.store.RevalidateWithTx(ctx, organisationID, s.now())
}

func (s *cloudTrialService) RevalidateIfExistsWithTx(ctx context.Context, organisationID string) (*models.TrialAllocation, *errors.ServiceError) {
	return s.store.RevalidateIfExistsWithTx(ctx, organisationID, s.now())
}

func (s *cloudTrialService) RequireActive(ctx context.Context, organisationID string) (*models.TrialAllocation, *errors.ServiceError) {
	return s.store.GetActiveByOrganisationID(ctx, organisationID, s.now())
}

func (s *cloudTrialService) EnsureNoAllocation(ctx context.Context, organisationID string) *errors.ServiceError {
	_, serr := s.store.GetByOrganisationID(ctx, organisationID)
	if serr == nil {
		return errors.BadRequest("cannot delete organisation while its cloud trial allocation exists")
	}
	if serr.Is404() {
		return nil
	}
	return serr
}
