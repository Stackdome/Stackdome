import { useState, useEffect, useRef } from 'react';
import type { LogEntry, ConnectionStatus, LogFilters } from './types';
import { parseLogEntry, getLogStreamParams } from './utils';
import { buildStackLogStreamUrl, buildStackResourceLogStreamUrl } from '@/api/observability';

interface UseLogStreamProps {
  stackId: string;
  organizationId: string;
  filters: LogFilters;
  enabled?: boolean;
}

interface UseLogStreamReturn {
  logs: LogEntry[];
  connectionStatus: ConnectionStatus;
  error: string | null;
}

export function useLogStream({
  stackId,
  organizationId,
  filters,
  enabled = true,
}: UseLogStreamProps): UseLogStreamReturn {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('disconnected');
  const [error, setError] = useState<string | null>(null);
  const eventSourcesRef = useRef<EventSource[]>([]);

  useEffect(() => {
    if (!enabled || !stackId || !organizationId) {
      return;
    }

    // Convert time range to backend parameters
    const streamParams = getLogStreamParams(filters.timeRange);

    // Clear previous logs when filters change
    setLogs([]);
    setConnectionStatus('connecting');
    setError(null);

    // Close any existing connections
    eventSourcesRef.current.forEach(source => source.close());
    eventSourcesRef.current = [];

    try {
      let activeConnections = 0;
      let connectedSources = 0;

      const handleConnectionOpen = () => {
        connectedSources++;
        if (connectedSources === activeConnections) {
          setConnectionStatus('connected');
          setError(null);
        }
      };

      const handleConnectionError = (sourceName?: string) => {
        setConnectionStatus('error');
        setError(`Connection lost${sourceName ? ` to ${sourceName}` : ''}. Please refresh to retry.`);
      };

      if (filters.sources.length === 0) {
        // No specific sources selected, get logs from the entire stack
        activeConnections = 1;
        const url = buildStackLogStreamUrl(organizationId, stackId, streamParams);
        const eventSource = new EventSource(url);
        eventSourcesRef.current.push(eventSource);

        eventSource.onopen = handleConnectionOpen;

        eventSource.onmessage = (event) => {
          console.log('SSE Event received (stack):', event.data);

          if (event.data && event.data.trim()) {
            const rawData = event.data.trim();
            console.log('Processing SSE data (stack):', rawData);

            const logEntry = parseLogEntry(rawData);
            console.log('Created log entry (stack):', logEntry);

            if (logEntry.message && logEntry.message.length > 0) {
              setLogs(prevLogs => [...prevLogs, logEntry]);
            } else {
              console.warn('Skipping empty log entry (stack):', logEntry);
            }
          }
        };

        eventSource.onerror = () => handleConnectionError();
      } else {
        // Specific sources selected, get logs from individual resources
        activeConnections = filters.sources.length;

        filters.sources.forEach(resourceName => {
          const url = buildStackResourceLogStreamUrl(organizationId, stackId, resourceName, streamParams);
          const eventSource = new EventSource(url);
          eventSourcesRef.current.push(eventSource);

          eventSource.onopen = handleConnectionOpen;

          eventSource.onmessage = (event) => {
            console.log(`SSE Event received (${resourceName}):`, event.data);

            if (event.data && event.data.trim()) {
              const rawData = event.data.trim();
              console.log(`Processing SSE data (${resourceName}):`, rawData);

              // Parse the log entry and ensure it has the resource name as source
              const logEntry = parseLogEntry(rawData);
              logEntry.source = resourceName; // Ensure source is set to the resource name
              console.log(`Created log entry (${resourceName}):`, logEntry);

              if (logEntry.message && logEntry.message.length > 0) {
                setLogs(prevLogs => [...prevLogs, logEntry]);
              } else {
                console.warn(`Skipping empty log entry (${resourceName}):`, logEntry);
              }
            }
          };

          eventSource.onerror = () => handleConnectionError(resourceName);
        });
      }

    } catch (err) {
      setConnectionStatus('error');
      setError(err instanceof Error ? err.message : 'Failed to connect to log stream');
    }

    return () => {
      eventSourcesRef.current.forEach(source => source.close());
      eventSourcesRef.current = [];
    };
  }, [stackId, organizationId, filters.timeRange, filters.sources, enabled]);

  return {
    logs,
    connectionStatus,
    error,
  };
}
