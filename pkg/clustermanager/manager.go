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

// ClusterManager defines the interface for managing clusters
type ClusterManager interface {
	RegisterCluster(cluster *models.Cluster) error
	GetClient(clusterID string) (client.Client, error)
	IsClusterRegistered(clusterID string) bool
	UnregisterCluster(clusterID string) error
	Start(ctx context.Context)
	Stop(ctx context.Context) error
	IsRunning() bool
}

// Controller defines the interface for controllers that can be added to a manager
type Controller interface {
	AddToManager(manager ctrl.Manager) error
}

// ClusterControl represents a control structure for a single cluster
type ClusterControl struct {
	cluster     *models.Cluster
	clusterID   string
	client      client.Client
	serviceID   suture.ServiceToken
	controllers []Controller
}

// ClusterManagerImpl implements the ClusterManager interface
type ClusterManagerImpl struct {
	mu                    sync.RWMutex
	supervisor            *suture.Supervisor
	leadershipFlag        *leadership.Flag
	registeredClusters    map[string]*ClusterControl
	controllersToRegister []Controller
	supervisorCancelFn    context.CancelFunc
	supervisorErr         error
	isRunning             bool
}

func (cc *ClusterControl) Serve(ctx context.Context) error {
	mgr, err := cc.createManager()
	if err != nil {
		return err
	}
	if err := mgr.Start(ctx); err != nil {
		return err
	}
	return nil
}

func (cc *ClusterControl) createManager() (ctrl.Manager, error) {
	restConfig, err := createRestConfig(cc.cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to create rest config: %w", err)
	}

	scheme, err := createScheme()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheme: %w", err)
	}

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:         scheme,
		Metrics:        metricsserver.Options{BindAddress: "0"},
		LeaderElection: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}
	if err := cc.registerControllers(mgr); err != nil {
		return nil, fmt.Errorf("failed to register controllers: %w", err)
	}
	return mgr, nil
}

func (cc *ClusterControl) registerControllers(manager ctrl.Manager) error {
	for _, controller := range cc.controllers {
		if err := controller.AddToManager(manager); err != nil {
			return err
		}
	}
	return nil
}

func (cc *ClusterControl) String() string {
	return fmt.Sprintf("supervisor for cluster %s", cc.clusterID)
}

// ClusterManagerConfig holds configuration for creating a new ClusterManager
type ClusterManagerConfig struct {
	LeadershipFlag        *leadership.Flag
	ControllersToRegister []Controller
}

// NewClusterManager creates a new ClusterManager instance
func NewClusterManager(config ClusterManagerConfig) ClusterManager {
	return &ClusterManagerImpl{
		leadershipFlag:        config.LeadershipFlag,
		controllersToRegister: config.ControllersToRegister,
		registeredClusters:    make(map[string]*ClusterControl),
		supervisor:            suture.NewSimple("cluster-manager"),
	}
}

// RegisterCluster registers a new cluster with the manager
func (cm *ClusterManagerImpl) RegisterCluster(cluster *models.Cluster) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.registeredClusters[cluster.ID]; exists {
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

	clusterCtrl := &ClusterControl{
		cluster:     cluster,
		clusterID:   cluster.ID,
		client:      client,
		controllers: cm.controllersToRegister,
	}

	serviceID := cm.supervisor.Add(clusterCtrl)
	clusterCtrl.serviceID = serviceID
	cm.registeredClusters[cluster.ID] = clusterCtrl
	return nil
}

// GetClient retrieves the client for a given cluster
func (cm *ClusterManagerImpl) GetClient(clusterID string) (client.Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clusterCtrl, ok := cm.registeredClusters[clusterID]
	if !ok {
		return nil, fmt.Errorf("cluster %s not registered", clusterID)
	}
	return clusterCtrl.client, nil
}

// IsClusterRegistered checks if a cluster is registered
func (cm *ClusterManagerImpl) IsClusterRegistered(clusterID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	_, ok := cm.registeredClusters[clusterID]
	return ok
}

// UnregisterCluster removes a cluster from the manager
func (cm *ClusterManagerImpl) UnregisterCluster(clusterID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

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

// Start begins the cluster manager operations
func (cm *ClusterManagerImpl) Start(ctx context.Context) {
	go cm.run(ctx)
}

func (cm *ClusterManagerImpl) run(ctx context.Context) {
	if err := wait.PollUntilContextCancel(ctx, 30*time.Second, true, func(ctx context.Context) (bool, error) {
		return cm.leadershipFlag.Raised(), nil
	}); err != nil {
		cm.supervisorErr = fmt.Errorf("leadership poll failed: %w", err)
		return
	}

	childCtx, cancelFn := context.WithCancel(ctx)
	cm.supervisorCancelFn = cancelFn
	cm.isRunning = true
	defer func() { cm.isRunning = false }()
	// Cancel the context when the parent context is done
	go func() {
		<-ctx.Done()
		cancelFn()
	}()
	if err := cm.supervisor.Serve(childCtx); err != nil {
		cm.supervisorErr = err
	}
}

// Stop halts the cluster manager operations
func (cm *ClusterManagerImpl) Stop(ctx context.Context) error {
	if cm.supervisorCancelFn != nil {
		cm.supervisorCancelFn()
	}
	return nil
}

// IsRunning checks if the cluster manager is currently running
func (cm *ClusterManagerImpl) IsRunning() bool {
	return cm.isRunning
}

func createRestConfig(cluster *models.Cluster) (*rest.Config, error) {
	cadata, err := base64.StdEncoding.DecodeString(cluster.ClusterCAData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cluster CA data: %w", err)
	}

	token, err := base64.StdEncoding.DecodeString(cluster.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cluster token: %w", err)
	}

	if _, err := certutil.NewPoolFromBytes(cadata); err != nil {
		return nil, fmt.Errorf("failed to parse cluster CA data: %w", err)
	}

	return &rest.Config{
		Host:        cluster.ClusterURL,
		BearerToken: string(token),
		TLSClientConfig: rest.TLSClientConfig{
			CAData: cadata,
		},
	}, nil
}

func createScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add client-go scheme: %w", err)
	}

	if err := workspacev1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add workspace v1alpha1 scheme: %w", err)
	}

	return scheme, nil
}

func createUncachedClient(restConfig *rest.Config, scheme *runtime.Scheme) (client.Client, error) {
	return client.New(restConfig, client.Options{Scheme: scheme})
}
