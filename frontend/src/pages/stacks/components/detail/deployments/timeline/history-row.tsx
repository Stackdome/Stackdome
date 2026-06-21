import { StatusPill, variantFromState } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import { causeLabel, releaseGitSha, formatDuration } from "../derive";
import type { ReleaseDetail } from "../use-release-detail";
import { ReleaseMenu } from "./release-menu";
import { ReleasePostMortem } from "./release-post-mortem";

export interface HistoryRowProps {
  release: StackRelease;
  prevReleaseId?: string; prevSeq?: number;
  detail: ReleaseDetail;
  isOpen: boolean;
  onToggle: (id: string) => void;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
  onCopyId: (id: string) => void;
}

export function HistoryRow({ release, prevReleaseId, prevSeq, detail, isOpen, onToggle, onRollback, onCancel, onCopyId }: HistoryRowProps) {
  const id = release.id ?? "";
  const state = release.state ?? "";
  const sha = releaseGitSha(release);
  const subline = state === "Failed"
    ? release.message
    : state === "Released"
      ? [sha && `git ${sha}`, formatDuration(release.rendered_at, release.completed_at)].filter(Boolean).join(" · ")
      : undefined;

  return (
    <div>
      <div className="-mx-2 flex cursor-pointer items-center gap-2.5 rounded-md px-2 py-1.5 hover:bg-muted" onClick={() => onToggle(id)}>
        <StatusPill variant={variantFromState(state)}>{state}</StatusPill>
        <span className="font-sans text-[13px] font-semibold text-foreground">#{release.sequence}</span>
        <span className="min-w-0 flex-1 truncate text-[13px] text-fg-muted">
          {causeLabel(release.cause)}
          {subline ? <span className={state === "Failed" ? "text-danger" : ""}> · {subline}</span> : null}
        </span>
        <ReleaseMenu release={release} onView={() => onToggle(id)} onRollback={onRollback} onCancel={onCancel} onCopyId={onCopyId} />
      </div>
      {isOpen && <ReleasePostMortem detail={detail} release={release} prevReleaseId={prevReleaseId} prevSeq={prevSeq} />}
    </div>
  );
}
