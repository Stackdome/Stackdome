import { useEffect } from "react";
import { StageTracker } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { deriveStages, deriveFailingResources, deriveRecovered, resourceSource, replicaLabel } from "../derive";
import { ReleaseState } from "../release-states";
import { diffSnapshots } from "../release-snapshot-diff";
import type { ReleaseDetail } from "../use-release-detail";
import { type ResourceRowVM, type LogContext } from "./resource-row";
import { ResourceOutcomeList } from "./resource-outcome-list";
import { ConfigChangesToggle } from "./config-changes-toggle";
import { DeployFailedBanner } from "./deploy-failed-banner";

export interface LiveReleaseBodyProps {
  release: StackRelease;
  stack: Stack;
  logContext?: LogContext;
  onOpenLogs?: (name: string) => void;
  detail?: ReleaseDetail;
  prevReleaseId?: string;
  prevSeq?: number;
}

/**
 * Detail-card body for the latest deploy (releases[0]). Renders LIVE progress from
 * release.live_status.resources + the derived tracker (vs a historical node's stored outcome).
 */
export function LiveReleaseBody({ release, stack, logContext, onOpenLogs, detail, prevReleaseId, prevSeq }: LiveReleaseBodyProps) {
  // Source image/repo from THIS release's snapshot (the frozen spec it ships), not the
  // live stack spec which may have been edited after deploy. Also prefetch prev snapshot for the diff.
  useEffect(() => {
    if (release.id) detail?.ensure(release.id);
    if (prevReleaseId) detail?.ensure(prevReleaseId);
  }, [detail, release.id, prevReleaseId]);
  const releaseSnapshot = detail?.peek(release.id).data?.snapshot;
  const liveStatus = release.live_status;
  const failing = deriveFailingResources(release, liveStatus);
  const canDiff = !!prevReleaseId && !!detail;
  // Suppress empty "no changes" until prev snapshot resolves — else the diff briefly lies.
  const prevLoaded = !prevReleaseId || !!detail?.peek(prevReleaseId).data;
  const diff = diffSnapshots(detail?.peek(prevReleaseId).data?.snapshot, detail?.peek(release.id).data?.snapshot);
  const recovered = deriveRecovered(release, liveStatus);
  const recoveredNames = new Set(recovered.map((r) => r.name));
  const failingByName = new Map(failing.map((f) => [f.name, f]));
  const stages = deriveStages(release, failing, liveStatus);
  const liveResources = Object.entries(liveStatus?.resources ?? {});
  const sourceByName = new Map((releaseSnapshot?.resources ?? stack.spec?.stack_resources ?? []).map((r) => [r.name, r]));
  const failureMsg = release.state === ReleaseState.Failed && failing.length === 0 ? release.message : undefined;
  const rows: ResourceRowVM[] = liveResources.map(([name, s]) => ({
    name,
    phase: s.state ?? "",
    replicas: replicaLabel(s.available_replicas, s.replicas),
    msg: s.message,
    tag: recoveredNames.has(name) ? "RECOVERED" : undefined,
    failure: failingByName.get(name),
    source: resourceSource(sourceByName.get(name)),
  }));

  return (
    <div>
      <StageTracker stages={stages} />

      {failureMsg && <DeployFailedBanner message={failureMsg} />}

      <ResourceOutcomeList rows={rows} logContext={logContext} onOpenLogs={onOpenLogs} />

      {canDiff && <ConfigChangesToggle diff={diff} prevSeq={prevSeq} loading={!prevLoaded} />}

      {recovered.length > 0 && (
        <div className="mt-4 rounded-md border border-warn-border bg-warn-bg px-3.5 py-2.5 text-[12.5px] text-fg-muted">
          {recovered.map((r) => (
            <div key={r.name}>
              <span className="font-medium text-foreground">{r.name}</span> recovered
              {r.restartCount != null ? ` after ${r.restartCount} ${r.restartCount === 1 ? "restart" : "restarts"}` : ""} — last failure{" "}
              <span className="text-warn">{r.reason}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
