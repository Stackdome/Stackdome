// Types for metrics based on the updated OpenAPI ResourceMetrics schema

export interface ResourceMetrics {
  cpu_usage?: string; // CPU usage in millicores
  memory_usage?: string; // Memory usage in Mi (mebibytes)
  timestamp?: string; // ISO date-time
}

export interface MetricsDisplayData {
  cpu: string;
  memory: string;
  status: 'running' | 'stopped' | 'pending' | 'failed';
}

export interface ResourceMetricsData {
  resourceName: string;
  metrics: ResourceMetrics;
  displayMetrics: MetricsDisplayData;
}
