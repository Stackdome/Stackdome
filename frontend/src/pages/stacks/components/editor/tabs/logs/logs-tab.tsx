import { LogViewer } from './log-viewer';
import type { LogsTabProps } from './types';

export function LogsTab({
  stackId,
  organizationId,
  resources,
  liveStatusResources,
  initialSources
}: LogsTabProps) {
  return (
    <div className="space-y-4">
      <LogViewer
        stackId={stackId}
        organizationId={organizationId}
        resources={resources}
        liveStatusResources={liveStatusResources}
        initialSources={initialSources}
      />
    </div>
  );
}
