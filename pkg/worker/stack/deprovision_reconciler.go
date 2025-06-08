package stack

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
	storagev1alpha1 "stackdome.io/cluster-agent/api/storage/v1alpha1"
)

const (
	deprovisionReconcilerName = "deprovision-reconciler"
)

type deprovisionReconciler struct {
	stackService     stackService
	volumeService    volumeService
	namespaceService namespaceService
	logger           logger.Logger
	clusterManager   clustermanager.ClusterManager
}

type DeprovisionReconcilerSpec struct {
	StackService     stackService
	NamespaceService namespaceService
	Logger           logger.Logger
	VolumeService    volumeService
	ClusterManager   clustermanager.ClusterManager
}

func NewDeprovisionReconciler(spec DeprovisionReconcilerSpec) *deprovisionReconciler {
	return &deprovisionReconciler{
		stackService:     spec.StackService,
		namespaceService: spec.NamespaceService,
		logger:           spec.Logger,
		volumeService:    spec.VolumeService,
		clusterManager:   spec.ClusterManager,
	}
}

func (r *deprovisionReconciler) Name() string {
	return deprovisionReconcilerName
}

func (r *deprovisionReconciler) Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	if stack.DeletionTimestamp == nil {
		// NOOP
		return resultNil, nil
	}

	r.logger.Infof("deprovisioning stack '%s' in namespace '%s'", stack.Name, stack.Namespace)
	if err := r.deleteClusterResources(ctx, stack); err != nil {
		return resultNil, fmt.Errorf("failed to delete cluster resources for stack '%s': %w", stack.Name, err)
	}
	r.logger.Infof("deleted cluster resources for stack '%s' in namespace '%s'", stack.Name, stack.Namespace)
	if err := r.deleteResourcesFromDB(ctx, stack); err != nil {
		return resultNil, fmt.Errorf("failed to delete resources from db for stack '%s': %w", stack.Name, err)
	}
	r.logger.Infof("deleted resources from db for stack '%s' in namespace '%s'", stack.Name, stack.Namespace)
	return resultStop, nil
}

func (r *deprovisionReconciler) deleteResourcesFromDB(ctx context.Context, stack *models.Stack) error {
	if err := r.stackService.InternalDeleteFromDB(ctx, stack.ID); err != nil {
		return fmt.Errorf("failed to delete stack '%s' from db: %w", stack.ID, err)
	}
	if err := r.volumeService.InternalDeleteVolumesUsedByStackFromDB(ctx, stack.ID); err != nil {
		return fmt.Errorf("failed to delete volumes for stack '%s' from db: %w", stack.ID, err)
	}

	if err := r.namespaceService.InternalDeleteFromDB(ctx, stack.NamespaceID); err != nil {
		return fmt.Errorf("failed to delete namespace '%s' from db: %w", stack.NamespaceID, err)
	}
	return nil
}

func (r *deprovisionReconciler) deleteClusterResources(ctx context.Context, stack *models.Stack) error {
	clusterClient, cerr := r.clusterManager.GetClient(stack.ClusterID)
	if cerr != nil {
		return cerr
	}

	// Delete the stack CR from the cluster
	stackCR := &corev1alpha1.Stack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      stack.Name,
			Namespace: stack.Namespace,
		},
	}
	if err := clusterClient.Delete(
		ctx, stackCR,
		&client.DeleteOptions{
			PropagationPolicy: ptr.To(metav1.DeletePropagationForeground),
		}); client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("failed to delete stack CR '%s' in namespace '%s': %w", stack.Name, stack.Namespace, err)
	}

	// delete volumes
	for _, volume := range stack.Volumes {
		volumeCR := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      volume.Name,
				Namespace: stack.Namespace,
			},
		}
		if err := clusterClient.Delete(
			ctx, volumeCR,
			&client.DeleteOptions{
				PropagationPolicy: ptr.To(metav1.DeletePropagationForeground),
			}); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to delete volume CR '%s' in namespace '%s': %w", volume.Name, stack.Namespace, err)
		}
	}
	// delete namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: stack.Namespace,
		},
	}
	if err := clusterClient.Delete(
		ctx, namespace,
		&client.DeleteOptions{
			PropagationPolicy: ptr.To(metav1.DeletePropagationForeground),
		}); client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("failed to delete namespace '%s': %w", stack.Namespace, err)
	}
	return nil
}
