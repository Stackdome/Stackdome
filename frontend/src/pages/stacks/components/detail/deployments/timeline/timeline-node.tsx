import { ChevronDown } from "lucide-react";
import { StatusPill } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { causeLabel, releaseGitSha, formatDuration, formatReleaseTime } from "../derive";
import type { ReleaseDetail } from "../use-release-detail";
import { ReleaseMenu } from "./release-menu";
import { ReleasePostMortem } from "./release-post-mortem";
import { LiveReleaseBody } from "./live-release-body";
import type { LogContext } from "./resource-row";

const DEPLOYING = new Set(["Pending", "InProgress"]);

export interface TimelineNodeProps {
  release: StackRelease;
  prevReleaseId?: string;
  prevSeq?: number;
  detail: ReleaseDetail;
  isOpen: boolean;
  onToggle: (id: string) => void;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
  onCopyId: (id: string) => void;
  /** releases[0] — render LIVE progress from the stack rather than the stored outcome. */
  isActive: boolean;
  /** This release currently serves traffic (stack.status.last_converged). */
  isLive: boolean;
  stack: Stack;
  logContext?: LogContext;
  onOpenLogs?: (name: string) => void;
}

/**
 * One node in the unified deploy timeline: a lean toggle row plus a detail card
 * that opens directly below it. The same shape is used for every release; only
 * the body differs (live progress for the latest deploy, stored post-mortem for
 * earlier ones) and the open card's border (green for the live release, amber
 * while deploying).
 */
export function TimelineNode(props: TimelineNodeProps) {
  const { release, prevReleaseId, prevSeq, detail, isOpen, onToggle, onRollback, onCancel, onCopyId, isActive, isLive, stack, logContext, onOpenLogs } = props;
  const id = release.id ?? "";
  const state = release.state ?? "";
  const deploying = DEPLOYING.has(state);

  const sha = releaseGitSha(release);
  const dur = formatDuration(release.rendered_at, release.completed_at);
  const subline = state === "Released"
    ? [sha && `git ${sha}`, dur !== "—" && `took ${dur}`].filter(Boolean).join(" · ")
    : release.message || (dur !== "—" ? `took ${dur}` : undefined);
  const ts = formatReleaseTime(release.completed_at ?? release.created_at);

  const chip = isLive
    ? <StatusPill variant="ready">Live</StatusPill>
    : state === "Failed"
      ? <StatusPill variant="error" withDot={false}>Failed</StatusPill>
      : deploying
        ? <StatusPill variant="pending">Deploying</StatusPill>
        : null;

  const cardBorder = isLive ? "border-success-border" : deploying ? "border-brand-border" : "border-border";

  return (
    <div>
      <div
        className="-mx-2 flex cursor-pointer items-center gap-2.5 rounded-md px-2 py-1.5 hover:bg-muted"
        onClick={() => onToggle(id)}
      >
        <span className="flex-none font-sans text-[13px] font-semibold text-foreground">#{release.sequence}</span>
        <span className="flex-none text-[13px] text-fg-2">{causeLabel(release.cause)}</span>
        {chip}
        <span className={`min-w-0 flex-1 truncate text-[13px] ${state === "Failed" ? "text-danger" : "text-fg-muted"}`}>
          {subline ? `· ${subline}` : ""}
        </span>
        {ts && <span className="flex-none font-mono text-[11px] text-fg-muted">{ts}</span>}
        <ChevronDown className={`h-3.5 w-3.5 flex-none text-fg-muted transition-transform ${isOpen ? "rotate-180" : ""}`} />
        <span onClick={(e) => e.stopPropagation()}>
          <ReleaseMenu release={release} onRollback={onRollback} onCancel={onCancel} onCopyId={onCopyId} />
        </span>
      </div>

      {isOpen && (
        <div className={`mb-1 mt-1.5 rounded-md border ${cardBorder} bg-card p-4`}>
          {isActive ? (
            <LiveReleaseBody
              release={release}
              stack={stack}
              logContext={logContext}
              onOpenLogs={onOpenLogs}
              detail={detail}
              prevReleaseId={prevReleaseId}
              prevSeq={prevSeq}
            />
          ) : (
            <ReleasePostMortem detail={detail} release={release} prevReleaseId={prevReleaseId} prevSeq={prevSeq} />
          )}
        </div>
      )}
    </div>
  );
}
