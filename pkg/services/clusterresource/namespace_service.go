package clusterresource

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceClusterResourceService is an interface for managing namespace resources in a cluster.
type NamespaceClusterResourceService interface {
	CreateNamespaceInCluster(ctx context.Context, ns *models.Namespace) *ClusterResourceError
	DeleteNamespaceInCluster(ctx context.Context, ns *models.Namespace) *ClusterResourceError
}

type namespaceClusterResourceService struct {
	clusterService DBClusterService
	clusterManager clustermanager.ClusterManager
	logger         logger.Logger
}

type NamespaceClusterResourceServiceSpec struct {
	ClusterService DBClusterService
	ClusterManager clustermanager.ClusterManager
	Logger         logger.Logger
}

// NewNamespaceClusterResourceService creates a new NamespaceClusterResourceService
func NewNamespaceClusterResourceService(spec NamespaceClusterResourceServiceSpec) NamespaceClusterResourceService {
	return &namespaceClusterResourceService{
		clusterService: spec.ClusterService,
		clusterManager: spec.ClusterManager,
		logger:         spec.Logger,
	}
}

func labelsToMap(labels models.Labels) map[string]string {
	m := make(map[string]string)
	for _, l := range labels {
		m[l.Key] = l.Value
	}
	return m
}

func annotationsToMap(annotations models.Annotations) map[string]string {
	m := make(map[string]string)
	for _, a := range annotations {
		m[a.Key] = a.Value
	}
	return m
}

// CreateNamespaceInCluster creates a namespace in the cluster
func (s *namespaceClusterResourceService) CreateNamespaceInCluster(ctx context.Context, ns *models.Namespace) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, ns.OrganisationID)
	if err != nil {
		s.logger.Error(ctx, "failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	client, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		return newError("failed to get cluster client", clientGetErr)
	}

	desiredObject := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ns.Name,
			Labels:      labelsToMap(ns.Labels),
			Annotations: annotationsToMap(ns.Annotations),
		},
	}

	createErr := client.Create(ctx, desiredObject)
	if createErr != nil {
		s.logger.Error(ctx, "failed to create namespace in cluster: %v", createErr)
		return newError("failed to create namespace in cluster", createErr)
	}

	return nil
}

// DeleteNamespaceInCluster deletes a namespace in the cluster
func (s *namespaceClusterResourceService) DeleteNamespaceInCluster(ctx context.Context, ns *models.Namespace) *ClusterResourceError {
	cluster, err := s.clusterService.GetClusterForOrg(ctx, ns.OrganisationID)
	if err != nil {
		s.logger.Error(ctx, "failed to get cluster for org: %v", err)
		return newError("failed to get cluster for org", err)
	}

	client, clientGetErr := s.clusterManager.GetClient(cluster.ID)
	if clientGetErr != nil {
		return newError("failed to get cluster client", clientGetErr)
	}

	desiredObject := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns.Name,
		},
	}

	deleteErr := client.Delete(ctx, desiredObject)
	if deleteErr != nil {
		s.logger.Error(ctx, "failed to delete namespace in cluster: %v", deleteErr)
		return newError("failed to delete namespace in cluster", deleteErr)
	}

	return nil
}
