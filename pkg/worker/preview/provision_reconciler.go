package preview

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
)

type stackfileCacheEntry struct {
	content []byte
	hash    string
}

type provisionReconciler struct {
	previewStackService previewStackService
	previewStackStore   previewStackStore
	configStore         configStore
	stackService        stackService
	releaseService      releaseService
	stackfileCache      *sync.Map
	previewCacheKeys    *sync.Map
	logger              logger.Logger
}

func newProvisionReconciler(spec PreviewWorkerSpec, stackfileCache, previewCacheKeys *sync.Map) *provisionReconciler {
	return &provisionReconciler{
		previewStackService: spec.PreviewStackService,
		previewStackStore:   spec.PreviewStackStore,
		configStore:         spec.ConfigStore,
		stackService:        spec.StackService,
		releaseService:      spec.ReleaseService,
		stackfileCache:      stackfileCache,
		previewCacheKeys:    previewCacheKeys,
		logger:              logger.NewLoggerWithPrefix(context.Background(), "preview-provision"),
	}
}

func (r *provisionReconciler) Name() string { return "provision" }

func (r *provisionReconciler) Reconcile(ctx context.Context, preview *models.PreviewStack) (subReconcilerResult, error) {
	config, sErr := r.configStore.GetByID(ctx, preview.StackPreviewConfigID)
	if sErr != nil {
		return resultNil, fmt.Errorf("failed to get config: %w", sErr)
	}

	content, hash, opErr := r.resolveStackfileContent(ctx, preview, config)
	if opErr != nil {
		if errors.IsRetryable(opErr) {
			return resultNil, opErr
		}
		return r.fail(ctx, preview, opErr.Reason, opErr.Message)
	}

	// Sync path: stack already exists
	if preview.StackID != nil {
		overridesHash := hashOverrides(preview.ImageOverrides)
		needsUpdate, needsRelease := r.needsUpdateOrRelease(preview, hash, overridesHash)
		if !needsRelease {
			// NOOP.
			return resultNil, nil
		}

		var model *models.Stack
		if needsUpdate {
			// Build the stack model from the content
			var buildErr *errors.OperationError
			model, buildErr = r.previewStackService.InternalBuildStackFromContent(ctx, config, preview, content)
			if buildErr != nil {
				if errors.IsRetryable(buildErr) {
					return resultNil, buildErr
				}
				return r.fail(ctx, preview, buildErr.Reason, buildErr.Message)
			}
		}

		if txErr := r.previewStackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
			if needsUpdate {
				if _, sErr := r.stackService.InternalUpdateStack(txCtx, *preview.StackID, model); sErr != nil {
					return sErr
				}
			}

			release, sErr := r.releaseService.InternalCreateRelease(txCtx, *preview.StackID, models.ReleaseCause{
				Kind:   models.ReleaseCausePreviewSync,
				Detail: fmt.Sprintf("PR #%s synced at %s", preview.PRNumber, shortSHA(preview.CommitSHA)),
			})
			if sErr != nil {
				return sErr
			}

			preview.ReconcilerStatus = models.PreviewReconcilerStatus{
				LastAppliedCommitSHA:     preview.CommitSHA,
				LastAppliedStackfileHash: hash,
				LastAppliedOverridesHash: overridesHash,
			}
			if preview.ForceSyncRequestedAt != nil {
				preview.ReconcilerStatus.LastProcessedForceSyncAt = *preview.ForceSyncRequestedAt
			}
			preview.ActiveReleaseID = &release.ID
			preview.Status = models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeploying, Reason: "SyncTriggered"}
			if _, sErr := r.previewStackStore.Update(txCtx, preview); sErr != nil {
				return sErr
			}
			return nil
		}); txErr != nil {
			return resultNil, fmt.Errorf("sync transaction failed: %w", txErr)
		}

		return resultRequeueAfter(convergePollInterval), nil
	}

	// Create path: no stack yet
	overridesHash := hashOverrides(preview.ImageOverrides)
	model, opErr := r.previewStackService.InternalBuildStackFromContent(ctx, config, preview, content)
	if opErr != nil {
		if errors.IsRetryable(opErr) {
			return resultNil, opErr
		}
		return r.fail(ctx, preview, opErr.Reason, opErr.Message)
	}

	var created *models.Stack
	if txErr := r.previewStackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		var sErr *errors.ServiceError
		created, sErr = r.stackService.InternalCreateStack(txCtx, model)
		if sErr != nil {
			return sErr
		}

		release, sErr := r.releaseService.InternalCreateRelease(txCtx, created.ID, models.ReleaseCause{
			Kind:   models.ReleaseCausePreviewSync,
			Detail: fmt.Sprintf("PR #%s opened at %s", preview.PRNumber, shortSHA(preview.CommitSHA)),
		})
		if sErr != nil {
			return sErr
		}

		preview.StackID = &created.ID
		preview.ActiveReleaseID = &release.ID
		preview.ReconcilerStatus = models.PreviewReconcilerStatus{
			LastAppliedCommitSHA:     preview.CommitSHA,
			LastAppliedStackfileHash: hash,
			LastAppliedOverridesHash: overridesHash,
		}
		if preview.ForceSyncRequestedAt != nil {
			preview.ReconcilerStatus.LastProcessedForceSyncAt = *preview.ForceSyncRequestedAt
		}
		preview.Status = models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeploying, Reason: "StackCreated"}
		if _, sErr := r.previewStackStore.Update(txCtx, preview); sErr != nil {
			return sErr
		}
		return nil
	}); txErr != nil {
		return resultNil, fmt.Errorf("provision transaction failed: %w", txErr)
	}

	r.logger.Info(ctx, "preview %s provisioned, stack %s created", preview.ID, created.ID)
	return resultRequeueAfter(convergePollInterval), nil
}

func (r *provisionReconciler) resolveStackfileContent(ctx context.Context, preview *models.PreviewStack, config *models.StackPreviewConfig) ([]byte, string, *errors.OperationError) {
	if preview.StackfileContent != nil {
		content := []byte(*preview.StackfileContent)
		h := sha256.Sum256(content)
		return content, hex.EncodeToString(h[:]), nil
	}

	cacheKey := config.GitRepository.RepoURL + ":" + preview.CommitSHA + ":" + config.StackfilePath
	if cached, ok := r.stackfileCache.Load(cacheKey); ok {
		entry := cached.(stackfileCacheEntry)
		return entry.content, entry.hash, nil
	}

	content, hash, opErr := r.previewStackService.InternalFetchStackfile(ctx, config, preview.CommitSHA)
	if opErr != nil {
		return nil, "", opErr
	}

	if oldKey, ok := r.previewCacheKeys.Load(preview.ID); ok {
		r.stackfileCache.Delete(oldKey.(string))
	}
	r.stackfileCache.Store(cacheKey, stackfileCacheEntry{content: content, hash: hash})
	r.previewCacheKeys.Store(preview.ID, cacheKey)
	return content, hash, nil
}

func (r *provisionReconciler) needsUpdateOrRelease(preview *models.PreviewStack, stackFileHash, overridesHash string) (needsUpdate, needsRelease bool) {
	needsUpdate = stackFileHash != preview.ReconcilerStatus.LastAppliedStackfileHash ||
		preview.CommitSHA != preview.ReconcilerStatus.LastAppliedCommitSHA ||
		overridesHash != preview.ReconcilerStatus.LastAppliedOverridesHash
	forceSyncPending := preview.ForceSyncRequestedAt != nil &&
		preview.ForceSyncRequestedAt.After(preview.ReconcilerStatus.LastProcessedForceSyncAt)
	needsRelease = needsUpdate || forceSyncPending
	return needsUpdate, needsRelease
}

func hashOverrides(overrides models.ImageOverrides) string {
	if len(overrides) == 0 {
		return ""
	}
	data, _ := json.Marshal(overrides)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func (r *provisionReconciler) fail(ctx context.Context, preview *models.PreviewStack, reason, message string) (subReconcilerResult, error) {
	preview.Status = models.PreviewStackStatus{
		Phase:   models.PreviewStackPhaseFailed,
		Reason:  reason,
		Message: message,
	}
	if _, sErr := r.previewStackStore.Update(ctx, preview); sErr != nil {
		return resultNil, fmt.Errorf("failed to mark preview %s as failed: %w", preview.ID, sErr)
	}
	r.logger.Error(ctx, "preview %s failed: %s: %s", preview.ID, reason, message)
	return resultStop, nil
}
