import { useMemo, useState } from 'react';
import { LazyLog } from 'react-lazylog';
import { Card, CardContent } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Loader2, AlertCircle, Check, Clock, ChevronDown, Layers, ScrollText, WifiOff, RefreshCw } from 'lucide-react';
import { StatusPill, EmptyState, type StatusVariant } from '@/components/branded';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem } from '@/components/ui/command';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { LogViewerProps, ConnectionStatus, TimeRangeOption, LogFilters } from './types';
import { useLogStream } from './use-log-stream';
import { convertLogsToLazyLogFormat, getTimeRangeLabel } from './utils';

function connectionStatusInfo(status: ConnectionStatus): { variant: StatusVariant; label: string } {
  switch (status) {
    case 'connecting':
      return { variant: 'pending', label: 'Connecting' };
    case 'connected':
      return { variant: 'ready', label: 'Connected' };
    case 'disconnected':
      return { variant: 'neutral', label: 'Disconnected' };
    case 'error':
      return { variant: 'error', label: 'Error' };
    default:
      return { variant: 'neutral', label: 'Unknown' };
  }
}

export function LogViewer({ stackId, organizationId, resources = [], className = '' }: LogViewerProps) {
  const [sourceSelectOpen, setSourceSelectOpen] = useState(false);
  const [filters, setFilters] = useState<LogFilters>({
    sources: [],
    timeRange: 'live-4h',
  });

  const { logs, connectionStatus, error, retry } = useLogStream({
    stackId,
    organizationId,
    filters,
    enabled: true,
  });

  const availableSources = useMemo(() => {
    return resources.map(resource => resource.name).filter(Boolean);
  }, [resources]);

  const filteredLogs = useMemo(() => {
    if (filters.sources.length === 0) {
      return logs;
    }
    return logs.filter(log =>
      !log.source || filters.sources.includes(log.source)
    );
  }, [logs, filters.sources]);

  const setSources = (sources: string[]) => {
    setFilters((prev: LogFilters) => ({ ...prev, sources }));
  };

  const setTimeRange = (timeRange: TimeRangeOption) => {
    setFilters((prev: LogFilters) => ({ ...prev, timeRange }));
  };

  const statusInfo = connectionStatusInfo(connectionStatus);

  const logText = useMemo(() => {
    return convertLogsToLazyLogFormat(filteredLogs);
  }, [filteredLogs]);

  const toggleSource = (source: string) => {
    const newSources = filters.sources.includes(source)
      ? filters.sources.filter(s => s !== source)
      : [...filters.sources, source];
    setSources(newSources);
  };

  const timeRangeOptions: TimeRangeOption[] = [
    'live-4h',
    '30m',
    '1h',
    '4h',
    '24h',
    'all',
  ];

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Header with integrated filter controls */}
      <div className="flex items-center justify-between p-4 border-b">
        <div className="flex items-center gap-4">
          <h3 className="text-lg font-semibold">Stack Logs</h3>
          <StatusPill variant={statusInfo.variant}>{statusInfo.label}</StatusPill>
        </div>

        <div className="flex items-center gap-2">
          {/* Resources Multi-Select */}
          {availableSources.length > 0 && (
            <Popover open={sourceSelectOpen} onOpenChange={setSourceSelectOpen}>
              <PopoverTrigger asChild>
                <Button variant="outline" size="sm" className="h-8 w-48 justify-start">
                  <Layers className="h-4 w-4" />
                  Resources
                  <ChevronDown className="ml-auto h-4 w-4" />
                  {filters.sources.length > 0 && (
                    <Badge variant="secondary" className="ml-2 h-5 px-1.5 text-xs">
                      {filters.sources.length}
                    </Badge>
                  )}
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-48 p-0" align="start">
                <Command>
                  <CommandInput placeholder="Search resources..." />
                  <CommandEmpty>No resources found.</CommandEmpty>
                  <CommandGroup>
                    {availableSources.map(source => (
                      <CommandItem
                        key={source}
                        onSelect={() => toggleSource(source)}
                      >
                        <Check
                          className={`mr-2 h-4 w-4 ${
                            filters.sources.includes(source) ? 'opacity-100' : 'opacity-0'
                          }`}
                        />
                        {source}
                      </CommandItem>
                    ))}
                  </CommandGroup>
                </Command>
              </PopoverContent>
            </Popover>
          )}

          {/* Time Range Selector */}
          <Select value={filters.timeRange} onValueChange={setTimeRange}>
            <SelectTrigger className="w-48 h-8">
              <Clock className="h-4 w-4" />
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {timeRangeOptions.map(option => (
                <SelectItem key={option} value={option}>
                  {getTimeRangeLabel(option)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Log Display */}
      {logText ? (
        <Card>
          <CardContent className="p-0">
            <div className="h-96 border rounded-md bg-gray-900">
              <LazyLog
                text={logText}
                extraLines={1}
                enableSearch
                caseInsensitive
                selectableLines
                follow={filters.timeRange === 'live-4h'}
                height={384}
                style={{
                  backgroundColor: '#111827',
                  color: '#f9fafb',
                  fontSize: '13px',
                  fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
                }}
              />
            </div>
          </CardContent>
        </Card>
      ) : connectionStatus === 'connecting' ? (
        <EmptyState
          icon={<Loader2 className="h-6 w-6 animate-spin" />}
          title="Connecting to log stream"
          description="Waiting for the first event to arrive."
        />
      ) : connectionStatus === 'connected' ? (
        <EmptyState
          icon={<ScrollText className="h-6 w-6" />}
          title="No logs yet"
          description="Logs will appear once the stack starts emitting them."
        />
      ) : (
        <EmptyState
          icon={<WifiOff className="h-6 w-6" />}
          title="Disconnected from log stream"
          description="The connection dropped. Try reconnecting."
          action={
            <Button variant="outline" size="sm" onClick={retry}>
              <RefreshCw className="h-4 w-4" />
              Retry
            </Button>
          }
        />
      )}
    </div>
  );
}
