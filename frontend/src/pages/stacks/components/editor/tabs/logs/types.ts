import type { ReleaseLiveStatus } from "@/api/releases";
import type { SseStreamStatus } from "@/api/observability";

export interface LogEntry {
  id: string;
  timestamp?: Date;
  message: string;
  raw: string;
  source?: string;
}

export type ConnectionStatus = SseStreamStatus;

export type TimeRangeOption =
  | 'live-4h'
  | '30m'
  | '1h'
  | '4h'
  | '24h'
  | 'all';

export interface LogFilters {
  sources: string[];
  timeRange: TimeRangeOption;
}

export interface StackResource {
  name: string;
  id?: string;
}

export interface LogViewerProps {
  stackId: string;
  organizationId: string;
  resources?: StackResource[];
  /** Live rollout status per resource name (release live_status); streams only
   *  open for Ready resources, since the backend rejects the rest pre-stream. */
  liveStatusResources?: ReleaseLiveStatus["resources"];
  /** Resource names pre-selected in the source filter on mount (e.g. arriving via a drawer's "View logs"). */
  initialSources?: string[];
  className?: string;
}

export interface LogsTabProps {
  stackId: string;
  organizationId: string;
  resources?: StackResource[];
  /** Live rollout status per resource name (release live_status). */
  liveStatusResources?: ReleaseLiveStatus["resources"];
  /** Resource names pre-selected in the source filter on mount. */
  initialSources?: string[];
}
