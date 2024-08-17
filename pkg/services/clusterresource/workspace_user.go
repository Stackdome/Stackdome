package clusterresource

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	workspacev1alpha1 "soradev.io/cluster-agent/api/v1alpha1"
)

// WorkspaceUserClusterResourceService is an interface for managing workspace user resources in a cluster.
type WorkspaceUserClusterResourceService interface {
	CreateWorkspaceUserInCluster(ctx context.Context, workspaceUser *models.WorkspaceUser) *ClusterResourceError
	DeleteWorkspaceUserInCluster(ctx context.Context, workspaceUser *models.WorkspaceUser) *ClusterResourceError
	UpdateWorkspaceUserInCluster(ctx context.Context, workspaceUser *models.WorkspaceUser) *ClusterResourceError
}

type workspaceUserClusterResourceService struct {
	clusterService DBClusterService
	userService    DBUserService
	clusterManager clustermanager.ClusterManager
	logger         logger.Logger
}

type WorkspaceUserClusterResourceServiceSpec struct {
	ClusterService DBClusterService
	UserService    DBUserService
	ClusterManager clustermanager.ClusterManager
	Logger         logger.Logger
}

// NewWorkspaceUserClusterResourceService creates a new WorkspaceUserClusterResourceService
func NewWorkspaceUserClusterResourceService(spec WorkspaceUserClusterResourceServiceSpec) WorkspaceUserClusterResourceService {
	return &workspaceUserClusterResourceService{
		clusterService: spec.ClusterService,
		userService:    spec.UserService,
		clusterManager: spec.ClusterManager,
		logger:         spec.Logger,
	}
}

// CreateWorkspaceUserInCluster creates a workspace user in the cluster
func (s *workspaceUserClusterResourceService) CreateWorkspaceUserInCluster(ctx context.Context, workspaceUser *models.WorkspaceUser) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, workspaceUser.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	user, err := s.userService.Get(ctx, workspaceUser.UserID)
	if err != nil {
		s.logger.Errorf("failed to get user: %v", err)
		return newError("failed to get user", err)
	}

	accessRules := user.ClusterAccessRules()

	desiredObject := s.desiredObjectInCluster(user, workspaceUser, accessRules)

	client, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		return newError("failed to get cluster client", clientGetErr)
	}

	// Create the user in the cluster
	createErr := client.Create(ctx, desiredObject)
	if createErr != nil {
		s.logger.Errorf("failed to create workspaceuser in cluster: %v", err)
		return newError("failed to create workspaceuser in cluster", createErr)
	}

	return nil
}

// DeleteWorkspaceUserInCluster deletes a workspace user in the cluster
func (s *workspaceUserClusterResourceService) DeleteWorkspaceUserInCluster(ctx context.Context, workspaceUser *models.WorkspaceUser) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, workspaceUser.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}
	user, err := s.userService.Get(ctx, workspaceUser.UserID)
	if err != nil {
		s.logger.Errorf("failed to get user: %v", err)
		return newError("failed to get user from db", err)
	}

	client, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		return newError("failed to get cluster client", clientGetErr)
	}

	desiredObject := &workspacev1alpha1.WorkspaceUser{
		ObjectMeta: metav1.ObjectMeta{
			Name: WorkspaceUserClusterObjectName(user),
		},
	}

	// Delete the user in the cluster
	deleteErr := client.Delete(ctx, desiredObject)
	if deleteErr != nil {
		s.logger.Errorf("failed to delete workspaceuser in cluster: %v", deleteErr)
		return newError("failed to delete workspaceuser in cluster", deleteErr)
	}

	return nil
}

// UpdateWorkspaceUserInCluster updates a workspace user in the cluster
func (s *workspaceUserClusterResourceService) UpdateWorkspaceUserInCluster(ctx context.Context, workspaceUser *models.WorkspaceUser) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, workspaceUser.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	user, err := s.userService.Get(ctx, workspaceUser.UserID)
	if err != nil {
		s.logger.Errorf("failed to get user: %v", err)
		return newError("failed to get user from db", err)
	}

	accessRules := user.ClusterAccessRules()

	desiredObject := s.desiredObjectInCluster(user, workspaceUser, accessRules)

	client, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		return newError("failed to get cluster client", clientGetErr)
	}

	existingObject := &workspacev1alpha1.WorkspaceUser{}
	if err := client.Get(ctx, k8sclient.ObjectKeyFromObject(desiredObject), existingObject); err != nil {
		return newError("failed to get workspaceuser from cluster", err)
	}

	// Update the user in the cluster
	desiredObject.ResourceVersion = existingObject.ResourceVersion
	updateErr := client.Update(ctx, desiredObject)
	if updateErr != nil {
		s.logger.Errorf("failed to update workspaceuser in cluster: %v", err)
		return newError("failed to update workspaceuser in cluster", updateErr)
	}

	return nil
}

func (s *workspaceUserClusterResourceService) desiredObjectInCluster(user *models.User, workspaceUser *models.WorkspaceUser, accessRules []rbacv1.PolicyRule) *workspacev1alpha1.WorkspaceUser {
	return &workspacev1alpha1.WorkspaceUser{
		ObjectMeta: metav1.ObjectMeta{
			Name: WorkspaceUserClusterObjectName(user),
			Labels: map[string]string{
				models.WorkspaceUserIDLabel:   workspaceUser.ID,
				models.ObjectServerGeneration: fmt.Sprintf("%d", workspaceUser.Version),
			},
		},
		Spec: workspacev1alpha1.WorkspaceUserSpec{
			Username:    user.Name,
			Namespaces:  getNamespaces(workspaceUser),
			AccessRules: accessRules,
		},
	}
}

func WorkspaceNamespaceFor(user *models.User) string {
	return fmt.Sprintf("%s-workspace", WorkspaceUserClusterObjectName(user))
}

func WorkspaceUserClusterObjectName(user *models.User) string {
	sanitizedName := sanitizeName(user.Name)

	objectName := fmt.Sprintf("%s-%s", sanitizedName, user.ID)
	// Ensure the object name meets the Kubernetes requirements
	return truncateObjectName(objectName)
}

func getNamespaces(workspaceUser *models.WorkspaceUser) []string {
	var namespaces []string
	for _, ns := range workspaceUser.WorkspaceNamespaces {
		if ns.Enabled {
			namespaces = append(namespaces, ns.Namespace)
		}
	}
	return namespaces
}
