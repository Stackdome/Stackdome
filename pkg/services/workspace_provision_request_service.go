package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker/workermanager"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkspaceProvisionRequestService interface {
	Get(ctx context.Context, ID string) (*models.WorkspaceProvisionRequest, *errors.ServiceError)
	GetProvisionRequestForUser(ctx context.Context, userID string) (*models.WorkspaceProvisionRequest, *errors.ServiceError)
	InternalList(ctx context.Context, query string, args ...any) ([]*models.WorkspaceProvisionRequest, *errors.ServiceError)
	Create(ctx context.Context, spec *models.WorkspaceProvisionRequest) (*models.WorkspaceProvisionRequest, *errors.ServiceError)
	Update(ctx context.Context, ID string, spec *models.WorkspaceProvisionRequest) (*models.WorkspaceProvisionRequest, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, spec *models.WorkspaceProvisionRequestStatus) (*models.WorkspaceProvisionRequest, *errors.ServiceError)
	InternalUpdate(ctx context.Context, ID string, spec *models.WorkspaceProvisionRequest) *errors.ServiceError
	Delete(ctx context.Context, ID string) *errors.ServiceError
}

var _ UserService = &usersService{}

func NewWorkspaceProvisionRequestService(spec WorkspaceProvisionRequestServiceSpec) WorkspaceProvisionRequestService {
	return &workspaceProvisionRequestService{
		wsProvisionRequestStore: pgstore.NewWorkspaceProvisionRequestStore(pgstore.WorkspaceProvisionRequestStoreSpec{
			SessionFactory:              spec.SessionFactory,
			ProvisionRequestStatusStore: pgstore.NewWorkspaceProvisionRequestStatusStore(),
		}),
		logger:                spec.Logger,
		clusterClient:         spec.ClusterClient,
		backgroundJobEnqueuer: spec.BackgroundJobEnqueuer,
	}
}

type WorkspaceProvisionRequestServiceSpec struct {
	SessionFactory        db.SessionFactory
	Logger                logger.Logger
	ClusterClient         client.Client
	BackgroundJobEnqueuer workermanager.BackgroundJobEnqueuer
}

type workspaceProvisionRequestService struct {
	wsProvisionRequestStore stores.WorkspaceProvisionRequestStore
	logger                  logger.Logger
	clusterClient           client.Client
	backgroundJobEnqueuer   workermanager.BackgroundJobEnqueuer
}

func (s *workspaceProvisionRequestService) Get(ctx context.Context, ID string) (*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	request, err := s.wsProvisionRequestStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get workspace provision request: %v", err)
		return nil, err
	}
	return request, nil
}
func (s *workspaceProvisionRequestService) GetProvisionRequestForUser(ctx context.Context, userID string) (*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	request, err := s.wsProvisionRequestStore.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Errorf("failed to get workspace provision request: %v", err)
		return nil, err
	}
	return request, nil
}

func (s *workspaceProvisionRequestService) InternalList(ctx context.Context, query string, args ...any) ([]*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	requests, err := s.wsProvisionRequestStore.InternalList(ctx, query, args...)
	if err != nil {
		s.logger.Errorf("failed to internal list workspace provision requests: %v", err)
		return nil, err
	}
	return requests, nil
}

func (s *workspaceProvisionRequestService) Create(ctx context.Context, spec *models.WorkspaceProvisionRequest) (*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	spec.Status = nil
	spec.State = models.ProvisionRequestPending
	spec.Message = "Pending object creation in cluster."
	request, err := s.wsProvisionRequestStore.Create(ctx, spec)
	if err != nil {
		s.logger.Errorf("failed to create workspace provision request: %v", err)
		return nil, err
	}
	if err := s.backgroundJobEnqueuer.Enqueue(request); err != nil {
		return nil, errors.GeneralError("failed to enqueue background job: %s", err.Error())
	}
	return request, nil
}

func (s *workspaceProvisionRequestService) Update(ctx context.Context, id string, spec *models.WorkspaceProvisionRequest) (*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	request, err := s.wsProvisionRequestStore.Update(ctx, id, spec)
	if err != nil {
		s.logger.Errorf("failed to update workspace provision request: %v", err)
		return nil, err
	}
	if err := s.backgroundJobEnqueuer.Enqueue(request); err != nil {
		return nil, errors.GeneralError("failed to enqueue background job: %s", err.Error())
	}
	return request, nil
}

func (s *workspaceProvisionRequestService) UpdateStatus(ctx context.Context, id string, spec *models.WorkspaceProvisionRequestStatus) (*models.WorkspaceProvisionRequest, *errors.ServiceError) {
	request, err := s.wsProvisionRequestStore.PatchStatus(ctx, id, spec)
	if err != nil {
		s.logger.Errorf("failed to update workspace provision request: %v", err)
		return nil, err
	}
	return request, nil
}

func (s *workspaceProvisionRequestService) InternalUpdate(ctx context.Context, id string, spec *models.WorkspaceProvisionRequest) *errors.ServiceError {
	_, err := s.wsProvisionRequestStore.Update(ctx, id, spec)
	if err != nil {
		s.logger.Errorf("failed to update workspace provision request: %v", err)
		return err
	}
	return nil
}

func (s *workspaceProvisionRequestService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	err := s.wsProvisionRequestStore.Delete(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to delete workspace provision request: %v", err)
		return err
	}
	return nil
}
