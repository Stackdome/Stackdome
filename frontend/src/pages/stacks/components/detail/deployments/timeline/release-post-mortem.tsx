import { useEffect } from "react";
import { StageTracker } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import type { ReleaseDetail } from "../use-release-detail";
import { diffSnapshots } from "../release-snapshot-diff";
import { resourceSource, replicaLabel, deriveStages } from "../derive";
import { ReleaseState } from "../release-states";
import { ResourceOutcomeList } from "./resource-outcome-list";
import { ConfigChangesToggle } from "./config-changes-toggle";
import { DeployFailedBanner } from "./deploy-failed-banner";
import type { ResourceRowVM } from "./resource-row";

export interface ReleasePostMortemProps {
  detail: ReleaseDetail;
  release: StackRelease;
  /** Live stack — only used to derive the Build→Deploy→Ready tracker (last_converged). */
  stack: Stack;
  prevReleaseId?: string;
  prevSeq?: number;
}

export function ReleasePostMortem({ detail, release, stack, prevReleaseId, prevSeq }: ReleasePostMortemProps) {
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

  // Build the same ResourceRowVM shape the live body uses, sourced from this
  // release's STORED outcome + its frozen snapshot, so the section is uniform.
  const sourceByName = new Map((snap?.resources ?? []).map((r) => [r.name, r]));
  const rows: ResourceRowVM[] = Object.entries(outcomes).map(([name, o]) => ({
    name,
    phase: o.phase ?? "",
    replicas: replicaLabel(o.ready_replicas, o.replicas),
    msg: o.message,
    source: resourceSource(sourceByName.get(name)),
  }));
  // Tracker reads from the release's own state/outcome; the live stack only
  // supplies last_converged (so the live release reads done/done/done). No live
  // failure set is threaded — an old node shouldn't surface current cluster crashes.
  const stages = deriveStages(stack, release, []);

  return (
    <div className="px-0.5 pb-1.5 pt-3.5">
      <StageTracker stages={stages} />
      {release.state === ReleaseState.Failed && release.message && <DeployFailedBanner message={release.message} />}
      <ResourceOutcomeList rows={rows} />
      {prevReleaseId && <ConfigChangesToggle diff={diffs} prevSeq={prevSeq} loading={!prev.data} />}
    </div>
  );
}
