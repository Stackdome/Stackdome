package worker

import (
	"fmt"

	"k8s.io/utils/keymutex"
)

// ClusterMutationCoordinator serializes Kubernetes writes within one namespace.
// All cluster-mutating workers in an environment must share one instance. This
// is an in-process fence: cloud must run one Hub replica until it has a
// distributed equivalent. Workers acquire their resource lock before this lock.
type ClusterMutationCoordinator struct {
	locks keymutex.KeyMutex
}

func NewClusterMutationCoordinator() *ClusterMutationCoordinator {
	return &ClusterMutationCoordinator{locks: keymutex.NewHashed(resourceLockShards)}
}

func (c *ClusterMutationCoordinator) LockClusterNamespace(clusterID, namespace string) func() {
	key := fmt.Sprintf("%d:%s:%d:%s", len(clusterID), clusterID, len(namespace), namespace)
	c.locks.LockKey(key)
	return func() {
		if err := c.locks.UnlockKey(key); err != nil {
			panic(fmt.Sprintf("unlock cluster namespace %q/%q: %v", clusterID, namespace, err))
		}
	}
}
