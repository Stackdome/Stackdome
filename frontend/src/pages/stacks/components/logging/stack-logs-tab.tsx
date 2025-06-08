import { LogViewer } from './log-viewer';
import type { StackLogsTabProps } from './types';

export default function StackLogsTab({
  stackId,
  organizationId
}: StackLogsTabProps) {
  return (
    <div className="space-y-4">
      <LogViewer
        stackId={stackId}
        organizationId={organizationId}
      />
    </div>
  );
}
