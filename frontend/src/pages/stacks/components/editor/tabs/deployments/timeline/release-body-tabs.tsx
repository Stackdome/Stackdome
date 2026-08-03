import { useState } from "react";
import { cn } from "@/lib/utils";
import type { SnapshotDiff } from "../release-snapshot-diff";
import { ConfigDiff } from "./config-diff";

type TabId = "outcomes" | "changes";

export interface ReleaseBodyTabsProps {
  /** Outcomes tab content — the split console. */
  children: React.ReactNode;
  diff: SnapshotDiff;
  /** A previous release exists to diff against. */
  hasPrev: boolean;
  /** Sequence the diff is taken against (the previous release). */
  prevSeq?: number;
  /** Previous snapshot still loading — show a placeholder instead of a false "no changes". */
  loading?: boolean;
}

/**
 * Outcomes | Changes tab row for a node's detail card. Outcomes carries the split
 * console; Changes carries the config diff vs the previous release. Identical
 * across live + historical bodies.
 */
export function ReleaseBodyTabs({ children, diff, hasPrev, prevSeq, loading = false }: ReleaseBodyTabsProps) {
  const [tab, setTab] = useState<TabId>("outcomes");
  const changeCount = diff.resources.length + diff.volumes.length;

  const tabClass = (id: TabId) =>
    cn(
      "-mb-px flex items-center gap-1.5 border-b-2 pb-2 text-[12.5px] font-medium",
      tab === id ? "border-foreground text-foreground" : "border-transparent text-fg-muted hover:text-fg-2",
    );

  return (
    <div className="mt-4">
      <div className="flex items-center gap-5 border-b border-border">
        <button onClick={() => setTab("outcomes")} className={tabClass("outcomes")}>Outcomes</button>
        <button onClick={() => setTab("changes")} className={tabClass("changes")}>
          Changes
          {changeCount > 0 && <span className="font-mono text-[10px] text-fg-muted">{changeCount}</span>}
        </button>
        {tab === "changes" && hasPrev && (
          <span className="ml-auto pb-2 font-mono text-[10.5px] text-fg-muted">vs #{prevSeq ?? "previous"}</span>
        )}
      </div>
      {tab === "outcomes" && children}
      {tab === "changes" && (
        <div className="mt-4">
          {loading
            ? <div className="text-[12.5px] text-fg-muted">Loading changes…</div>
            : <ConfigDiff diff={diff} hasPrev={hasPrev} prevSeq={prevSeq} />}
        </div>
      )}
    </div>
  );
}
