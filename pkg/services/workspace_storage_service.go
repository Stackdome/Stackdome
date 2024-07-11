package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type WorkspaceStorageService interface {
	Get(ctx context.Context, ID string) (*models.WorkspaceStorage, *errors.ServiceError)
	ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.WorkspaceStorage, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.WorkspaceStorage, *errors.ServiceError)
	Create(ctx context.Context, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError)
	Update(ctx context.Context, ID string, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.WorkspaceStorageStatus) (*models.WorkspaceStorage, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
}

type WorkspaceStorageServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

func NewWorkspaceStorageService(spec WorkspaceStorageServiceSpec) WorkspaceStorageService {
	return &workspaceStorageService{
		wsStorageStore: pgstore.NewWorkspaceStorageStore(pgstore.WorkspaceStorageStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

type workspaceStorageService struct {
	wsStorageStore stores.WorkspaceStorageStore
	logger         logger.Logger
}

func (s *workspaceStorageService) Get(ctx context.Context, ID string) (*models.WorkspaceStorage, *errors.ServiceError) {
	storage, err := s.wsStorageStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get workspace storage: %v", err)
		return nil, err
	}
	return storage, nil
}

func (s *workspaceStorageService) ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.WorkspaceStorage, *errors.ServiceError) {
	storages, err := s.wsStorageStore.ListByOrganisationID(ctx, organisationID)
	if err != nil {
		s.logger.Errorf("failed to list workspace storages: %v", err)
		return nil, err
	}
	return storages, nil
}

func (s *workspaceStorageService) ListByUserID(ctx context.Context, userID string) ([]*models.WorkspaceStorage, *errors.ServiceError) {
	storages, err := s.wsStorageStore.ListByUserID(ctx, userID)
	if err != nil {
		s.logger.Errorf("failed to list workspace storages by user ID: %v", err)
		return nil, err
	}
	return storages, nil
}

func (s *workspaceStorageService) Create(ctx context.Context, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError) {
	storage, err := s.wsStorageStore.Create(ctx, spec)
	if err != nil {
		s.logger.Errorf("failed to create workspace storage: %v", err)
		return nil, err
	}
	return storage, nil
}

func (s *workspaceStorageService) Update(ctx context.Context, ID string, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError) {
	storage, err := s.wsStorageStore.Update(ctx, ID, spec)
	if err != nil {
		s.logger.Errorf("failed to update workspace storage: %v", err)
		return nil, err
	}
	return storage, nil
}

func (s *workspaceStorageService) UpdateStatus(ctx context.Context, ID string, status *models.WorkspaceStorageStatus) (*models.WorkspaceStorage, *errors.ServiceError) {
	storage, err := s.wsStorageStore.UpsertStatus(ctx, ID, status)
	if err != nil {
		s.logger.Errorf("failed to update workspace storage status: %v", err)
		return nil, err
	}
	return storage, nil
}

func (s *workspaceStorageService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	err := s.wsStorageStore.Delete(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to delete workspace storage: %v", err)
		return err
	}
	return nil
}
