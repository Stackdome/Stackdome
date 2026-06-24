import { useEffect } from "react";
import { StageTracker } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { deriveStages, deriveFailingResources, deriveRecovered, resourceSource, replicaLabel } from "../derive";
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
 * The detail-card body for the latest deploy (releases[0]). Unlike a historical
 * node — which reads its stored outcome — this renders LIVE progress from
 * stack.status.resources plus the derived Build→Deploy→Ready tracker, so an
 * in-flight or just-failed deploy reflects the current cluster state.
 */
export function LiveReleaseBody({ release, stack, logContext, onOpenLogs, detail, prevReleaseId, prevSeq }: LiveReleaseBodyProps) {
  // Source the image/repo from THIS release's snapshot — not the live stack spec,
  // which may have been edited after the deploy started. The snapshot is the
  // frozen spec the release is actually shipping. Also ensure the previous
  // release's snapshot so the config diff is ready when expanded.
  useEffect(() => {
    if (release.id) detail?.ensure(release.id);
    if (prevReleaseId) detail?.ensure(prevReleaseId);
  }, [detail, release.id, prevReleaseId]);
  const releaseSnapshot = detail?.peek(release.id).data?.snapshot;
  const failing = deriveFailingResources(stack);
  const canDiff = !!prevReleaseId && !!detail;
  const diff = diffSnapshots(detail?.peek(prevReleaseId).data?.snapshot, detail?.peek(release.id).data?.snapshot);
  const recovered = deriveRecovered(stack);
  const recoveredNames = new Set(recovered.map((r) => r.name));
  const failingByName = new Map(failing.map((f) => [f.name, f]));
  const stages = deriveStages(stack, release, failing);
  const summaries = stack.status?.resources ?? [];
  const sourceByName = new Map((releaseSnapshot?.resources ?? stack.spec?.stack_resources ?? []).map((r) => [r.name, r]));
  const releaseLevelError = release.state === "Failed" && failing.length === 0 && release.message;
  const rows: ResourceRowVM[] = summaries.map((s) => ({
    name: s.name ?? "",
    phase: s.phase ?? "",
    replicas: replicaLabel(s.available_replicas, s.replicas),
    msg: s.message,
    tag: recoveredNames.has(s.name ?? "") ? "RECOVERED" : undefined,
    failure: failingByName.get(s.name ?? ""),
    source: resourceSource(sourceByName.get(s.name ?? "")),
  }));

  return (
    <div>
      <StageTracker stages={stages} />

      {releaseLevelError && <DeployFailedBanner message={release.message!} />}

      <ResourceOutcomeList rows={rows} logContext={logContext} onOpenLogs={onOpenLogs} />

      {canDiff && <ConfigChangesToggle diff={diff} prevSeq={prevSeq} />}

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
