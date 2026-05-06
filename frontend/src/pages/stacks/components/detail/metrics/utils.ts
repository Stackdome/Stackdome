import type { ResourceMetrics, MetricsDisplayData } from './types';

export function formatMemoryUsage(mebibytes: string | number): string {
  const numMi = typeof mebibytes === 'string' ? parseInt(mebibytes, 10) : mebibytes;

  if (isNaN(numMi) || numMi === 0) return '0 Mi';

  // Convert to appropriate unit
  if (numMi >= 1024) {
    return `${(numMi / 1024).toFixed(1)} Gi`;
  }

  return `${numMi} Mi`;
}


export function formatCpuUsage(millicores: string | number): string {
  const numMillicores = typeof millicores === 'string' ? parseInt(millicores, 10) : millicores;

  if (isNaN(numMillicores)) return '0m';

  if (numMillicores >= 1000) {
    return `${(numMillicores / 1000).toFixed(1)} cores`;
  }

  return `${numMillicores}m`;
}

export function convertToDisplayMetrics(metrics: ResourceMetrics): MetricsDisplayData {
  const cpuUsage = metrics.cpu_usage || '0';
  const memoryUsage = metrics.memory_usage || '0';

  return {
    cpu: formatCpuUsage(cpuUsage),
    memory: formatMemoryUsage(memoryUsage),
    status: 'running' // Default status - could be enhanced based on actual resource state
  };
}

export function getStatusColor(status: MetricsDisplayData['status']): string {
  switch (status) {
    case 'running':
      return 'bg-success';
    case 'pending':
      return 'bg-warn';
    case 'failed':
      return 'bg-danger';
    case 'stopped':
      return 'bg-muted-foreground';
    default:
      return 'bg-muted-foreground';
  }
}

export function getStatusText(status: MetricsDisplayData['status']): string {
  switch (status) {
    case 'running':
      return 'Running';
    case 'pending':
      return 'Pending';
    case 'failed':
      return 'Failed';
    case 'stopped':
      return 'Stopped';
    default:
      return 'Unknown';
  }
}
