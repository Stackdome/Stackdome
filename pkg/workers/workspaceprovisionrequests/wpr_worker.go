package workspaceprovisionrequests

import (
	"context"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	workerlib "github.com/ashishmax31/stackdome-api-server/pkg/worker"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	workspacev1alpha1 "soradev.io/cluster-agent/api/v1alpha1"
)

func (w *workspaceProvisionRequestReconcileWorker) Execute(ctx context.Context, operand workerlib.Operand) (workerlib.Result, *errors.ServiceError) {
	provisionRequestID, ok := operand.(*models.WorkspaceProvisionRequest)
	if !ok {
		return workerlib.Result{}, w.WorkerError.NewError("failed to type cast workspace provision request")
	}

	provisionRequest, err := w.wprService.Get(ctx, provisionRequestID.ID)
	if err != nil {
		if err.Code == errors.ErrorNotFound {
			return workerlib.Result{}, nil
		}
		return workerlib.Result{}, err
	}

	user, err := w.userService.Get(ctx, provisionRequest.UserID)
	if err != nil {
		return workerlib.Result{}, err
	}

	desiredClusterObject, err := w.desiredWPRObjectInCluster(user)
	if err != nil {
		return workerlib.Result{}, err
	}

	existingClusterObject := &workspacev1alpha1.WorkspaceConfiguration{}

	if err := w.clusterClient.Get(ctx, client.ObjectKeyFromObject(desiredClusterObject), existingClusterObject); err != nil {
		if k8serrors.IsNotFound(err) {
			createErr := w.clusterClient.Create(ctx, desiredClusterObject)
			if createErr != nil {
				return workerlib.Result{}, w.WorkerError.NewError("failed to create workspace provision request object in cluster: %s", createErr.Error())
			}
			provisionRequest.State = models.ProvisionRequestPending
			provisionRequest.Message = "Object Created in cluster"
			return workerlib.Result{RequeueAfter: time.Second * 2}, w.wprService.InternalUpdate(ctx, provisionRequest.ID, provisionRequest)
		}
		return workerlib.Result{}, w.WorkerError.NewError("failed to get wpr object from cluster: %s", err)
	}

	if !equality.Semantic.DeepEqual(existingClusterObject.Spec, desiredClusterObject.Spec) {
		existingClusterObject.Spec = desiredClusterObject.Spec
		updateErr := w.clusterClient.Update(ctx, existingClusterObject)
		if updateErr != nil {
			return workerlib.Result{}, w.WorkerError.NewError("failed to update workspace provision request object in cluster: %s", updateErr.Error())
		}
		return workerlib.Result{RequeueAfter: time.Second * 2}, nil
	}

	availableCond := meta.FindStatusCondition(existingClusterObject.Status.Conditions, workspacev1alpha1.WorkspaceConfigurationAvailable)
	if availableCond == nil || availableCond.Status != metav1.ConditionTrue {
		w.Logger().Infof("workspace provision request: %s not yet in available condition in cluster", provisionRequest.ID)
		return workerlib.Result{RequeueAfter: time.Second * 2}, nil
	}

	cluster, err := w.clusterService.GetClusterForOrg(ctx, user.OrganisationID)
	if err != nil {
		return workerlib.Result{}, err
	}

	organisation, err := w.organisationService.Get(ctx, user.OrganisationID)
	if err != nil {
		return workerlib.Result{}, err
	}
	if provisionRequest.Status == nil {
		provisionRequest.Status = &models.WorkspaceProvisionRequestStatus{}
	}
	provisionRequest.Status.ClusterCACert = &cluster.ClusterCAData
	provisionRequest.Status.ClusterUrl = &cluster.ClusterURL
	provisionRequest.Status.Domain = &organisation.DomainName
	provisionRequest.Status.WorkspaceServiceAccountName = existingClusterObject.Status.ServiceAccountName
	provisionRequest.Status.WorkspaceServiceAccountToken = existingClusterObject.Status.ServiceAccountToken
	provisionRequest.Status.WorkspaceNamespace = existingClusterObject.Status.Namespace
	provisionRequest.State = models.ProvisionRequestCompleted
	provisionRequest.Message = "Provision Completed"
	return workerlib.Result{}, w.wprService.InternalUpdate(ctx, provisionRequest.ID, provisionRequest)
}

func (w *workspaceProvisionRequestReconcileWorker) desiredWPRObjectInCluster(user *models.User) (*workspacev1alpha1.WorkspaceConfiguration, *errors.ServiceError) {
	return &workspacev1alpha1.WorkspaceConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: WPRClusterObjectName(user),
		},
		Spec: workspacev1alpha1.WorkspaceConfigurationSpec{
			WorkspaceNamespace: WorkspaceNamespaceFor(user),
			Username:           user.Name,
		},
	}, nil
}

func (w *workspaceProvisionRequestReconcileWorker) GetInput(ctx context.Context) ([]workerlib.Operand, *errors.ServiceError) {
	res, err := w.wprService.InternalList(ctx, "state NOT IN ?", []models.ProvisionRequestState{models.ProvisionRequestCompleted})
	if err != nil {
		return nil, err
	}
	var candidates []*models.WorkspaceProvisionRequest
	for _, item := range res {
		candidates = append(candidates, item)
	}
	return workerlib.ToOperandList(candidates...), nil
}
