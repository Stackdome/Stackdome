package stack

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

const ResourceRestartReconcilerName = "resource-restart"

type ResourceRestartReconcilerSpec struct {
	ClusterManager clustermanager.ClusterManager
	Logger         logger.Logger
}

// resourceRestartReconciler applies restart requests already accepted and
// persisted by the API. Compute-access decisions stay at the request boundary.
type resourceRestartReconciler struct {
	clusterManager clustermanager.ClusterManager
	logger         logger.Logger
}

func NewResourceRestartReconciler(spec ResourceRestartReconcilerSpec) subReconciler {
	return &resourceRestartReconciler{
		clusterManager: spec.ClusterManager,
		logger:         spec.Logger,
	}
}

func (*resourceRestartReconciler) Name() string {
	return ResourceRestartReconcilerName
}

func (r *resourceRestartReconciler) Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	hasRestartRequest := false
	for _, resource := range stack.StackResources {
		if resource != nil && resource.LifecycleConfig != nil && resource.LifecycleConfig.RestartRequestTime != nil {
			hasRestartRequest = true
			break
		}
	}
	if !hasRestartRequest {
		return resultNil, nil
	}

	clusterClient, err := r.clusterManager.GetClient(stack.ClusterID)
	if err != nil {
		return resultNil, err
	}

	for _, resource := range stack.StackResources {
		if resource == nil || resource.LifecycleConfig == nil || resource.LifecycleConfig.RestartRequestTime == nil {
			continue
		}

		existing := &corev1alpha1.StackResource{}
		if err := clusterClient.Get(ctx, client.ObjectKey{
			Name:      resource.Name,
			Namespace: resource.Namespace,
		}, existing); err != nil {
			return resultNil, err
		}

		restartTime := resource.LifecycleConfig.RestartRequestTime
		if existing.Spec.RestartRequest != nil && existing.Spec.RestartRequest.Time.Equal(*restartTime) {
			continue
		}
		existing.Spec.RestartRequest = &metav1.Time{Time: *restartTime}
		if err := clusterClient.Update(ctx, existing); err != nil {
			return resultNil, err
		}
		r.logger.WithField("resource_id", resource.ID).Info(ctx, "applied stack resource restart request")
	}

	return resultNil, nil
}
