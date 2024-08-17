package validation

import (
	goerrors "errors"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation"
)

func ValidateWorkspaceStorage(in *openapi.WorkspaceStorage) Validate {
	return ValidateAll([]Validate{
		validateEmpty(in, "Id", "id"),
		validateEmpty(in, "OrganisationId", "organisation_id"),
		validateEmpty(in, "Status", "status"),
		validateLabels(&in.Labels),
		validateAnnotations(&in.Annotations),
		validateNotEmpty(in, "Name", "name"),
		validateEmpty(in, "Namespace", "namespace"),
		validateWorkspaceStorageSpec(in),
		validateVolumes(in),
		func() *errors.ServiceError {
			if !ValidateName(in.Name) {
				return errors.Validation("name is not a valid name")
			}
			return nil
		},
	})
}
func ValidateWorkspaceStorageUpdate(in *openapi.WorkspaceStorage) Validate {
	return ValidateAll([]Validate{
		validateEmpty(in, "Id", "id"),
		validateEmpty(in, "OrganisationId", "organisation_id"),
		validateEmpty(in, "Status", "status"),
		validateLabels(&in.Labels),
		validateAnnotations(&in.Annotations),
		validateNotEmpty(in, "Name", "name"),
		validateEmpty(in, "Namespace", "namespace"),
		validateWorkspaceStorageSpec(in),
		validateVolumes(in),
		func() *errors.ServiceError {
			if !ValidateName(in.Name) {
				return errors.Validation("name is not a valid name")
			}
			return nil
		},
	})
}

func validateWorkspaceStorageSpec(in *openapi.WorkspaceStorage) Validate {
	return func() *errors.ServiceError {
		if in.Spec.WorkspaceName == "" {
			return errors.Validation("workspace name cannot be empty")
		}
		if err := validateVolumes(in)(); err != nil {
			return errors.Validation("validation error in volumes: %s", err.Error())
		}
		return nil
	}
}

func validateVolumes(in *openapi.WorkspaceStorage) Validate {
	return func() *errors.ServiceError {
		for _, volume := range in.Spec.Volumes {
			if volume.Name == "" {
				return errors.Validation("volume name cannot be empty")
			}
			if !ValidateName(volume.Name) {
				return errors.Validation("volume name is not a valid name")
			}
			if err := validateLabels(&volume.Labels)(); err != nil {
				return errors.Validation("validation error in volume labels: %s", err.Error())
			}
			if err := validateAnnotations(&volume.Annotations)(); err != nil {
				return errors.Validation("validation error in volume annotations: %s", err.Error())
			}
			if err := validateVolumeSpec(volume.Spec); err != nil {
				return errors.Validation("validation error in volume spec: %s", err.Error())
			}
		}
		return nil
	}
}

func validateVolumeSpec(spec openapi.WorkspaceVolumeSpec) error {
	if spec.Size == "" {
		return goerrors.New("size is required")
	}

	if _, err := resource.ParseQuantity(spec.Size); err != nil {
		return goerrors.New("volume size is not a valid quantity")
	}

	if spec.Source != nil {
		switch spec.Source.SourceType {
		case openapi.LOCAL:
			if spec.Source.LocalSource == nil {
				return goerrors.New("local_source is required when source_type is 'Local'")
			}
			if spec.Source.BuildSource != nil {
				return goerrors.New("build_source should not be set when source_type is 'Local'")
			}
		case openapi.BUILD_ARTIFACT:
			if spec.Source.BuildSource == nil || len(spec.Source.BuildSource) == 0 {
				return goerrors.New("build_source is required when source_type is 'BuildArtifact'")
			}
			if spec.Source.LocalSource != nil {
				return goerrors.New("local_source should not be set when source_type is 'BuildArtifact'")
			}
		default:
			return goerrors.New("source_type is not valid")
		}

		if err := validateLocalSource(spec.Source.LocalSource); err != nil {
			return fmt.Errorf("validation error in volume local source: %s", err.Error())
		}
		if err := validateBuildSource(spec.Source.BuildSource); err != nil {
			return fmt.Errorf("validation error in volume build source: %s", err.Error())
		}
	}

	return nil
}

func validateBuildSource(buildSource []openapi.BuildArtifact) error {
	if len(buildSource) == 0 {
		return nil
	}
	for _, source := range buildSource {
		if len(source.ResourceRef) == 0 {
			return errors.Validation("build source resource ref cannot be empty")
		}
		if len(source.DestinationPath) == 0 {
			return errors.Validation("build source destination path cannot be empty")
		}
		if len(source.SourcePath) == 0 {
			return errors.Validation("build source source path cannot be empty")
		}
	}
	return nil
}

func validateLocalSource(localSource *openapi.LocalSource) error {
	if localSource == nil {
		return nil
	}
	if localSource.Path == "" {
		return errors.Validation("local source path cannot be empty")
	}
	return nil
}

func validateLabels(labelsPtr *[]openapi.Label) Validate {
	return func() *errors.ServiceError {
		if labelsPtr == nil {
			return nil
		}
		labels := *labelsPtr
		if len(labels) == 0 {
			return nil
		}
		for _, label := range labels {
			if label.Key == "" {
				return errors.Validation("label key cannot be empty")
			}
			if !ValidateLabelKey(label.Key) {
				return errors.Validation("label key '%s' is not a valid label key", label.Key)
			}
		}
		return nil
	}
}

func validateAnnotations(annotationsPtr *[]openapi.Annotation) Validate {
	return func() *errors.ServiceError {
		if annotationsPtr == nil {
			return nil
		}
		annotations := *annotationsPtr
		if len(annotations) == 0 {
			return nil
		}
		for _, annotation := range annotations {
			if annotation.Key == "" {
				return errors.Validation("annotation key cannot be empty")
			}
			if !ValidateAnnotationKey(annotation.Key) {
				return errors.Validation("annotation key '%s' is not a valid annotation key", annotation.Key)
			}
		}
		return nil
	}
}

// K8s string validations
func ValidateLabelKey(key string) bool {
	errors := validation.IsQualifiedName(key)
	return len(errors) == 0
}

func ValidateAnnotationKey(key string) bool {
	errors := validation.IsQualifiedName(key)
	return len(errors) == 0
}

func ValidateName(name string) bool {
	errors := validation.IsDNS1123Subdomain(name)
	return len(errors) == 0
}

func ValidateNamespace(namespace string) bool {
	errors := validation.IsDNS1123Label(namespace)
	return len(errors) == 0
}
