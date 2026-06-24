import { useState } from "react";
import { ChevronDown } from "lucide-react";
import type { SnapshotDiff } from "../release-snapshot-diff";
import { ConfigDiff } from "./config-diff";

export interface ConfigChangesToggleProps {
  diff: SnapshotDiff;
  /** Sequence the diff is taken against (the previous release). */
  prevSeq?: number;
  defaultOpen?: boolean;
}

/**
 * The "CONFIG CHANGES · vs #N" collapsible shown at the foot of every node's
 * detail card. Brand-amber to read as the deploy accent (per the design), and
 * identical across live + historical nodes so the section is uniform.
 */
export function ConfigChangesToggle({ diff, prevSeq, defaultOpen = false }: ConfigChangesToggleProps) {
  const [open, setOpen] = useState(defaultOpen);
  return (
    <div className="mt-4">
      <button
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-1.5 font-mono text-[11px] uppercase tracking-wide text-brand hover:text-brand/80"
      >
        Config changes · vs #{prevSeq ?? "previous"}
        <ChevronDown className={`h-3 w-3 transition-transform ${open ? "rotate-180" : ""}`} />
      </button>
      {open && <div className="mt-3"><ConfigDiff diff={diff} hasPrev prevSeq={prevSeq} /></div>}
    </div>
  );
}
