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

const fieldSourceImageRef = "source.image.ref"

// validationReconciler runs the expensive checks (image existence, registry
// pull/push auth) once a release is InProgress, before render. Successful
// probes are remembered by fingerprint so unchanged targets are skipped.
type validationReconciler struct {
	releaseService     releaseService
	credentialResolver credentials.Resolver
	registryClients    registryClientProvider
	validationRecords  stores.ResourceValidationRecordStore
	eventRecorder      eventRecorder
	logger             logger.Logger
}

func newValidationReconciler(spec ReleaseWorkerSpec) *validationReconciler {
	// RegistryClients is optional: nil substitutes the real (network)
	// registry client provider.
	registryClients := spec.RegistryClients
	if registryClients == nil {
		registryClients = defaultRegistryClientProvider{}
	}
	return &validationReconciler{
		releaseService:     spec.ReleaseService,
		credentialResolver: spec.CredentialResolver,
		registryClients:    registryClients,
		validationRecords:  spec.ValidationRecords,
		eventRecorder:      spec.EventRecorder,
		logger:             logger.NewLoggerWithPrefix(context.Background(), "release-validation"),
	}
}

func (r *validationReconciler) Name() string { return "validation" }

func (r *validationReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	if release.Manifest != nil {
		// Rollbacks carry a pre-rendered manifest; their pins were already
		// validated when the source release went out. No check events are
		// emitted for the rollback path.
		return resultNil, nil
	}

	// checks_started is dedupe-keyed server-side (release:checks_started), so a
	// requeue that re-enters here re-calls harmlessly. Log-only on error: the
	// validation outcome is authoritative, not the event trail.
	if recErr := r.eventRecorder.RecordReleaseChecksStarted(ctx, release); recErr != nil {
		r.logger.Error(ctx, "release %s: failed to record release_checks_started event: %v", release.ID, recErr)
	}

	var verrs models.ReleaseValidationErrors
	var anyRateLimited bool

	for _, res := range release.Snapshot.Resources {
		resErrs, rateLimited, err := r.validateResource(ctx, release, res)
		if err != nil {
			return resultNil, err
		}
		verrs = append(verrs, resErrs...)
		anyRateLimited = anyRateLimited || rateLimited
	}

	if len(verrs) > 0 {
		msg := fmt.Sprintf("release validation failed: %d error(s)", len(verrs))
		// The release service records the per-check and terminal failure
		// events on the CAS win.
		if _, serr := r.releaseService.MarkFailedWithValidationErrors(ctx, release.ID, msg, verrs); serr != nil {
			return resultNil, fmt.Errorf("failed to mark release failed: %w", serr)
		}
		return resultStop, nil
	}

	// checks_passed asserts every check genuinely passed. A rate-limited check
	// is skipped (not verified), so the release still proceeds but we withhold
	// checks_passed until a later re-entry verifies it.
	if !anyRateLimited {
		if recErr := r.eventRecorder.RecordReleaseChecksPassed(ctx, release); recErr != nil {
			r.logger.Error(ctx, "release %s: failed to record release_checks_passed event: %v", release.ID, recErr)
		}
	}
	return resultNil, nil
}

func (r *validationReconciler) validateResource(ctx context.Context, release *models.StackRelease, res *models.StackResource) (models.ReleaseValidationErrors, bool, error) {
	var verrs models.ReleaseValidationErrors
	var rateLimited bool

	if res.ImageConfig != nil && res.ImageConfig.Image != "" && !isClusterLocalRegistryRef(res.ImageConfig.Image) {
		errsHere, rl, err := r.checkImagePull(ctx, release, res)
		if err != nil {
			return nil, false, err
		}
		verrs = append(verrs, errsHere...)
		rateLimited = rateLimited || rl
	}

	if res.BuildConfig != nil {
		repo := res.BuildConfig.BuildImageRepository
		if repo.ExternalImageRef != "" && !repo.InsecureRegistry && !isClusterLocalRegistryRef(repo.ExternalImageRef) {
			errsHere, rl, err := r.checkPushAccess(ctx, release, res)
			if err != nil {
				return nil, false, err
			}
			verrs = append(verrs, errsHere...)
			rateLimited = rateLimited || rl
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
		// Registry rate limits can persist for hours; requeueing would hang
		// the release until the deploy timeout even though the image would
		// most likely deploy fine. Warn and skip the check so the release
		// proceeds; no success fingerprint is recorded since nothing was
		// verified. The skip is reported so checks_passed is withheld.
		r.logger.Warn(ctx, "release %s: resource %s: registry rate limited while checking image '%s'; skipping check", release.ID, res.Name, imageRef)
		return nil, true, nil
	case stderrors.Is(err, clients.ErrAuthFailed) && resolved.Source == credentials.SourceAnonymous:
		// No credentials were resolved for this registry and it rejects
		// anonymous access: the user needs to ADD credentials, which is a
		// different failure from configured credentials being rejected.
		return models.ReleaseValidationErrors{{
			ResourceName: res.Name, Field: fieldSourceImageRef,
			Code:    errors.VErrRegistryCredentialsRequired,
			Message: fmt.Sprintf("image '%s' requires credentials for registry '%s', but none are configured", imageRef, registryHostForRef(imageRef)),
		}}, false, nil
	case stderrors.Is(err, clients.ErrAuthFailed):
		return models.ReleaseValidationErrors{{
			ResourceName: res.Name, Field: fieldSourceImageRef,
			Code:    errors.VErrRegistryAuthFailed,
			Message: fmt.Sprintf("registry '%s' rejected the configured credentials for image '%s'", registryHostForRef(imageRef), imageRef),
		}}, false, nil
	case err != nil:
		// transient network problem: let the worker retry the whole reconcile
		return nil, false, fmt.Errorf("resource %s: image probe: %w", res.Name, err)
	case !exists:
		return models.ReleaseValidationErrors{{
			ResourceName: res.Name, Field: fieldSourceImageRef,
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
		// Same rationale as checkImagePull: skip instead of requeueing so a
		// registry rate limit cannot hang the release to deploy timeout. The
		// skip is reported so checks_passed is withheld.
		r.logger.Warn(ctx, "release %s: resource %s: registry rate limited while checking push access to '%s'; skipping check", release.ID, res.Name, pushRef)
		return nil, true, nil
	case stderrors.Is(err, clients.ErrAuthFailed) && resolved.Source == credentials.SourceAnonymous:
		// Same distinction as checkImagePull: pushing without any resolved
		// credentials means the user must add push credentials, not fix
		// rejected ones.
		return models.ReleaseValidationErrors{{
			ResourceName: res.Name, Field: "source.git.push.repository",
			Code:    errors.VErrRegistryCredentialsRequired,
			Message: fmt.Sprintf("pushing to '%s' requires credentials for registry '%s', but none are configured", pushRef, registryHostForRef(pushRef)),
		}}, false, nil
	case stderrors.Is(err, clients.ErrAuthFailed):
		return models.ReleaseValidationErrors{{
			ResourceName: res.Name, Field: "source.git.push.repository",
			Code:    errors.VErrPushAccessDenied,
			Message: fmt.Sprintf("cannot push to '%s': registry rejected the configured credentials", pushRef),
		}}, false, nil
	case err != nil:
		// transient network problem: let the worker retry the whole reconcile
		return nil, false, fmt.Errorf("resource %s: push probe: %w", res.Name, err)
	}

	r.rememberSuccess(ctx, release.StackID, res.Name, models.ValidationCheckPushAccess, fp)
	return nil, false, nil
}

func (r *validationReconciler) probeCached(ctx context.Context, stackID, resourceName string, kind models.ResourceValidationCheckKind, fingerprint string) bool {
	rec, serr := r.validationRecords.Get(ctx, stackID, resourceName, kind)
	return serr == nil && rec != nil && rec.Fingerprint == fingerprint
}

func (r *validationReconciler) rememberSuccess(ctx context.Context, stackID, resourceName string, kind models.ResourceValidationCheckKind, fingerprint string) {
	if serr := r.validationRecords.Upsert(ctx, &models.ResourceValidationRecord{
		StackID: stackID, ResourceName: resourceName, CheckKind: kind,
		Fingerprint: fingerprint, ValidatedAt: time.Now().UTC(),
	}); serr != nil {
		r.logger.Warn(ctx, "failed to record validation success for %s/%s: %v", stackID, resourceName, serr)
	}
}

// checkFingerprint hashes the probe inputs; a matching fingerprint on a past
// success means nothing relevant changed (ref, pinned credential ID, or the
// resolved credential's own data hash) and the probe can be skipped.
func checkFingerprint(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}
