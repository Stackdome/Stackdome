import { useState, useEffect, useRef, useCallback } from 'react';
import type { ResourceMetrics } from './types';
import { useResourceProjects } from '@/hooks/use-resource-projects';
import {
  buildStackMetricsStreamUrl,
  buildStackResourceMetricsStreamUrl,
  attachSseHandlers,
  aggregateStreamStatus,
  STACK_STREAM_KEY,
  type SseStreamStatus,
} from '@/api/observability';

interface UseMetricsStreamProps {
  stackId: string;
  organizationId: string;
  /** Names must already be filtered to ready resources by the caller. */
  resourceNames: string[];
  enabled?: boolean;
}

interface UseMetricsStreamReturn {
  stackMetrics: ResourceMetrics | null;
  resourceMetrics: Map<string, ResourceMetrics>;
  connectionStatus: SseStreamStatus;
  /** Set only from the backend's `event: error` SSE events — never from
   *  connection drops, which surface through connectionStatus instead. */
  error: string | null;
  retry: () => void;
}

export function useMetricsStream({
  stackId,
  organizationId,
  resourceNames,
  enabled = true,
}: UseMetricsStreamProps): UseMetricsStreamReturn {
  const [stackMetrics, setStackMetrics] = useState<ResourceMetrics | null>(null);
  const [resourceMetrics, setResourceMetrics] = useState<Map<string, ResourceMetrics>>(new Map());
  const [connectionStatus, setConnectionStatus] = useState<SseStreamStatus>('disconnected');
  const [error, setError] = useState<string | null>(null);
  const [retryNonce, setRetryNonce] = useState(0);
  const eventSourcesRef = useRef<EventSource[]>([]);
  const { defaultProjectName } = useResourceProjects();

  const retry = useCallback(() => setRetryNonce((n) => n + 1), []);

  // resourceNames identity changes on unrelated re-renders; key on content.
  const namesKey = resourceNames.join(',');

  useEffect(() => {
    if (!enabled || !stackId || !organizationId || !defaultProjectName) {
      setConnectionStatus('disconnected');
      return;
    }

    setConnectionStatus('connecting');
    setError(null);

    const names = namesKey ? namesKey.split(',') : [];
    const streamKeys = [STACK_STREAM_KEY, ...names];
    const statuses = new Map<string, SseStreamStatus>(streamKeys.map((key) => [key, 'connecting']));

    const applyStatus = (key: string, status: SseStreamStatus) => {
      statuses.set(key, status);
      setConnectionStatus(aggregateStreamStatus([...statuses.values()]));
    };

    const parseMetrics = (data: string): ResourceMetrics | null => {
      try {
        return JSON.parse(data);
      } catch {
        return null;
      }
    };

    eventSourcesRef.current = streamKeys.map((key) => {
      const url =
        key === STACK_STREAM_KEY
          ? buildStackMetricsStreamUrl(organizationId, defaultProjectName, stackId)
          : buildStackResourceMetricsStreamUrl(organizationId, defaultProjectName, stackId, key);
      const eventSource = new EventSource(url);

      attachSseHandlers(eventSource, {
        onData: (data) => {
          const metrics = parseMetrics(data);
          if (!metrics) return;
          setError(null); // flowing data supersedes a stale stream error
          if (key === STACK_STREAM_KEY) {
            setStackMetrics(metrics);
          } else {
            setResourceMetrics((prev) => new Map(prev).set(key, metrics));
          }
        },
        onStreamError: (message) => setError(message),
        onStatusChange: (status) => applyStatus(key, status),
      });

      return eventSource;
    });

    return () => {
      eventSourcesRef.current.forEach((source) => source.close());
      eventSourcesRef.current = [];
    };
  }, [stackId, organizationId, defaultProjectName, namesKey, enabled, retryNonce]);

  return {
    stackMetrics,
    resourceMetrics,
    connectionStatus,
    error,
    retry,
  };
}
