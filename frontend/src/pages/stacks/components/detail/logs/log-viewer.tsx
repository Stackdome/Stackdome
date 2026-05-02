import { useMemo, useState } from 'react';
import { LazyLog } from 'react-lazylog';
import { Card, CardContent } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Loader2, Wifi, WifiOff, AlertCircle, Check, Clock, ChevronDown, Layers } from 'lucide-react';
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

export function LogViewer({ stackId, organizationId, resources = [], className = '' }: LogViewerProps) {
  const [sourceSelectOpen, setSourceSelectOpen] = useState(false);
  const [filters, setFilters] = useState<LogFilters>({
    sources: [],
    timeRange: 'live-4h',
  });

  const { logs, connectionStatus, error } = useLogStream({
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

  const statusInfo = getConnectionStatusInfo(connectionStatus);
  const StatusIcon = statusInfo.icon;

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

          {/* Connection Status */}
          <Badge variant="outline" className={statusInfo.className}>
            <StatusIcon className={`mr-2 h-3 w-3 ${statusInfo.iconClass}`} />
            {statusInfo.text}
          </Badge>
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
      <Card>
        <CardContent className="p-0">
          <div className="h-96 border rounded-md bg-gray-900">
            {logText ? (
              <LazyLog
                text={logText}
                extraLines={1}
                enableSearch
                caseInsensitive
                selectableLines
                follow={filters.timeRange === 'live-4h'}
                height={384} // 24rem in pixels
                style={{
                  backgroundColor: '#111827',
                  color: '#f9fafb',
                  fontSize: '13px',
                  fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
                }}
              />
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
