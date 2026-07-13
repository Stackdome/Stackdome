package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/google/uuid"
	"k8s.io/utils/ptr"
)

type WorkspaceUserService interface {
	GetByID(ctx context.Context, ID string) (*models.WorkspaceUser, *errors.ServiceError)
	InternalGetByID(ctx context.Context, ID string) (*models.WorkspaceUser, *errors.ServiceError)
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
		permissions:      spec.Permissions,
	}
}

type WorkspaceUserServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
	ClusterService ClusterService
	UserService    UserService
	Permissions    auth.PermissionService
}

type workspaceUserService struct {
	workspaceUserStore     stores.WorkspaceUserStore
	dbClusterService       ClusterService
	logger                 logger.Logger
	usersService           UserService
	permissions            auth.PermissionService
	clusterResourceService clusterresource.WorkspaceUserClusterResourceService
}

func (s *workspaceUserService) InjectClusterResourceService(clusterResourceService clusterresource.WorkspaceUserClusterResourceService) {
	s.clusterResourceService = clusterResourceService
}

func (s *workspaceUserService) GetByID(ctx context.Context, ID string) (*models.WorkspaceUser, *errors.ServiceError) {
	request, err := s.workspaceUserStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Error(ctx, "failed to get workspace user: %v", err)
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, request.ProjectID, auth.ResourceWorkspaceUsers, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return request, nil
}

func (s *workspaceUserService) InternalGetByID(ctx context.Context, ID string) (*models.WorkspaceUser, *errors.ServiceError) {
	return s.workspaceUserStore.GetByID(ctx, ID)
}

func (s *workspaceUserService) GetWorkspaceUser(ctx context.Context, userID string) (*models.WorkspaceUser, *errors.ServiceError) {
	request, err := s.workspaceUserStore.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "failed to get workspace user: %v", err)
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, request.ProjectID, auth.ResourceWorkspaceUsers, request.ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return request, nil
}

func (s *workspaceUserService) InternalList(ctx context.Context, query string, args ...any) ([]*models.WorkspaceUser, *errors.ServiceError) {
	requests, err := s.workspaceUserStore.InternalList(ctx, query, args...)
	if err != nil {
		s.logger.Error(ctx, "failed to internal list workspace users: %v", err)
		return nil, err
	}
	return requests, nil
}

func (s *workspaceUserService) Create(ctx context.Context, spec *models.WorkspaceUser, user *models.User) (*models.WorkspaceUser, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, spec.ProjectID, auth.ResourceWorkspaceUsers, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}
	spec.Status.State = models.WorkspaceUserProvisionPending
	spec.Status.Message = "Provision pending"
	s.setNamespacesForCreate(spec, user)
	cluster, serr := s.dbClusterService.GetClusterForOrg(ctx, user.OrganisationID)
	if serr != nil {
		s.logger.Error(ctx, "failed to get cluster for org: %v", serr)
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
	if createErr != nil {
		return nil, createErr
	}
	return createdWorkspaceUser, nil
}

func (s *workspaceUserService) Update(ctx context.Context, id string, spec *models.WorkspaceUser, user *models.User) (*models.WorkspaceUser, *errors.ServiceError) {
	current, err := s.workspaceUserStore.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to get workspace user: %v", err)
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, current.ProjectID, auth.ResourceWorkspaceUsers, id, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	s.setNamespacesForUpdate(spec, current, user)
	cluster, serr := s.dbClusterService.GetClusterForOrg(ctx, user.OrganisationID)
	if serr != nil {
		s.logger.Error(ctx, "failed to get cluster for org: %v", serr)
		return nil, serr
	}
	spec.ClusterID = cluster.ID

	var updatedWorkspaceUser *models.WorkspaceUser
	updateErr := s.workspaceUserStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		updatedWorkspaceUser, serr = s.workspaceUserStore.UpdateWithTx(ctx, id, spec)
		if serr != nil {
			s.logger.Error(ctx, "failed to update workspace user: %v", serr)
			return serr
		}
		s.logger.Info(ctx, "updated workspace user: %s", updatedWorkspaceUser.ID)
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
		s.logger.Error(ctx, "failed to update workspace user: %v", err)
		return err
	}
	return nil
}

func (s *workspaceUserService) InternalUpdate(ctx context.Context, id string, spec *models.WorkspaceUser) *errors.ServiceError {
	_, err := s.workspaceUserStore.Update(ctx, id, spec)
	if err != nil {
		s.logger.Error(ctx, "failed to update workspace user: %v", err)
		return err
	}
	return nil
}

func (s *workspaceUserService) InternalDelete(ctx context.Context, id string) *errors.ServiceError {
	err := s.workspaceUserStore.Delete(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "failed to delete workspace user: %v", err)
		return err
	}
	return nil
}

func (s *workspaceUserService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	workspaceUser, serr := s.workspaceUserStore.GetByID(ctx, ID)
	if serr != nil {
		return serr
	}
	if permErr := s.permissions.Check(ctx, workspaceUser.ProjectID, auth.ResourceWorkspaceUsers, ID, auth.ActionDelete); permErr != nil {
		return permErr
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
	if deleteErr != nil {
		return deleteErr
	}
	return nil
}

func (s *workspaceUserService) setNamespacesForCreate(spec *models.WorkspaceUser, user *models.User) {
	for i := range spec.WorkspaceNamespaces {
		spec.WorkspaceNamespaces[i].Namespace = buildNamepaceForUser(user)
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
			spec.WorkspaceNamespaces[i].Namespace = buildNamepaceForUser(user)
		}
		spec.WorkspaceNamespaces[i].Enabled = true
	}
}

func buildNamepaceForUser(user *models.User) string {
	namespaceUUID := uuid.NewString()
	return k8sValidString(fmt.Sprintf("%s-%s", user.Name, namespaceUUID))
}

func k8sValidString(name string) string {
	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	sanitized := reg.ReplaceAllString(name, "-")

	// remove all spaces
	sanitized = strings.ReplaceAll(sanitized, " ", "")
	// Remove leading and trailing hyphens
	sanitized = strings.TrimPrefix(sanitized, "-")
	sanitized = strings.TrimSuffix(sanitized, "-")

	return truncateObjectName(strings.ToLower(sanitized))
}

func truncateObjectName(name string) string {
	// Truncate the object name if it exceeds the maximum length
	maxLength := 63
	if len(name) > maxLength {
		name = name[:maxLength]
	}

	name = strings.TrimSuffix(name, "-")
	return name
}
