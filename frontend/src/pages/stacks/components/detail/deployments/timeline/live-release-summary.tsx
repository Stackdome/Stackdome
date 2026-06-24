import { ArrowDown } from "lucide-react";
import { StatusPill } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { causeLabel, deriveFailingResources, formatReleaseTime, releaseGitSha } from "../derive";

export interface LiveReleaseSummaryProps {
  /** The release currently serving traffic (stack.status.last_converged). */
  release: StackRelease;
  stack: Stack;
  /** Scroll the matching timeline node into view. Omitted when there's nothing to jump to. */
  onJump?: () => void;
}

/**
 * A compact, pinned anchor for the release that's currently live. It leads the
 * Deployments tab so you can see WHAT serves traffic without scrolling — the
 * live release sinks below newer in-flight/failed deploys in the newest-first
 * timeline, so when it's buried this strip surfaces it (and jumps to its node).
 */
export function LiveReleaseSummary({ release, stack, onJump }: LiveReleaseSummaryProps) {
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
    <button
      type="button"
      onClick={onJump}
      className="group flex w-full items-center gap-2.5 rounded-md border border-success-border bg-success-bg px-3 py-2 text-left transition-colors hover:border-success"
    >
      <StatusPill variant="ready" className="flex-none gap-1 px-2 py-0.5 text-[10px] tracking-[0.06em]">Live</StatusPill>
      <span className="flex-none font-sans text-[13px] font-semibold text-foreground">#{release.sequence}</span>
      <span className="flex-none text-[13px] text-fg-2">{causeLabel(release.cause)}</span>
      <span className={`min-w-0 flex-1 truncate text-[12.5px] ${allHealthy ? "text-fg-muted" : "text-danger"}`}>
        {subline && `· ${subline}`}
      </span>
      {ts && <span className="flex-none font-mono text-[11px] text-fg-muted">{ts}</span>}
      {onJump && (
        <span className="flex flex-none items-center gap-1 font-mono text-[11px] text-fg-muted opacity-70 transition-opacity group-hover:opacity-100">
          Jump <ArrowDown className="h-3 w-3" />
        </span>
      )}
    </button>
  );
}
