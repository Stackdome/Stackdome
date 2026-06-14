package builders

import (
	"context"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/davecgh/go-spew/spew"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
	storagev1alpha1 "stackdome.io/cluster-agent/api/storage/v1alpha1"
)

type ClusterResourceBuilder interface {
	BuildStackCR(stack *models.Stack) (*corev1alpha1.Stack, error)
	GetStackCRHash(stack *models.Stack) (string, error)
	BuildStackResourceCR(stackResource *models.StackResource) (*corev1alpha1.StackResource, error)
	BuildVolumeCR(ctx context.Context, volume *models.Volume) (*storagev1alpha1.Volume, error)
}

type clusterResourceBuilder struct {
	secretService secretFetcher
}

type ClusterResourceBuilderSpec struct {
	SecretService secretFetcher
}

func NewClusterResourceBuilder(spec ClusterResourceBuilderSpec) ClusterResourceBuilder {
	return &clusterResourceBuilder{
		secretService: spec.SecretService,
	}
}

func (b *clusterResourceBuilder) GetStackCRHash(stack *models.Stack) (string, error) {
	stackCR, err := b.BuildStackCR(stack)
	if err != nil {
		return "", fmt.Errorf("failed to build stack CR: %w", err)
	}

	hasher := fnv.New32a()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", stackCR)

	for _, sr := range stack.StackResources {
		srCR, err := b.BuildStackResourceCR(sr)
		if err != nil {
			return "", fmt.Errorf("failed to build stack resource CR for '%s': %w", sr.Name, err)
		}
		printer.Fprintf(hasher, "%#v", srCR)
	}

	return rand.SafeEncodeString(fmt.Sprint(hasher.Sum32())), nil
}

func (b *clusterResourceBuilder) BuildVolumeCR(ctx context.Context, volume *models.Volume) (*storagev1alpha1.Volume, error) {
	volumeCRLabels := volume.Labels.ToMap()
	volumeCRLabels[models.VolumeIDLabel] = volume.ID
	volumeCRLabels[models.CreatedForUserLabel] = volume.UserID
	// storageCRLabels[models.ObjectServerGeneration] = fmt.Sprintf("%d", volume.Version)
	res := storagev1alpha1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        volume.Name,
			Namespace:   volume.Namespace,
			Labels:      volumeCRLabels,
			Annotations: volume.Annotations.ToMap(),
		},
		Spec: storagev1alpha1.VolumeSpec{
			Annotations:        volume.Annotations.ToMap(),
			Labels:             volumeCRLabels,
			Size:               volume.Size,
			StorageClass:       volume.StorageClass,
			AccessMode:         corev1.PersistentVolumeAccessMode(volume.AccessMode),
			NeedsSyncBeforeUse: volume.SyncBeforeUse,
		},
	}

	if volume.VolumeSource != nil {
		switch {
		case volume.VolumeSource.RemoteDirSource != nil:
			res.Spec.Source = &storagev1alpha1.VolumeSource{
				RemoteDir: &storagev1alpha1.RemoteDirSource{
					Path:                 volume.VolumeSource.RemoteDirSource.Path,
					CurrentDirectoryHash: volume.VolumeSource.RemoteDirSource.CurrentDirectoryHash,
				},
			}
		case len(volume.VolumeSource.BuildSource) > 0:
			buildSrcs := make([]storagev1alpha1.BuildArtifactSource, len(volume.VolumeSource.BuildSource))
			for i, buildSrc := range volume.VolumeSource.BuildSource {
				buildSrcs[i] = storagev1alpha1.BuildArtifactSource{
					BuildSource: storagev1alpha1.StackResourceReference{
						Name: buildSrc.ResourceName,
					},
					SourcePath:      buildSrc.SourcePath,
					DestinationPath: buildSrc.DestinationPath,
				}
			}
			res.Spec.Source = &storagev1alpha1.VolumeSource{
				BuildArtifacts: buildSrcs,
			}
		case volume.VolumeSource.GitRepoSource != nil:
			gitRevision := corev1alpha1.GitRepoRevision{}

			switch volume.VolumeSource.GitRepoSource.Revision.Type() {
			case models.Branch:
				gitRevision.Branch = &corev1alpha1.GitBranch{
					Name: volume.VolumeSource.GitRepoSource.Revision.Branch.Name,
				}
			case models.Tag:
				gitRevision.Tag = volume.VolumeSource.GitRepoSource.Revision.Tag
			case models.Commit:
				gitRevision.Commit = volume.VolumeSource.GitRepoSource.Revision.Commit
			default:
				return nil, fmt.Errorf("unknown git revision type: %s", volume.VolumeSource.GitRepoSource.Revision.Type())
			}

			res.Spec.Source = &storagev1alpha1.VolumeSource{
				GitRepo: &storagev1alpha1.GitRepoSource{
					RepoUrl:  volume.VolumeSource.GitRepoSource.RepoUrl,
					Revision: gitRevision,
				},
			}
		}
	}
	return &res, nil
}

func (b *clusterResourceBuilder) BuildStackCR(stack *models.Stack) (*corev1alpha1.Stack, error) {
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
		spec, err := b.buildStackResourceSpec(sr)
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

func (b *clusterResourceBuilder) BuildStackResourceCR(stackResource *models.StackResource) (*corev1alpha1.StackResource, error) {
	stackResourceSpec, err := b.buildStackResourceSpec(stackResource)
	if err != nil {
		return nil, err
	}

	stackResourceCR := &corev1alpha1.StackResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      stackResource.Name,
			Namespace: stackResource.Namespace,
			Labels: map[string]string{
				corev1alpha1.LabelStackID:      stackResource.StackID,
				corev1alpha1.LabelResourceName: stackResource.Name,
				corev1alpha1.LabelResourceID:   stackResource.ID,
				corev1alpha1.LabelManagedBy:    corev1alpha1.ManagedByStackdome,
			},
		},
		Spec: *stackResourceSpec,
	}

	if hasTLSPorts(stackResourceSpec) {
		if stackResourceCR.Annotations == nil {
			stackResourceCR.Annotations = map[string]string{}
		}
		stackResourceCR.Annotations[corev1alpha1.ClusterIssuerAnnotation] = models.DefaultClusterIssuerName
	}

	return stackResourceCR, nil
}

func hasTLSPorts(spec *corev1alpha1.StackResourceSpec) bool {
	for _, port := range spec.Ports {
		if port.TLS {
			return true
		}
	}
	return false
}

func (b *clusterResourceBuilder) buildStackResourceSpec(stackResource *models.StackResource) (*corev1alpha1.StackResourceSpec, error) {
	workloadType := corev1alpha1.WorkloadTypeService
	if stackResource.StateFul {
		workloadType = corev1alpha1.WorkloadTypeStatefulService
	}
	resourceSpec := corev1alpha1.StackResourceSpec{
		WorkloadType: workloadType,
		DependsOn:    stackResource.DependsOn,
	}

	if stackResource.ExecutionConfig != nil {
		resourceSpec.Command = stackResource.ExecutionConfig.Command
		resourceSpec.Args = stackResource.ExecutionConfig.Args
	}

	if stackResource.LifecycleConfig != nil && stackResource.LifecycleConfig.RestartRequestTime != nil {
		resourceSpec.RestartRequest = &metav1.Time{Time: stackResource.LifecycleConfig.RestartRequestTime.UTC()}
	}

	buildSpec, err := b.buildStackResourceBuildSpec(stackResource)
	if err != nil {
		return nil, fmt.Errorf("failed to build stack resource build spec: %w", err)
	}
	resourceSpec.BuildSpec = buildSpec
	imageSpec, err := b.buildStackResourceImageSpec(stackResource)
	if err != nil {
		return nil, fmt.Errorf("failed to build stack resource image spec: %w", err)
	}
	resourceSpec.ImageSpec = imageSpec
	setInitSpec(&resourceSpec, stackResource)
	setVolumeMounts(&resourceSpec, stackResource)
	setPorts(&resourceSpec, stackResource)
	setEnvVars(&resourceSpec, stackResource)
	return &resourceSpec, nil
}

func (b *clusterResourceBuilder) buildStackResourceBuildSpec(stackResource *models.StackResource) (*corev1alpha1.StackResourceBuildSpec, error) {
	if stackResource.BuildConfig != nil {
		buildSourceCtx, err := b.buildBuildSourceContext(stackResource.BuildConfig.SourceContext)
		if err != nil {
			return nil, fmt.Errorf("failed to build source context: %w", err)
		}

		registrySpec, err := b.buildRegistrySpec(stackResource.BuildConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build registry spec: %w", err)
		}
		sourceRevision, err := b.buildBuildSourceRevision(stackResource.BuildConfig.SourceRevision)
		if err != nil {
			return nil, fmt.Errorf("failed to build source revision: %w", err)
		}

		res := &corev1alpha1.StackResourceBuildSpec{
			SourceContext:  *buildSourceCtx,
			SourceRevision: sourceRevision,
			BuildContext:   stackResource.BuildConfig.ContextPathWithinSource,
			DockerFilePath: stackResource.BuildConfig.DockerfilePath,
			Registry:       *registrySpec,
		}
		return res, nil
	}
	return nil, nil
}
func (b *clusterResourceBuilder) buildStackResourceImageSpec(stackResource *models.StackResource) (*corev1alpha1.ImageSpec, error) {
	if stackResource.ImageConfig != nil {
		res := &corev1alpha1.ImageSpec{
			Image: stackResource.ImageConfig.Image,
		}
		if stackResource.ImageConfig.PullSecretRef != nil {
			pullSecret, err := b.secretService.InternalGetByID(context.Background(), stackResource.ImageConfig.PullSecretRef.SecretID)
			if err != nil {
				return nil, fmt.Errorf("failed to get pull secret: %w", err)
			}
			res.PullAuth = &corev1alpha1.RegistryAuth{
				DockerConfigAuth: &corev1alpha1.DockerConfigAuth{
					SecretKey: corev1.DockerConfigJsonKey,
					SecretRef: &corev1.SecretReference{
						Name: pullSecret.ClusterSecretName(),
					},
				},
				//TODO: This is not required. Refactor crd to remove this field.
				Type: corev1alpha1.RegistryAuthTypeDockerHub,
			}
		}
		return res, nil
	}
	return nil, nil
}

func (b *clusterResourceBuilder) buildBuildSourceContext(sourceContext models.BuildContextSource) (*corev1alpha1.BuildContextSource, error) {
	if sourceContext.Volume != nil {
		return &corev1alpha1.BuildContextSource{
			Volume: &corev1alpha1.VolumeSource{
				Name: sourceContext.Volume.SourceVolumeName,
			},
		}, nil
	} else if sourceContext.Git != nil {
		res := corev1alpha1.BuildContextSource{
			Git: &corev1alpha1.GitRepoSource{
				RepoUrl: sourceContext.Git.RepoURL,
			},
		}
		if sourceContext.Git.GitSecretRef != nil {
			gitSecret, err := b.secretService.InternalGetByID(context.Background(), sourceContext.Git.GitSecretRef.SecretID)
			if err != nil {
				return nil, fmt.Errorf("failed to get git secret: %w", err)
			}
			res.Git.Auth = &corev1alpha1.GitAuth{
				UsernamePasswordAuthRef: &corev1alpha1.CredentialSecretKeyPair{
					SecretRef: corev1.SecretReference{
						Name: gitSecret.ClusterSecretName(),
					},
					UsernameKey: models.UsernameSecretKey,
					PasswordKey: models.PasswordSecretKey,
				},
			}
			return &res, nil
		}
		return &res, nil
	}
	return nil, fmt.Errorf("invalid build source context: must specify either volume or git")
}

func (b *clusterResourceBuilder) buildRegistrySpec(buildConfig *models.BuildConfigSpec) (*corev1alpha1.RegistrySpec, error) {
	res := &corev1alpha1.RegistrySpec{
		RepositoryURL: buildConfig.ImageRepositoryUrl,
	}
	if buildConfig.BuildImageRepository.InsecureRegistry {
		res.Insecure = buildConfig.BuildImageRepository.InsecureRegistry
	}

	if buildConfig.RegistrySecretRef != nil {
		pushSecret, err := b.secretService.InternalGetByID(context.Background(), buildConfig.RegistrySecretRef.SecretID)
		if err != nil {
			return nil, fmt.Errorf("failed to get registry secret: %w", err)
		}
		res.Auth = &corev1alpha1.RegistryAuth{
			DockerConfigAuth: &corev1alpha1.DockerConfigAuth{
				SecretKey: corev1.DockerConfigJsonKey,
				SecretRef: &corev1.SecretReference{
					Name: pushSecret.ClusterSecretName(),
				},
			},
		}
	}

	return res, nil
}

func (b *clusterResourceBuilder) buildBuildSourceRevision(sourceRevision models.BuildSourceRevision) (corev1alpha1.SourceRevisionSpec, error) {
	if sourceRevision.Volume != nil {
		return corev1alpha1.SourceRevisionSpec{
			Volume: &corev1alpha1.VolumeRevision{
				CurrentVolumeHash: sourceRevision.Volume.CurrentVolumeHash,
			},
		}, nil
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
		return res, nil
	}
	return corev1alpha1.SourceRevisionSpec{}, fmt.Errorf("invalid build source revision: must specify either volume or git")
}

func setInitSpec(resourceSpecCr *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) {
	if stackResource.Init != nil {
		resourceSpecCr.Init = &corev1alpha1.InitSpec{
			Command: stackResource.Init.Command,
			Args:    stackResource.Init.Args,
		}
	}
}

func setVolumeMounts(resourceSpecCr *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) {
	if len(stackResource.VolumeMounts) > 0 {
		resourceSpecCr.VolumeMounts = make([]corev1alpha1.VolumeMount, len(stackResource.VolumeMounts))
		for i, volumeMount := range stackResource.VolumeMounts {
			resourceSpecCr.VolumeMounts[i] = corev1alpha1.VolumeMount{
				SourceVolume:  volumeMount.SourceVolumeName,
				SourceSubPath: volumeMount.SourceSubPath,
				Destination:   volumeMount.TargetPath,
			}
		}
	}
}

func setPorts(resourceSpecCr *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) {
	if len(stackResource.Ports) > 0 {
		resourceSpecCr.Ports = make([]corev1alpha1.Port, len(stackResource.Ports))
		for i, port := range stackResource.Ports {
			resourceSpecCr.Ports[i] = corev1alpha1.Port{
				Name:           port.Name,
				Number:         int32(port.Number),
				Protocol:       strings.ToLower(port.Protocol),
				ExposeToPublic: port.ExposedToPublic,
				TLS:            port.ExposedToPublic && shouldEnableTLS(port.ExposedFqdn),
			}
			if port.ExposedToPublic {
				resourceSpecCr.Ports[i].FQDN = port.ExposedFqdn
			}
		}
	}
}

func shouldEnableTLS(fqdn string) bool {
	if fqdn == "" {
		return false
	}
	nonTLSDomains := []string{".nip.io", ".sslip.io", ".local", ".localhost"}
	for _, suffix := range nonTLSDomains {
		if strings.HasSuffix(fqdn, suffix) {
			return false
		}
	}
	return true
}

func setEnvVars(resourceSpecCr *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) {
	if stackResource.ExecutionConfig == nil || len(stackResource.ExecutionConfig.Env) == 0 {
		return
	}
	resourceSpecCr.EnvironmentVariables = make([]corev1alpha1.EnvironmentVariable, len(stackResource.ExecutionConfig.Env))
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
		resourceSpecCr.EnvironmentVariables[i] = ev
	}
}
