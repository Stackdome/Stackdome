package clusterinfo

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

const controllerName = "cluster-info-controller"

type clusterInfoReconciler struct {
	Client         client.Client
	ClusterID      string
	ClusterService services.ClusterService
	Log            logger.Logger
}

type ClusterInfoReconcilerSpec struct {
	ClusterID      string
	ClusterService services.ClusterService
	Log            logger.Logger
}

func NewClusterInfoReconciler(spec ClusterInfoReconcilerSpec) *clusterInfoReconciler {
	return &clusterInfoReconciler{
		ClusterID:      spec.ClusterID,
		ClusterService: spec.ClusterService,
		Log:            spec.Log,
	}
}

func (r *clusterInfoReconciler) Name() string {
	return controllerName
}

func (r *clusterInfoReconciler) AddToManager(manager manager.Manager) error {
	r.Client = manager.GetClient()
	c, err := controller.New(controllerName, manager, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	src := source.Kind(
		manager.GetCache(),
		&corev1alpha1.ClusterInfo{},
		&handler.TypedEnqueueRequestForObject[*corev1alpha1.ClusterInfo]{},
		singletonPredicate(),
	)

	return c.Watch(src)
}

func singletonPredicate() predicate.TypedFuncs[*corev1alpha1.ClusterInfo] {
	return predicate.NewTypedPredicateFuncs(func(object *corev1alpha1.ClusterInfo) bool {
		return object.GetName() == corev1alpha1.ClusterInfoSingletonName
	})
}

func (r *clusterInfoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cr := &corev1alpha1.ClusterInfo{}
	if err := r.Client.Get(ctx, req.NamespacedName, cr); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	info := r.toClusterInfo(ctx, cr.Status)
	if serr := r.ClusterService.InternalUpdateClusterInfo(ctx, r.ClusterID, info); serr != nil {
		return ctrl.Result{}, fmt.Errorf("failed to persist cluster info: %w", serr)
	}

	r.Log.WithField("storage_classes", len(info.StorageClasses)).Debug(ctx, "synced cluster info")
	return ctrl.Result{}, nil
}

func (r *clusterInfoReconciler) toClusterInfo(ctx context.Context, status corev1alpha1.ClusterInfoStatus) *models.ClusterInfo {
	info := &models.ClusterInfo{
		Phase:             models.ClusterInfoPhase(status.Phase),
		KubernetesVersion: status.KubernetesVersion,
		TotalNodes:        status.TotalNodes,
		ReadyNodes:        status.ReadyNodes,
		AvailabilityZones: status.AvailabilityZones,
	}
	if status.LastRefreshedAt != nil {
		t := status.LastRefreshedAt.Time
		info.LastRefreshedAt = &t
	}
	for _, n := range status.Nodes {
		info.Nodes = append(info.Nodes, models.ClusterNode{
			Name:                     n.Name,
			Ready:                    n.Ready,
			AllocatableCPU:           r.parseQuantity(ctx, n.Name, n.AllocatableCPU),
			AllocatableMemory:        r.parseQuantity(ctx, n.Name, n.AllocatableMemory),
			AllocatableEphemeralDisk: r.parseQuantity(ctx, n.Name, n.AllocatableEphemeralDisk),
			CapacityEphemeralDisk:    r.parseQuantity(ctx, n.Name, n.CapacityEphemeralDisk),
			Zone:                     n.Topology.Zone,
			Region:                   n.Topology.Region,
		})
	}
	for _, sc := range status.StorageClasses {
		info.StorageClasses = append(info.StorageClasses, models.ClusterStorageClass{
			Name:        sc.Name,
			Provisioner: sc.Provisioner,
			IsDefault:   sc.IsDefault,
		})
	}
	for _, lb := range status.LoadBalancers {
		info.LoadBalancers = append(info.LoadBalancers, models.ClusterLoadBalancer{
			ServiceName:      lb.ServiceName,
			ServiceNamespace: lb.ServiceNamespace,
			IngressIPs:       lb.IngressIPs,
			IngressHostnames: lb.IngressHostnames,
			HasIP:            lb.HasIP,
		})
	}
	for _, ic := range status.IngressClasses {
		info.IngressClasses = append(info.IngressClasses, models.ClusterIngressClass{
			Name:       ic.Name,
			Controller: ic.Controller,
			IsDefault:  ic.IsDefault,
		})
	}
	return info
}

func (r *clusterInfoReconciler) parseQuantity(ctx context.Context, node, value string) *resource.Quantity {
	if value == "" {
		return nil
	}
	q, err := resource.ParseQuantity(value)
	if err != nil {
		r.Log.WithFields(map[string]interface{}{"node": node, "value": value}).Warn(ctx, "unparseable node quantity")
		return nil
	}
	return &q
}
