import { useMemo, useState } from 'react';
import { LazyLog } from 'react-lazylog';
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
import { isResourceReady } from '@/pages/stacks/lib/resource-readiness';

function connectionStatusInfo(status: ConnectionStatus): { variant: StatusVariant; label: string } {
  switch (status) {
    case 'connecting':
      return { variant: 'pending', label: 'Connecting' };
    case 'connected':
      return { variant: 'ready', label: 'Connected' };
    case 'reconnecting':
      return { variant: 'pending', label: 'Reconnecting' };
    case 'disconnected':
      return { variant: 'neutral', label: 'Disconnected' };
    default:
      return { variant: 'neutral', label: 'Unknown' };
  }
}

export function LogViewer({ stackId, organizationId, resources = [], liveStatusResources, initialSources, className = '' }: LogViewerProps) {
  const [sourceSelectOpen, setSourceSelectOpen] = useState(false);
  const [filters, setFilters] = useState<LogFilters>({
    sources: initialSources ?? [],
    timeRange: 'live-4h',
  });

  // Streams only open for Ready resources — the backend rejects the rest with
  // a pre-stream HTTP error that EventSource can't distinguish from an outage.
  const readySources = useMemo(() => {
    return new Set(
      resources
        .map((r) => r.name)
        .filter((name) => name && isResourceReady(liveStatusResources?.[name]?.state)),
    );
  }, [resources, liveStatusResources]);

  const streamSources = useMemo(
    () => filters.sources.filter((s) => readySources.has(s)),
    [filters.sources, readySources],
  );
  const canStream =
    readySources.size > 0 && (filters.sources.length === 0 || streamSources.length > 0);

  const { logs, connectionStatus, error, retry } = useLogStream({
    stackId,
    organizationId,
    filters: useMemo(() => ({ ...filters, sources: streamSources }), [filters, streamSources]),
    enabled: canStream,
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

  const statusInfo = canStream
    ? connectionStatusInfo(connectionStatus)
    : { variant: 'neutral' as StatusVariant, label: 'Waiting' };

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
    <div className={`mx-auto max-w-[1100px] px-[30px] py-6 ${className}`}>
      {/* Header with integrated filter controls */}
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h2 className="text-[18px] font-medium tracking-[-0.01em] text-foreground">Stack logs</h2>
          <StatusPill variant={statusInfo.variant}>{statusInfo.label}</StatusPill>
        </div>

        <div className="flex items-center gap-2.5">
          {/* Resources Multi-Select */}
          {availableSources.length > 0 && (
            <Popover open={sourceSelectOpen} onOpenChange={setSourceSelectOpen}>
              <PopoverTrigger asChild>
                <Button variant="outline" size="sm" className="h-8 w-48 justify-start rounded-sm text-[12.5px] font-medium">
                  <Layers className="h-3.5 w-3.5" />
                  Resources
                  <ChevronDown className="ml-auto h-3.5 w-3.5" />
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
                        <span className="truncate">{source}</span>
                        {!readySources.has(source) && (
                          <span className="ml-auto pl-2 text-[10px] text-fg-muted">not ready</span>
                        )}
                      </CommandItem>
                    ))}
                  </CommandGroup>
                </Command>
              </PopoverContent>
            </Popover>
          )}

          {/* Time Range Selector */}
          <Select value={filters.timeRange} onValueChange={setTimeRange}>
            <SelectTrigger className="h-8 w-48 rounded-sm text-[12.5px] font-medium">
              <Clock className="h-3.5 w-3.5" />
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
        <Alert variant="destructive" className="mb-4">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Log Display — near-black terminal panel */}
      {!canStream ? (
        <EmptyState
          icon={<Clock className="h-6 w-6" />}
          title="No ready resources yet"
          description={
            filters.sources.length > 0
              ? 'The selected resources are still starting. Logs will stream once they are ready.'
              : 'Logs will stream once at least one resource is ready.'
          }
        />
      ) : logText ? (
        <div className="overflow-hidden rounded-md border border-border bg-[#070a0f]">
          <div className="h-[560px]">
            <LazyLog
              text={logText}
              extraLines={1}
              enableSearch
              caseInsensitive
              selectableLines
              follow={filters.timeRange === 'live-4h'}
              height={560}
              style={{
                backgroundColor: '#070a0f',
                color: '#94a3b8',
                fontSize: '12px',
                lineHeight: '1.7',
                fontFamily: 'var(--font-mono), ui-monospace, SFMono-Regular, Menlo, monospace',
              }}
            />
          </div>
        </div>
      ) : connectionStatus === 'connecting' || connectionStatus === 'reconnecting' ? (
        <EmptyState
          icon={<Loader2 className="h-6 w-6 animate-spin" />}
          title={connectionStatus === 'connecting' ? 'Connecting to log stream' : 'Reconnecting to log stream'}
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
