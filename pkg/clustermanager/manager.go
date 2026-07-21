package clustermanager

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	barmancloudv1 "github.com/cloudnative-pg/plugin-barman-cloud/api/v1"
	"github.com/openshift-online/ocm-sdk-go/leadership"
	suture "github.com/thejerf/suture/v4"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	addonsv1alpha1 "stackdome.io/cluster-agent/api/addons/v1alpha1"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
	registryv1alpha1 "stackdome.io/cluster-agent/api/registry/v1alpha1"
	storagev1alpha1 "stackdome.io/cluster-agent/api/storage/v1alpha1"
	usersv1alpha1 "stackdome.io/cluster-agent/api/users/v1alpha1"
)

type CredentialDecryptor interface {
	DecryptData(in string) ([]byte, error)
}

//go:generate mockgen -destination=../mocks/mock_cluster_manager.go -package=mocks github.com/Stackdome/stackdome/pkg/clustermanager ClusterManager

// ClusterManager defines the interface for managing clusters
type ClusterManager interface {
	RegisterCluster(cluster *models.Cluster) error
	GetClient(clusterID string) (client.Client, error)
	GetRestConfig(clusterID string) (*rest.Config, error)
	IsClusterRegistered(clusterID string) bool
	UnregisterCluster(clusterID string) error
	ReRegisterCluster(cluster *models.Cluster) error
	Start(ctx context.Context)
	Stop(ctx context.Context) error
	IsRunning() bool
}

// Controller defines the interface for controllers that can be added to a manager
type Controller interface {
	AddToManager(manager ctrl.Manager) error
	Name() string
}

type ControllerFn func(clusterID string) Controller

// ClusterControl represents a control structure for a single cluster
type ClusterControl struct {
	cluster     *models.Cluster
	clusterID   string
	client      client.Client
	restConfig  *rest.Config
	serviceID   suture.ServiceToken
	controllers []Controller
}

// ClusterManagerImpl implements the ClusterManager interface
type ClusterManagerImpl struct {
	mu                    sync.RWMutex
	supervisor            *suture.Supervisor
	leadershipFlag        *leadership.Flag
	registeredClusters    map[string]*ClusterControl
	controllersToRegister []ControllerFn
	credentialDecryptor   CredentialDecryptor
	supervisorCancelFn    context.CancelFunc
	supervisorErr         error
	isRunning             bool
	log                   logger.Logger
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
		Controller: config.Controller{
			SkipNameValidation: ptr.To(true),
		},
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
			return fmt.Errorf("failed to register controller %s: %w", controller.Name(), err)
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
	ControllersToRegister []ControllerFn
	CredentialDecryptor   CredentialDecryptor
	Logger                logger.Logger
}

// NewClusterManager creates a new ClusterManager instance
func NewClusterManager(config ClusterManagerConfig) ClusterManager {
	return &ClusterManagerImpl{
		leadershipFlag:        config.LeadershipFlag,
		controllersToRegister: config.ControllersToRegister,
		credentialDecryptor:   config.CredentialDecryptor,
		registeredClusters:    make(map[string]*ClusterControl),
		supervisor:            suture.NewSimple("cluster-manager"),
		log:                   config.Logger,
	}
}

// RegisterCluster registers a new cluster with the manager
func (cm *ClusterManagerImpl) RegisterCluster(cluster *models.Cluster) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.registeredClusters[cluster.ID]; exists {
		return nil
	}

	if err := cm.decryptClusterCredentials(cluster); err != nil {
		return fmt.Errorf("failed to decrypt credentials for cluster %s: %w", cluster.ID, err)
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

	controllers := make([]Controller, len(cm.controllersToRegister))
	for i, fn := range cm.controllersToRegister {
		controllers[i] = fn(cluster.ID)
	}

	clusterCtrl := &ClusterControl{
		cluster:     cluster,
		clusterID:   cluster.ID,
		client:      client,
		restConfig:  restConfig,
		controllers: controllers,
	}

	serviceID := cm.supervisor.Add(clusterCtrl)
	clusterCtrl.serviceID = serviceID
	cm.registeredClusters[cluster.ID] = clusterCtrl
	cm.log.WithFields(map[string]interface{}{
		logger.FieldClusterID: cluster.ID,
		"cluster_name":        cluster.Name,
		"controllers":         len(controllers),
	}).Infof("registered cluster")
	return nil
}

func (cm *ClusterManagerImpl) decryptClusterCredentials(cluster *models.Cluster) error {
	if cluster.Token != "" && cluster.ClusterCAData != "" {
		return nil
	}
	if cluster.EncryptedToken == "" || cluster.EncryptedClusterCAData == "" {
		return fmt.Errorf("cluster %s has no encrypted credentials", cluster.ID)
	}
	token, err := cm.credentialDecryptor.DecryptData(cluster.EncryptedToken)
	if err != nil {
		return fmt.Errorf("failed to decrypt token: %w", err)
	}
	caData, err := cm.credentialDecryptor.DecryptData(cluster.EncryptedClusterCAData)
	if err != nil {
		return fmt.Errorf("failed to decrypt CA data: %w", err)
	}
	cluster.Token = string(token)
	cluster.ClusterCAData = string(caData)
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

// GetRestConfig retrieves the rest config for a given cluster
func (cm *ClusterManagerImpl) GetRestConfig(clusterID string) (*rest.Config, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clusterCtrl, ok := cm.registeredClusters[clusterID]
	if !ok {
		return nil, fmt.Errorf("cluster %s not registered", clusterID)
	}
	return clusterCtrl.restConfig, nil
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
	cm.log.WithField(logger.FieldClusterID, clusterID).Infof("unregistered cluster")
	return nil
}

// ReRegisterCluster tears down an existing cluster registration and registers it again
func (cm *ClusterManagerImpl) ReRegisterCluster(cluster *models.Cluster) error {
	if err := cm.UnregisterCluster(cluster.ID); err != nil {
		return fmt.Errorf("failed to unregister cluster %s for re-registration: %w", cluster.ID, err)
	}
	return cm.RegisterCluster(cluster)
}

// Start begins the cluster manager operations
func (cm *ClusterManagerImpl) Start(ctx context.Context) {
	go cm.run(ctx)
}

func (cm *ClusterManagerImpl) run(ctx context.Context) {
	cm.log.Info(ctx, "waiting for leadership before starting cluster supervisor")
	if err := wait.PollUntilContextCancel(ctx, 30*time.Second, true, func(ctx context.Context) (bool, error) {
		return cm.leadershipFlag.Raised(), nil
	}); err != nil {
		cm.supervisorErr = fmt.Errorf("leadership poll failed: %w", err)
		cm.log.Error(ctx, "leadership poll failed, cluster supervisor not started: %v", err)
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
	cm.log.Info(ctx, "leadership acquired, cluster supervisor serving")
	if err := cm.supervisor.Serve(childCtx); err != nil {
		cm.supervisorErr = err
		cm.log.Error(ctx, "cluster supervisor terminated with error: %v", err)
	}
}

// Stop halts the cluster manager operations
func (cm *ClusterManagerImpl) Stop(ctx context.Context) error {
	cm.log.Info(ctx, "stopping cluster manager")
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

	if err := corev1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add corev1alpha1 scheme: %w", err)
	}

	if err := buildsv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add buildsv1alpha1 scheme: %w", err)
	}

	if err := usersv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add userv1alpha1 scheme: %w", err)
	}

	if err := storagev1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add storagev1alpha1 scheme: %w", err)
	}

	if err := registryv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add registryv1alpha1 scheme: %w", err)
	}

	if err := addonsv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add addonsv1alpha1 scheme: %w", err)
	}

	if err := cnpgv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add cnpgv1 scheme: %w", err)
	}

	if err := barmancloudv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add barmancloudv1 scheme: %w", err)
	}

	if err := cmv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add cert-manager scheme: %w", err)
	}

	return scheme, nil
}

func createUncachedClient(restConfig *rest.Config, scheme *runtime.Scheme) (client.Client, error) {
	return client.New(restConfig, client.Options{Scheme: scheme})
}
