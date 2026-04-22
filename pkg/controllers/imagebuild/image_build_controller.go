package imagebuild

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/controllers"
	apperrors "github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
)

const (
	controllerName = "image-build-controller"
)

type ImageBuildReconciler struct {
	Client              client.Client
	DBImageBuildService services.ImageBuildService
	DBResourceService   services.StackResourceService
	DBVolumeService     services.VolumeService
	Logger              logger.Logger
}

type ImageBuildReconcilerSpec struct {
	Client              client.Client
	DBImageBuildService services.ImageBuildService
	DBResourceService   services.StackResourceService
	Log                 logger.Logger
}

func NewImageBuildReconciler(spec ImageBuildReconcilerSpec) *ImageBuildReconciler {
	return &ImageBuildReconciler{
		Client:              spec.Client,
		DBImageBuildService: spec.DBImageBuildService,
		DBResourceService:   spec.DBResourceService,
		Logger:              spec.Log,
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

	stackID, ok := imageBuild.Labels[models.StackIDLabel]
	if !ok {
		r.Logger.Errorf("imageBuild %v does not have stack ID label", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	dbStackResouce, err := r.DBResourceService.GetByStackIDAndResourceName(ctx, stackID, imageBuild.Spec.ResourceName)
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

	dbResourceBuild, serr := r.DBImageBuildService.GetByID(ctx, imageBuild.Name)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			r.Logger.Infof("imageBuild %s not found in DB, creating a new build", imageBuild.Name)
			return ctrl.Result{Requeue: true}, r.createImageBuildInDB(ctx, imageBuild, dbStackResouce)
		}
		return ctrl.Result{}, fmt.Errorf("failed to get image build from db: %v", serr)
	}

	if dbResourceBuild.Status == nil || dbResourceBuild.Status.LastObservedStatusHash != imageBuild.Status.StatusHash {
		dbResourceBuild.Status = mapClusterStatusToServerStatus(imageBuild.Status)
		if serr := r.DBImageBuildService.UpdateStatus(ctx, dbResourceBuild.ID, dbResourceBuild.Status); serr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update image build status: %v", serr)
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
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
			ImageRepositoryUrl:      imageBuildCr.Spec.RegistryURL,
			SourceContext:           *dbSourceContext,
			SourceRevision:          dbSourceRevision,
		},
		Status: mapClusterStatusToServerStatus(imageBuildCr.Status),
	}

	_, serr := r.DBImageBuildService.Create(ctx, dbImageBuild)
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
			CurrentVolumeHash: imageBuildCr.Spec.SourceRevision.Volume.CurrentVolumeHash,
		}
	case imageBuildCr.Spec.SourceRevision.GitRepo != nil:
		repoRevision := imageBuildCr.Spec.SourceRevision.GitRepo
		res.Git = &models.GitRevision{}
		switch {
		case repoRevision.Tag != "":
			res.Git.Tag = repoRevision.Tag
		case repoRevision.Branch != nil:
			res.Git.Branch = &models.GitBranch{
				Name:    repoRevision.Branch.Name,
				HeadSha: repoRevision.Branch.HeadSha,
			}
		case repoRevision.Commit != "":
			res.Git.Commit = repoRevision.Commit
		default:
			return res, apperrors.GeneralError(
				"exactly one of source_revision.git.tag, source_revision.git.branch or source_revision.git.commit must be specified",
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
	}
}
