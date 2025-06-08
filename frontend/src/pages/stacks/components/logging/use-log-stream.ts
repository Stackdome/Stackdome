import { useState, useEffect, useRef } from 'react';
import type { LogEntry, ConnectionStatus } from './types';
import { parseLogEntry } from './utils';
import { buildStackLogStreamUrl } from '@/api/observability';

interface UseLogStreamProps {
  stackId: string;
  organizationId: string;
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
  enabled = true,
}: UseLogStreamProps): UseLogStreamReturn {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('disconnected');
  const [error, setError] = useState<string | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!enabled || !stackId || !organizationId) {
      return;
    }

    const url = buildStackLogStreamUrl(organizationId, stackId);
    setConnectionStatus('connecting');
    setError(null);

    try {
      const eventSource = new EventSource(url);
      eventSourceRef.current = eventSource;

      eventSource.onopen = () => {
        setConnectionStatus('connected');
        setError(null);
      };

      eventSource.onmessage = (event) => {
        console.log('SSE Event received:', event.data);

        // The event.data contains the actual log line from SSE
        if (event.data && event.data.trim()) {
          const rawData = event.data.trim();
          console.log('Processing SSE data:', rawData);

          // Parse the log entry directly from the raw SSE data
          const logEntry = parseLogEntry(rawData);
          console.log('Created log entry:', logEntry);

          // Only add if we have a valid message
          if (logEntry.message && logEntry.message.length > 0) {
            setLogs(prevLogs => [...prevLogs, logEntry]);
          } else {
            console.warn('Skipping empty log entry:', logEntry);
          }
        }
      };

      eventSource.onerror = () => {
        setConnectionStatus('error');
        setError('Connection lost. Please refresh to retry.');
      };

    } catch (err) {
      setConnectionStatus('error');
      setError(err instanceof Error ? err.message : 'Failed to connect to log stream');
    }

    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
    };
  }, [stackId, organizationId, enabled]);

  return {
    logs,
    connectionStatus,
    error,
  };
}
