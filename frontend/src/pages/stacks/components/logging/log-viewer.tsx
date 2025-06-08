import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Loader2, Wifi, WifiOff, AlertCircle } from 'lucide-react';
import type { LogViewerProps, ConnectionStatus } from './types';
import { useLogStream } from './use-log-stream';
import { formatLogTimestamp } from './utils';
import { useEffect, useRef } from 'react';

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

export function LogViewer({ stackId, organizationId, className = '' }: LogViewerProps) {
  const { logs, connectionStatus, error } = useLogStream({
    stackId,
    organizationId,
    enabled: true,
  });

  const statusInfo = getConnectionStatusInfo(connectionStatus);
  const StatusIcon = statusInfo.icon;

  // Auto-scroll reference
  const logEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (logEndRef.current) {
      logEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs.length]);

  // Debug logging
  console.log('LogViewer Debug:', {
    logsCount: logs.length,
    connectionStatus,
    error
  });

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Connection Status */}
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Stack Logs</h3>
        <Badge variant="outline" className={statusInfo.className}>
          <StatusIcon className={`mr-2 h-3 w-3 ${statusInfo.iconClass}`} />
          {statusInfo.text}
        </Badge>
      </div>

      {/* Error Display */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Log Display */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-sm text-muted-foreground">
            Showing {logs.length} log entries
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <div className="h-96 border rounded-md bg-gray-900 text-gray-100">
            {logs.length > 0 ? (
              <div className="h-full overflow-auto p-4 font-mono text-sm">
                {logs.map((log, index) => (
                  <div key={log.id || index} className="mb-1 whitespace-pre-wrap leading-snug">
                    {log.timestamp && <span className="text-blue-400">[{formatLogTimestamp(log.timestamp)}]</span>}
                    {log.source && <span className="text-green-400 ml-1">[{log.source}]</span>}
                    <span className="text-gray-100 ml-1">{log.message}</span>
                  </div>
                ))}
                <div ref={logEndRef} />
              </div>
            ) : (
              <div className="h-96 flex items-center justify-center text-gray-400">
                {connectionStatus === 'connecting' ? (
                  <div className="flex items-center gap-2">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    <span>Connecting to log stream...</span>
                  </div>
                ) : connectionStatus === 'connected' ? (
                  <span>No logs available</span>
                ) : (
                  <span>Disconnected from log stream</span>
                )}
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
