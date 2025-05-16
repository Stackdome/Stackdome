package clusterresource

import (
	"context"
	"fmt"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

type ClusterStackService interface {
	CreateStackInCluster(ctx context.Context, stack *models.Stack) *ClusterResourceError
	DeleteStackInCluster(ctx context.Context, stack *models.Stack) *ClusterResourceError
	UpdateStackInCluster(ctx context.Context, stack *models.Stack) *ClusterResourceError
}

type clusterStackService struct {
	clusterService      DBClusterService
	clusterManager      clustermanager.ClusterManager
	organisationService DBOrganisationService
	logger              logger.Logger
}

type ClusterStackServiceSpec struct {
	ClusterService      DBClusterService
	OrganisationService DBOrganisationService
	UserService         DBUserService
	ClusterManager      clustermanager.ClusterManager
	Logger              logger.Logger
}

func NewClusterStackService(spec ClusterStackServiceSpec) ClusterStackService {
	return &clusterStackService{
		clusterService:      spec.ClusterService,
		clusterManager:      spec.ClusterManager,
		organisationService: spec.OrganisationService,
		logger:              spec.Logger,
	}
}

func (s *clusterStackService) CreateStackInCluster(ctx context.Context, stack *models.Stack) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, stack.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	clusterClient, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		s.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	desiredCR := s.desiredObjectInCluster(stack)

	if err := clusterClient.Create(ctx, desiredCR); err != nil {
		s.logger.Errorf("failed to create workspaceCR in cluster: %v", err)
		return newError("failed to create workspaceCR in cluster", err)
	}
	return nil
}

func (s *clusterStackService) DeleteStackInCluster(ctx context.Context, stack *models.Stack) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, stack.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	clusterClient, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		s.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	stackCr := &corev1alpha1.Stack{}
	if err := clusterClient.Get(ctx, client.ObjectKey{
		Name:      stack.Name,
		Namespace: stack.Namespace,
	}, stackCr); err != nil {
		if k8sapierrors.IsNotFound(err) {
			s.logger.Warn(ctx, "stack '%s' not found in cluster", stack.ID)
			return nil
		}
		s.logger.Errorf("failed to get stackCR in cluster: %v", err)
		return newError("failed to get stackCR in cluster", err)
	}

	if err := clusterClient.Delete(ctx, stackCr); err != nil {
		if k8sapierrors.IsNotFound(err) {
			s.logger.Warn(ctx, "stack '%s' not found in cluster", stack.ID)
			return nil
		}
		s.logger.Errorf("failed to delete stackCR in cluster: %v", err)
		return newError("failed to delete stackCR`1 in cluster", err)
	}
	return nil
}

func (s *clusterStackService) UpdateStackInCluster(ctx context.Context, stack *models.Stack) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, stack.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	clusterClient, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		s.logger.Errorf("failed to get cluster client: %v", clientGetErr)
		return newError("failed to get cluster client", clientGetErr)
	}

	desiredstackCR := s.desiredObjectInCluster(stack)
	var existingstackCR corev1alpha1.Stack

	if err := clusterClient.Get(ctx, client.ObjectKeyFromObject(desiredstackCR), &existingstackCR); err != nil {
		s.logger.Errorf("failed to get stackCR in cluster: %v", err)
		return newError("failed to get stackCR in cluster", err)
	}

	desiredstackCR.ResourceVersion = existingstackCR.ResourceVersion
	if err := clusterClient.Update(ctx, desiredstackCR); err != nil {
		s.logger.Errorf("failed to update workspaceCR in cluster: %v", err)
		return newError("failed to update workspaceCR in cluster", err)
	}
	return nil
}

func (s *clusterStackService) desiredObjectInCluster(stack *models.Stack) *corev1alpha1.Stack {
	stackCR := &corev1alpha1.Stack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      stack.Name,
			Namespace: stack.Namespace,
			Labels: map[string]string{
				models.StackIDLabel:           stack.ID,
				models.ObjectServerGeneration: fmt.Sprintf("%d", stack.Version),
			},
		},
		Spec: corev1alpha1.StackSpec{
			StackResources: desiredStackResources(stack),
		},
	}
	return stackCR
}

func desiredStackResources(stack *models.Stack) []corev1alpha1.StackResourceTemplate {
	var resources []corev1alpha1.StackResourceTemplate
	for _, stackResource := range stack.StackResources {
		resource := corev1alpha1.StackResourceTemplate{
			Name: stackResource.Name,
			Spec: corev1alpha1.StackResourceSpec{
				StateFul:  stackResource.StateFul,
				DependsOn: stackResource.DependsOn,
			},
		}

		if stackResource.ExecutionConfig != nil {
			resource.Spec.Command = stackResource.ExecutionConfig.Command
			resource.Spec.Args = stackResource.ExecutionConfig.Args
		}

		if stackResource.LifecycleConfig != nil && stackResource.LifecycleConfig.RestartRequestTime != nil {
			resource.Spec.RestartRequest = &metav1.Time{Time: stackResource.LifecycleConfig.RestartRequestTime.UTC()}
		}

		setBuildSpec(&resource, stackResource)
		setImageSpec(&resource, stackResource)
		setInitSpec(&resource, stackResource)
		setVolumeMounts(&resource, stackResource)
		setPorts(&resource, stackResource)
		setEnvVars(&resource, stackResource)
		resources = append(resources, resource)
	}
	return resources
}

func setBuildSpec(resourceTemplateCr *corev1alpha1.StackResourceTemplate, stackResource *models.StackResource) {
	if stackResource.BuildConfig != nil {
		resourceTemplateCr.Spec.BuildSpec = &corev1alpha1.StackResourceBuildSpec{
			SourceContext:  buildBuildSourceContext(stackResource.BuildConfig.SourceContext),
			SourceRevision: buildBuildSourceRevision(stackResource.BuildConfig.SourceRevision),
			BuildContext:   stackResource.BuildConfig.ContextPathWithinSource,
			DockerFilePath: stackResource.BuildConfig.DockerfilePath,
			Registry: corev1alpha1.RegistrySpec{
				RepositoryURL: stackResource.BuildConfig.ImageRepositoryUrl,
				Insecure:      stackResource.BuildConfig.InsecureRegistry,
			},
		}
	}
}

func buildBuildSourceRevision(sourceRevision models.BuildSourceRevision) corev1alpha1.SourceRevisionSpec {
	if sourceRevision.Volume != nil {
		return corev1alpha1.SourceRevisionSpec{
			Volume: &corev1alpha1.VolumeRevision{
				CurrentVolumeHash: sourceRevision.Volume.CurrentVolumeHash,
			},
		}
	} else if sourceRevision.Git != nil {
		res := corev1alpha1.SourceRevisionSpec{
			GitRepo: &corev1alpha1.GitRepoRevision{},
		}
		switch {
		case sourceRevision.Git.Branch != nil:
			res.GitRepo.Branch = &corev1alpha1.GitBranch{
				Name:    sourceRevision.Git.Branch.Name,
				HeadSha: "head",
			}
			if sourceRevision.Git.Branch.HeadSha != "" {
				res.GitRepo.Branch.HeadSha = sourceRevision.Git.Branch.HeadSha
			}
		case len(sourceRevision.Git.Tag) > 0:
			res.GitRepo.Tag = sourceRevision.Git.Tag
		case len(sourceRevision.Git.Commit) > 0:
			res.GitRepo.Commit = sourceRevision.Git.Commit
		}
		return res
	}
	return corev1alpha1.SourceRevisionSpec{}
}

func buildBuildSourceContext(sourceContext models.BuildContextSource) corev1alpha1.BuildContextSource {
	if sourceContext.Volume != nil {
		return corev1alpha1.BuildContextSource{
			Volume: &corev1alpha1.VolumeSource{
				Name: sourceContext.Volume.SourceVolumeName,
			},
		}
	} else if sourceContext.Git != nil {
		return corev1alpha1.BuildContextSource{
			Git: &corev1alpha1.GitRepoSource{
				RepoUrl: sourceContext.Git.RepoURL,
			},
		}
	}
	return corev1alpha1.BuildContextSource{}
}

func setImageSpec(resourceTemplateCr *corev1alpha1.StackResourceTemplate, stackResource *models.StackResource) {
	if stackResource.ImageConfig != nil {
		resourceTemplateCr.Spec.ImageSpec = &corev1alpha1.ImageSpec{
			Image: stackResource.ImageConfig.Image,
		}
	}
}

func setInitSpec(resourceTemplateCr *corev1alpha1.StackResourceTemplate, stackResource *models.StackResource) {
	if stackResource.Init != nil {
		resourceTemplateCr.Spec.Init = &corev1alpha1.InitSpec{
			Command: stackResource.Init.Command,
			Args:    stackResource.Init.Args,
		}
		if stackResource.Init.ImageConfig != nil {
			resourceTemplateCr.Spec.Init.ImageSpec = &corev1alpha1.ImageSpec{
				Image: stackResource.Init.ImageConfig.Image,
			}
		}
	}
}

func setVolumeMounts(resourceTemplateCr *corev1alpha1.StackResourceTemplate, stackResource *models.StackResource) {
	if len(stackResource.VolumeMounts) > 0 {
		resourceTemplateCr.Spec.VolumeMounts = make([]corev1alpha1.VolumeMount, len(stackResource.VolumeMounts))
		for i, volumeMount := range stackResource.VolumeMounts {
			resourceTemplateCr.Spec.VolumeMounts[i] = corev1alpha1.VolumeMount{
				SourceVolume:  volumeMount.SourceVolumeID,
				SourceSubPath: volumeMount.SourceSubPath,
				Destination:   volumeMount.TargetPath,
			}
		}
	}
}

func setPorts(resourceTemplateCr *corev1alpha1.StackResourceTemplate, stackResource *models.StackResource) {
	if len(stackResource.Ports) > 0 {
		resourceTemplateCr.Spec.Ports = make([]corev1alpha1.Port, len(stackResource.Ports))
		for i, port := range stackResource.Ports {
			resourceTemplateCr.Spec.Ports[i] = corev1alpha1.Port{
				Number:         int32(port.Number),
				ExposeToPublic: port.ExposedToPublic,
				IsHttp:         strings.ToLower(port.Protocol) == "http",
			}
			if port.ExposedToPublic {
				resourceTemplateCr.Spec.Ports[i].FQDN = port.ExposedFqdn
			}
		}
	}
}

func setEnvVars(resourceTemplateCr *corev1alpha1.StackResourceTemplate, stackResource *models.StackResource) {
	if stackResource.ExecutionConfig != nil && len(stackResource.ExecutionConfig.Env) > 0 {
		resourceTemplateCr.Spec.EnvironmentVariables = make([]corev1alpha1.EnvironmentVariables, len(stackResource.ExecutionConfig.Env))
		for i, envVar := range stackResource.ExecutionConfig.Env {
			resourceTemplateCr.Spec.EnvironmentVariables[i] = corev1alpha1.EnvironmentVariables{
				Name:  envVar.Name,
				Value: envVar.Value,
			}
		}
	}
}
