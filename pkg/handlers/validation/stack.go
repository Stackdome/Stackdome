package validation

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
)

func ValidateStack(in *openapi.Stack) Validate {
	return ValidateAll([]Validate{
		validateEmpty(in, "Id", "id"),
		validateEmpty(in, "OrganisationId", "organisation_id"),
		validateEmpty(in, "Status", "status"),
		validateLabels(&in.Labels),
		validateAnnotations(&in.Annotations),
		validateNotEmpty(in, "Name", "name"),
		validateEmpty(in, "Namespace", "namespace"),
		validateStackSpec(in),
	})
}

func validateStackSpec(in *openapi.Stack) Validate {
	return func() *errors.ServiceError {
		if len(in.Spec.StackResources) == 0 {
			return errors.BadRequest("spec.stack_resources: %s", "spec.resources is required")
		}
		for _, wr := range in.Spec.StackResources {
			if err := validateStackResource(&wr, in); err != nil {
				return err
			}
		}
		return nil
	}
}

func validateStackResource(in *openapi.StackResource, stack *openapi.Stack) *errors.ServiceError {
	if err := validateEmpty(in, "Id", "id")(); err != nil {
		return err
	}
	if err := validateNotEmpty(in, "Name", "name")(); err != nil {
		return err
	}

	if err := validateLabels(&in.Labels)(); err != nil {
		return err
	}
	if err := validateAnnotations(&in.Annotations)(); err != nil {
		return err
	}

	if in.BuildSpec == nil && in.ImageSpec == nil {
		return errors.BadRequest("build_spec or image_spec is required")
	}

	if in.BuildSpec != nil {
		if err := validateBuildConfig(in.BuildSpec); err != nil {
			return err
		}
	}
	if in.ImageSpec != nil {
		if err := validateImageConfig(in.ImageSpec); err != nil {
			return err
		}
	}

	if in.InitSpec != nil {
		if err := validateInitConfig(in.InitSpec); err != nil {
			return err
		}
	}

	if in.ExecutionConfig != nil {
		if err := validateExecutionConfig(in.ExecutionConfig); err != nil {
			return err
		}
	}

	if len(in.VolumeMounts) != 0 {
		if err := validateVolumeMounts(in.VolumeMounts); err != nil {
			return err
		}
	}

	if len(in.DependsOn) != 0 {
		if err := validateDependencies(in, stack); err != nil {
			return err
		}
	}

	if in.LifecycleConfig != nil {
		if err := validateLifecycleConfig(in.LifecycleConfig); err != nil {
			return err
		}
	}
	if len(in.Ports) != 0 {
		if err := validatePorts(in.Ports); err != nil {
			return err
		}
	}
	return nil
}

func validateBuildConfig(in *openapi.StackResourceBuildSpec) *errors.ServiceError {
	if in == nil {
		return errors.BadRequest("stack_resource_build_spec: %s", "stack_resource_build_spec is required")
	}
	if in.ContextPathWithinSource == "" {
		return errors.BadRequest("stack_resource_build_spec.context_path_within_source: %s", "stack_resource_build_spec.context_path_within_source is required")
	}
	if in.DockerfilePath == "" {
		return errors.BadRequest("stack_resource_build_spec.dockerfile_path: %s", "stack_resource_build_spec.dockerfile_path is required")
	}

	if in.ImageRepository.GetExternalImageRepoUrl() == "" && in.ImageRepository.UseInternalRegistry == nil {
		return errors.BadRequest("stack_resource_build_spec.image_repository.external_image_repo_url or stack_resource_build_spec.image_repository.use_external_registry is required")
	}

	if in.ImageRepository.GetExternalImageRepoUrl() == "" && !in.ImageRepository.GetUseInternalRegistry() {
		// If external_image_repo_url is empty, we need to check if the in-cluster registry is set to true
		// If it is not, we need to return an error
		return errors.BadRequest("if external_image_repo_url is empty, use_internal_registry must be true")
	}

	if in.ImageRepository.GetExternalImageRepoUrl() != "" && in.ImageRepository.GetUseInternalRegistry() {
		// If external_image_repo_url is not empty, we need to check if the in-cluster registry is set to false
		// If it is not, we need to return an error
		return errors.BadRequest("if external_image_repo_url is not empty, use_internal_registry must be false")
	}

	if err := validateBuildSourceContext(in.SourceContext); err != nil {
		return err
	}
	if err := validateBuildSourceRevision(in.SourceRevision); err != nil {
		return err
	}
	return nil
}

func validateBuildSourceRevision(in openapi.BuildSourceRevision) *errors.ServiceError {
	if in.VolumeSourceRevision == nil && in.GitRepoRevision == nil {
		return errors.BadRequest("stack_resource_build_spec.source_revision: %s", "stack_resource_build_spec.source_revision is required")
	}

	setValues := 0

	if in.VolumeSourceRevision != nil {
		setValues++
	}
	if in.GitRepoRevision != nil {
		setValues++
	}
	if setValues > 1 {
		return errors.BadRequest(
			"stack_resource_build_spec.source_revision: %s",
			"stack_resource_build_spec.source_revision can only have one of volume_source_revision or git_repo_revision",
		)
	}

	if in.VolumeSourceRevision != nil {
		if in.VolumeSourceRevision.CurrentVolumeHash == "" {
			return errors.BadRequest(
				"stack_resource_build_spec.source_revision.volume_source_revision.current_volume_hash: %s",
				"stack_resource_build_spec.source_revision.volume_source_revision.current_volume_hash is required",
			)
		}
	}
	if in.GitRepoRevision != nil {
		return validateGitRepoRevision(in.GitRepoRevision)
	}
	return nil
}

func validateBuildSourceContext(in openapi.BuildSourceContext) *errors.ServiceError {
	if in.Volume == nil && in.GitRepo == nil {
		return errors.BadRequest("stack_resource_build_spec.source_context: %s", "stack_resource_build_spec.source_context is required")
	}

	setValues := 0

	if in.Volume != nil {
		setValues++
	}
	if in.GitRepo != nil {
		setValues++
	}
	if setValues > 1 {
		return errors.BadRequest("stack_resource_build_spec.source_context: %s", "stack_resource_build_spec.source_context can only have one of volume or git_repo")
	}

	if in.Volume != nil {
		if in.Volume.Id == "" {
			return errors.BadRequest("stack_resource_build_spec.source_context.volume.id: %s", "stack_resource_build_spec.source_context.volume.id is required")
		}
	}
	if in.GitRepo != nil {
		if in.GitRepo.RepoUrl == "" {
			return errors.BadRequest("stack_resource_build_spec.source_context.git.repo_url: %s", "stack_resource_build_spec.source_context.git.repo_url is required")
		}
	}
	return nil
}

func validateImageConfig(in *openapi.ImageSpec) *errors.ServiceError {
	if in == nil {
		return errors.BadRequest("image_spec: %s", "image_spec is required")
	}
	if in.Image == "" {
		return errors.BadRequest("image_spec.image: %s", "image_spec.image is required")
	}
	return nil
}

// TODO: implement validation for init config
func validateInitConfig(in *openapi.InitSpec) *errors.ServiceError {
	return nil
}

func validateExecutionConfig(in *openapi.ExecutionConfig) *errors.ServiceError {
	return nil
}

func validateVolumeMounts(in []openapi.VolumeMount) *errors.ServiceError {
	return nil
}

func validateDependencies(in *openapi.StackResource, stack *openapi.Stack) *errors.ServiceError {
	peerResources := make(map[string]struct{})
	for _, wr := range stack.Spec.StackResources {
		peerResources[wr.Name] = struct{}{}
	}

	for _, dep := range in.DependsOn {
		if _, ok := peerResources[dep]; !ok {
			return errors.BadRequest("depends_on: %s", "depends_on resource not found")
		}
	}
	return nil
}

func validateLifecycleConfig(in *openapi.LifecycleConfig) *errors.ServiceError {
	return nil
}

func validatePorts(in []openapi.Port) *errors.ServiceError {
	return nil
}
