import { useState } from "react";
import { ChevronDown } from "lucide-react";
import type { SnapshotDiff } from "../release-snapshot-diff";
import { RailNode } from "./rail-node";
import { ConfigDiff } from "./config-diff";

export interface DraftNodeProps {
  /** editing = unsaved edits in flight; staged = saved draft, not deployed. */
  phase: "editing" | "staged";
  /** Saved spec vs the comparison release. Absent when it couldn't be diffed. */
  diff?: SnapshotDiff;
  /** Sequence the draft is diffed against — live, or the in-flight release when superseding it. */
  vsSeq?: number;
  /** True when no release nodes follow (hides the rail connector). */
  isLast?: boolean;
  defaultOpen?: boolean;
}

function changedNames(diff?: SnapshotDiff): string[] {
  if (!diff) return [];
  const names: string[] = [];
  for (const r of diff.resources) {
    if (r.change === "renamed" && r.fromName) names.push(r.fromName, r.name);
    else names.push(r.name);
  }
  for (const v of diff.volumes) names.push(v.name);
  return names;
}

/**
 * Leads the rail when there are saved-but-undeployed changes. Mirrors a release node's
 * shape but with a dashed amber ring/border ("not deployed") and shows the staged diff.
 */
export function DraftNode({ phase, diff, vsSeq, isLast, defaultOpen = false }: DraftNodeProps) {
  const [open, setOpen] = useState(defaultOpen);
  const names = changedNames(diff);
  const hasChanges = names.length > 0;
  const namesLabel = hasChanges
    ? `${names.slice(0, 2).join(", ")}${names.length > 2 ? ` +${names.length - 2}` : ""} changed`
    : "";
  const chipLabel = phase === "editing" ? "Unsaved" : "Draft";

  return (
    <RailNode tone="amber" shape="dashed" isLast={isLast}>
      <div>
        <div
          className="-mx-2 flex cursor-pointer items-center gap-2.5 rounded-md px-2 py-1.5 hover:bg-muted"
          onClick={() => setOpen((o) => !o)}
        >
          <span className="flex-none rounded-full border border-brand-border bg-brand-bg px-2 py-0.5 font-mono text-[10px] font-bold uppercase tracking-[0.06em] text-brand">
            {chipLabel}
          </span>
          <span className="min-w-0 flex-1 truncate text-[13px] text-fg-muted">
            <span className="font-medium text-foreground">Staged changes</span>
            {namesLabel && <span className="text-fg-muted"> {namesLabel}</span>}
          </span>
          {vsSeq != null && <span className="flex-none font-mono text-[11px] text-fg-muted">vs #{vsSeq}</span>}
          <ChevronDown className={`h-3.5 w-3.5 flex-none text-fg-muted transition-transform ${open ? "rotate-180" : ""}`} />
        </div>

        {open && (
          <div className="mb-1 mt-1.5 rounded-md border border-dashed border-brand bg-card p-4">
            {hasChanges && diff ? (
              <ConfigDiff diff={diff} hasPrev prevSeq={vsSeq} />
            ) : (
              <div className="text-[12.5px] text-fg-muted">Saved changes are staged for deploy.</div>
            )}
          </div>
        )}
      </div>
    </RailNode>
  );
}
