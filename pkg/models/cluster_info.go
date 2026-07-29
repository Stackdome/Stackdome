package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

type ClusterInfoPhase string

const (
	ClusterInfoPhaseReady      ClusterInfoPhase = "Ready"
	ClusterInfoPhaseRefreshing ClusterInfoPhase = "Refreshing"
	ClusterInfoPhaseUnknown    ClusterInfoPhase = "Unknown"
)

type ClusterInfo struct {
	Phase             ClusterInfoPhase      `json:"phase,omitempty"`
	LastRefreshedAt   *time.Time            `json:"lastRefreshedAt,omitempty"`
	KubernetesVersion string                `json:"kubernetesVersion,omitempty"`
	TotalNodes        int                   `json:"totalNodes,omitempty"`
	ReadyNodes        int                   `json:"readyNodes,omitempty"`
	AvailabilityZones []string              `json:"availabilityZones,omitempty"`
	Nodes             []ClusterNode         `json:"nodes,omitempty"`
	StorageClasses    []ClusterStorageClass `json:"storageClasses,omitempty"`
	LoadBalancers     []ClusterLoadBalancer `json:"loadBalancers,omitempty"`
	IngressClasses    []ClusterIngressClass `json:"ingressClasses,omitempty"`
}

type ClusterNode struct {
	Name                     string             `json:"name"`
	Ready                    bool               `json:"ready"`
	AllocatableCPU           *resource.Quantity `json:"allocatableCpu,omitempty"`
	AllocatableMemory        *resource.Quantity `json:"allocatableMemory,omitempty"`
	AllocatableEphemeralDisk *resource.Quantity `json:"allocatableEphemeralDisk,omitempty"`
	CapacityEphemeralDisk    *resource.Quantity `json:"capacityEphemeralDisk,omitempty"`
	Zone                     string             `json:"zone,omitempty"`
	Region                   string             `json:"region,omitempty"`
}

type ClusterStorageClass struct {
	Name        string `json:"name"`
	Provisioner string `json:"provisioner,omitempty"`
	IsDefault   bool   `json:"isDefault"`
}

type ClusterLoadBalancer struct {
	ServiceName      string   `json:"serviceName"`
	ServiceNamespace string   `json:"serviceNamespace"`
	IngressIPs       []string `json:"ingressIPs,omitempty"`
	IngressHostnames []string `json:"ingressHostnames,omitempty"`
	HasIP            bool     `json:"hasIP"`
}

type ClusterIngressClass struct {
	Name       string `json:"name"`
	Controller string `json:"controller,omitempty"`
	IsDefault  bool   `json:"isDefault"`
}

func (c *ClusterInfo) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &c)
}

func (c ClusterInfo) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// DefaultStorageClass returns the name of the cluster's default StorageClass,
// or "" when none is marked default.
func (c *ClusterInfo) DefaultStorageClass() string {
	if c == nil {
		return ""
	}
	for _, sc := range c.StorageClasses {
		if sc.IsDefault {
			return sc.Name
		}
	}
	return ""
}
