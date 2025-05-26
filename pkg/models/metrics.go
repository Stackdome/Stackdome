package models

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	AssignedNodeAnnotation = "assigned-node"
)

type ResourceMetrics struct {
	AssignedNodes  []string          `json:"assigned_node"` // Node where the resource is assigned
	CPUUsage       resource.Quantity `json:"cpu_usage"`     // CPU usage in percentage
	CPULimit       resource.Quantity `json:"cpu_limit"`     // CPU limit in cores
	MemoryUsage    resource.Quantity `json:"memory_usage"`  // Memory usage in percentage
	MemoryLimit    resource.Quantity `json:"memory_limit"`  // Memory limit in bytes
	NetworkIO      float64           `json:"network_io"`    // Network I/O in bytes
	StorageIO      float64           `json:"storage_io"`    // Storage I/O in bytes
	NodeCapacities []*NodeCapacity   `json:"node_capacity"` // Node capacity metrics
	TimeStamp      time.Time         `json:"time_stamp"`    // Timestamp of the metrics
}

type NodeCapacity struct {
	NodeName string             `json:"node_name"` // Name of the node
	CPU      *resource.Quantity `json:"cpu"`       // Total CPU capacity
	Memory   *resource.Quantity `json:"memory"`    // Total memory capacity
	Storage  *resource.Quantity `json:"storage"`   // Total storage capacity
}
