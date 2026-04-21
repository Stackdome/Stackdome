package stack

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/clients"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator"
)

type organisationDomainService interface {
	ListByOrganisationID(ctx context.Context, orgID string) ([]*models.OrganisationDomain, *errors.ServiceError)
}

type secretService interface {
	ValidateImageRegistrySecretForStackResource(ctx context.Context, secretID string) *errors.ServiceError
	ValidateGitSecretForStackResource(ctx context.Context, secretID string) *errors.ServiceError
	ValidateSecretHasKeys(ctx context.Context, secretID string, requiredKeys []string) (bool, []string, *errors.ServiceError)
	ValidateSecretExists(ctx context.Context, secretID string) (bool, *errors.ServiceError)
	InternalGetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError)
}

// Add only validations that take reasonable time to complete.
// Avoid validations that require network calls or long-running operations.
// Long running validations should be handled in the stack worker.
type stackValidator struct {
	interpolationValidator validator.InterpolationValidation
	domainService          organisationDomainService
	secretService          secretService
}

type StackValidatorSpec struct {
	DomainService organisationDomainService
	SecretService secretService
}

func NewStackValidator(
	spec StackValidatorSpec,
) validator.StackValidator {
	return &stackValidator{
		interpolationValidator: NewInterpolationValidation(),
		domainService:          spec.DomainService,
		secretService:          spec.SecretService,
	}
}

func (v *stackValidator) ValidateForCreate(ctx context.Context, spec *models.Stack) *errors.ServiceError {
	if err := v.validateUniqueResourceNames(spec); err != nil {
		return err
	}
	if err := v.validateImageSource(spec); err != nil {
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
	if err := v.validateBuildSourceVolumes(spec); err != nil {
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

	if err := v.validateUniqueResourceNames(spec); err != nil {
		return err
	}
	if err := v.validateImageSource(spec); err != nil {
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
	if err := v.validateBuildSourceVolumes(spec); err != nil {
		return err
	}
	return nil
}

func (v *stackValidator) validateUniqueResourceNames(spec *models.Stack) *errors.ServiceError {
	seen := make(map[string]struct{}, len(spec.StackResources))
	for _, r := range spec.StackResources {
		if _, exists := seen[r.Name]; exists {
			return errors.BadRequest("duplicate stack resource name '%s'", r.Name)
		}
		seen[r.Name] = struct{}{}
	}
	return nil
}

func (v *stackValidator) validateImageSource(spec *models.Stack) *errors.ServiceError {
	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]

		// Validate that resource has exactly one config type
		if err := v.validateResourceConfigType(currentResource); err != nil {
			return err
		}

		// Validate image config if present
		if currentResource.ImageConfig != nil {
			if err := v.validateImageConfig(currentResource); err != nil {
				return err
			}
		}

		// Validate build config if present
		if currentResource.BuildConfig != nil {
			if err := v.validateBuildConfig(currentResource); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *stackValidator) validateResourceConfigType(resource *models.StackResource) *errors.ServiceError {
	hasBuild := resource.BuildConfig != nil
	hasImage := resource.ImageConfig != nil

	if hasBuild && hasImage {
		return errors.BadRequest("stack resource '%s' cannot have both build and image config", resource.Name)
	}
	if !hasBuild && !hasImage {
		return errors.BadRequest("stack resource '%s' must have either build or image config", resource.Name)
	}
	return nil
}

func (v *stackValidator) validateImageConfig(resource *models.StackResource) *errors.ServiceError {
	if err := resource.ImageConfig.Validate(); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid image config: %s", resource.Name, err.Error())
	}

	if resource.ImageConfig.PullSecretRef != nil {
		if err := v.validateImageWithSecret(resource); err != nil {
			return err
		}
	}

	return nil
}

func (v *stackValidator) validateImageWithSecret(resource *models.StackResource) *errors.ServiceError {
	secretRef := resource.ImageConfig.PullSecretRef

	if secretRef.SecretID == "" {
		return errors.BadRequest("stack resource '%s' has empty pull secret ID", resource.Name)
	}

	if err := v.secretService.ValidateImageRegistrySecretForStackResource(context.Background(), secretRef.SecretID); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid pull secret: %s", resource.Name, err.Error())
	}

	return nil
}

func (v *stackValidator) validateImageAnonymously(resource *models.StackResource) *errors.ServiceError {
	client, err := clients.NewRegistryClientAnonymous()
	if err != nil {
		return errors.GeneralError("failed to create anonymous registry client for stack resource '%s': %s", resource.Name, err.Error())
	}

	return v.checkImageExists(client, resource, "does not exist or is not pullable")
}

func (v *stackValidator) checkImageExists(client clients.RegistryClient, resource *models.StackResource, errorSuffix string) *errors.ServiceError {
	exists, err := client.CheckImage(context.Background(), resource.ImageConfig.Image)
	if err != nil {
		return errors.GeneralError("failed to check image for stack resource '%s': %s", resource.Name, err.Error())
	}
	if !exists {
		return errors.BadRequest("stack resource '%s' image '%s' %s", resource.Name, resource.ImageConfig.Image, errorSuffix)
	}
	return nil
}

func (v *stackValidator) validateBuildConfig(resource *models.StackResource) *errors.ServiceError {
	if err := resource.BuildConfig.Validate(); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid build config: %s", resource.Name, err.Error())
	}

	if resource.BuildConfig.SourceContext.Git != nil {
		return v.validateGitSource(resource)
	}

	return nil
}

func (v *stackValidator) validateGitSource(resource *models.StackResource) *errors.ServiceError {
	git := resource.BuildConfig.SourceContext.Git

	if git.GitSecretRef != nil {
		return v.validateGitWithSecret(resource)
	}

	return nil
}

func (v *stackValidator) validateGitWithSecret(resource *models.StackResource) *errors.ServiceError {
	secretRef := resource.BuildConfig.SourceContext.Git.GitSecretRef

	if secretRef.SecretID == "" {
		return errors.BadRequest("stack resource '%s' has empty git secret ID", resource.Name)
	}

	if err := v.secretService.ValidateGitSecretForStackResource(context.Background(), secretRef.SecretID); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid git secret: %s", resource.Name, err.Error())
	}

	return nil
}

func (v *stackValidator) validateBuildSourceVolumes(spec *models.Stack) *errors.ServiceError {
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
		if currentResource.ExecutionConfig == nil || (currentResource.ExecutionConfig.Env == nil && currentResource.ExecutionConfig.EnvVarsFromSecrets == nil) {
			continue
		}
		currentEnvVars := currentResource.ExecutionConfig.Env
		keys := make(map[string]struct{})
		for _, envVar := range currentEnvVars {
			if len(envVar.Name) == 0 {
				return errors.BadRequest("stack resource '%s' has empty env var name", currentResource.Name)
			}
			if len(envVar.Value) == 0 {
				return errors.BadRequest("stack resource '%s' has empty env var value", currentResource.Name)
			}
			if _, exists := keys[envVar.Name]; exists {
				return errors.BadRequest("stack resource '%s' has duplicate env var name '%s'", currentResource.Name, envVar.Name)
			}
			keys[envVar.Name] = struct{}{}
		}

		keys = make(map[string]struct{})
		currentEnvVarsFromSecrets := currentResource.ExecutionConfig.EnvVarsFromSecrets
		for _, envVarFromSecret := range currentEnvVarsFromSecrets {
			if err := v.validateEnvVarFromSecret(currentResource, envVarFromSecret); err != nil {
				return err
			}
			if _, exists := keys[envVarFromSecret.EnvName]; exists {
				return errors.BadRequest("stack resource '%s' has duplicate env var from secret name '%s'", currentResource.Name, envVarFromSecret.EnvName)
			}
			keys[envVarFromSecret.EnvName] = struct{}{}
		}
	}

	if err := v.interpolationValidator.ValidateStackInterpolations(spec); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid interpolation: %s", spec.Name, err.Error())
	}

	return nil
}

func (v *stackValidator) validateEnvVarFromSecret(currentResource *models.StackResource, envVarFromSecret models.EnvSecretReference) *errors.ServiceError {
	if len(envVarFromSecret.SecretID) == 0 {
		return errors.BadRequest("stack resource '%s' has empty secret ID in env var from secret", currentResource.Name)
	}
	if len(envVarFromSecret.SecretKey) == 0 {
		return errors.BadRequest("stack resource '%s' has empty secret key in env var from secret", currentResource.Name)
	}
	if len(envVarFromSecret.EnvName) == 0 {
		return errors.BadRequest("stack resource '%s' has empty environment variable name in env var from secret", currentResource.Name)
	}

	// Validate that the secret exists
	secretExists, err := v.secretService.ValidateSecretExists(context.Background(), envVarFromSecret.SecretID)
	if err != nil {
		return errors.GeneralError("failed to validate secret existence for stack resource '%s': %s", currentResource.Name, err.Error())
	}
	if !secretExists {
		return errors.BadRequest("stack resource '%s' references non-existent secret '%s'", currentResource.Name, envVarFromSecret.SecretID)
	}
	hasKeys, _, err := v.secretService.ValidateSecretHasKeys(context.Background(), envVarFromSecret.SecretID, []string{envVarFromSecret.SecretKey})
	if err != nil {
		return errors.GeneralError("failed to validate secret keys in environment variables for stack resource '%s': %s", currentResource.Name, err.Error())
	}
	if !hasKeys {
		return errors.BadRequest("stack resource '%s' references secret '%s' but is missing required key '%s'", currentResource.Name, envVarFromSecret.SecretID, envVarFromSecret.SecretKey)
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
