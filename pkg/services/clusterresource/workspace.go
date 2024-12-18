package clusterresource

import (
	"context"
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"strconv"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	workspacev1alpha1 "soradev.io/cluster-agent/api/v1alpha1"
)

type ClusterWorkspaceService interface {
	CreateWorkspaceInCluster(ctx context.Context, workspace *models.Workspace) *ClusterResourceError
	DeleteWorkspaceInCluster(ctx context.Context, workspace *models.Workspace) *ClusterResourceError
	UpdateWorkspaceInCluster(ctx context.Context, workspace *models.Workspace) *ClusterResourceError
}

type clusterWorkspaceService struct {
	clusterService      DBClusterService
	clusterManager      clustermanager.ClusterManager
	organisationService DBOrganisationService
	logger              logger.Logger
}

type ClusterWorkspaceServiceSpec struct {
	ClusterService      DBClusterService
	OrganisationService DBOrganisationService
	UserService         DBUserService
	ClusterManager      clustermanager.ClusterManager
	Logger              logger.Logger
}

func NewClusterWorkspaceService(spec ClusterWorkspaceServiceSpec) ClusterWorkspaceService {
	return &clusterWorkspaceService{
		clusterService:      spec.ClusterService,
		clusterManager:      spec.ClusterManager,
		organisationService: spec.OrganisationService,
		logger:              spec.Logger,
	}
}

func (s *clusterWorkspaceService) CreateWorkspaceInCluster(ctx context.Context, workspace *models.Workspace) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, workspace.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	organisation, err := s.organisationService.Get(ctx, workspace.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to get organisation: %v", err)
		return newError("failed to get organisation", err)
	}

	clusterClient, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		s.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	desiredWorkspaceCR := s.desiredObjectInCluster(workspace, organisation)

	if err := clusterClient.Create(ctx, desiredWorkspaceCR); err != nil {
		s.logger.Errorf("failed to create workspaceCR in cluster: %v", err)
		return newError("failed to create workspaceCR in cluster", err)
	}
	return nil
}

func (s *clusterWorkspaceService) DeleteWorkspaceInCluster(ctx context.Context, workspace *models.Workspace) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, workspace.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}
	organisation, err := s.organisationService.Get(ctx, workspace.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to get organisation: %v", err)
		return newError("failed to get organisation", err)
	}

	clusterClient, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		s.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	desiredWorkspaceCR := s.desiredObjectInCluster(workspace, organisation)

	if err := clusterClient.Delete(ctx, desiredWorkspaceCR); err != nil {
		if k8sapierrors.IsNotFound(err) {
			s.logger.Warn(ctx, "workspace storage '%s' not found in cluster", workspace.ID)
			return nil
		}
		s.logger.Errorf("failed to delete workspaceCR in cluster: %v", err)
		return newError("failed to delete workspaceCR in cluster", err)
	}
	return nil
}

func (s *clusterWorkspaceService) UpdateWorkspaceInCluster(ctx context.Context, workspace *models.Workspace) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, workspace.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	organisation, err := s.organisationService.Get(ctx, workspace.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to get organisation: %v", err)
		return newError("failed to get organisation", err)
	}

	clusterClient, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		s.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	desiredWorkspaceCR := s.desiredObjectInCluster(workspace, organisation)
	var existingWorkspaceCR workspacev1alpha1.Workspace

	if err := clusterClient.Get(ctx, client.ObjectKeyFromObject(desiredWorkspaceCR), &existingWorkspaceCR); err != nil {
		s.logger.Errorf("failed to get workspaceCR in cluster: %v", err)
		return newError("failed to get workspaceCR in cluster", err)
	}

	desiredWorkspaceCR.ResourceVersion = existingWorkspaceCR.ResourceVersion
	if err := clusterClient.Update(ctx, desiredWorkspaceCR); err != nil {
		s.logger.Errorf("failed to update workspaceCR in cluster: %v", err)
		return newError("failed to update workspaceCR in cluster", err)
	}
	return nil
}

func (s *clusterWorkspaceService) desiredObjectInCluster(workspace *models.Workspace, organisation *models.Organisation) *workspacev1alpha1.Workspace {
	workspaceCR := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workspace.Name,
			Namespace: workspace.Namespace,
			Labels: map[string]string{
				models.WorkspaceIDLabel:       workspace.ID,
				models.ObjectServerGeneration: fmt.Sprintf("%d", workspace.Version),
			},
		},
		Spec: workspacev1alpha1.WorkspaceSpec{
			Organisation: organisation.Name,
			Domain:       organisation.DomainName,
			Resources:    desiredWorkspaceResources(workspace),
		},
	}
	return workspaceCR
}

func desiredWorkspaceResources(workspace *models.Workspace) []workspacev1alpha1.ResourceSpec {
	var resources []workspacev1alpha1.ResourceSpec
	for _, workspaceResource := range workspace.WorkspaceResources {
		resource := workspacev1alpha1.ResourceSpec{
			Name: workspaceResource.Name,
			Spec: workspacev1alpha1.WorkspaceResourceSpec{
				ImageRegistry: workspaceResource.ImageRegistry,
				StateFul:      workspaceResource.StateFul,
				DependsOn:     workspaceResource.DependsOn,
			},
		}

		if workspaceResource.ExecutionConfig != nil {
			resource.Spec.Command = workspaceResource.ExecutionConfig.Command
			resource.Spec.Args = workspaceResource.ExecutionConfig.Args
		}

		if workspaceResource.LifecycleConfig != nil && workspaceResource.LifecycleConfig.RestartRequestTime != nil {
			resource.Spec.RestartRequest = &metav1.Time{Time: workspaceResource.LifecycleConfig.RestartRequestTime.UTC()}
		}

		setBuildSpec(&resource, workspaceResource)
		setInitSpec(&resource, workspaceResource)
		setVolumeMounts(&resource, workspaceResource)
		setPorts(&resource, workspaceResource)
		setEnvVars(&resource, workspaceResource)
		resources = append(resources, resource)
	}
	return resources
}

func setBuildSpec(resource *workspacev1alpha1.ResourceSpec, workspaceResource *models.WorkspaceResource) {
	if workspaceResource.Build != nil {
		resource.Spec.ApplicationBuildSpec = &workspacev1alpha1.ApplicationBuildSpec{
			VolumeName:      workspaceResource.Build.SourceVolumeID,
			Context:         workspaceResource.Build.ContextPath,
			DockerFile:      workspaceResource.Build.DockerfilePath,
			BuildSourceHash: workspaceResource.Build.SourceHash,
		}
	} else {
		resource.Spec.PrebuiltApplicationSpec = &workspacev1alpha1.PrebuiltApplicationSpec{
			Image: fmt.Sprintf("%s:%s", workspaceResource.Prebuilt.ImageName, workspaceResource.Prebuilt.Tag),
		}
	}
}

func setInitSpec(resource *workspacev1alpha1.ResourceSpec, workspaceResource *models.WorkspaceResource) {
	if workspaceResource.Init != nil {
		resource.Spec.Init = &workspacev1alpha1.WorkspaceResourceInit{
			Command: workspaceResource.Init.Command,
			Args:    workspaceResource.Init.Args,
		}
	}
}

func setVolumeMounts(resource *workspacev1alpha1.ResourceSpec, workspaceResource *models.WorkspaceResource) {
	if len(workspaceResource.VolumeMounts) > 0 {
		resource.Spec.VolumeMounts = make([]workspacev1alpha1.VolumeMount, len(workspaceResource.VolumeMounts))
		for i, volumeMount := range workspaceResource.VolumeMounts {
			resource.Spec.VolumeMounts[i] = workspacev1alpha1.VolumeMount{
				SourceWorkspaceVolume: volumeMount.SourceVolumeID,
				SourceSubPath:         volumeMount.SourceSubPath,
				Destination:           volumeMount.TargetPath,
			}
		}
	}
}

func setPorts(resource *workspacev1alpha1.ResourceSpec, workspaceResource *models.WorkspaceResource) {
	if len(workspaceResource.Ports) > 0 {
		resource.Spec.Ports = make([]workspacev1alpha1.Port, len(workspaceResource.Ports))
		for i, port := range workspaceResource.Ports {
			resource.Spec.Ports[i] = workspacev1alpha1.Port{
				Number:         int32(port.Number),
				ExposeToPublic: port.ExposedToPublic,
				IsHttp:         strings.ToLower(port.Protocol) == "http",
			}
			if len(port.SubdomainPrefix) == 0 {
				resource.Spec.Ports[i].Subdomain = encodeUUIDAndPort(workspaceResource.ID, port.Number)
			} else {
				resource.Spec.Ports[i].Subdomain = port.SubdomainPrefix
			}
		}
	}
}

func setEnvVars(resource *workspacev1alpha1.ResourceSpec, workspaceResource *models.WorkspaceResource) {
	if workspaceResource.ExecutionConfig != nil && len(workspaceResource.ExecutionConfig.Env) > 0 {
		resource.Spec.EnvironmentVariables = make([]workspacev1alpha1.EnvironmentVariables, len(workspaceResource.ExecutionConfig.Env))
		for i, envVar := range workspaceResource.ExecutionConfig.Env {
			resource.Spec.EnvironmentVariables[i] = workspacev1alpha1.EnvironmentVariables{
				Name:  envVar.Name,
				Value: envVar.Value,
			}
		}
	}
}

func encodeUUIDAndPort(uuid string, port int) string {
	// Combine UUID and port into a single string
	input := uuid + ":" + strconv.Itoa(port)

	// Hash the combined string using MD5 (128-bit hash)
	hasher := md5.New()
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)

	// Encode the hash using Base32 (URL-safe) and trim padding
	base32Encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash)

	// Truncate the result to 16 characters for a shorter subdomain
	// Adjust the length as needed for your use case
	if len(base32Encoded) > 16 {
		base32Encoded = base32Encoded[:16]
	}

	return strings.ToLower(base32Encoded)
}
