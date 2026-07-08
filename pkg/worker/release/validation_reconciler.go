package release

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/Stackdome/stackdome/pkg/clients"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
)

const rateLimitRequeueDelay = time.Minute

// validationReconciler runs the expensive checks (image existence, registry
// pull/push auth) once a release is InProgress, before render. Successful
// probes are remembered by fingerprint so unchanged targets are skipped.
type validationReconciler struct {
	releaseService     releaseService
	credentialResolver credentials.Resolver
	registryClients    registryClientProvider
	validationRecords  stores.ResourceValidationRecordStore
	logger             logger.Logger
}

func newValidationReconciler(spec ReleaseWorkerSpec) *validationReconciler {
	registryClients := spec.RegistryClients
	if registryClients == nil {
		registryClients = defaultRegistryClientProvider{}
	}
	return &validationReconciler{
		releaseService:     spec.ReleaseService,
		credentialResolver: spec.CredentialResolver,
		registryClients:    registryClients,
		validationRecords:  spec.ValidationRecords,
		logger:             logger.NewLoggerWithPrefix(context.Background(), "release-validation"),
	}
}

func (r *validationReconciler) Name() string { return "validation" }

func (r *validationReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	if release.Manifest != nil {
		// Rollbacks carry a pre-rendered manifest; their pins were already
		// validated when the source release went out.
		return resultNil, nil
	}

	var verrs models.ReleaseValidationErrors
	rateLimited := false

	for _, res := range release.Snapshot.Resources {
		resErrs, limited, err := r.validateResource(ctx, release, res)
		if err != nil {
			return resultNil, err
		}
		rateLimited = rateLimited || limited
		verrs = append(verrs, resErrs...)
	}

	if rateLimited && len(verrs) == 0 {
		r.logger.Infof("release %s: registry rate limited, requeueing", release.ID)
		return resultRequeueAfter(rateLimitRequeueDelay), nil
	}

	if len(verrs) > 0 {
		msg := fmt.Sprintf("release validation failed: %d error(s)", len(verrs))
		if _, serr := r.releaseService.MarkFailedWithValidationErrors(ctx, release.ID, msg, verrs); serr != nil {
			return resultNil, fmt.Errorf("failed to mark release failed: %v", serr)
		}
		return resultStop, nil
	}
	return resultNil, nil
}

func (r *validationReconciler) validateResource(ctx context.Context, release *models.StackRelease, res *models.StackResource) (models.ReleaseValidationErrors, bool, error) {
	var verrs models.ReleaseValidationErrors
	rateLimited := false

	if res.ImageConfig != nil && res.ImageConfig.Image != "" && !isClusterLocalRegistryRef(res.ImageConfig.Image) {
		errsHere, limited, err := r.checkImagePull(ctx, release, res)
		if err != nil {
			return nil, false, err
		}
		rateLimited = rateLimited || limited
		verrs = append(verrs, errsHere...)
	}

	if res.BuildConfig != nil {
		repo := res.BuildConfig.BuildImageRepository
		if repo.ExternalImageRef != "" && !repo.InsecureRegistry && !isClusterLocalRegistryRef(repo.ExternalImageRef) {
			errsHere, limited, err := r.checkPushAccess(ctx, release, res)
			if err != nil {
				return nil, false, err
			}
			rateLimited = rateLimited || limited
			verrs = append(verrs, errsHere...)
		}
	}
	return verrs, rateLimited, nil
}

func (r *validationReconciler) checkImagePull(ctx context.Context, release *models.StackRelease, res *models.StackResource) (models.ReleaseValidationErrors, bool, error) {
	imageRef := res.ImageConfig.Image
	credID := res.RegistryPullCredentialID()

	resolved, serr := r.credentialResolver.RegistryCredentials(ctx, release.Snapshot.Stack.OrganisationID, imageRef,
		credentials.RegistryPurposePull, credentials.RegistryAuthSelector{RegistryCredentialID: credID})
	if serr != nil {
		if serr.Is404() {
			return models.ReleaseValidationErrors{{
				ResourceName: res.Name, Field: "source.image.registry_credentials_id",
				Code: errors.VErrRegistryCredentialNotFound, Message: "registry credential not found",
			}}, false, nil
		}
		return nil, false, fmt.Errorf("resource %s: resolve pull credentials: %w", res.Name, serr)
	}

	fp := checkFingerprint(imageRef, credID, resolved.DataHash)
	if r.probeCached(ctx, release.StackID, res.Name, models.ValidationCheckImagePull, fp) {
		return nil, false, nil
	}

	client, err := r.registryClients.ClientFor(resolved)
	if err != nil {
		return nil, false, fmt.Errorf("resource %s: build registry client: %w", res.Name, err)
	}

	exists, err := client.CheckImage(ctx, imageRef)
	switch {
	case stderrors.Is(err, clients.ErrRateLimited):
		return nil, true, nil
	case stderrors.Is(err, clients.ErrAuthFailed):
		return models.ReleaseValidationErrors{{
			ResourceName: res.Name, Field: "source.image.ref",
			Code:    errors.VErrRegistryAuthFailed,
			Message: fmt.Sprintf("registry rejected credentials for image '%s'", imageRef),
		}}, false, nil
	case err != nil:
		// transient network problem: let the worker retry the whole reconcile
		return nil, false, fmt.Errorf("resource %s: image probe: %w", res.Name, err)
	case !exists:
		return models.ReleaseValidationErrors{{
			ResourceName: res.Name, Field: "source.image.ref",
			Code:    errors.VErrImageNotFound,
			Message: fmt.Sprintf("image '%s' does not exist or is not accessible", imageRef),
		}}, false, nil
	}

	r.rememberSuccess(ctx, release.StackID, res.Name, models.ValidationCheckImagePull, fp)
	return nil, false, nil
}

func (r *validationReconciler) checkPushAccess(ctx context.Context, release *models.StackRelease, res *models.StackResource) (models.ReleaseValidationErrors, bool, error) {
	pushRef := res.BuildConfig.BuildImageRepository.ExternalImageRef
	credID := res.RegistryPushCredentialID()

	resolved, serr := r.credentialResolver.RegistryCredentials(ctx, release.Snapshot.Stack.OrganisationID, pushRef,
		credentials.RegistryPurposePush, credentials.RegistryAuthSelector{RegistryCredentialID: credID})
	if serr != nil {
		if serr.Is404() {
			return models.ReleaseValidationErrors{{
				ResourceName: res.Name, Field: "source.git.push.registry_credentials_id",
				Code: errors.VErrRegistryCredentialNotFound, Message: "registry credential not found",
			}}, false, nil
		}
		return nil, false, fmt.Errorf("resource %s: resolve push credentials: %w", res.Name, serr)
	}

	fp := checkFingerprint(pushRef, credID, resolved.DataHash)
	if r.probeCached(ctx, release.StackID, res.Name, models.ValidationCheckPushAccess, fp) {
		return nil, false, nil
	}

	client, err := r.registryClients.ClientFor(resolved)
	if err != nil {
		return nil, false, fmt.Errorf("resource %s: build registry client: %w", res.Name, err)
	}

	err = client.CheckPushAccess(ctx, pushRef)
	switch {
	case stderrors.Is(err, clients.ErrRateLimited):
		return nil, true, nil
	case stderrors.Is(err, clients.ErrAuthFailed):
		return models.ReleaseValidationErrors{{
			ResourceName: res.Name, Field: "source.git.push.repository",
			Code:    errors.VErrPushAccessDenied,
			Message: fmt.Sprintf("cannot push to '%s': registry rejected credentials", pushRef),
		}}, false, nil
	case err != nil:
		// transient network problem: let the worker retry the whole reconcile
		return nil, false, fmt.Errorf("resource %s: push probe: %w", res.Name, err)
	}

	r.rememberSuccess(ctx, release.StackID, res.Name, models.ValidationCheckPushAccess, fp)
	return nil, false, nil
}

func (r *validationReconciler) probeCached(ctx context.Context, stackID, resourceName string, kind models.ResourceValidationCheckKind, fingerprint string) bool {
	if r.validationRecords == nil {
		return false
	}
	rec, serr := r.validationRecords.Get(ctx, stackID, resourceName, kind)
	return serr == nil && rec != nil && rec.Fingerprint == fingerprint
}

func (r *validationReconciler) rememberSuccess(ctx context.Context, stackID, resourceName string, kind models.ResourceValidationCheckKind, fingerprint string) {
	if r.validationRecords == nil {
		return
	}
	if serr := r.validationRecords.Upsert(ctx, &models.ResourceValidationRecord{
		StackID: stackID, ResourceName: resourceName, CheckKind: kind,
		Fingerprint: fingerprint, ValidatedAt: time.Now().UTC(),
	}); serr != nil {
		r.logger.Warnf("failed to record validation success for %s/%s: %v", stackID, resourceName, serr)
	}
}

// checkFingerprint hashes the probe inputs; a matching fingerprint on a past
// success means nothing relevant changed (ref, pinned credential ID, or the
// resolved credential's own data hash) and the probe can be skipped.
func checkFingerprint(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}
