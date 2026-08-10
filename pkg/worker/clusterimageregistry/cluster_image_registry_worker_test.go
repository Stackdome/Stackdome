package clusterimageregistry

import (
	"context"
	"testing"
	"time"

	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/worker"
	"go.uber.org/mock/gomock"
)

func TestWorkerEnsuresPendingRegistries(t *testing.T) {
	ctrl := gomock.NewController(t)
	clusterStore := mocks.NewMockClusterStore(ctrl)
	registryStore := mocks.NewMockClusterImageRegistryStore(ctrl)
	clusterResource := mocks.NewMockClusterResourceImageRegistryService(ctrl)
	clusterManager := mocks.NewMockClusterManager(ctrl)
	cluster := &models.Cluster{ID: "cluster-1"}
	registry := &models.ClusterImageRegistry{
		ID:        "registry-1",
		ClusterID: cluster.ID,
		Status:    &models.ClusterImageRegistryStatus{State: models.RegistryStatePending},
	}

	clusterStore.EXPECT().Get(gomock.Any(), cluster.ID).Return(cluster, nil)
	registryStore.EXPECT().ListByClusterIDInternal(gomock.Any(), cluster.ID).
		Return([]*models.ClusterImageRegistry{registry}, nil)
	clusterResource.EXPECT().EnsureImageRegistryInCluster(gomock.Any(), registry).Return(nil)

	w := NewClusterImageRegistryWorker(ClusterImageRegistryWorkerSpec{
		ClusterStore:       clusterStore,
		ImageRegistryStore: registryStore,
		ClusterManager:     clusterManager,
		ClusterResource:    clusterResource,
	})
	result, err := w.Execute(context.Background(), models.ClusterImageRegistryOperand{ClusterID: cluster.ID})
	if err != nil {
		t.Fatalf("Execute returned an error: %v", err)
	}
	if result != (worker.Result{}) {
		t.Fatalf("Execute returned unexpected result: %#v", result)
	}
}

func TestWorkerFinalizesClusterAfterRegistriesAreAbsent(t *testing.T) {
	ctrl := gomock.NewController(t)
	clusterStore := mocks.NewMockClusterStore(ctrl)
	registryStore := mocks.NewMockClusterImageRegistryStore(ctrl)
	clusterResource := mocks.NewMockClusterResourceImageRegistryService(ctrl)
	clusterManager := mocks.NewMockClusterManager(ctrl)
	deletionTime := time.Now().UTC()
	cluster := &models.Cluster{ID: "cluster-1", DeletionTimestamp: &deletionTime}
	registry := &models.ClusterImageRegistry{ID: "registry-1", ClusterID: cluster.ID}

	clusterStore.EXPECT().Get(gomock.Any(), cluster.ID).Return(cluster, nil)
	registryStore.EXPECT().ListByClusterIDInternal(gomock.Any(), cluster.ID).
		Return([]*models.ClusterImageRegistry{registry}, nil)
	clusterResource.EXPECT().DeleteImageRegistryInCluster(gomock.Any(), registry).Return(true, nil)
	registryStore.EXPECT().Delete(gomock.Any(), registry.ID).Return(nil)
	clusterManager.EXPECT().UnregisterCluster(cluster.ID).Return(nil)
	clusterStore.EXPECT().Delete(gomock.Any(), cluster.ID).Return(nil)

	w := NewClusterImageRegistryWorker(ClusterImageRegistryWorkerSpec{
		ClusterStore:       clusterStore,
		ImageRegistryStore: registryStore,
		ClusterManager:     clusterManager,
		ClusterResource:    clusterResource,
	})
	result, err := w.Execute(context.Background(), models.ClusterImageRegistryOperand{ClusterID: cluster.ID})
	if err != nil {
		t.Fatalf("Execute returned an error: %v", err)
	}
	if result != (worker.Result{}) {
		t.Fatalf("Execute returned unexpected result: %#v", result)
	}
}
