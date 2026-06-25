import { useState } from "react";
import { ChevronDown } from "lucide-react";
import { StatusPill } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { causeLabel, deriveFailingResources, formatReleaseTime, releaseGitSha } from "../derive";
import { useReleaseDetail } from "../use-release-detail";
import { LiveReleaseBody } from "./live-release-body";
import type { LogContext } from "./resource-row";

export interface LiveReleaseSummaryProps {
  /** The release currently serving traffic (stack.status.last_converged). */
  release: StackRelease;
  stack: Stack;
  /** Previous release, for the config diff inside the expanded body. */
  prevReleaseId?: string;
  prevSeq?: number;
  logContext?: LogContext;
  onOpenLogs?: (name: string) => void;
}

/**
 * A compact, pinned card for the release that's currently live. It leads the
 * Deployments tab so what serves traffic is visible without scrolling — the
 * live release sinks below newer in-flight/failed deploys in the newest-first
 * timeline. The row expands IN PLACE into the live body (tracker + resource
 * outcome + config changes), so you never have to hunt for the live node.
 */
export function LiveReleaseSummary({ release, stack, prevReleaseId, prevSeq, logContext, onOpenLogs }: LiveReleaseSummaryProps) {
  const [open, setOpen] = useState(false);
  const detail = useReleaseDetail(logContext?.orgId ?? "", logContext?.teamName ?? "", logContext?.stackId ?? "");
  const failing = deriveFailingResources(stack);
  const total = stack.status?.resources?.length ?? stack.spec?.stack_resources?.length ?? 0;
  const allHealthy = failing.length === 0;
  const health = total === 0
    ? ""
    : allHealthy
      ? `${total} ${total === 1 ? "resource" : "resources"} healthy`
      : `${failing.length} of ${total} unhealthy`;
  const sha = releaseGitSha(release);
  const ts = formatReleaseTime(release.completed_at ?? release.created_at);
  const subline = [health, sha && `git ${sha}`].filter(Boolean).join(" · ");

  return (
    <div className="overflow-hidden rounded-md border border-success-border bg-card">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center gap-2.5 bg-success-bg px-3 py-2 text-left transition-colors"
      >
        <StatusPill variant="ready" className="flex-none gap-1 px-2 py-0.5 text-[10px] tracking-[0.06em]">Live</StatusPill>
        <span className="flex-none font-sans text-[13px] font-semibold text-foreground">#{release.sequence}</span>
        <span className="flex-none text-[13px] text-fg-2">{causeLabel(release.cause)}</span>
        <span className={`min-w-0 flex-1 truncate text-[12.5px] ${allHealthy ? "text-fg-muted" : "text-danger"}`}>
          {subline && `· ${subline}`}
        </span>
        {ts && <span className="flex-none font-mono text-[11px] text-fg-muted">{ts}</span>}
        <ChevronDown className={`h-3.5 w-3.5 flex-none text-fg-muted transition-transform ${open ? "rotate-180" : ""}`} />
      </button>

      {open && (
        <div className="border-t border-success-border p-4">
          <LiveReleaseBody
            release={release}
            stack={stack}
            detail={detail}
            prevReleaseId={prevReleaseId}
            prevSeq={prevSeq}
            logContext={logContext}
            onOpenLogs={onOpenLogs}
          />
        </div>
      )}
    </div>
  );
}
