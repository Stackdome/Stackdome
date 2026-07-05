package validation

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
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

	if err := validateSource(in.Source); err != nil {
		return err
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

func validateSource(in *openapi.SourceSpec) *errors.ServiceError {
	if in == nil {
		return errors.BadRequest("source: %s", "source is required")
	}

	setValues := 0
	if in.Git != nil {
		setValues++
	}
	if in.Image != nil {
		setValues++
	}
	if in.Volume != nil {
		setValues++
	}
	if setValues != 1 {
		return errors.BadRequest("source: %s", "exactly one of git, image, or volume must be set")
	}

	switch {
	case in.Git != nil:
		return validateGitSource(in.Git)
	case in.Image != nil:
		return validateImageSource(in.Image)
	default:
		return validateVolumeSource(in.Volume)
	}
}

func validateGitSource(in *openapi.GitSource) *errors.ServiceError {
	if in.RepoUrl == "" {
		return errors.BadRequest("source.git.repo_url: %s", "source.git.repo_url is required")
	}
	if in.GetBranch() != "" && in.GetTag() != "" {
		return errors.BadRequest("source.git: %s", "branch and tag cannot both be set")
	}
	if in.GetCommit() != "" && in.GetBranch() == "" && in.GetTag() == "" {
		return errors.BadRequest("source.git: %s", "commit requires a branch or tag")
	}
	if in.Push != nil {
		if in.Push.Repository == "" {
			return errors.BadRequest("source.git.push.repository: %s", "source.git.push.repository is required")
		}
	}
	return nil
}

func validateImageSource(in *openapi.ImageSource) *errors.ServiceError {
	if in.Ref == "" {
		return errors.BadRequest("source.image.ref: %s", "source.image.ref is required")
	}
	return nil
}

func validateVolumeSource(in *openapi.VolumeBuildSource) *errors.ServiceError {
	if in.GetVolumeId() == "" && in.GetVolumeName() == "" {
		return errors.BadRequest("source.volume: %s", "volume_id or volume_name is required")
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
