package clusterresource

import (
	"context"
	"fmt"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	corev1 "k8s.io/api/core/v1"
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
	secretService       DBSecretService
	logger              logger.Logger
}

type ClusterStackServiceSpec struct {
	ClusterService      DBClusterService
	OrganisationService DBOrganisationService
	UserService         DBUserService
	ClusterManager      clustermanager.ClusterManager
	SecretService       DBSecretService
	Logger              logger.Logger
}

func NewClusterStackService(spec ClusterStackServiceSpec) ClusterStackService {
	return &clusterStackService{
		clusterService:      spec.ClusterService,
		clusterManager:      spec.ClusterManager,
		organisationService: spec.OrganisationService,
		secretService:       spec.SecretService,
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

	desiredCR, berr := s.desiredObjectInCluster(stack)
	if berr != nil {
		return newError("failed to build desired stackCR", berr)
	}

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

	desiredstackCR, berr := s.desiredObjectInCluster(stack)
	if berr != nil {
		return newError("failed to build desired stackCR", berr)
	}
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

func (s *clusterStackService) desiredObjectInCluster(stack *models.Stack) (*corev1alpha1.Stack, error) {
	stackCR := &corev1alpha1.Stack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      stack.Name,
			Namespace: stack.Namespace,
			Labels: map[string]string{
				models.StackIDLabel:           stack.ID,
				models.ObjectServerGeneration: stack.CrRevision,
			},
		},
		Spec: corev1alpha1.StackSpec{},
	}

	stackResources, err := s.desiredStackResources(stack)
	if err != nil {
		return nil, err
	}
	stackCR.Spec.StackResources = stackResources
	return stackCR, nil
}

func (s *clusterStackService) desiredStackResources(stack *models.Stack) ([]corev1alpha1.StackResourceTemplate, error) {
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

		if err := s.setBuildSpec(&resource, stackResource); err != nil {
			return nil, err
		}
		if err := s.setImageSpec(&resource, stackResource); err != nil {
			return nil, err
		}
		setInitSpec(&resource, stackResource)
		setVolumeMounts(&resource, stackResource)
		setPorts(&resource, stackResource)
		if err := s.setEnvVars(&resource, stackResource); err != nil {
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func (s *clusterStackService) setBuildSpec(resourceTemplateCr *corev1alpha1.StackResourceTemplate, stackResource *models.StackResource) error {
	if stackResource.BuildConfig != nil {
		buildSourceCtx, err := s.buildBuildSourceContext(stackResource.BuildConfig.SourceContext)
		if err != nil {
			return err
		}

		resourceTemplateCr.Spec.BuildSpec = &corev1alpha1.StackResourceBuildSpec{
			SourceContext:  *buildSourceCtx,
			SourceRevision: buildBuildSourceRevision(stackResource.BuildConfig.SourceRevision),
			BuildContext:   stackResource.BuildConfig.ContextPathWithinSource,
			DockerFilePath: stackResource.BuildConfig.DockerfilePath,
			Registry: corev1alpha1.RegistrySpec{
				RepositoryURL: stackResource.BuildConfig.ImageRepositoryUrl,
			},
		}
		if stackResource.BuildConfig.BuildImageRepository.UseInClusterRegistry {
			resourceTemplateCr.Spec.BuildSpec.Registry.Insecure = true
		}
		if stackResource.BuildConfig.RegistrySecretRef != nil {
			secret, err := s.secretService.InternalGetByID(context.Background(), stackResource.BuildConfig.RegistrySecretRef.SecretID)
			if err != nil {
				return fmt.Errorf("failed to get registry secret: %w", err)
			}
			resourceTemplateCr.Spec.BuildSpec.Registry.Auth = &corev1alpha1.RegistryAuth{
				DockerConfigAuth: &corev1alpha1.DockerConfigAuth{
					SecretKey: corev1.DockerConfigJsonKey,
					SecretRef: &corev1.SecretReference{
						Name: secret.ClusterSecretName(),
					},
				},
			}
		}
	}
	return nil
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

func (s *clusterStackService) buildBuildSourceContext(sourceContext models.BuildContextSource) (*corev1alpha1.BuildContextSource, error) {
	if sourceContext.Volume != nil {
		return &corev1alpha1.BuildContextSource{
			Volume: &corev1alpha1.VolumeSource{
				Name: sourceContext.Volume.SourceVolumeName,
			},
		}, nil
	} else if sourceContext.Git != nil {
		if sourceContext.Git.GitSecretRef != nil {
			secret, err := s.secretService.InternalGetByID(context.Background(), sourceContext.Git.GitSecretRef.SecretID)
			if err != nil {
				return nil, err
			}
			return &corev1alpha1.BuildContextSource{
				Git: &corev1alpha1.GitRepoSource{
					RepoUrl: sourceContext.Git.RepoURL,
					Auth: &corev1alpha1.GitAuth{
						UsernamePasswordAuthRef: &corev1alpha1.CredentialSecretKeyPair{
							SecretRef: corev1.SecretReference{
								Name: secret.ClusterSecretName(),
							},
							UsernameKey: models.UsernameSecretKey,
							PasswordKey: models.PasswordSecretKey,
						},
					},
				},
			}, nil
		}
		return &corev1alpha1.BuildContextSource{
			Git: &corev1alpha1.GitRepoSource{
				RepoUrl: sourceContext.Git.RepoURL,
			},
		}, nil
	}
	return nil, nil
}

func (s *clusterStackService) setImageSpec(resourceTemplateCr *corev1alpha1.StackResourceTemplate, stackResource *models.StackResource) error {
	if stackResource.ImageConfig != nil {
		resourceTemplateCr.Spec.ImageSpec = &corev1alpha1.ImageSpec{
			Image: stackResource.ImageConfig.Image,
		}

		if stackResource.ImageConfig.PullSecretRef != nil {
			secret, err := s.secretService.InternalGetByID(context.Background(), stackResource.ImageConfig.PullSecretRef.SecretID)
			if err != nil {
				return fmt.Errorf("failed to get image pull secret: %w", err)
			}
			resourceTemplateCr.Spec.ImageSpec.PullAuth = &corev1alpha1.RegistryAuth{
				DockerConfigAuth: &corev1alpha1.DockerConfigAuth{
					SecretKey: corev1.DockerConfigJsonKey,
					SecretRef: &corev1.SecretReference{
						Name: secret.ClusterSecretName(),
					},
				},
			}
		}
	}
	return nil
}

func (s *clusterStackService) setEnvVars(resourceTemplateCr *corev1alpha1.StackResourceTemplate, stackResource *models.StackResource) error {
	if stackResource.ExecutionConfig != nil && len(stackResource.ExecutionConfig.Env) > 0 {
		resourceTemplateCr.Spec.EnvironmentVariables = make([]corev1alpha1.EnvironmentVariables, len(stackResource.ExecutionConfig.Env))
		for i, envVar := range stackResource.ExecutionConfig.Env {
			resourceTemplateCr.Spec.EnvironmentVariables[i] = corev1alpha1.EnvironmentVariables{
				Name:  envVar.Name,
				Value: envVar.Value,
			}
		}
	}

	return nil
}

func setInitSpec(resourceTemplateCr *corev1alpha1.StackResourceTemplate, stackResource *models.StackResource) {
	if stackResource.Init != nil {
		resourceTemplateCr.Spec.Init = &corev1alpha1.InitSpec{
			Command: stackResource.Init.Command,
			Args:    stackResource.Init.Args,
		}
	}
}

func setVolumeMounts(resourceTemplateCr *corev1alpha1.StackResourceTemplate, stackResource *models.StackResource) {
	if len(stackResource.VolumeMounts) > 0 {
		resourceTemplateCr.Spec.VolumeMounts = make([]corev1alpha1.VolumeMount, len(stackResource.VolumeMounts))
		for i, volumeMount := range stackResource.VolumeMounts {
			resourceTemplateCr.Spec.VolumeMounts[i] = corev1alpha1.VolumeMount{
				SourceVolume:  volumeMount.SourceVolumeName,
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
