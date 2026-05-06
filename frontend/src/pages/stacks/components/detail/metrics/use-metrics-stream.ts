import { useState, useEffect, useRef, useCallback } from 'react';
import type { ResourceMetrics } from './types';

type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

interface UseMetricsStreamProps {
  stackId: string;
  organizationId: string;
  enabled?: boolean;
}

interface UseMetricsStreamReturn {
  stackMetrics: ResourceMetrics | null;
  resourceMetrics: Map<string, ResourceMetrics>;
  isStreaming: boolean;
  connectionStatus: ConnectionStatus;
  error: string | null;
  startStreaming: () => void;
  stopStreaming: () => void;
  retry: () => void;
  lastUpdated: Date | null;
  updateResources: (resources: string[]) => void;
}

export function useMetricsStream({
  stackId,
  organizationId,
  enabled = true,
}: UseMetricsStreamProps): UseMetricsStreamReturn {
  const [stackMetrics, setStackMetrics] = useState<ResourceMetrics | null>(null);
  const [resourceMetrics, setResourceMetrics] = useState<Map<string, ResourceMetrics>>(new Map());
  const [isStreaming, setIsStreaming] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('disconnected');
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const eventSourcesRef = useRef<EventSource[]>([]);
  const resourcesRef = useRef<string[]>([]);
  const enabledRef = useRef(enabled);

  // Update enabled ref when prop changes
  useEffect(() => {
    enabledRef.current = enabled;
  }, [enabled]);

  const buildStackMetricsStreamUrl = useCallback(() => {
    const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
    return `${baseUrl}/organizations/${organizationId}/stacks/${stackId}/metrics?stream=true`;
  }, [organizationId, stackId]);

  const buildResourceMetricsStreamUrl = useCallback((resourceName: string) => {
    const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
    return `${baseUrl}/organizations/${organizationId}/stacks/${stackId}/resources/${resourceName}/metrics?stream=true`;
  }, [organizationId, stackId]);

  const parseMetricsData = useCallback((data: string): ResourceMetrics | null => {
    try {
      const parsed = JSON.parse(data);
      return parsed;
    } catch (err) {
      console.warn('Failed to parse metrics data:', err);
      return null;
    }
  }, []);

  const handleConnectionError = useCallback(() => {
    console.warn('Metrics stream connection error, attempting reconnection...');
    setConnectionStatus('error');
    setError('Connection lost. Please refresh to retry.');
    // The browser will automatically attempt to reconnect EventSource
  }, []);

  const stopStreaming = useCallback(() => {
    console.log('Stopping metrics streaming...');

    // Close all event sources
    eventSourcesRef.current.forEach(eventSource => {
      eventSource.close();
    });
    eventSourcesRef.current = [];

    setIsStreaming(false);
    setConnectionStatus('disconnected');
    setError(null);
  }, []);

  const startStreaming = useCallback(() => {
    if (isStreaming) {
      return;
    }

    console.log('Starting metrics streaming...');
    setIsStreaming(true);
    setConnectionStatus('connecting');
    setError(null);

    const activeConnections = 1 + resourcesRef.current.length; // Stack + resources
    let connectedSources = 0;

    const handleConnectionOpen = () => {
      connectedSources++;
      if (connectedSources === activeConnections) {
        setConnectionStatus('connected');
        setError(null);
      }
    };

    // Stream stack-level metrics
    const stackUrl = buildStackMetricsStreamUrl();
    const stackEventSource = new EventSource(stackUrl);
    eventSourcesRef.current.push(stackEventSource);

    stackEventSource.onopen = () => {
      console.log('Stack metrics stream connected');
      handleConnectionOpen();
    };

    stackEventSource.onmessage = (event) => {
      const metrics = parseMetricsData(event.data);
      if (metrics) {
        setStackMetrics(metrics);
        setLastUpdated(new Date());
      }
    };

    stackEventSource.onerror = () => {
      handleConnectionError();
    };

    // Stream resource-level metrics for each resource
    resourcesRef.current.forEach(resourceName => {
      const resourceUrl = buildResourceMetricsStreamUrl(resourceName);
      const resourceEventSource = new EventSource(resourceUrl);
      eventSourcesRef.current.push(resourceEventSource);

      resourceEventSource.onopen = () => {
        console.log(`Resource metrics stream connected for ${resourceName}`);
        handleConnectionOpen();
      };

      resourceEventSource.onmessage = (event) => {
        const metrics = parseMetricsData(event.data);
        if (metrics) {
          setResourceMetrics(prev => {
            const newMap = new Map(prev);
            newMap.set(resourceName, metrics);
            return newMap;
          });
          setLastUpdated(new Date());
        }
      };

      resourceEventSource.onerror = () => {
        console.warn(`Resource metrics stream error for ${resourceName}`);
        handleConnectionError();
      };
    });
  }, [isStreaming, buildStackMetricsStreamUrl, buildResourceMetricsStreamUrl, parseMetricsData, handleConnectionError]);

  // Update resources list - simplified to avoid infinite loops
  const updateResources = useCallback((resources: string[]) => {
    resourcesRef.current = resources;
  }, []);

  // Auto-start streaming when enabled
  useEffect(() => {
    if (enabled && !isStreaming) {
      const timer = setTimeout(() => {
        if (enabledRef.current && !isStreaming) {
          startStreaming();
        }
      }, 100);
      return () => clearTimeout(timer);
    }
  }, [enabled, isStreaming, startStreaming]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      stopStreaming();
    };
  }, [stopStreaming]);

  // Tearing down clears connections + flips isStreaming to false. The
  // auto-start effect above then sees enabled && !isStreaming and reconnects.
  const retry = useCallback(() => {
    stopStreaming();
  }, [stopStreaming]);

  return {
    stackMetrics,
    resourceMetrics,
    isStreaming,
    connectionStatus,
    error,
    startStreaming,
    stopStreaming,
    retry,
    lastUpdated,
    updateResources,
  };
}
