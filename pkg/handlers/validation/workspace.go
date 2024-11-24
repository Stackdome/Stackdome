package validation

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
)

func ValidateWorkspace(in *openapi.Workspace) Validate {
	return ValidateAll([]Validate{
		validateEmpty(in, "Id", "id"),
		validateEmpty(in, "OrganisationId", "organisation_id"),
		validateEmpty(in, "Status", "status"),
		validateLabels(&in.Labels),
		validateAnnotations(&in.Annotations),
		validateNotEmpty(in, "Name", "name"),
		validateEmpty(in, "Namespace", "namespace"),
		validateWorkspaceSpec(in),
	})
}

func validateWorkspaceSpec(in *openapi.Workspace) Validate {
	return func() *errors.ServiceError {
		if in.Spec.Resources == nil || len(in.Spec.Resources) == 0 {
			return errors.BadRequest("spec.resources: %s", "spec.resources is required")
		}
		for _, wr := range in.Spec.Resources {
			if err := validateWorkspaceResource(&wr, in); err != nil {
				return err
			}
		}
		return nil
	}
}

func validateWorkspaceResource(in *openapi.WorkspaceResource, workspace *openapi.Workspace) *errors.ServiceError {
	if err := validateEmpty(in, "Id", "id")(); err != nil {
		return err
	}
	if err := validateEmpty(in, "WorkspaceId", "workspace_id")(); err != nil {
		return err
	}
	if err := validateEmpty(in, "Version", "version")(); err != nil {
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

	if in.Build == nil && in.Prebuilt == nil {
		return errors.BadRequest("build or prebuilt is required")
	}

	if in.Build != nil {
		if in.ImageRegistry == nil || len(*in.ImageRegistry) == 0 {
			return errors.BadRequest("image_registry is required")
		}
		if err := validateBuildConfig(in.Build); err != nil {
			return err
		}
	}
	if in.Prebuilt != nil {
		if err := validatePrebuiltConfig(in.Prebuilt); err != nil {
			return err
		}
	}

	if in.Init != nil {
		if err := validateInitConfig(in.Init); err != nil {
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
		if err := validateDependencies(in, workspace); err != nil {
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

func validateBuildConfig(in *openapi.BuildConfig) *errors.ServiceError {
	if in == nil {
		return errors.BadRequest("build: %s", "build is required")
	}
	if in.ContextPath == "" {
		return errors.BadRequest("build.context_path: %s", "build.context_path is required")
	}
	if in.DockerfilePath == "" {
		return errors.BadRequest("build.dockerfile_path: %s", "build.dockerfile_path is required")
	}
	if in.SourceHash == "" {
		return errors.BadRequest("build.source_hash: %s", "build.source_hash is required")
	}
	return nil
}

func validatePrebuiltConfig(in *openapi.PrebuiltConfig) *errors.ServiceError {
	if in == nil {
		return errors.BadRequest("prebuilt: %s", "prebuilt is required")
	}
	if in.Image == "" {
		return errors.BadRequest("prebuilt.image: %s", "prebuilt.image is required")
	}
	return nil
}

// TODO: implement validation for init config
func validateInitConfig(in *openapi.InitConfig) *errors.ServiceError {
	return nil
}

func validateExecutionConfig(in *openapi.ExecutionConfig) *errors.ServiceError {
	return nil
}

func validateVolumeMounts(in []openapi.VolumeMount) *errors.ServiceError {
	return nil
}

func validateDependencies(in *openapi.WorkspaceResource, workspace *openapi.Workspace) *errors.ServiceError {
	peerResources := make(map[string]struct{})
	for _, wr := range workspace.Spec.Resources {
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
