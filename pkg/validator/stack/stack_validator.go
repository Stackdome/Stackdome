package stack

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type OrganisationDomainService interface {
	ListByOrganisationID(ctx context.Context, orgID string) ([]*models.OrganisationDomain, *errors.ServiceError)
}

type StackValidator interface {
	ValidateForCreate(ctx context.Context, spec *models.Stack) *errors.ServiceError
	ValidateForUpdate(ctx context.Context, existing *models.Stack, spec *models.Stack) *errors.ServiceError
}

type stackValidator struct {
	interpolationValidator InterpolationValidation
	domainService          OrganisationDomainService
}

func NewStackValidator(
	domainService OrganisationDomainService,
) StackValidator {
	return &stackValidator{
		interpolationValidator: NewInterpolationValidation(),
		domainService:          domainService,
	}
}

func (v *stackValidator) ValidateForCreate(ctx context.Context, spec *models.Stack) *errors.ServiceError {
	if err := v.validateImageToRun(spec); err != nil {
		return err
	}
	if err := v.validateStackEnvVars(spec); err != nil {
		return err
	}
	if err := v.validateStackPorts(spec); err != nil {
		return err
	}
	if err := v.validateVolumeMounts(spec); err != nil {
		return err
	}
	if err := v.validateDomainExistence(ctx, spec); err != nil {
		return err
	}
	if err := v.validateBuildConfig(spec); err != nil {
		return err
	}

	return nil
}

func (v *stackValidator) ValidateForUpdate(ctx context.Context, existing *models.Stack, spec *models.Stack) *errors.ServiceError {
	// Validate immutable fields
	if spec.Name != existing.Name {
		return errors.BadRequest("stack name cannot be updated")
	}
	if spec.UserID != existing.UserID {
		return errors.BadRequest("stack user cannot be updated")
	}
	if spec.OrganisationID != existing.OrganisationID {
		return errors.BadRequest("stack organisation cannot be updated")
	}

	if err := v.validateImageToRun(spec); err != nil {
		return err
	}
	if err := v.validateStackEnvVars(spec); err != nil {
		return err
	}
	if err := v.validateStackPorts(spec); err != nil {
		return err
	}
	if err := v.validateVolumeMounts(spec); err != nil {
		return err
	}
	if err := v.validateDomainExistence(ctx, spec); err != nil {
		return err
	}
	if err := v.validateBuildConfig(spec); err != nil {
		return err
	}
	return nil
}

func (v *stackValidator) validateImageToRun(spec *models.Stack) *errors.ServiceError {
	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		if currentResource.BuildConfig != nil && currentResource.ImageConfig != nil {
			return errors.BadRequest("stack resource '%s' cannot have both build and image config", currentResource.Name)
		}
		if currentResource.ImageConfig == nil && currentResource.BuildConfig == nil {
			return errors.BadRequest("stack resource '%s' must have either build or image config", currentResource.Name)
		}
		if currentResource.ImageConfig != nil {
			if err := currentResource.ImageConfig.Validate(); err != nil {
				return errors.BadRequest("stack resource '%s' has invalid image config: %s", currentResource.Name, err.Error())
			}
		}

		if currentResource.BuildConfig != nil {
			if err := currentResource.BuildConfig.Validate(); err != nil {
				return errors.BadRequest("stack resource '%s' has invalid build config: %s", currentResource.Name, err.Error())
			}
		}
	}
	return nil
}

func (v *stackValidator) validateBuildConfig(spec *models.Stack) *errors.ServiceError {
	// Populate the volume mounts with the source volume names.
	definedVolumesMap := spec.VolumesMap()
	for i := range spec.StackResources {
		spec.StackResources[i].UserID = spec.UserID
		if spec.StackResources[i].BuildConfig != nil {
			buildConfig := spec.StackResources[i].BuildConfig
			if buildConfig.SourceContext.Volume != nil {
				volume, found := definedVolumesMap[buildConfig.SourceContext.Volume.SourceVolumeName]
				if !found {
					return errors.BadRequest("volume '%s' does not exist", buildConfig.SourceContext.Volume.SourceVolumeName)
				}
				buildConfig.SourceContext.Volume.SourceVolumeName = volume.Name
			}
		}
	}
	return nil
}

func (v *stackValidator) validateVolumeMounts(spec *models.Stack) *errors.ServiceError {
	if len(spec.Volumes) == 0 && spec.HasVolumeMounts() {
		return errors.BadRequest("stack '%s' has volume mounts but no volumes defined", spec.Name)
	}
	if !spec.HasVolumeMounts() {
		return nil
	}

	definedVolumes := spec.Volumes
	definedVolumesMap := make(map[string]*models.Volume)
	for i := range definedVolumes {
		definedVolumesMap[definedVolumes[i].Name] = definedVolumes[i]
	}

	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		for j := range spec.StackResources[i].VolumeMounts {
			currentVolumeMount := currentResource.VolumeMounts[j]
			if _, found := definedVolumesMap[currentVolumeMount.SourceVolumeName]; !found {
				return errors.BadRequest("volume '%s' does not exist", currentVolumeMount.SourceVolumeName)
			}
		}
	}
	return nil
}

func (v *stackValidator) validateStackEnvVars(spec *models.Stack) *errors.ServiceError {
	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		if currentResource.ExecutionConfig == nil || currentResource.ExecutionConfig.Env == nil {
			continue
		}
		currentEnvVars := currentResource.ExecutionConfig.Env

		for _, envVar := range currentEnvVars {
			if len(envVar.Name) == 0 {
				return errors.BadRequest("stack resource '%s' has empty env var name", currentResource.Name)
			}
			if len(envVar.Value) == 0 {
				return errors.BadRequest("stack resource '%s' has empty env var value", currentResource.Name)
			}
		}
	}

	if err := v.interpolationValidator.ValidateStackInterpolations(spec); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid interpolation: %s", spec.Name, err.Error())
	}

	return nil
}

func (v *stackValidator) validateDomainExistence(ctx context.Context, spec *models.Stack) *errors.ServiceError {
	if !spec.HasExposedPorts() {
		return nil
	}

	orgDomains, err := v.domainService.ListByOrganisationID(ctx, spec.OrganisationID)
	if err != nil {
		return errors.GeneralError("failed to list domains for organisation '%s': %s", spec.OrganisationID, err.Error())
	}

	if len(orgDomains) == 0 {
		return errors.BadRequest("stack '%s' has publicly exposed ports but no domains defined for organisation '%s'", spec.Name, spec.OrganisationID)
	}
	return nil
}

func (v *stackValidator) validateStackPorts(spec *models.Stack) *errors.ServiceError {
	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		if len(currentResource.Ports) == 0 {
			continue
		}
		currentPorts := currentResource.Ports

		for _, port := range currentPorts {
			if port.Number <= 0 {
				return errors.BadRequest("stack resource '%s' has invalid port number", currentResource.Name)
			}
		}
	}
	return nil
}
