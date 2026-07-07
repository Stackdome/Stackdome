package stackresource

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

// validateReferences checks that everything the resource points at outside
// itself actually exists: volumes, secrets, credentials, org domains.
// DB lookups only — never the network.
func (v *validator) validateReferences(ctx context.Context, stack *models.Stack, resource *models.StackResource) []errors.FieldError {
	var errs []errors.FieldError
	errs = append(errs, v.validateMountedVolumes(ctx, stack, resource)...)
	errs = append(errs, v.validateEnvSecrets(ctx, stack, resource)...)
	errs = append(errs, v.validateCredentialRefs(ctx, stack, resource)...)
	errs = append(errs, v.validateExposedPortDomain(ctx, stack, resource)...)
	return errs
}

func (v *validator) validateMountedVolumes(ctx context.Context, stack *models.Stack, resource *models.StackResource) []errors.FieldError {
	if v.volumes == nil {
		return nil
	}
	var errs []errors.FieldError
	for i, m := range resource.VolumeMounts {
		if m.SourceVolumeName == "" {
			continue // shape error already reported by input rules
		}
		if _, serr := v.volumes.GetByVolumeNameAndNamespace(ctx, m.SourceVolumeName, stack.Namespace); serr != nil {
			errs = append(errs, fieldErr(fmt.Sprintf("volume_mounts[%d].source_volume", i), errors.VErrVolumeNotFound,
				"volume '%s' does not exist", m.SourceVolumeName))
		}
	}
	if resource.BuildConfig != nil && resource.BuildConfig.SourceContext.Volume != nil {
		name := resource.BuildConfig.SourceContext.Volume.SourceVolumeName
		if name != "" {
			if _, serr := v.volumes.GetByVolumeNameAndNamespace(ctx, name, stack.Namespace); serr != nil {
				errs = append(errs, fieldErr("source.volume", errors.VErrVolumeNotFound,
					"build source volume '%s' does not exist", name))
			}
		}
	}
	return errs
}

func (v *validator) validateEnvSecrets(ctx context.Context, stack *models.Stack, resource *models.StackResource) []errors.FieldError {
	if v.secrets == nil || resource.ExecutionConfig == nil {
		return nil
	}
	var errs []errors.FieldError
	checked := map[string]bool{}
	for i, env := range resource.ExecutionConfig.Env {
		ref := env.SecretKeyRef
		if ref == nil || ref.SecretName == "" || checked[ref.SecretName] {
			continue
		}
		checked[ref.SecretName] = true
		if _, serr := v.secrets.GetByName(ctx, stack.OrganisationID, ref.SecretName); serr != nil {
			errs = append(errs, fieldErr(fmt.Sprintf("execution_config.env[%d].secret_key_ref.secret_name", i),
				errors.VErrSecretNotFound, "secret '%s' does not exist", ref.SecretName))
		}
	}
	return errs
}

func (v *validator) validateCredentialRefs(ctx context.Context, stack *models.Stack, resource *models.StackResource) []errors.FieldError {
	if v.credentials == nil {
		return nil
	}
	var errs []errors.FieldError

	if id := resource.RegistryPullCredentialID(); id != "" {
		if _, serr := v.credentials.RegistryCredentials(ctx, stack.OrganisationID, resource.ImageConfig.Image,
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{RegistryCredentialID: id}); serr != nil {
			errs = append(errs, fieldErr("source.image.registry_credentials_id",
				errors.VErrRegistryCredentialNotFound, "registry credential '%s' does not exist", id))
		}
	}
	if id := resource.RegistryPushCredentialID(); id != "" {
		if _, serr := v.credentials.RegistryCredentials(ctx, stack.OrganisationID,
			resource.BuildConfig.BuildImageRepository.ExternalImageRef,
			credentials.RegistryPurposePush, credentials.RegistryAuthSelector{RegistryCredentialID: id}); serr != nil {
			errs = append(errs, fieldErr("source.git.push.registry_credentials_id",
				errors.VErrRegistryCredentialNotFound, "registry credential '%s' does not exist", id))
		}
	}
	if id := resource.GitIntegrationID(); id != "" {
		if _, serr := v.credentials.GitCredentials(ctx, stack.OrganisationID,
			resource.BuildConfig.SourceContext.Git.RepoURL,
			credentials.GitAuthSelector{IntegrationID: id}); serr != nil {
			errs = append(errs, fieldErr("source.git.integration_id",
				errors.VErrGitIntegrationNotFound, "git integration '%s' does not exist", id))
		}
	}
	return errs
}

func (v *validator) validateExposedPortDomain(ctx context.Context, stack *models.Stack, resource *models.StackResource) []errors.FieldError {
	if v.domains == nil {
		return nil
	}
	exposedIdx := -1
	for i, p := range resource.Ports {
		if p.ExposedToPublic {
			exposedIdx = i
			break
		}
	}
	if exposedIdx == -1 {
		return nil
	}
	domains, serr := v.domains.ListByOrganisationID(ctx, stack.OrganisationID)
	if serr != nil || len(domains) == 0 {
		return []errors.FieldError{fieldErr(fmt.Sprintf("ports[%d]", exposedIdx), errors.VErrDomainNotConfigured,
			"organisation has no domain configured; cannot expose ports to public")}
	}
	return nil
}
