import { useState, useEffect, useRef, useCallback } from 'react';
import type { LogEntry, ConnectionStatus, LogFilters } from './types';
import { parseLogEntry, getLogStreamParams } from './utils';
import {
  buildStackLogStreamUrl,
  buildStackResourceLogStreamUrl,
  attachSseHandlers,
  aggregateStreamStatus,
  type SseStreamStatus,
} from '@/api/observability';
import { useResourceProjects } from '@/hooks/use-resource-projects';

interface UseLogStreamProps {
  stackId: string;
  organizationId: string;
  /** Sources must already be filtered to ready resources by the caller. */
  filters: LogFilters;
  enabled?: boolean;
}

interface UseLogStreamReturn {
  logs: LogEntry[];
  connectionStatus: ConnectionStatus;
  /** Set only from the backend's `event: error` SSE events — never from
   *  connection drops, which surface through connectionStatus instead. */
  error: string | null;
  retry: () => void;
}

// Key for the whole-stack stream in the per-source status map.
const STACK_STREAM_KEY = '__stack__';

export function useLogStream({
  stackId,
  organizationId,
  filters,
  enabled = true,
}: UseLogStreamProps): UseLogStreamReturn {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('disconnected');
  const [error, setError] = useState<string | null>(null);
  const [retryNonce, setRetryNonce] = useState(0);
  const eventSourcesRef = useRef<EventSource[]>([]);
  const { defaultProjectName } = useResourceProjects();

  const retry = useCallback(() => setRetryNonce((n) => n + 1), []);

  // filters.sources identity changes on unrelated re-renders; key on content.
  const sourcesKey = filters.sources.join(',');

  useEffect(() => {
    if (!enabled || !stackId || !organizationId || !defaultProjectName) {
      setConnectionStatus('disconnected');
      return;
    }

    const streamParams = getLogStreamParams(filters.timeRange);

    // Clear previous logs when filters change
    setLogs([]);
    setConnectionStatus('connecting');
    setError(null);

    const sourceNames = sourcesKey ? sourcesKey.split(',') : [];
    const streamKeys = sourceNames.length === 0 ? [STACK_STREAM_KEY] : sourceNames;
    const statuses = new Map<string, SseStreamStatus>(streamKeys.map((key) => [key, 'connecting']));

    const applyStatus = (key: string, status: SseStreamStatus) => {
      statuses.set(key, status);
      setConnectionStatus(aggregateStreamStatus([...statuses.values()]));
    };

    eventSourcesRef.current = streamKeys.map((key) => {
      const url =
        key === STACK_STREAM_KEY
          ? buildStackLogStreamUrl(organizationId, defaultProjectName, stackId, streamParams)
          : buildStackResourceLogStreamUrl(organizationId, defaultProjectName, stackId, key, streamParams);
      const eventSource = new EventSource(url);

      attachSseHandlers(eventSource, {
        onData: (data) => {
          if (!data || !data.trim()) return;
          const logEntry = parseLogEntry(data.trim());
          if (key !== STACK_STREAM_KEY) {
            logEntry.source = key;
          }
          if (logEntry.message && logEntry.message.length > 0) {
            setLogs((prevLogs) => [...prevLogs, logEntry]);
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
  }, [stackId, organizationId, defaultProjectName, filters.timeRange, sourcesKey, enabled, retryNonce]);

  return {
    logs,
    connectionStatus,
    error,
    retry,
  };
}
