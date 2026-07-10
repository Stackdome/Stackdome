package release

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stackdeploy"
	"github.com/Stackdome/stackdome/pkg/stackrelease"
)

type renderReconciler struct {
	releaseService releaseService
	stackService   stackService
	crBuilder      builders.ClusterResourceBuilder
	resolver       *stackdeploy.Resolver
	logger         logger.Logger
}

func newRenderReconciler(spec ReleaseWorkerSpec) *renderReconciler {
	return &renderReconciler{
		releaseService: spec.ReleaseService,
		stackService:   spec.StackService,
		crBuilder:      spec.CRBuilder,
		resolver:       spec.Resolver,
		logger:         logger.NewLoggerWithPrefix(context.Background(), "release-render"),
	}
}

func (r *renderReconciler) Name() string { return "render" }

func (r *renderReconciler) Reconcile(ctx context.Context, release *models.StackRelease) (subReconcilerResult, error) {
	if release.Manifest != nil {
		return resultNil, nil
	}

	draft := release.Snapshot.ToStack()

	effective, err := r.resolver.Resolve(ctx, draft)
	if err != nil {
		var depErr stackdeploy.DependencyNotReadyError
		if isNotReady(err, &depErr) {
			r.logger.Infof("release %s: dependency not ready, requeueing: %s", release.ID, depErr.Message)
			return resultRequeueAfter(convergencePollInterval), nil
		}
		failRelease(ctx, r.releaseService, r.logger, release, fmt.Sprintf("render failed: %v", err))
		return resultStop, nil
	}

	stackCR, err := r.crBuilder.BuildStackCR(effective)
	if err != nil {
		failRelease(ctx, r.releaseService, r.logger, release, fmt.Sprintf("failed to build stack CR: %v", err))
		return resultStop, nil
	}

	stackCRBytes, err := json.Marshal(stackCR)
	if err != nil {
		failRelease(ctx, r.releaseService, r.logger, release, fmt.Sprintf("failed to marshal stack CR: %v", err))
		return resultStop, nil
	}

	resourceCRs := make(map[string]json.RawMessage)
	resourceNames := make([]string, 0, len(effective.StackResources))
	resourceRevisions := make(map[string]string)

	for _, sr := range effective.StackResources {
		srCR, buildErr := r.crBuilder.BuildStackResourceCR(sr, effective.Name, effective.OrganisationID)
		if buildErr != nil {
			failRelease(ctx, r.releaseService, r.logger, release, fmt.Sprintf("failed to build CR for resource '%s': %v", sr.Name, buildErr))
			return resultStop, nil
		}

		srBytes, marshalErr := json.Marshal(srCR)
		if marshalErr != nil {
			failRelease(ctx, r.releaseService, r.logger, release, fmt.Sprintf("failed to marshal CR for resource '%s': %v", sr.Name, marshalErr))
			return resultStop, nil
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

	ok, serr := r.releaseService.SaveManifest(ctx, release.ID, manifest, manifestRevision, release.Pins, stackrelease.RendererVersion)
	if serr != nil {
		return resultNil, fmt.Errorf("failed to save manifest: %w", serr)
	}
	if !ok {
		r.logger.Infof("release %s: SaveManifest CAS failed", release.ID)
		return resultStop, nil
	}

	if sErr := r.stackService.UpdateStackCrRevision(ctx, release.StackID, manifestRevision); sErr != nil {
		r.logger.Errorf("failed to update stack CrRevision: %v", sErr)
	}

	return resultRequeue, nil
}

func isNotReady(err error, target *stackdeploy.DependencyNotReadyError) bool {
	return stderrors.As(err, target)
}
