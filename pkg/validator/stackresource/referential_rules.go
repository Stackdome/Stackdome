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
// DB lookups only - never the network. A non-nil ServiceError means a
// lookup failed for a reason other than not-found and validation was
// aborted.
func (v *validator) validateReferences(ctx context.Context, stack *models.Stack, resource *models.StackResource) ([]errors.FieldError, *errors.ServiceError) {
	var errs []errors.FieldError

	volumeErrs, serr := v.validateMountedVolumes(ctx, stack, resource)
	if serr != nil {
		return nil, serr
	}
	errs = append(errs, volumeErrs...)

	secretErrs, serr := v.validateEnvSecrets(ctx, stack, resource)
	if serr != nil {
		return nil, serr
	}
	errs = append(errs, secretErrs...)

	credErrs, serr := v.validateCredentialRefs(ctx, stack, resource)
	if serr != nil {
		return nil, serr
	}
	errs = append(errs, credErrs...)

	domainErrs, serr := v.validateExposedPortDomain(ctx, stack, resource)
	if serr != nil {
		return nil, serr
	}
	errs = append(errs, domainErrs...)

	return errs, nil
}

func (v *validator) validateMountedVolumes(ctx context.Context, stack *models.Stack, resource *models.StackResource) ([]errors.FieldError, *errors.ServiceError) {
	if v.volumes == nil {
		return nil, nil
	}
	var errs []errors.FieldError
	for i, m := range resource.VolumeMounts {
		if m.SourceVolumeName == "" && m.SourceVolumeID == "" {
			continue // shape error already reported by input rules
		}
		ok, serr := v.volumeExists(ctx, stack.Namespace, m.SourceVolumeName, m.SourceVolumeID)
		if serr != nil {
			return nil, serr
		}
		if !ok {
			errs = append(errs, fieldErr(fmt.Sprintf("volume_mounts[%d].source_volume", i), errors.VErrVolumeNotFound,
				"volume '%s' does not exist", volumeRef(m.SourceVolumeName, m.SourceVolumeID)))
		}
	}
	if resource.BuildConfig != nil && resource.BuildConfig.SourceContext.Volume != nil {
		src := resource.BuildConfig.SourceContext.Volume
		if src.SourceVolumeName != "" || src.SourceVolumeID != "" {
			ok, serr := v.volumeExists(ctx, stack.Namespace, src.SourceVolumeName, src.SourceVolumeID)
			if serr != nil {
				return nil, serr
			}
			if !ok {
				errs = append(errs, fieldErr("source.volume", errors.VErrVolumeNotFound,
					"build source volume '%s' does not exist", volumeRef(src.SourceVolumeName, src.SourceVolumeID)))
			}
		}
	}
	return errs, nil
}

// volumeExists resolves a volume by name (preferred) or, when the name is
// empty, by ID. Returns false (no error) on a 404 from either lookup path;
// any other error is propagated so the caller aborts validation.
func (v *validator) volumeExists(ctx context.Context, namespace, name, id string) (bool, *errors.ServiceError) {
	if name != "" {
		if _, serr := v.volumes.GetByVolumeNameAndNamespace(ctx, name, namespace); serr != nil {
			if serr.Is404() {
				return false, nil
			}
			return false, serr
		}
		return true, nil
	}
	var (
		vol  *models.Volume
		serr *errors.ServiceError
	)
	if vol, serr = v.volumes.GetByID(ctx, id); serr != nil {
		if serr.Is404() {
			return false, nil
		}
		return false, serr
	}
	if vol.Namespace != namespace {
		// If the volume is not in the same namespace as the resource, it is not valid.
		return false, nil
	}
	return true, nil
}

// volumeRef returns the identifier to surface in a not-found message,
// preferring the human-readable name when present.
func volumeRef(name, id string) string {
	if name != "" {
		return name
	}
	return id
}

func (v *validator) validateEnvSecrets(ctx context.Context, stack *models.Stack, resource *models.StackResource) ([]errors.FieldError, *errors.ServiceError) {
	if v.secrets == nil || resource.ExecutionConfig == nil {
		return nil, nil
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
			if !serr.Is404() {
				return nil, serr
			}
			errs = append(errs, fieldErr(fmt.Sprintf("execution_config.env[%d].secret_key_ref.secret_name", i),
				errors.VErrSecretNotFound, "secret '%s' does not exist", ref.SecretName))
		}
	}
	return errs, nil
}

func (v *validator) validateCredentialRefs(ctx context.Context, stack *models.Stack, resource *models.StackResource) ([]errors.FieldError, *errors.ServiceError) {
	var errs []errors.FieldError
	if id := resource.RegistryPullCredentialID(); id != "" && v.credentials != nil {
		if _, serr := v.credentials.RegistryCredentials(ctx, stack.OrganisationID, resource.ImageConfig.Image,
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{RegistryCredentialID: id}); serr != nil {
			if !serr.Is404() {
				return nil, serr
			}
			errs = append(errs, fieldErr("source.image.registry_credentials_id",
				errors.VErrRegistryCredentialNotFound, "registry credential '%s' does not exist", id))
		}
	}
	if id := resource.RegistryPushCredentialID(); id != "" && v.credentials != nil {
		if _, serr := v.credentials.RegistryCredentials(ctx, stack.OrganisationID,
			resource.BuildConfig.BuildImageRepository.ExternalImageRef,
			credentials.RegistryPurposePush, credentials.RegistryAuthSelector{RegistryCredentialID: id}); serr != nil {
			if !serr.Is404() {
				return nil, serr
			}
			errs = append(errs, fieldErr("source.git.push.registry_credentials_id",
				errors.VErrRegistryCredentialNotFound, "registry credential '%s' does not exist", id))
		}
	}

	// Existence check only - deliberately bypasses credentials.Resolver here.
	// The resolver's GitCredentials mints a GitHub App installation token
	// over the network for github_app integrations; this seam is DB-only.
	if id := resource.GitIntegrationID(); id != "" && v.gitIntegrations != nil {
		integration, serr := v.gitIntegrations.InternalGetByID(ctx, id)
		if serr != nil && !serr.Is404() {
			return nil, serr
		}
		if serr != nil || integration.OrganisationID != stack.OrganisationID {
			errs = append(errs, fieldErr("source.git.integration_id",
				errors.VErrGitIntegrationNotFound, "git integration '%s' does not exist", id))
		}
	}
	return errs, nil
}

func (v *validator) validateExposedPortDomain(ctx context.Context, stack *models.Stack, resource *models.StackResource) ([]errors.FieldError, *errors.ServiceError) {
	if v.domains == nil {
		return nil, nil
	}
	exposedIdx := -1
	for i, p := range resource.Ports {
		if p.ExposedToPublic {
			exposedIdx = i
			break
		}
	}
	if exposedIdx == -1 {
		return nil, nil
	}
	domains, serr := v.domains.ListByOrganisationID(ctx, stack.OrganisationID)
	if serr != nil && !serr.Is404() {
		return nil, serr
	}
	if len(domains) == 0 {
		return []errors.FieldError{fieldErr(fmt.Sprintf("ports[%d]", exposedIdx), errors.VErrDomainNotConfigured,
			"organisation has no domain configured; cannot expose ports to public")}, nil
	}
	return nil, nil
}
