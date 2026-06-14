package release

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stackdeploy"
	"github.com/ashishmax31/stackdome-api-server/pkg/stackrelease"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker"
)

func (w *releaseWorker) render(ctx context.Context, release *models.StackRelease) (worker.Result, *errors.ServiceError) {
	// Rollback releases carry a pre-rendered manifest; skip render.
	if release.Manifest != nil {
		ok, serr := w.releaseService.MarkApplyingDirect(ctx, release.ID)
		if serr != nil {
			return worker.Result{}, serr
		}
		if !ok {
			return worker.Result{Requeue: true}, nil
		}
		release.State = models.ReleaseStateApplying
		return w.applyAndConverge(ctx, release)
	}

	// CAS transition Pending -> Rendering
	if release.State == models.ReleaseStatePending {
		ok, serr := w.releaseService.MarkRendering(ctx, release.ID)
		if serr != nil {
			return worker.Result{}, serr
		}
		if !ok {
			return worker.Result{Requeue: true}, nil
		}
	}

	// Reconstruct draft from snapshot.
	draft := release.Snapshot.ToStack()

	// Resolve connections to produce the effective stack.
	effective, err := w.resolver.Resolve(ctx, draft)
	if err != nil {
		var depErr stackdeploy.DependencyNotReadyError
		if isNotReady(err, &depErr) {
			w.Logger().Infof("release %s: dependency not ready, requeueing: %s", release.ID, depErr.Message)
			return worker.Result{RequeueAfter: convergencePollInterval}, nil
		}
		w.fail(ctx, release, fmt.Sprintf("render failed: %v", err))
		return worker.Result{}, nil
	}

	// Build CRs.
	stackCR, err := w.crBuilder.BuildStackCR(effective)
	if err != nil {
		w.fail(ctx, release, fmt.Sprintf("failed to build stack CR: %v", err))
		return worker.Result{}, nil
	}

	stackCRBytes, err := json.Marshal(stackCR)
	if err != nil {
		w.fail(ctx, release, fmt.Sprintf("failed to marshal stack CR: %v", err))
		return worker.Result{}, nil
	}

	resourceCRs := make(map[string]json.RawMessage)
	resourceNames := make([]string, 0, len(effective.StackResources))
	resourceRevisions := make(map[string]string)

	for _, sr := range effective.StackResources {
		srCR, buildErr := w.crBuilder.BuildStackResourceCR(sr)
		if buildErr != nil {
			w.fail(ctx, release, fmt.Sprintf("failed to build CR for resource '%s': %v", sr.Name, buildErr))
			return worker.Result{}, nil
		}

		srBytes, marshalErr := json.Marshal(srCR)
		if marshalErr != nil {
			w.fail(ctx, release, fmt.Sprintf("failed to marshal CR for resource '%s': %v", sr.Name, marshalErr))
			return worker.Result{}, nil
		}

		resourceCRs[sr.Name] = srBytes
		resourceNames = append(resourceNames, sr.Name)
		resourceRevisions[sr.Name] = stackrelease.ComputeResourceRevision(srCR)
	}

	volumeBuildSources := make(map[string][]models.BuildArtifactSource)
	for _, volume := range effective.Volumes {
		if volume.VolumeSource != nil && len(volume.VolumeSource.BuildSource) > 0 {
			volumeBuildSources[volume.Name] = volume.VolumeSource.BuildSource
		}
	}

	manifest := &models.ReleaseManifest{
		StackCR:            stackCRBytes,
		StackResourceCRs:   resourceCRs,
		ResourceNames:      resourceNames,
		ResourceRevisions:  resourceRevisions,
		VolumeBuildSources: volumeBuildSources,
	}

	manifestRevision := stackrelease.ComputeManifestRevision(manifest)

	// CAS transition Rendering -> Applying and persist manifest.
	ok, serr := w.releaseService.SaveManifest(ctx, release.ID, manifest, manifestRevision, release.Pins, stackrelease.RendererVersion)
	if serr != nil {
		return worker.Result{}, serr
	}
	if !ok {
		return worker.Result{Requeue: true}, nil
	}

	// Update the stack's CrRevision so the stack controller knows the target.
	if sErr := w.stackService.UpdateStackCrRevision(ctx, release.StackID, manifestRevision); sErr != nil {
		w.Logger().Errorf("failed to update stack CrRevision: %v", sErr)
	}

	release.Manifest = manifest
	release.ManifestRevision = manifestRevision
	release.State = models.ReleaseStateApplying
	return w.applyAndConverge(ctx, release)
}

func isNotReady(err error, target *stackdeploy.DependencyNotReadyError) bool {
	return stderrors.As(err, target)
}
