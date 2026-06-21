import { useEffect } from "react";
import type { StackRelease } from "@/api/releases";
import type { ReleaseDetail } from "../use-release-detail";
import { diffSnapshots } from "../release-snapshot-diff";
import { OutcomesTable } from "./outcomes-table";
import { ConfigDiff } from "./config-diff";

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
  const outcomes = (data as unknown as { outcome?: { resources?: Record<string, { phase?: string; ready_replicas?: number; replicas?: number; message?: string }> } })?.outcome?.resources ?? {};
  const snap = (data as unknown as { snapshot?: unknown })?.snapshot;
  const prevSnap = (prev.data as unknown as { snapshot?: unknown })?.snapshot;
  const diffs = diffSnapshots(prevSnap, snap);

  return (
    <div className="space-y-4 px-0.5 pb-1.5 pt-3.5">
      {release.state === "Failed" && release.message && (
        <div>
          <Marker tone="text-danger">Why it failed</Marker>
          <div className="flex items-start gap-2 font-mono text-[11.5px] leading-relaxed text-foreground">
            <span className="flex-none text-danger">⊘</span><span>{release.message}</span>
          </div>
        </div>
      )}
      {Object.keys(outcomes).length > 0 && (
        <div><Marker>Resource outcomes</Marker><OutcomesTable outcomes={outcomes} /></div>
      )}
      {prevReleaseId && (
        <div><Marker>Config changes · vs #{prevSeq ?? "previous"}</Marker><ConfigDiff diffs={diffs} prevSeq={prevSeq} /></div>
      )}
    </div>
  );
}
