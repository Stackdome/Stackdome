package builders

import (
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"strconv"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

type ClusterResourceBuilder interface {
	BuildStackCR(stack *models.Stack, organisation *models.Organisation) (*corev1alpha1.Stack, error)
	BuildStackResourceCR(stackResource *models.StackResource, organisation *models.Organisation) (*corev1alpha1.StackResource, error)
}

type clusterResourceBuilder struct{}

func NewClusterResourceBuilder() ClusterResourceBuilder {
	return &clusterResourceBuilder{}
}

func (b *clusterResourceBuilder) BuildStackCR(stack *models.Stack, organisation *models.Organisation) (*corev1alpha1.Stack, error) {
	stackCR := &corev1alpha1.Stack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      stack.Name,
			Namespace: stack.Namespace,
			Labels: map[string]string{
				models.StackIDLabel:           stack.ID,
				models.ObjectServerGeneration: fmt.Sprintf("%d", stack.Version),
			},
		},
		Spec: corev1alpha1.StackSpec{},
	}

	stackResourcesTemplates := make([]corev1alpha1.StackResourceTemplate, len(stack.StackResources))
	for i, stackResource := range stack.StackResources {
		stackResourcesTemplates[i] = corev1alpha1.StackResourceTemplate{
			Name: stackResource.Name,
			Spec: b.buildStackResourceSpec(stackResource, organisation),
		}
	}
	stackCR.Spec.StackResources = stackResourcesTemplates
	return stackCR, nil
}

func (b *clusterResourceBuilder) BuildStackResourceCR(stackResource *models.StackResource, organisation *models.Organisation) (*corev1alpha1.StackResource, error) {
	stackResourceCR := &corev1alpha1.StackResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      stackResource.Name,
			Namespace: stackResource.Namespace,
			Labels: map[string]string{
				models.StackIDLabel:           stackResource.StackID,
				models.ObjectServerGeneration: fmt.Sprintf("%d", stackResource.Version),
			},
		},
		Spec: b.buildStackResourceSpec(stackResource, organisation),
	}
	return stackResourceCR, nil
}

func (b *clusterResourceBuilder) buildStackResourceSpec(stackResource *models.StackResource, organisation *models.Organisation) corev1alpha1.StackResourceSpec {
	resourceSpec := corev1alpha1.StackResourceSpec{
		StateFul:  stackResource.StateFul,
		DependsOn: stackResource.DependsOn,
	}

	if stackResource.ExecutionConfig != nil {
		resourceSpec.Command = stackResource.ExecutionConfig.Command
		resourceSpec.Args = stackResource.ExecutionConfig.Args
	}

	if stackResource.LifecycleConfig != nil && stackResource.LifecycleConfig.RestartRequestTime != nil {
		resourceSpec.RestartRequest = &metav1.Time{Time: stackResource.LifecycleConfig.RestartRequestTime.UTC()}
	}

	setBuildSpec(&resourceSpec, stackResource)
	setImageSpec(&resourceSpec, stackResource)
	setInitSpec(&resourceSpec, stackResource)
	setVolumeMounts(&resourceSpec, stackResource)
	setPorts(&resourceSpec, stackResource, organisation)
	setEnvVars(&resourceSpec, stackResource)
	return resourceSpec
}

func setBuildSpec(resourceSpecCr *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) {
	if stackResource.BuildConfig != nil {
		resourceSpecCr.BuildSpec = &corev1alpha1.StackResourceBuildSpec{
			SourceContext:  buildBuildSourceContext(stackResource.BuildConfig.SourceContext),
			SourceRevision: buildBuildSourceRevision(stackResource.BuildConfig.SourceRevision),
			BuildContext:   stackResource.BuildConfig.ContextPathWithinSource,
			DockerFilePath: stackResource.BuildConfig.DockerfilePath,
			Registry: corev1alpha1.RegistrySpec{
				RepositoryURL: stackResource.BuildConfig.ImageRepositoryUrl,
			},
		}
		if stackResource.BuildConfig.BuildImageRepository.InsecureRegistry {
			resourceSpecCr.BuildSpec.Registry.Insecure = stackResource.BuildConfig.BuildImageRepository.InsecureRegistry
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

func setImageSpec(resourceSpecCr *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) {
	if stackResource.ImageConfig != nil {
		resourceSpecCr.ImageSpec = &corev1alpha1.ImageSpec{
			Image: stackResource.ImageConfig.Image,
		}
	}
}

func setInitSpec(resourceSpecCr *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) {
	if stackResource.Init != nil {
		resourceSpecCr.Init = &corev1alpha1.InitSpec{
			Command: stackResource.Init.Command,
			Args:    stackResource.Init.Args,
		}
		if stackResource.Init.ImageConfig != nil {
			resourceSpecCr.Init.ImageSpec = &corev1alpha1.ImageSpec{
				Image: stackResource.Init.ImageConfig.Image,
			}
		}
	}
}

func setVolumeMounts(resourceSpecCr *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) {
	if len(stackResource.VolumeMounts) > 0 {
		resourceSpecCr.VolumeMounts = make([]corev1alpha1.VolumeMount, len(stackResource.VolumeMounts))
		for i, volumeMount := range stackResource.VolumeMounts {
			resourceSpecCr.VolumeMounts[i] = corev1alpha1.VolumeMount{
				SourceVolume:  volumeMount.SourceVolumeID,
				SourceSubPath: volumeMount.SourceSubPath,
				Destination:   volumeMount.TargetPath,
			}
		}
	}
}

func setPorts(resourceSpecCr *corev1alpha1.StackResourceSpec, stackResource *models.StackResource, organisation *models.Organisation) {
	if len(stackResource.Ports) > 0 {
		resourceSpecCr.Ports = make([]corev1alpha1.Port, len(stackResource.Ports))
		for i, port := range stackResource.Ports {
			resourceSpecCr.Ports[i] = corev1alpha1.Port{
				Number:         int32(port.Number),
				ExposeToPublic: port.ExposedToPublic,
				IsHttp:         strings.ToLower(port.Protocol) == "http",
			}
			if port.ExposedToPublic {
				resourceSpecCr.Ports[i].FQDN = port.ExposedFqdn
			}
		}
	}
}

func setEnvVars(resourceSpecCr *corev1alpha1.StackResourceSpec, stackResource *models.StackResource) {
	if stackResource.ExecutionConfig != nil && len(stackResource.ExecutionConfig.Env) > 0 {
		resourceSpecCr.EnvironmentVariables = make([]corev1alpha1.EnvironmentVariables, len(stackResource.ExecutionConfig.Env))
		for i, envVar := range stackResource.ExecutionConfig.Env {
			resourceSpecCr.EnvironmentVariables[i] = corev1alpha1.EnvironmentVariables{
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
	if len(base32Encoded) > 16 {
		base32Encoded = base32Encoded[:16]
	}

	return strings.ToLower(base32Encoded)
}
