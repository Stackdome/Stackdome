export interface LogEntry {
  id: string;
  timestamp?: Date;
  message: string;
  raw: string;
  source?: string;
}

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export interface LogViewerProps {
  stackId: string;
  organizationId: string;
  className?: string;
}

export interface StackLogsTabProps {
  stackId: string;
  organizationId: string;
}
