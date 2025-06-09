import { format, isValid, parseISO } from 'date-fns';
import type { LogEntry, TimeRangeOption } from './types';

export function generateLogId(): string {
  return `log_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

export function parseLogEntry(rawMessage: string, source?: string): LogEntry {
  const id = generateLogId();
  let timestamp: Date | undefined;
  let message = rawMessage.trim();
  let extractedSource = source;

  // Handle SSE data format that might have "data: " prefix
  if (message.startsWith('data: ')) {
    message = message.substring(6).trim();
  }

  // Try to extract service name from format like [service]: message
  const serviceRegex = /^\[([^\]]+)\]:\s*(.*)$/;
  const serviceMatch = message.match(serviceRegex);

  if (serviceMatch) {
    extractedSource = serviceMatch[1];
    message = serviceMatch[2].trim();
  }

  // Try to extract timestamp from common formats
  const timestampRegex = /^(\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}(?:\.\d{3})?(?:Z|[+-]\d{2}:\d{2})?)\s*(.*)$/;
  const timestampMatch = message.match(timestampRegex);

  if (timestampMatch) {
    const parsedDate = parseISO(timestampMatch[1]);
    if (isValid(parsedDate)) {
      timestamp = parsedDate;
      message = timestampMatch[2] || message;
    }
  }

  // If no timestamp found, try extracting from nginx format: 2025/06/01 04:20:08
  const nginxTimestampRegex = /^(\d{4}\/\d{2}\/\d{2}\s+\d{2}:\d{2}:\d{2})\s*(.*)$/;
  const nginxMatch = message.match(nginxTimestampRegex);

  if (nginxMatch && !timestamp) {
    // Convert nginx timestamp format to ISO format
    const nginxTime = nginxMatch[1].replace('/', '-').replace('/', '-').replace(' ', 'T') + 'Z';
    const parsedDate = parseISO(nginxTime);
    if (isValid(parsedDate)) {
      timestamp = parsedDate;
      message = nginxMatch[2] || message;
    }
  }

  return {
    id,
    timestamp,
    message,
    raw: rawMessage,
    source: extractedSource,
  };
}


export function formatLogTimestamp(timestamp?: Date): string {
  if (!timestamp || !isValid(timestamp)) {
    return '';
  }

  return format(timestamp, 'HH:mm:ss.SSS');
}

/**
 * Get unique sources from logs
 */
export function getUniqueSources(logs: LogEntry[]): string[] {
  const sources = logs
    .map(log => log.source)
    .filter((source): source is string => !!source);
  return Array.from(new Set(sources)).sort();
}

/**
 * Convert logs to LazyLog format for react-lazylog
 */
export function convertLogsToLazyLogFormat(logs: LogEntry[]): string {
  return logs.map(log => {
    const timestamp = log.timestamp ? `[${formatLogTimestamp(log.timestamp)}]` : '';
    const source = log.source ? `[${log.source}]` : '';
    return `${timestamp}${source} ${log.message}`;
  }).join('\n');
}

/**
 * Get time range filter date based on option
 */
export function getTimeRangeDate(option: TimeRangeOption): Date | null {
  if (option === 'all') {
    return null; // No filter
  }

  const now = new Date();

  switch (option) {
    case 'live-4h':
    case '4h':
      return new Date(now.getTime() - 4 * 60 * 60 * 1000);
    case '30m':
      return new Date(now.getTime() - 30 * 60 * 1000);
    case '1h':
      return new Date(now.getTime() - 60 * 60 * 1000);
    case '24h':
      return new Date(now.getTime() - 24 * 60 * 60 * 1000);
    default:
      return null;
  }
}

/**
 * Get display label for time range option
 */
export function getTimeRangeLabel(option: TimeRangeOption): string {
  switch (option) {
    case 'live-4h':
      return 'Live Tail (4h)';
    case '30m':
      return 'Last 30 minutes';
    case '1h':
      return 'Last 1 hour';
    case '4h':
      return 'Last 4 hours';
    case '24h':
      return 'Last 24 hours';
    case 'all':
      return 'Since beginning';
    default:
      return 'Unknown';
  }
}

/**
 * Convert TimeRangeOption to backend API parameters
 */
export function getLogStreamParams(option: TimeRangeOption): { follow: boolean; since?: string } {
  switch (option) {
    case 'live-4h':
      return { follow: true, since: '4h' };
    case '30m':
      return { follow: true, since: '30m' };
    case '1h':
      return { follow: true, since: '1h' };
    case '4h':
      return { follow: true, since: '4h' };
    case '24h':
      return { follow: true, since: '24h' };
    case 'all':
      return { follow: true }; // No since parameter means all logs
    default:
      return { follow: true };
  }
}
