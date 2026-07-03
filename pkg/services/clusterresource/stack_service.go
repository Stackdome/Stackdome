package clusterresource

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/google/go-containerregistry/pkg/name"
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
		return newError("failed to delete stackCR in cluster", err)
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
	resourceNames := make([]string, len(stack.StackResources))
	for i, sr := range stack.StackResources {
		resourceNames[i] = sr.Name
	}
	sort.Strings(resourceNames)

	stackCR := &corev1alpha1.Stack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      stack.Name,
			Namespace: stack.Namespace,
			Labels: map[string]string{
				corev1alpha1.LabelStackID:   stack.ID,
				corev1alpha1.LabelManagedBy: corev1alpha1.ManagedByStackdome,
			},
		},
		Spec: corev1alpha1.StackSpec{
			ResourceNames: resourceNames,
		},
	}

	for _, sr := range stack.StackResources {
		spec, err := s.buildStackResourceSpec(sr, stack.OrganisationID, stack.Name)
		if err != nil {
			return nil, err
		}
		if hasTLSPorts(spec) {
			if stackCR.Annotations == nil {
				stackCR.Annotations = map[string]string{}
			}
			stackCR.Annotations[corev1alpha1.ClusterIssuerAnnotation] = models.DefaultClusterIssuerName
			break
		}
	}

	return stackCR, nil
}

func (s *clusterStackService) buildStackResourceSpec(stackResource *models.StackResource, orgID, stackName string) (*corev1alpha1.StackResourceSpec, error) {
	workloadType := corev1alpha1.WorkloadType(stackResource.WorkloadType)
	if workloadType == "" {
		workloadType = corev1alpha1.WorkloadTypeService
	}
	spec := &corev1alpha1.StackResourceSpec{
		WorkloadType: workloadType,
		Schedule:     stackResource.Schedule,
		Replicas:     stackResource.Replicas,
		DependsOn:    stackResource.DependsOn,
	}

	if stackResource.ExecutionConfig != nil {
		spec.Command = stackResource.ExecutionConfig.Command
		spec.Args = stackResource.ExecutionConfig.Args
	}

	if stackResource.LifecycleConfig != nil && stackResource.LifecycleConfig.RestartRequestTime != nil {
		spec.RestartRequest = &metav1.Time{Time: stackResource.LifecycleConfig.RestartRequestTime.UTC()}
	}

	if err := s.setBuildSpec(spec, stackResource, orgID, stackName); err != nil {
		return nil, err
	}
	if err := s.setImageSpec(spec, stackResource); err != nil {
		return nil, err
	}
	setInitSpec(spec, stackResource)
	setVolumeMounts(spec, stackResource)
	setPorts(spec, stackResource)
	if err := s.setEnvVars(spec, stackResource); err != nil {
		return nil, err
	}
	return spec, nil
}

func hasTLSPorts(spec *corev1alpha1.StackResourceSpec) bool {
	for _, port := range spec.Ports {
		if port.TLS {
			return true
		}
	}
	return false
}

func (s *clusterStackService) setBuildSpec(spec *corev1alpha1.StackResourceSpec, stackResource *models.StackResource, orgID, stackName string) error {
	if stackResource.BuildConfig != nil {
		buildSourceCtx, err := s.buildBuildSourceContext(stackResource.BuildConfig.SourceContext)
		if err != nil {
			return err
		}

		spec.BuildSpec = &corev1alpha1.StackResourceBuildSpec{
			SourceContext:  *buildSourceCtx,
			SourceRevision: buildBuildSourceRevision(stackResource.BuildConfig.SourceRevision),
			BuildContext:   stackResource.BuildConfig.ContextPathWithinSource,
			DockerFilePath: stackResource.BuildConfig.DockerfilePath,
		}

		if stackResource.BuildConfig.BuildImageRepository.UseInClusterRegistry {
			spec.BuildSpec.Repository = corev1alpha1.ImageRepositorySpec{
				ClusterRegistryRef: &corev1.LocalObjectReference{
					Name: stackResource.BuildConfig.BuildImageRepository.ClusterRegistryName,
				},
				Repository: fmt.Sprintf("%s/%s/%s", orgID, stackName, stackResource.Name),
			}
		} else {
			var opts []name.Option
			if stackResource.BuildConfig.BuildImageRepository.InsecureRegistry {
				opts = append(opts, name.Insecure)
			}
			repo, err := name.NewRepository(stackResource.BuildConfig.BuildImageRepository.ExternalImageRef, opts...)
			if err != nil {
				return fmt.Errorf("failed to parse image repository URL: %w", err)
			}

			spec.BuildSpec.Repository = corev1alpha1.ImageRepositorySpec{
				External: &corev1alpha1.ExternalRegistrySpec{
					Host: repo.Registry.RegistryStr(),
				},
				Repository: repo.RepositoryStr(),
			}

			if stackResource.BuildConfig.BuildImageRepository.InsecureRegistry {
				spec.BuildSpec.Repository.External.TLS = &corev1alpha1.RegistryTLSSpec{Insecure: true}
			}

			if stackResource.BuildConfig.RegistrySecretRef != nil {
				secret, err := s.secretService.InternalGetByID(context.Background(), stackResource.BuildConfig.RegistrySecretRef.SecretID)
				if err != nil {
					return fmt.Errorf("failed to get registry secret: %w", err)
				}
				spec.BuildSpec.Repository.Auth = &corev1alpha1.RegistryCredentialsSpec{
					Basic: &corev1alpha1.BasicAuthCredentials{
						SecretRef: corev1.SecretReference{
							Name: secret.ClusterSecretName(),
						},
						UsernameKey: models.UsernameSecretKey,
						PasswordKey: models.PasswordSecretKey,
					},
				}
			}
		}
	}
	return nil
}

func buildBuildSourceRevision(sourceRevision models.BuildSourceRevision) corev1alpha1.SourceRevisionSpec {
	if sourceRevision.Volume != nil {
		return corev1alpha1.SourceRevisionSpec{
			Volume: &corev1alpha1.VolumeRevision{
				RevisionString: sourceRevision.Volume.CurrentVolumeHash,
			},
		}
	} else if sourceRevision.Git != nil {
		res := corev1alpha1.SourceRevisionSpec{
			GitRepo: &corev1alpha1.GitRepoRevision{},
		}
		if sourceRevision.Git.Branch != "" {
			res.GitRepo.Branch = sourceRevision.Git.Branch
		}
		if sourceRevision.Git.Tag != "" {
			res.GitRepo.Tag = sourceRevision.Git.Tag
		}
		if sourceRevision.Git.Commit != "" {
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

func (s *clusterStackService) setImageSpec(spec *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) error {
	if stackResource.ImageConfig != nil {
		spec.ImageSpec = &corev1alpha1.ImageSpec{
			Image: stackResource.ImageConfig.Image,
		}

		if stackResource.ImageConfig.PullSecretRef != nil {
			secret, err := s.secretService.InternalGetByID(context.Background(), stackResource.ImageConfig.PullSecretRef.SecretID)
			if err != nil {
				return fmt.Errorf("failed to get image pull secret: %w", err)
			}
			spec.ImageSpec.PullAuth = &corev1alpha1.RegistryAuth{
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

func (s *clusterStackService) setEnvVars(spec *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) error {
	if stackResource.ExecutionConfig == nil || len(stackResource.ExecutionConfig.Env) == 0 {
		return nil
	}
	spec.EnvironmentVariables = make([]corev1alpha1.EnvironmentVariable, len(stackResource.ExecutionConfig.Env))
	for i, envVar := range stackResource.ExecutionConfig.Env {
		ev := corev1alpha1.EnvironmentVariable{
			Name: envVar.Name,
		}
		if envVar.SecretKeyRef != nil {
			ev.ValueFrom = &corev1alpha1.EnvVarSource{
				SecretKeyRef: corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: envVar.SecretKeyRef.SecretName,
					},
					Key: envVar.SecretKeyRef.Key,
				},
			}
		} else {
			ev.Value = envVar.Value
		}
		spec.EnvironmentVariables[i] = ev
	}
	return nil
}

func setInitSpec(spec *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) {
	if stackResource.Init != nil {
		spec.Init = &corev1alpha1.InitSpec{
			Command: stackResource.Init.Command,
			Args:    stackResource.Init.Args,
		}
	}
}

func setVolumeMounts(spec *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) {
	if len(stackResource.VolumeMounts) > 0 {
		spec.VolumeMounts = make([]corev1alpha1.VolumeMount, len(stackResource.VolumeMounts))
		for i, volumeMount := range stackResource.VolumeMounts {
			spec.VolumeMounts[i] = corev1alpha1.VolumeMount{
				SourceVolume:  volumeMount.SourceVolumeName,
				SourceSubPath: volumeMount.SourceSubPath,
				Destination:   volumeMount.TargetPath,
			}
		}
	}
}

func setPorts(spec *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) {
	if len(stackResource.Ports) > 0 {
		spec.Ports = make([]corev1alpha1.Port, len(stackResource.Ports))
		for i, port := range stackResource.Ports {
			spec.Ports[i] = corev1alpha1.Port{
				Name:           port.Name,
				Number:         int32(port.Number),
				Protocol:       strings.ToLower(port.Protocol),
				ExposeToPublic: port.ExposedToPublic,
			}
			if port.ExposedToPublic {
				spec.Ports[i].FQDN = port.ExposedFqdn
			}
		}
	}
}
