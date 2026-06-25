import { EmptyState } from "@/components/branded";
import type { Stack } from "@/api/stacks";
import type { StackRelease } from "@/api/releases";
import type { DeployLifecycle } from "./use-deploy-lifecycle";
import { TimelineRail } from "./timeline/timeline-rail";
import { DraftNode } from "./timeline/draft-node";
import { LiveReleaseSummary } from "./timeline/live-release-summary";

export interface DeploymentsTabProps {
  orgId: string;
  teamName: string;
  stackId: string;
  stack: Stack;
  onOpenLogs?: (resourceName?: string) => void;
  // Deploy lifecycle + release data are owned by the page (the status bar owns
  // deploy); this tab is presentational.
  releases: StackRelease[];
  activeRelease?: StackRelease;
  loading: boolean;
  error: string | null;
  lifecycle: DeployLifecycle;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
  onCopyId: (id: string) => void;
}

export function DeploymentsTab({ orgId, teamName, stackId, stack, onOpenLogs, releases, activeRelease, loading, error, lifecycle, onRollback, onCancel, onCopyId }: DeploymentsTabProps) {
  if (error) return <EmptyState title="Could not load deployments" description={error} />;

  const logContext = { orgId, teamName, stackId };
  const openLogs = onOpenLogs ? (name: string) => onOpenLogs(name) : undefined;

  const draftNode = lifecycle.phase === "editing" || lifecycle.phase === "staged"
    ? <DraftNode phase={lifecycle.phase} diff={lifecycle.stagedDiff} vsSeq={lifecycle.vsSeq} nextSeq={lifecycle.nextSeq} isLast={releases.length === 0} />
    : undefined;

  // Anchor the live release at the top only when it's buried (not already the newest node).
  const liveReleaseId = stack.status?.last_converged?.release_id;
  const liveIdx = liveReleaseId ? releases.findIndex((r) => r.id === liveReleaseId) : -1;
  const liveRelease = liveIdx >= 0 ? releases[liveIdx] : undefined;
  const liverev = liveIdx >= 0 ? releases[liveIdx + 1] : undefined;
  const showLiveAnchor = liveRelease && releases[0]?.id !== liveRelease.id;

  return (
    <div className="space-y-4">
      {showLiveAnchor && liveRelease && (
        <LiveReleaseSummary
          release={liveRelease}
          stack={stack}
          prevReleaseId={liverev?.id}
          prevSeq={liverev?.sequence}
          logContext={logContext}
          onOpenLogs={openLogs}
        />
      )}

      <div className="font-mono text-[11px] uppercase tracking-wide text-fg-muted">Deploy timeline</div>

      {loading && releases.length === 0 && !draftNode ? (
        <p className="text-[13px] text-fg-muted">Loading deployments…</p>
      ) : (
        <TimelineRail
          releases={releases}
          activeRelease={activeRelease}
          stack={stack}
          logContext={logContext}
          onOpenLogs={openLogs}
          draftNode={draftNode}
          onRollback={onRollback}
          onCancel={onCancel}
          onCopyId={onCopyId}
        />
      )}
    </div>
  );
}
