import { StatusPill, StageTracker, variantFromState } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { deriveStages, deriveFailingResources, deriveRecovered, formatDuration, formatReleaseTime } from "../derive";
import { ResourceRow, type LogContext } from "./resource-row";

export interface CurrentReleaseNodeProps {
  release: StackRelease; stack: Stack; logContext?: LogContext;
  onOpenLogs?: (name: string) => void;
}

function meta(release: StackRelease): string {
  const parts: string[] = [];
  const when = formatReleaseTime(release.completed_at);
  if (when) parts.push(`deployed ${when}`);
  const dur = formatDuration(release.rendered_at, release.completed_at);
  if (dur !== "—") parts.push(`took ${dur}`);
  if (release.snapshot_revision) parts.push(`config ${release.snapshot_revision.slice(0, 7)}`);
  return parts.join(" · ");
}

export function CurrentReleaseNode({ release, stack, logContext, onOpenLogs }: CurrentReleaseNodeProps) {
  const failing = deriveFailingResources(stack);
  const recovered = deriveRecovered(stack);
  const recoveredNames = new Set(recovered.map((r) => r.name));
  const failingByName = new Map(failing.map((f) => [f.name, f]));
  const stages = deriveStages(stack, release, failing);
  const summaries = stack.status?.resources ?? [];
  const releaseLevelError = release.state === "Failed" && failing.length === 0 && release.message;

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex flex-wrap items-baseline justify-between gap-3">
        <div className="flex min-w-0 items-center gap-2.5">
          <StatusPill variant={variantFromState(release.state ?? "")}>{release.state}</StatusPill>
          <span className="font-sans text-[16px] font-semibold text-foreground">#{release.sequence}</span>
        </div>
        <span className="font-mono text-[11px] text-fg-muted">{meta(release)}</span>
      </div>

      <div className="mt-3.5"><StageTracker stages={stages} /></div>

      {releaseLevelError && (
        <div className="mt-3.5 rounded-md border border-danger-border bg-danger-bg p-3.5">
          <div className="mb-1.5 flex items-center gap-2 font-sans text-[13px] font-semibold text-danger">
            <span>⊘</span> Release failed
          </div>
          <div className="font-mono text-[11.5px] leading-relaxed text-foreground">{release.message}</div>
        </div>
      )}

      {summaries.length > 0 && (
        <div className="mt-4 divide-y divide-border">
          {summaries.map((s, i) => (
            <ResourceRow
              key={s.name ?? i}
              vm={{
                name: s.name ?? "",
                phase: s.phase ?? "",
                replicas: `${s.available_replicas ?? 0}/${s.replicas ?? 0}`,
                msg: s.message,
                tag: recoveredNames.has(s.name ?? "") ? "RECOVERED" : undefined,
                failure: failingByName.get(s.name ?? ""),
              }}
              logContext={logContext}
              onOpenLogs={onOpenLogs}
            />
          ))}
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
