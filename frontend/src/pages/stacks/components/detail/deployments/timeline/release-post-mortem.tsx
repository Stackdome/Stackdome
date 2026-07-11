import { useState, useEffect } from "react";
import { StageTracker } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import type { EditSessionTab } from "@/pages/stacks/hooks/use-stack-edit-session";
import { ValidationBanner } from "@/pages/stacks/components/detail/ValidationBanner";
import type { ReleaseDetail } from "../use-release-detail";
import { diffSnapshots } from "../release-snapshot-diff";
import { resourceSource, replicaLabel, deriveStages } from "../derive";
import { ReleaseState } from "../release-states";
import { releaseValidationBannerItems } from "../release-errors";
import { ResourceOutcomeList } from "./resource-outcome-list";
import { ConfigChangesToggle } from "./config-changes-toggle";
import { DeployFailedBanner } from "./deploy-failed-banner";
import type { ResourceRowVM } from "./resource-row";

export interface ReleasePostMortemProps {
  detail: ReleaseDetail;
  release: StackRelease;
  stack: Stack;
  prevReleaseId?: string;
  prevSeq?: number;
  onJumpToResource?: (resourceIndex: number, tab: EditSessionTab) => void;
}

export function ReleasePostMortem({ detail, release, stack, prevReleaseId, prevSeq, onJumpToResource }: ReleasePostMortemProps) {
  const [validationDismissed, setValidationDismissed] = useState(false);
  useEffect(() => {
    if (release.id) detail.ensure(release.id);
    if (prevReleaseId) detail.ensure(prevReleaseId);
  }, [detail, release.id, prevReleaseId]);

  const cur = detail.peek(release.id);
  const prev = detail.peek(prevReleaseId);

  if (cur.loading && !cur.data) return <div className="px-0.5 py-3 text-[12.5px] text-fg-muted">Loading release detail…</div>;
  if (cur.error) return <div className="px-0.5 py-3 text-[12.5px] text-danger">Could not load detail: {cur.error}</div>;

  const data = cur.data;
  const outcomes = data?.outcome?.resources ?? {};
  const snap = data?.snapshot;
  const prevSnap = prev.data?.snapshot;
  const diffs = diffSnapshots(prevSnap, snap);

  // Same ResourceRowVM shape as the live body, but from this release's STORED outcome + frozen snapshot.
  const sourceByName = new Map((snap?.resources ?? []).map((r) => [r.name, r]));
  const rows: ResourceRowVM[] = Object.entries(outcomes).map(([name, o]) => ({
    name,
    phase: o.phase ?? "",
    replicas: replicaLabel(o.ready_replicas, o.replicas),
    msg: o.message,
    source: resourceSource(sourceByName.get(name)),
  }));
  // Tracker reads the release's own state/outcome/live_status.
  // Empty failure set: an old node must not surface current cluster crashes.
  const stages = deriveStages(release, [], release.live_status);
  const validationItems = release.state === ReleaseState.Failed
    ? releaseValidationBannerItems(release, stack.spec?.stack_resources ?? [])
    : [];

  return (
    <div className="px-0.5 pb-1.5 pt-3.5">
      <StageTracker stages={stages} />
      {release.state === ReleaseState.Failed && release.message && <DeployFailedBanner message={release.message} />}
      {!validationDismissed && validationItems.length > 0 && (
        <ValidationBanner
          items={validationItems}
          onJump={onJumpToResource}
          onDismiss={() => setValidationDismissed(true)}
        />
      )}
      <ResourceOutcomeList rows={rows} />
      {prevReleaseId && <ConfigChangesToggle diff={diffs} prevSeq={prevSeq} loading={!prev.data} />}
    </div>
  );
}
