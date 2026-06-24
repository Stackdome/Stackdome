import { useEffect } from "react";
import type { StackRelease } from "@/api/releases";
import type { ReleaseDetail } from "../use-release-detail";
import { diffSnapshots } from "../release-snapshot-diff";
import { resourceSource, replicaLabel } from "../derive";
import { ResourceOutcomeList } from "./resource-outcome-list";
import { ConfigChangesToggle } from "./config-changes-toggle";
import type { ResourceRowVM } from "./resource-row";

const Marker = ({ children, tone }: { children: React.ReactNode; tone?: string }) => (
  <div className={`mb-2.5 font-mono text-[11px] uppercase tracking-wide ${tone ?? "text-fg-muted"}`}>{children}</div>
);

export interface ReleasePostMortemProps {
  detail: ReleaseDetail;
  release: StackRelease;
  prevReleaseId?: string;
  prevSeq?: number;
}

export function ReleasePostMortem({ detail, release, prevReleaseId, prevSeq }: ReleasePostMortemProps) {
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

  return (
    <div className="px-0.5 pb-1.5 pt-3.5">
      {release.state === "Failed" && release.message && (
        <div>
          <Marker tone="text-danger">Why it failed</Marker>
          <div className="flex items-start gap-2 font-mono text-[11.5px] leading-relaxed text-foreground">
            <span className="flex-none text-danger">⊘</span><span>{release.message}</span>
          </div>
        </div>
      )}
      <ResourceOutcomeList rows={rows} />
      {prevReleaseId && <ConfigChangesToggle diff={diffs} prevSeq={prevSeq} />}
    </div>
  );
}
