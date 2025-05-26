package clients

import (
	corev1 "k8s.io/api/core/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type LogStreamObject struct {
	Data  string
	Error error
}

type MetricsStreamObject struct {
	NamespaceMetrics []metricsv1beta1.PodMetrics
	NodeCapacityMap  map[string]corev1.ResourceList
	Error            error
}
