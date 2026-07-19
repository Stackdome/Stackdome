import { useState } from "react";
import { ChevronDown } from "lucide-react";
import { StatusPill } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { causeLabel, formatReleaseTime } from "../derive";
import { useReleaseDetailContext } from "../use-release-detail";
import { LiveReleaseBody } from "./live-release-body";
import type { LogContext } from "./split-console";

export interface LiveReleaseSummaryProps {
  /** The release currently serving traffic (stack.converged_release). */
  release: StackRelease;
  stack: Stack;
  /** Previous release, for the config diff inside the expanded body. */
  prevReleaseId?: string;
  prevSeq?: number;
  logContext?: LogContext;
}

/**
 * Compact pinned card for the live release, leading the tab so what serves traffic is
 * visible without scrolling (it sinks below newer deploys in the newest-first timeline).
 * Expands in place into the live body.
 */
export function LiveReleaseSummary({ release, stack, prevReleaseId, prevSeq, logContext }: LiveReleaseSummaryProps) {
  const [open, setOpen] = useState(false);
  const detail = useReleaseDetailContext();
  const ts = formatReleaseTime(release.completed_at ?? release.created_at);

  return (
    <div className="overflow-hidden rounded-md border border-success-border bg-card">
      <button
        type="button"
        aria-label={`Live release #${release.sequence}`}
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center gap-2.5 bg-success-bg px-3 py-2 text-left transition-colors"
      >
        <StatusPill variant="ready" className="flex-none gap-1 px-2 py-0.5 text-[10px] tracking-[0.06em]">Live</StatusPill>
        <span className="flex-none font-sans text-[13px] font-semibold text-foreground">#{release.sequence}</span>
        <span className="flex-none text-[13px] text-fg-2">{causeLabel(release.cause)}</span>
        <span className="min-w-0 flex-1" />
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
          />
        </div>
      )}
    </div>
  );
}
