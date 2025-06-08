import { format, isValid, parseISO } from 'date-fns';
import type { LogEntry } from './types';

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
