import { useState } from "react";
import { ChevronDown } from "lucide-react";
import type { SnapshotDiff } from "../release-snapshot-diff";
import { ConfigDiff } from "./config-diff";

export interface ConfigChangesToggleProps {
  diff: SnapshotDiff;
  /** Sequence the diff is taken against (the previous release). */
  prevSeq?: number;
  defaultOpen?: boolean;
  /** Previous snapshot still loading — show a placeholder instead of a false "no changes". */
  loading?: boolean;
}

/** "CONFIG CHANGES · vs #N" collapsible at the foot of every node's detail card; identical across live + historical. */
export function ConfigChangesToggle({ diff, prevSeq, defaultOpen = false, loading = false }: ConfigChangesToggleProps) {
  const [open, setOpen] = useState(defaultOpen);
  return (
    <div className="mt-4 border-t border-border pt-4">
      <button
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-1.5 font-mono text-[11px] uppercase tracking-wide text-brand hover:text-brand/80"
      >
        Config changes · vs #{prevSeq ?? "previous"}
        <ChevronDown className={`h-3 w-3 transition-transform ${open ? "rotate-180" : ""}`} />
      </button>
      {open && (
        loading
          ? <div className="mt-3 text-[12.5px] text-fg-muted">Loading changes…</div>
          : <div className="mt-3"><ConfigDiff diff={diff} hasPrev prevSeq={prevSeq} /></div>
      )}
    </div>
  );
}
