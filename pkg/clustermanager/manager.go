package clustermanager

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/openshift-online/ocm-sdk-go/leadership"
	suture "github.com/thejerf/suture/v4"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	workspacev1alpha1 "soradev.io/cluster-agent/api/v1alpha1"
)

type ClusterManager interface {
	RegisterCluster(cluster *models.Cluster) error
	GetClient(clusterID string) (client.Client, error)
	IsClusterRegistered(clusterID string) bool
	UnregisterCluster(clusterID string) error
	Start(ctx context.Context)
	Stop(ctx context.Context) error
	Running() bool
	IsManagerForClusterRunning(clusterID string) (bool, error)
}

type Controller interface {
	AddToManager(manager ctrl.Manager) error
}

type DBClusterStateAccessor interface {
	IsManagerForClusterRunning(ctx context.Context, clusterID string) (bool, error)
	PersistManagerState(ctx context.Context, clusterID string, running bool) error
}

type clusterCtrl struct {
	clusterStateAccessor DBClusterStateAccessor
	cluster              *models.Cluster
	clusterID            string
	manager              ctrl.Manager
	client               client.Client
	serviceID            suture.ServiceToken
}

type clusterManager struct {
	sync.RWMutex
	supervisor            *suture.Supervisor
	dbStateAccessor       DBClusterStateAccessor
	leadershipFlag        *leadership.Flag
	registeredClusters    map[string]*clusterCtrl
	controllersToRegister []Controller
	supervisorCancelFn    context.CancelFunc
	supervisiorErr        error
	supervisiorRunning    bool
}

func (c *clusterCtrl) Serve(ctx context.Context) error {
	c.clusterStateAccessor.PersistManagerState(ctx, c.clusterID, true)
	err := c.manager.Start(ctx)
	c.clusterStateAccessor.PersistManagerState(ctx, c.clusterID, false)
	return err
}

func (c *clusterCtrl) String() string {
	return fmt.Sprintf("clusterCtrl for cluster %s", c.clusterID)
}

type ClusterManagerSpec struct {
	LeadershipFlag       *leadership.Flag
	ClusterStateAccessor DBClusterStateAccessor
}

func NewClusterManager(spec ClusterManagerSpec) ClusterManager {
	return &clusterManager{
		dbStateAccessor:    spec.ClusterStateAccessor,
		leadershipFlag:     spec.LeadershipFlag,
		registeredClusters: make(map[string]*clusterCtrl),
		supervisor:         suture.NewSimple("cluster-manager"),
	}
}

func (cm *clusterManager) RegisterCluster(cluster *models.Cluster) error {
	cm.Lock()
	defer cm.Unlock()

	if _, ok := cm.registeredClusters[cluster.ID]; ok {
		return nil
	}

	restConfig, err := createRestConfig(cluster)
	if err != nil {
		return fmt.Errorf("failed to create rest config for cluster %s: %w", cluster.ID, err)
	}
	scheme, err := createScheme()
	if err != nil {
		return fmt.Errorf("failed to create scheme for cluster %s: %w", cluster.ID, err)
	}
	client, err := createUncachedClient(restConfig, scheme)
	if err != nil {
		return fmt.Errorf("failed to create client for cluster %s: %w", cluster.ID, err)
	}

	manager, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:         scheme,
		Metrics:        metricsserver.Options{BindAddress: "0"},
		LeaderElection: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create manager for cluster %s: %w", cluster.ID, err)
	}

	if err := cm.registerControllers(manager); err != nil {
		return fmt.Errorf("failed to register controllers for cluster %s: %w", cluster.ID, err)
	}

	clusterCtrl := &clusterCtrl{
		clusterStateAccessor: cm.dbStateAccessor,
		cluster:              cluster,
		clusterID:            cluster.ID,
		client:               client,
		manager:              manager,
	}

	serviceID := cm.supervisor.Add(clusterCtrl)
	clusterCtrl.serviceID = serviceID
	cm.registeredClusters[cluster.ID] = clusterCtrl
	return nil
}

func (cm *clusterManager) GetClient(clusterID string) (client.Client, error) {
	cm.RLock()
	defer cm.RUnlock()

	clusterCtrl, ok := cm.registeredClusters[clusterID]
	if !ok {
		return nil, fmt.Errorf("cluster %s not registered", clusterID)
	}
	return clusterCtrl.client, nil
}

func (cm *clusterManager) IsClusterRegistered(clusterID string) bool {
	cm.RLock()
	defer cm.RUnlock()

	_, ok := cm.registeredClusters[clusterID]
	return ok
}

func (cm *clusterManager) IsManagerForClusterRunning(clusterID string) (bool, error) {
	cm.RLock()
	defer cm.RUnlock()

	_, ok := cm.registeredClusters[clusterID]
	if !ok {
		return false, nil
	}
	running, err := cm.dbStateAccessor.IsManagerForClusterRunning(context.Background(), clusterID)
	if err != nil {
		return false, fmt.Errorf("failed to get manager state for cluster %s: %w", clusterID, err)
	}
	return running, nil
}

func (cm *clusterManager) UnregisterCluster(clusterID string) error {
	cm.Lock()
	defer cm.Unlock()

	clusterCtrl, ok := cm.registeredClusters[clusterID]
	if !ok {
		return nil
	}

	if err := cm.supervisor.Remove(clusterCtrl.serviceID); err != nil {
		return fmt.Errorf("failed to remove cluster %s from supervisor: %w", clusterID, err)
	}
	delete(cm.registeredClusters, clusterID)
	return nil
}

func (cm *clusterManager) Start(ctx context.Context) {
	go cm.start(ctx)
}

func (cm *clusterManager) start(ctx context.Context) error {
	wait.PollUntilContextCancel(ctx, time.Second*30, true, func(ctx context.Context) (bool, error) {
		return cm.leadershipFlag.Raised(), nil
	})
	childCtx, cancelFn := context.WithCancel(ctx)
	cm.supervisorCancelFn = cancelFn
	cm.supervisiorRunning = true
	defer func() {
		cm.supervisiorRunning = false
	}()
	if err := cm.supervisor.Serve(childCtx); err != nil {
		cm.supervisiorErr = err
	}
	return cm.supervisiorErr
}

func (cm *clusterManager) Stop(ctx context.Context) error {
	cm.supervisorCancelFn()
	return nil
}

func (cm *clusterManager) Running() bool {
	return cm.supervisiorRunning
}

func (cm *clusterManager) registerControllers(manager ctrl.Manager) error {
	for _, controller := range cm.controllersToRegister {
		if err := controller.AddToManager(manager); err != nil {
			return err
		}
	}
	return nil
}

func createRestConfig(cluster *models.Cluster) (*rest.Config, error) {
	cadata, err := base64.StdEncoding.DecodeString(cluster.ClusterCAData)
	if err != nil {
		return nil, err
	}
	token, err := base64.StdEncoding.DecodeString(cluster.Token)
	if err != nil {
		return nil, err
	}

	_, err = certutil.NewPoolFromBytes(cadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cluster CA data: %w", err)
	}

	restConfig := &rest.Config{
		Host:        cluster.ClusterURL,
		BearerToken: string(token),
		TLSClientConfig: rest.TLSClientConfig{
			CAData: cadata,
		},
	}
	return restConfig, nil
}

func createScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := workspacev1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func createUncachedClient(restConfig *rest.Config, scheme *runtime.Scheme) (client.Client, error) {
	clientset, err := client.New(restConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
