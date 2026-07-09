package imagebuild

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/controllers"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

const (
	controllerName = "image-build-controller"
)

//go:generate mockgen -source=image_build_controller.go -destination=image_build_controller_mock_test.go -package=imagebuild

// releaseActiveChecker is a narrow interface to check for an active release
// without importing the full StackReleaseService (mirrors the same seam the
// stack controller uses).
type releaseActiveChecker interface {
	InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *apperrors.ServiceError)
}

// buildEventRecorder records release lifecycle events for builds. It is declared
// locally (rather than reusing services.ReleaseEventRecorder) so the tests can
// mock it in-package: the services gomock mock lives in the services test
// package, and depending on it from here would create an import cycle.
type buildEventRecorder interface {
	RecordBuildEvent(
		ctx context.Context,
		release *models.StackRelease,
		resourceName string,
		eventType models.ReleaseEventType,
		buildID string,
		attribution string,
		reason string,
	) *apperrors.ServiceError
}

type ImageBuildReconciler struct {
	Client                client.Client
	DBImageBuildService   services.ImageBuildService
	DBResourceService     services.StackResourceService
	DBVolumeService       services.VolumeService
	GitIntegrationService services.GitIntegrationService
	Logger                logger.Logger
	releaseChecker        releaseActiveChecker
	eventRecorder         buildEventRecorder

	// clock is injectable for tests; defaults to time.Now.
	clock func() time.Time
}

type ImageBuildReconcilerSpec struct {
	Client                client.Client
	DBImageBuildService   services.ImageBuildService
	DBResourceService     services.StackResourceService
	GitIntegrationService services.GitIntegrationService
	Log                   logger.Logger
	ReleaseChecker        releaseActiveChecker
	EventRecorder         buildEventRecorder
}

func NewImageBuildReconciler(spec ImageBuildReconcilerSpec) *ImageBuildReconciler {
	if spec.ReleaseChecker == nil {
		panic("imagebuild: ReleaseChecker is required")
	}
	if spec.EventRecorder == nil {
		panic("imagebuild: EventRecorder is required")
	}
	return &ImageBuildReconciler{
		Client:                spec.Client,
		DBImageBuildService:   spec.DBImageBuildService,
		DBResourceService:     spec.DBResourceService,
		GitIntegrationService: spec.GitIntegrationService,
		Logger:                spec.Log,
		releaseChecker:        spec.ReleaseChecker,
		eventRecorder:         spec.EventRecorder,
		clock:                 time.Now,
	}
}

// AddToManager adds the reconciler to the manager
func (r *ImageBuildReconciler) AddToManager(manager manager.Manager) error {
	r.Client = manager.GetClient()
	controller, err := controller.New(controllerName, manager, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	src := source.Kind(
		manager.GetCache(),
		&buildsv1alpha1.ImageBuild{},
		&handler.TypedEnqueueRequestForObject[*buildsv1alpha1.ImageBuild]{},
		controllers.StackIDLabelPresentPredicate[*buildsv1alpha1.ImageBuild](),
	)

	return controller.Watch(src)
}

func (r *ImageBuildReconciler) Name() string {
	return controllerName
}

func (r *ImageBuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	imageBuild := &buildsv1alpha1.ImageBuild{}
	if err := r.Client.Get(ctx, req.NamespacedName, imageBuild); err != nil {
		if errors.IsNotFound(err) {
			r.Logger.Infof("imageBuild %v not found", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	r.Logger.Infof("reconciling image build: %v", req.NamespacedName)

	stackID, ok := imageBuild.Labels[corev1alpha1.LabelStackID]
	if !ok {
		r.Logger.Errorf("imageBuild %v does not have stack ID label", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	dbStackResouce, err := r.DBResourceService.InternalGetByStackIDAndResourceName(ctx, stackID, imageBuild.Spec.ResourceName)
	if err != nil {
		if err.Code == apperrors.ErrorNotFound {
			// stack might have gotten deleted. We log and ignore this event.
			r.Logger.Infof(
				"stack resource with name '%s' for stack '%s' not found, it might have been deleted. Ignoring image build '%s'",
				imageBuild.Spec.ResourceName,
				stackID,
				client.ObjectKeyFromObject(imageBuild).String(),
			)
			return ctrl.Result{}, nil
		}
		r.Logger.Errorf("failed to get stack resource %s for build '%s'", imageBuild.Spec.ResourceName, client.ObjectKeyFromObject(imageBuild).String())
		return ctrl.Result{}, err
	}

	dbResourceBuild, serr := r.DBImageBuildService.InternalGetByID(ctx, imageBuild.Name)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			r.Logger.Infof("imageBuild %s not found in DB, creating a new build", imageBuild.Name)
			return ctrl.Result{Requeue: true}, r.createImageBuildInDB(ctx, imageBuild, dbStackResouce)
		}
		return ctrl.Result{}, fmt.Errorf("failed to get image build from db: %v", serr)
	}

	if dbResourceBuild.Status == nil || dbResourceBuild.Status.LastObservedStatusHash != imageBuild.Status.StatusHash {
		// Propagate before persisting the build status: the hash guard above
		// means this block never re-runs for the same build status, so a
		// propagation failure must abort the reconcile before the new hash is
		// recorded — otherwise the requeued retry would skip this block and
		// the failure would be dropped for good.
		if err := r.propagateBuildFailureToStackResource(ctx, dbStackResouce, imageBuild.Status); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to propagate build failure to stack resource: %v", err)
		}
		dbResourceBuild.Status = mapClusterStatusToServerStatus(imageBuild.Status)
		if serr := r.DBImageBuildService.InternalUpdateStatus(ctx, dbResourceBuild.ID, dbResourceBuild.Status); serr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update image build status: %v", serr)
		}
		r.recordBuildEvent(ctx, stackID, imageBuild, dbResourceBuild)
	}

	// Keep minted GitHub App tokens fresh for in-flight builds. Watch events
	// re-establish the requeue chain after hub restarts (the informer's
	// initial list reconciles every build).
	return r.reconcileGitTokenRefresh(ctx, imageBuild, dbResourceBuild)
}

func (r *ImageBuildReconciler) createImageBuildInDB(
	ctx context.Context, imageBuildCr *buildsv1alpha1.ImageBuild, dbStackResource *models.StackResource) error {
	dbSourceContext, err := r.buildDBBuildSrcContextFromClusterObject(ctx, imageBuildCr)
	if err != nil {
		return err
	}

	dbSourceRevision, err := r.buildDBBuildSrcRevisionFromClusterObject(ctx, imageBuildCr)
	if err != nil {
		return err
	}

	dbImageBuild := &models.ImageBuild{
		ID:                imageBuildCr.Name,
		StackResourceID:   dbStackResource.ID,
		StackResourceName: dbStackResource.Name,
		Namespace:         imageBuildCr.Namespace,
		StackID:           dbStackResource.StackID,
		Spec: models.BuildConfigSpec{
			DockerfilePath:          imageBuildCr.Spec.BuildContext.DockerfilePath,
			ContextPathWithinSource: imageBuildCr.Spec.BuildContext.ContextPath,
			BuildImageRepository:    deriveImageRepository(imageBuildCr.Spec.Repository),
			SourceContext:           *dbSourceContext,
			SourceRevision:          dbSourceRevision,
		},
		Status: mapClusterStatusToServerStatus(imageBuildCr.Status),
	}

	_, serr := r.DBImageBuildService.InternalCreate(ctx, dbImageBuild)
	if serr != nil {
		r.Logger.Errorf("Failed to create image build '%s': %s", imageBuildCr.Name, serr)
		return serr.AsError()
	}
	return nil
}

func (r *ImageBuildReconciler) buildDBBuildSrcRevisionFromClusterObject(
	ctx context.Context,
	imageBuildCr *buildsv1alpha1.ImageBuild) (models.BuildSourceRevision, error) {
	res := models.BuildSourceRevision{}
	switch {
	case imageBuildCr.Spec.SourceRevision.Volume != nil:
		res.Volume = &models.VolumeRevision{
			CurrentVolumeHash: imageBuildCr.Spec.SourceRevision.Volume.RevisionString,
		}
	case imageBuildCr.Spec.SourceRevision.GitRepo != nil:
		repoRevision := imageBuildCr.Spec.SourceRevision.GitRepo
		res.Git = &models.GitRevision{
			Commit: repoRevision.Commit,
		}
		if repoRevision.Branch != "" {
			res.Git.Branch = repoRevision.Branch
		}
		if repoRevision.Tag != "" {
			res.Git.Tag = repoRevision.Tag
		}
		if repoRevision.Branch == "" && repoRevision.Tag == "" && repoRevision.Commit == "" {
			return res, apperrors.GeneralError(
				"source_revision.git requires at least a branch or tag with a commit",
			)
		}
	default:
		return res, apperrors.GeneralError(
			"exactly one of source_revision.volume or source_revision.git must be specified",
		)
	}
	return res, nil
}

func (r *ImageBuildReconciler) buildDBBuildSrcContextFromClusterObject(
	ctx context.Context,
	imageBuildCr *buildsv1alpha1.ImageBuild) (*models.BuildContextSource, error) {
	clusterBuildContext := imageBuildCr.Spec.BuildContext
	res := models.BuildContextSource{}
	switch {
	case clusterBuildContext.ContextSource.Volume != nil:
		volume, err := r.DBVolumeService.GetByVolumeNameAndNamespace(
			ctx,
			clusterBuildContext.ContextSource.Volume.Name,
			imageBuildCr.Namespace,
		)
		if err != nil {
			return nil, err
		}
		res.Volume = &models.VolumeBuildSource{
			SourceVolumeID:   volume.ID,
			SourceVolumeName: volume.Name,
		}
	case clusterBuildContext.ContextSource.Git != nil:
		res.Git = &models.GitBuildSource{
			RepoURL: clusterBuildContext.ContextSource.Git.RepoUrl,
		}
	default:
		return nil, apperrors.GeneralError("exactly one of source_context.volume or source_context.git must be specified")
	}

	return &res, nil
}

func mapClusterStatusToServerStatus(clusterStatus buildsv1alpha1.ImageBuildStatus) *models.ImageBuildStatus {
	return &models.ImageBuildStatus{
		Conditions:             models.ConvertConditions(clusterStatus.Conditions),
		State:                  string(clusterStatus.Phase),
		ImageURL:               clusterStatus.ImageUrl,
		BuildSourceRevision:    clusterStatus.BuildSourceRevision,
		LastObservedStatusHash: clusterStatus.StatusHash,
		LastBuildFailureDetail: controllers.MapBuildFailureDetail(clusterStatus.LastBuildFailureDetail),
	}
}

// recordBuildEvent emits a release event for an observed build phase transition.
// It is best-effort: recorder failures and a missing active release window are
// logged and swallowed so they never fail the reconcile.
func (r *ImageBuildReconciler) recordBuildEvent(
	ctx context.Context,
	stackID string,
	cr *buildsv1alpha1.ImageBuild,
	build *models.ImageBuild,
) {
	var eventType models.ReleaseEventType
	var reason string
	switch cr.Status.Phase {
	case buildsv1alpha1.BuildPhasePending:
		eventType = models.ReleaseEventTypeBuildStarted
	case buildsv1alpha1.BuildPhaseSuccess:
		eventType = models.ReleaseEventTypeBuildSucceeded
	case buildsv1alpha1.BuildPhaseFailed:
		eventType = models.ReleaseEventTypeBuildFailed
		reason = buildFailureReason(cr.Status)
	default:
		// Cancelled and unknown phases emit nothing.
		return
	}

	active, serr := r.releaseChecker.InternalGetActiveByStackID(ctx, stackID)
	if serr != nil {
		r.Logger.Debugf("no active release lookup for stack %s: %v", stackID, serr)
		return
	}
	if active == nil {
		r.Logger.Debugf("no active release for stack %s; skipping build event", stackID)
		return
	}

	attribution := models.ReleaseEventAttributionActiveRelease
	if buildMatchesReleasePins(active, cr.Spec.ResourceName, build.Spec.SourceRevision) {
		// The active release's pins deterministically identify this build, so no
		// best-effort attribution marker is needed.
		attribution = ""
	}

	if recErr := r.eventRecorder.RecordBuildEvent(
		ctx, active, cr.Spec.ResourceName, eventType, build.ID, attribution, reason,
	); recErr != nil {
		r.Logger.Errorf("failed to record build event for build %s: %v", build.ID, recErr)
	}
}

// buildFailureReason extracts a human-readable failure reason from the build
// status the controller already observes.
func buildFailureReason(status buildsv1alpha1.ImageBuildStatus) string {
	d := status.LastBuildFailureDetail
	if d == nil {
		return ""
	}
	if d.LastTerminationMessage != "" {
		return d.LastTerminationMessage
	}
	return d.LastTerminationReason
}

// buildMatchesReleasePins reports whether the active release's pins
// deterministically identify this build's source revision for the resource. It
// only returns true when the pins carry a matching per-resource revision;
// absent per-resource revision data it returns false so the caller falls back
// to best-effort active-release attribution.
func buildMatchesReleasePins(
	active *models.StackRelease,
	resourceName string,
	rev models.BuildSourceRevision,
) bool {
	pins, ok := active.Pins.Resources[resourceName]
	if !ok {
		return false
	}
	switch {
	case rev.Git != nil && rev.Git.Commit != "":
		return pins.GitSHA != "" && pins.GitSHA == rev.Git.Commit
	case rev.Volume != nil && rev.Volume.CurrentVolumeHash != "":
		return pins.VolumeHash != "" && pins.VolumeHash == rev.Volume.CurrentVolumeHash
	default:
		return false
	}
}

func buildStackResourceFailureFromBuild(d *corev1alpha1.LastFailureDetail) *models.StackResourceFailure {
	if d == nil {
		return nil
	}
	return &models.StackResourceFailure{
		Type:  models.FailureTypeBuildFailure,
		Build: controllers.MapBuildFailureDetail(d),
	}
}

// computeStackResourceStatusAfterBuild returns the updated status and true if an
// update is needed, or nil/false if nothing should be written to the DB.
func computeStackResourceStatusAfterBuild(
	current models.StackResourceStatus,
	clusterBuildStatus buildsv1alpha1.ImageBuildStatus,
) (*models.StackResourceStatus, bool) {
	updated := current
	switch {
	case clusterBuildStatus.LastBuildFailureDetail != nil:
		updated.LastFailure = buildStackResourceFailureFromBuild(clusterBuildStatus.LastBuildFailureDetail)
	case clusterBuildStatus.Phase == buildsv1alpha1.BuildPhaseSuccess:
		if current.LastFailure != nil && current.LastFailure.Type == models.FailureTypeBuildFailure {
			updated.LastFailure = nil
		} else {
			return nil, false
		}
	default:
		return nil, false
	}
	return &updated, true
}

func (r *ImageBuildReconciler) propagateBuildFailureToStackResource(
	ctx context.Context,
	dbStackResource *models.StackResource,
	clusterBuildStatus buildsv1alpha1.ImageBuildStatus,
) error {
	if dbStackResource.Status == nil {
		// The stack resource has no status row yet (its CR status hasn't been
		// observed). Nothing to clear in that case, but a build failure must
		// not be silently dropped: return an error so the reconcile requeues
		// and propagation retries once the first status write lands. The
		// caller runs this before recording the new build status hash, so the
		// retry re-enters the propagation path.
		if clusterBuildStatus.LastBuildFailureDetail == nil {
			return nil
		}
		return fmt.Errorf("stack resource '%s' has no status yet; requeueing build failure propagation", dbStackResource.ID)
	}
	newStatus, changed := computeStackResourceStatusAfterBuild(*dbStackResource.Status, clusterBuildStatus)
	if !changed {
		return nil
	}
	if serr := r.DBResourceService.UpdateStatus(ctx, dbStackResource.ID, newStatus); serr != nil {
		return serr.AsError()
	}
	return nil
}

func deriveImageRepository(repo corev1alpha1.ImageRepositorySpec) models.BuildImageRepository {
	if repo.ClusterRegistryRef != nil {
		return models.BuildImageRepository{
			UseInClusterRegistry: true,
			ClusterRegistryName:  repo.ClusterRegistryRef.Name,
		}
	}
	insecure := false
	if repo.External != nil && repo.External.TLS != nil {
		insecure = repo.External.TLS.Insecure
	}
	var externalRef string
	if repo.External != nil {
		externalRef = repo.External.Host + "/" + repo.Repository
	}
	return models.BuildImageRepository{
		InsecureRegistry: insecure,
		ExternalImageRef: externalRef,
	}
}
