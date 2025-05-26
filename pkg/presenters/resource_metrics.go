package presenters

import (
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

func PresentResourceMetrics(in *models.ResourceMetrics) *openapi.ResourceMetrics {
	if in == nil {
		return nil
	}
	res := &openapi.ResourceMetrics{
		CpuUsage:       presentCPUUnit(&in.CPUUsage),
		MemoryUsage:    presentMemoryUnit(&in.MemoryUsage),
		Timestamp:      &in.TimeStamp,
		NodeCapacities: make([]openapi.ResourceMetricsNodeCapacitiesInner, 0),
	}

	if len(in.AssignedNodes) != 0 {
		res.AssignedNodes = in.AssignedNodes
	}

	if len(in.NodeCapacities) != 0 {
		for _, nc := range in.NodeCapacities {
			res.NodeCapacities = append(res.NodeCapacities, openapi.ResourceMetricsNodeCapacitiesInner{
				NodeName:        &nc.NodeName,
				CpuCapacity:     presentCPUUnit(nc.CPU),
				MemoryCapacity:  presentMemoryUnit(nc.Memory),
				StorageCapacity: ptr.To(nc.Storage.String()),
			})
		}
	}
	return res
}

func presentMemoryUnit(in *resource.Quantity) *string {
	if in == nil {
		return nil
	}
	unit := fmt.Sprintf("%dmi", in.ScaledValue(resource.Mega))
	return &unit
}

func presentCPUUnit(in *resource.Quantity) *string {
	if in == nil {
		return nil
	}
	unit := fmt.Sprintf("%dm", in.ScaledValue(resource.Milli))
	return &unit
}
