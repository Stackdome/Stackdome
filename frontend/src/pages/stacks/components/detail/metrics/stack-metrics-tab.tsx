import { useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Activity, Cpu, MemoryStick, Clock, Loader2, Wifi, WifiOff, AlertCircle } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { Badge } from "@/components/ui/badge";
import { useMetricsStream } from "./use-metrics-stream";
import type { StackResource } from "@/pages/stacks/types";
import type { ResourceMetricsData } from "./types";
import {
  convertToDisplayMetrics,
  getStatusColor,
  getStatusText,
  formatMemoryUsage,
  formatCpuUsage
} from "./utils";

type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

function getConnectionStatusInfo(status: ConnectionStatus) {
  switch (status) {
    case 'connecting':
      return {
        icon: Loader2,
        text: 'Connecting...',
        className: 'text-yellow-600 bg-yellow-50 border-yellow-200',
        iconClass: 'animate-spin',
      };
    case 'connected':
      return {
        icon: Wifi,
        text: 'Connected',
        className: 'text-green-600 bg-green-50 border-green-200',
        iconClass: '',
      };
    case 'disconnected':
      return {
        icon: WifiOff,
        text: 'Disconnected',
        className: 'text-gray-600 bg-gray-50 border-gray-200',
        iconClass: '',
      };
    case 'error':
      return {
        icon: AlertCircle,
        text: 'Error',
        className: 'text-red-600 bg-red-50 border-red-200',
        iconClass: '',
      };
    default:
      return {
        icon: WifiOff,
        text: 'Unknown',
        className: 'text-gray-600 bg-gray-50 border-gray-200',
        iconClass: '',
      };
  }
}

interface StackMetricsTabProps {
  stackId: string;
  organizationId: string;
  resources: StackResource[];
}

export function StackMetricsTab({ stackId, organizationId, resources }: StackMetricsTabProps) {
  // Use the metrics streaming hook - enabled immediately
  const {
    stackMetrics,
    resourceMetrics,
    connectionStatus,
    error,
    updateResources
  } = useMetricsStream({
    stackId,
    organizationId,
    enabled: true
  });

  // Update streaming resources when resources prop changes
  useEffect(() => {
    updateResources(resources.map(r => r.name));
  }, [resources, updateResources]);

  // Convert streamed resource metrics to display format
  const currentResourceMetrics: ResourceMetricsData[] = Array.from(resourceMetrics.entries()).map(([resourceName, metrics]) => ({
    resourceName,
    metrics,
    displayMetrics: convertToDisplayMetrics(metrics)
  }));

  const statusInfo = getConnectionStatusInfo(connectionStatus);
  const StatusIcon = statusInfo.icon;

  return (
    <div className="space-y-4">
      {/* Header with connection status */}
      <div className="flex items-center justify-between p-4 border-b">
        <div className="flex items-center gap-4">
          <h3 className="text-lg font-semibold">Stack Metrics</h3>

          {/* Connection Status Badge */}
          <Badge variant="outline" className={statusInfo.className}>
            <StatusIcon className={`mr-2 h-3 w-3 ${statusInfo.iconClass}`} />
            {statusInfo.text}
          </Badge>
        </div>
      </div>

      {/* Error Display */}
      {error && connectionStatus === 'error' && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Stack-level metrics */}
      {stackMetrics && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Activity className="h-5 w-5" />
              <span>Stack Overview</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* CPU Usage */}
              <div className="space-y-2">
                <div className="flex items-center space-x-2">
                  <Cpu className="h-4 w-4 text-blue-500" />
                  <span className="text-sm font-medium">CPU Usage</span>
                </div>
                <div className="text-2xl font-bold">
                  {stackMetrics.cpu_usage ? formatCpuUsage(stackMetrics.cpu_usage) : 'N/A'}
                </div>
              </div>

              {/* Memory Usage */}
              <div className="space-y-2">
                <div className="flex items-center space-x-2">
                  <MemoryStick className="h-4 w-4 text-green-500" />
                  <span className="text-sm font-medium">Memory Usage</span>
                </div>
                <div className="text-2xl font-bold">
                  {stackMetrics.memory_usage ? formatMemoryUsage(stackMetrics.memory_usage) : 'N/A'}
                </div>
              </div>
            </div>

            {/* Timestamp */}
            {stackMetrics.timestamp && (
              <div className="mt-4 flex items-center space-x-2 text-sm text-muted-foreground">
                <Clock className="h-4 w-4" />
                <span>Last updated: {new Date(stackMetrics.timestamp).toLocaleString()}</span>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Resource-level metrics */}
      <div>
        <h3 className="text-md font-semibold mb-4">Resource Metrics</h3>
        {currentResourceMetrics.length === 0 ? (
          <Card>
            <CardContent className="py-8">
              <div className="text-center text-muted-foreground">
                <Activity className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p>No resource metrics available</p>
              </div>
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {currentResourceMetrics.map((resourceData) => (
              <Card key={resourceData.resourceName} className="relative">
                <CardHeader className="pb-3">
                  <CardTitle className="flex items-center justify-between text-base">
                    <span className="truncate">{resourceData.resourceName}</span>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <div className={`h-3 w-3 rounded-full ${getStatusColor(resourceData.displayMetrics.status)}`} />
                      </TooltipTrigger>
                      <TooltipContent>
                        <p>{getStatusText(resourceData.displayMetrics.status)}</p>
                      </TooltipContent>
                    </Tooltip>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center space-x-2">
                        <Cpu className="h-4 w-4 text-blue-500" />
                        <span className="text-muted-foreground">CPU</span>
                      </div>
                      <span className="font-medium">{resourceData.displayMetrics.cpu}</span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center space-x-2">
                        <MemoryStick className="h-4 w-4 text-green-500" />
                        <span className="text-muted-foreground">Memory</span>
                      </div>
                      <span className="font-medium">{resourceData.displayMetrics.memory}</span>
                    </div>
                  </div>

                  {/* Last updated timestamp */}
                  {resourceData.metrics.timestamp && (
                    <div className="text-xs text-muted-foreground flex items-center space-x-1 border-t pt-2">
                      <Clock className="h-3 w-3" />
                      <span>{new Date(resourceData.metrics.timestamp).toLocaleString()}</span>
                    </div>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
