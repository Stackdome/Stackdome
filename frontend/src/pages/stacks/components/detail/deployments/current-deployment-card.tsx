import { Panel, StatusPill, StageTracker, variantFromState } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { deriveStages, deriveFailingResources, deriveRecovered, formatDuration } from "./derive";
import { FailingResourcesAccordion } from "./failing-resources-accordion";

export interface CurrentDeploymentCardProps {
  release: StackRelease;
  stack: Stack;
}

export function CurrentDeploymentCard({ release, stack }: CurrentDeploymentCardProps) {
  const failing = deriveFailingResources(stack);
  const recovered = deriveRecovered(stack);
  const stages = deriveStages(stack, release, failing);
  const summaries = stack.status?.resources ?? [];

  return (
    <Panel title="Current deployment" bodyClassName="p-0">
      <div className="flex items-center justify-between px-4 py-3">
        <div className="flex items-center gap-3">
          <StatusPill variant={variantFromState(release.state ?? "")}>{release.state}</StatusPill>
          <span className="font-mono text-[13px]">#{release.sequence}</span>
          {release.snapshot_revision && (
            <span className="text-[12px] text-muted-foreground">config {release.snapshot_revision.slice(0, 7)}</span>
          )}
        </div>
        <span className="text-[12px] text-muted-foreground">
          {formatDuration(release.rendered_at, release.completed_at)}
        </span>
      </div>

      <div className="px-4 pb-3"><StageTracker stages={stages} /></div>

      <div className="divide-y divide-border border-t border-border">
        {summaries.map((r, idx) => (
          <div key={r.name ?? idx} className="flex items-center justify-between px-4 py-2 text-[13px]">
            <span className="font-medium">{r.name}</span>
            <span className="font-mono text-muted-foreground">
              {r.phase} · {r.available_replicas ?? 0}/{r.replicas ?? 0}
            </span>
          </div>
        ))}
      </div>

      {recovered.length > 0 && (
        <div className="border-t border-warn-border bg-warn-bg px-4 py-2 text-[12px] text-muted-foreground">
          {recovered.map((r) => (
            <div key={r.name}>
              <span className="font-medium text-foreground">{r.name}</span>{" "}
              recovered{r.restartCount != null ? ` after ${r.restartCount} ${r.restartCount === 1 ? "restart" : "restarts"}` : ""} — last failure{" "}
              <span className="text-warn">{r.reason}</span>
            </div>
          ))}
        </div>
      )}

      {failing.length > 0 && (
        <div className="border-t border-border p-4">
          <FailingResourcesAccordion failing={failing} releaseMessage={release.message} />
        </div>
      )}
    </Panel>
  );
}
