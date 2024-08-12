package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services/clusterresource"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/google/uuid"
	"k8s.io/utils/ptr"
)

type WorkspaceUserService interface {
	GetByID(ctx context.Context, ID string) (*models.WorkspaceUser, *errors.ServiceError)
	GetWorkspaceUser(ctx context.Context, userID string) (*models.WorkspaceUser, *errors.ServiceError)
	InternalList(ctx context.Context, query string, args ...any) ([]*models.WorkspaceUser, *errors.ServiceError)
	Create(ctx context.Context, spec *models.WorkspaceUser, user *models.User) (*models.WorkspaceUser, *errors.ServiceError)
	Update(ctx context.Context, ID string, spec *models.WorkspaceUser, user *models.User) (*models.WorkspaceUser, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, spec *models.WorkspaceUser) *errors.ServiceError
	InternalUpdate(ctx context.Context, ID string, spec *models.WorkspaceUser) *errors.ServiceError
	Delete(ctx context.Context, ID string) *errors.ServiceError
	InternalDelete(ctx context.Context, ID string) *errors.ServiceError
	InjectClusterResourceService(clusterResourceService clusterresource.WorkspaceUserClusterResourceService)
}

var _ UserService = &usersService{}

func NewWorkspaceUserService(spec WorkspaceUserServiceSpec) WorkspaceUserService {
	return &workspaceUserService{
		workspaceUserStore: pgstore.NewWorkspaceUserStore(pgstore.WorkspaceUserStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger:           spec.Logger,
		dbClusterService: spec.ClusterService,
		usersService:     spec.UserService,
	}
}

type WorkspaceUserServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
	ClusterService ClusterService
	UserService    UserService
}

type workspaceUserService struct {
	workspaceUserStore     stores.WorkspaceUserStore
	dbClusterService       ClusterService
	logger                 logger.Logger
	usersService           UserService
	clusterResourceService clusterresource.WorkspaceUserClusterResourceService
}

func (s *workspaceUserService) InjectClusterResourceService(clusterResourceService clusterresource.WorkspaceUserClusterResourceService) {
	s.clusterResourceService = clusterResourceService
}

func (s *workspaceUserService) GetByID(ctx context.Context, ID string) (*models.WorkspaceUser, *errors.ServiceError) {
	request, err := s.workspaceUserStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get workspace user: %v", err)
		return nil, err
	}
	return request, nil
}
func (s *workspaceUserService) GetWorkspaceUser(ctx context.Context, userID string) (*models.WorkspaceUser, *errors.ServiceError) {
	request, err := s.workspaceUserStore.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Errorf("failed to get workspace user: %v", err)
		return nil, err
	}
	return request, nil
}

func (s *workspaceUserService) InternalList(ctx context.Context, query string, args ...any) ([]*models.WorkspaceUser, *errors.ServiceError) {
	requests, err := s.workspaceUserStore.InternalList(ctx, query, args...)
	if err != nil {
		s.logger.Errorf("failed to internal list workspace users: %v", err)
		return nil, err
	}
	return requests, nil
}

func (s *workspaceUserService) Create(ctx context.Context, spec *models.WorkspaceUser, user *models.User) (*models.WorkspaceUser, *errors.ServiceError) {
	spec.Status.State = models.WorkspaceUserProvisionPending
	spec.Status.Message = "Provision pending"
	s.setNamespacesForCreate(spec, user)
	cluster, serr := s.dbClusterService.GetClusterForOrg(ctx, user.OrganisationID)
	if serr != nil {
		s.logger.Errorf("failed to get cluster for org: %v", serr)
		return nil, serr
	}
	spec.ClusterID = cluster.ID
	var createdWorkspaceUser *models.WorkspaceUser
	createErr := s.workspaceUserStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		createdWorkspaceUser, serr = s.workspaceUserStore.CreateWithTx(ctx, spec)
		if serr != nil {
			return serr
		}
		if err := s.clusterResourceService.CreateWorkspaceUserInCluster(ctx, createdWorkspaceUser); err != nil {
			return errors.GeneralError("failed to create workspace user in cluster: %s", err.Error())
		}
		return nil
	})
	return createdWorkspaceUser, createErr
}

func (s *workspaceUserService) Update(ctx context.Context, id string, spec *models.WorkspaceUser, user *models.User) (*models.WorkspaceUser, *errors.ServiceError) {
	current, err := s.workspaceUserStore.GetByID(ctx, id)
	if err != nil {
		s.logger.Errorf("failed to get workspace user: %v", err)
		return nil, err
	}
	s.setNamespacesForUpdate(spec, current, user)
	cluster, serr := s.dbClusterService.GetClusterForOrg(ctx, user.OrganisationID)
	if serr != nil {
		s.logger.Errorf("failed to get cluster for org: %v", serr)
		return nil, serr
	}
	spec.ClusterID = cluster.ID

	var updatedWorkspaceUser *models.WorkspaceUser
	updateErr := s.workspaceUserStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		updatedWorkspaceUser, serr = s.workspaceUserStore.UpdateWithTx(ctx, id, spec)
		if serr != nil {
			s.logger.Errorf("failed to update workspace user: %v", serr)
			return serr
		}
		s.logger.Infof("updated workspace user: %s", updatedWorkspaceUser.ID)
		if err := s.clusterResourceService.UpdateWorkspaceUserInCluster(ctx, updatedWorkspaceUser); err != nil {
			return errors.GeneralError("failed to update workspace user in cluster: %s", err.Error())
		}
		return nil
	})
	return updatedWorkspaceUser, updateErr
}

func (s *workspaceUserService) UpdateStatus(ctx context.Context, id string, WorkspaceUser *models.WorkspaceUser) *errors.ServiceError {
	_, err := s.workspaceUserStore.PatchStatus(ctx, id, WorkspaceUser.Status)
	if err != nil {
		s.logger.Errorf("failed to update workspace user: %v", err)
		return err
	}
	return nil
}

func (s *workspaceUserService) InternalUpdate(ctx context.Context, id string, spec *models.WorkspaceUser) *errors.ServiceError {
	_, err := s.workspaceUserStore.Update(ctx, id, spec)
	if err != nil {
		s.logger.Errorf("failed to update workspace user: %v", err)
		return err
	}
	return nil
}

func (s *workspaceUserService) InternalDelete(ctx context.Context, id string) *errors.ServiceError {
	err := s.workspaceUserStore.Delete(ctx, id)
	if err != nil {
		s.logger.Errorf("failed to delete workspace user: %v", err)
		return err
	}
	return nil
}

func (s *workspaceUserService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	workspaceUser, serr := s.GetByID(ctx, ID)
	if serr != nil {
		return serr
	}
	workspaceUser.DeletionTimeStamp = ptr.To(time.Now().UTC())

	deleteErr := s.workspaceUserStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		if err := s.clusterResourceService.DeleteWorkspaceUserInCluster(ctx, workspaceUser); err != nil {
			if k8sapierrors.IsNotFound(err.Err) {
				s.logger.Warn(ctx, "workspace user already gone from cluster")
			} else {
				return errors.GeneralError("failed to delete workspace user in cluster: %s", err.Error())
			}
		}
		return s.workspaceUserStore.DeleteWithTx(ctx, ID)
	})
	return deleteErr
}

func (s *workspaceUserService) setNamespacesForCreate(spec *models.WorkspaceUser, user *models.User) {
	for i := range spec.WorkspaceNamespaces {
		namespaceUUID := uuid.NewString()
		namespaceString := fmt.Sprintf("%s-%s", user.Name, namespaceUUID)
		spec.WorkspaceNamespaces[i].Namespace = namespaceString
		spec.WorkspaceNamespaces[i].Enabled = true
	}
}

func (s *workspaceUserService) setNamespacesForUpdate(spec *models.WorkspaceUser, current *models.WorkspaceUser, user *models.User) {
	exitingWorkspaceNamespaceMap := make(map[string]string)
	for _, wn := range current.WorkspaceNamespaces {
		exitingWorkspaceNamespaceMap[wn.Workspace] = wn.Namespace
	}
	for i, ws := range spec.WorkspaceNamespaces {
		if ns, ok := exitingWorkspaceNamespaceMap[ws.Workspace]; ok {
			spec.WorkspaceNamespaces[i].Namespace = ns
		} else {
			namespaceUUID := uuid.NewString()
			namespaceString := fmt.Sprintf("%s-%s", user.Name, namespaceUUID)
			spec.WorkspaceNamespaces[i].Namespace = namespaceString
		}
		spec.WorkspaceNamespaces[i].Enabled = true
	}
}
