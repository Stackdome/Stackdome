import { LogViewer } from './log-viewer';
import type { LogsTabProps } from './types';

export function LogsTab({
  stackId,
  organizationId,
  resources,
  initialSources
}: LogsTabProps) {
  return (
    <div className="space-y-4">
      <LogViewer
        stackId={stackId}
        organizationId={organizationId}
        resources={resources}
        initialSources={initialSources}
      />
    </div>
  );
}
