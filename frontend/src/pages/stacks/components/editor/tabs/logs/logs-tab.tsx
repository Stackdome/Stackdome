import { LogViewer } from './log-viewer';
import type { LogsTabProps } from './types';

export function LogsTab({
  stackId,
  organizationId,
  resources
}: LogsTabProps) {
  return (
    <div className="space-y-4">
      <LogViewer
        stackId={stackId}
        organizationId={organizationId}
        resources={resources}
      />
    </div>
  );
}
