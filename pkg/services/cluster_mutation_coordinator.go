package services

// ClusterMutationCoordinator serializes Kubernetes writes for one cluster
// namespace. The Hub wires one shared implementation into workers and direct
// service mutation paths. This remains an in-process, single-Hub-replica fence.
type ClusterMutationCoordinator interface {
	LockClusterNamespace(clusterID, namespace string) func()
}
