import { useState, useMemo } from 'react';
import type { LogEntry, LogFilters, TimeRangeOption } from './types';
import { getUniqueSources } from './utils';

interface UseLogFiltersProps {
  logs: LogEntry[];
}

interface UseLogFiltersReturn {
  filters: LogFilters;
  filteredLogs: LogEntry[];
  setSources: (sources: string[]) => void;
  setTimeRange: (timeRange: TimeRangeOption) => void;
  availableSources: string[];
}

const defaultFilters: LogFilters = {
  sources: [],
  timeRange: 'live-4h',
};

export function useLogFilters({ logs }: UseLogFiltersProps): UseLogFiltersReturn {
  const [filters, setFilters] = useState<LogFilters>(defaultFilters);

  const filteredLogs = useMemo(() => {
    let filtered = logs;

    // Filter by sources (frontend filtering)
    if (filters.sources.length > 0) {
      filtered = filtered.filter(log =>
        !log.source || filters.sources.includes(log.source)
      );
    }

    // Note: Time range filtering is handled by the backend via API parameters
    return filtered;
  }, [logs, filters.sources]);

  const availableSources = useMemo(() => {
    return getUniqueSources(logs);
  }, [logs]);

  const setSources = (sources: string[]) => {
    setFilters(prev => ({ ...prev, sources }));
  };

  const setTimeRange = (timeRange: TimeRangeOption) => {
    setFilters(prev => ({ ...prev, timeRange }));
  };

  return {
    filters,
    filteredLogs,
    setSources,
    setTimeRange,
    availableSources,
  };
}
