import { Panel, EmptyState } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import { ReleaseRow } from "./release-row";

export interface ReleaseHistoryProps {
  releases: StackRelease[];
  onViewDetails: (id: string) => void;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
}

export function ReleaseHistory({ releases, onViewDetails, onRollback, onCancel }: ReleaseHistoryProps) {
  return (
    <Panel title="History" count={releases.length} bodyClassName="p-0">
      {releases.length === 0 ? (
        <EmptyState title="No deployments yet" description="Deploy this stack to create your first release." />
      ) : (
        releases.map((r) => (
          <ReleaseRow key={r.id} release={r} onViewDetails={onViewDetails} onRollback={onRollback} onCancel={onCancel} />
        ))
      )}
    </Panel>
  );
}
