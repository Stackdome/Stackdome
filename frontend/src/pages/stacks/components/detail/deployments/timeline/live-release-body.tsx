import { useEffect, useState } from "react";
import { StageTracker } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { deriveStages, deriveFailingResources, deriveRecovered, resourceSource, replicaLabel } from "../derive";
import { diffSnapshots } from "../release-snapshot-diff";
import type { ReleaseDetail } from "../use-release-detail";
import { ResourceRow, type LogContext } from "./resource-row";
import { ConfigDiff } from "./config-diff";

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
  const [showDiff, setShowDiff] = useState(false);
  // Source the image/repo from THIS release's snapshot — not the live stack spec,
  // which may have been edited after the deploy started. The snapshot is the
  // frozen spec the release is actually shipping.
  useEffect(() => { if (release.id) detail?.ensure(release.id); }, [detail, release.id]);
  const releaseSnapshot = detail?.peek(release.id).data?.snapshot;
  const failing = deriveFailingResources(stack);
  const canDiff = !!prevReleaseId && !!detail;
  const onToggleDiff = () => {
    if (!detail || !prevReleaseId) return;
    if (release.id) detail.ensure(release.id);
    detail.ensure(prevReleaseId);
    setShowDiff((v) => !v);
  };
  const diff = diffSnapshots(detail?.peek(prevReleaseId).data?.snapshot, detail?.peek(release.id).data?.snapshot);
  const recovered = deriveRecovered(stack);
  const recoveredNames = new Set(recovered.map((r) => r.name));
  const failingByName = new Map(failing.map((f) => [f.name, f]));
  const stages = deriveStages(stack, release, failing);
  const summaries = stack.status?.resources ?? [];
  const sourceByName = new Map((releaseSnapshot?.resources ?? stack.spec?.stack_resources ?? []).map((r) => [r.name, r]));
  const releaseLevelError = release.state === "Failed" && failing.length === 0 && release.message;

  return (
    <div>
      <StageTracker stages={stages} />

      {releaseLevelError && (
        <div className="mt-4 rounded-md border border-danger-border bg-danger-bg p-3.5">
          <div className="mb-1.5 flex items-center gap-2 font-sans text-[13px] font-semibold text-danger">
            <span>⊘</span> Deploy failed
          </div>
          <div className="font-mono text-[11.5px] leading-relaxed text-foreground">{release.message}</div>
        </div>
      )}

      {summaries.length > 0 && (
        <div className="mt-4">
          <div className="mb-0.5 font-mono text-[11px] uppercase tracking-wide text-fg-muted">Resource outcome</div>
          <div className="divide-y divide-border">
            {summaries.map((s, i) => (
              <ResourceRow
                key={s.name ?? i}
                vm={{
                  name: s.name ?? "",
                  phase: s.phase ?? "",
                  replicas: replicaLabel(s.available_replicas, s.replicas),
                  msg: s.message,
                  tag: recoveredNames.has(s.name ?? "") ? "RECOVERED" : undefined,
                  failure: failingByName.get(s.name ?? ""),
                  source: resourceSource(sourceByName.get(s.name ?? "")),
                }}
                logContext={logContext}
                onOpenLogs={onOpenLogs}
              />
            ))}
          </div>
        </div>
      )}

      {canDiff && (
        <div className="mt-4">
          <button
            onClick={onToggleDiff}
            className="font-mono text-[11px] uppercase tracking-wide text-fg-muted hover:text-foreground"
          >
            {showDiff ? "Hide config changes" : `Config changes · vs #${prevSeq ?? "previous"}`}
          </button>
          {showDiff && <div className="mt-3"><ConfigDiff diff={diff} hasPrev prevSeq={prevSeq} /></div>}
        </div>
      )}

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
