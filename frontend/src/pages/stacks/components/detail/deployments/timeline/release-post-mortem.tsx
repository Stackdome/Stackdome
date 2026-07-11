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
import { useReleaseEvents } from "../use-release-events";
import { ResourceOutcomeList } from "./resource-outcome-list";
import { ConfigChangesToggle } from "./config-changes-toggle";
import { DeployFailedBanner } from "./deploy-failed-banner";
import { ReleaseActivityFeed } from "./release-activity-feed";
import type { ResourceRowVM, LogContext } from "./resource-row";

export interface ReleasePostMortemProps {
  detail: ReleaseDetail;
  release: StackRelease;
  stack: Stack;
  prevReleaseId?: string;
  prevSeq?: number;
  logContext?: LogContext;
  onJumpToResource?: (resourceName: string, tab: EditSessionTab) => void;
}

export function ReleasePostMortem({ detail, release, stack, prevReleaseId, prevSeq, logContext, onJumpToResource }: ReleasePostMortemProps) {
  const [validationDismissed, setValidationDismissed] = useState(false);
  // Historical release: one-shot event fetch only, no streaming — this component only
  // mounts once its timeline node is expanded (TimelineNode gates on isOpen).
  const { events } = useReleaseEvents({
    orgId: logContext?.orgId ?? "",
    teamName: logContext?.teamName ?? "",
    stackId: logContext?.stackId ?? "",
    releaseId: release.id,
    terminal: true,
  });
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
  const bannerResources = stack.spec?.stack_resources ?? [];
  const validationItems = release.state === ReleaseState.Failed
    ? releaseValidationBannerItems(release, bannerResources)
    : [];

  return (
    <div className="px-0.5 pb-1.5 pt-3.5">
      <StageTracker stages={stages} />
      {release.state === ReleaseState.Failed && release.message && <DeployFailedBanner message={release.message} />}
      {!validationDismissed && validationItems.length > 0 && (
        <ValidationBanner
          items={validationItems}
          onJump={onJumpToResource && ((idx, tab) => {
            // Round-trip index → name via the SAME list the items were built
            // against; the receiver re-resolves the name at click time.
            const name = bannerResources[idx]?.name;
            if (name) onJumpToResource(name, tab);
          })}
          onDismiss={() => setValidationDismissed(true)}
        />
      )}
      <ResourceOutcomeList rows={rows} />
      <div className="mt-4">
        <ReleaseActivityFeed events={events} streaming={false} />
      </div>
      {prevReleaseId && <ConfigChangesToggle diff={diffs} prevSeq={prevSeq} loading={!prev.data} />}
    </div>
  );
}
