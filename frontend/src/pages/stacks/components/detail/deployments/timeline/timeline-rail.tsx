import { useState } from "react";
import { EmptyState } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { stateTone } from "../derive";
import { useReleaseDetail } from "../use-release-detail";
import { RailNode } from "./rail-node";
import { CurrentReleaseNode } from "./current-release-node";
import { HistoryRow } from "./history-row";
import type { LogContext } from "./resource-row";

const TERMINAL = new Set(["Released", "Failed", "Superseded", "Cancelled"]);

export interface TimelineRailProps {
  releases: StackRelease[];
  activeRelease?: StackRelease;
  stack: Stack;
  logContext?: LogContext;
  onOpenLogs?: (name: string) => void;
  banner?: React.ReactNode;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
  onCopyId: (id: string) => void;
  initialWindow?: number;
}

export function TimelineRail(props: TimelineRailProps) {
  const { releases, activeRelease, stack, logContext, onOpenLogs, banner, onRollback, onCancel, onCopyId, initialWindow = 15 } = props;
  const detail = useReleaseDetail(logContext?.orgId ?? "", logContext?.teamName ?? "", logContext?.stackId ?? "");
  const [openIds, setOpenIds] = useState<Set<string>>(() => new Set());
  const [windowN, setWindowN] = useState(initialWindow);
  // Multiple release details can be open at once — not an accordion.
  const toggle = (id: string) =>
    setOpenIds((cur) => {
      const next = new Set(cur);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  const earlier = releases.slice(1); // activeRelease = releases[0]
  const shown = earlier.slice(0, windowN);
  const prevIdFor = (idx: number) => releases[idx + 1]?.id;   // idx is the index in `releases`
  const prevSeqFor = (idx: number) => releases[idx + 1]?.sequence;

  return (
    <div className="space-y-0">
      {banner && <div className="mb-5">{banner}</div>}

      {activeRelease && (
        <div className="ml-12 pb-2 font-mono text-[11px] uppercase tracking-wide text-fg-muted">Current deployment</div>
      )}

      {activeRelease ? (
        <RailNode tone={stateTone(activeRelease.state ?? "")} big pulse={!TERMINAL.has(activeRelease.state ?? "")} isLast={earlier.length === 0}>
          <CurrentReleaseNode
            release={activeRelease}
            stack={stack}
            logContext={logContext}
            onOpenLogs={onOpenLogs}
            onCancel={onCancel}
            detail={detail}
            prevReleaseId={prevIdFor(0)}
            prevSeq={prevSeqFor(0)}
          />
        </RailNode>
      ) : releases.length === 0 ? (
        <RailNode tone="muted" isLast>
          <EmptyState title="No deployments yet" description="Deploy this stack to create your first release." />
        </RailNode>
      ) : null}

      {earlier.length > 0 && (
        <div className="ml-12 py-2 font-mono text-[11px] uppercase tracking-wide text-fg-muted">Earlier deployments</div>
      )}

      {shown.map((r, i) => {
        const idx = i + 1; // position in `releases`
        return (
          <RailNode key={r.id ?? idx} tone={stateTone(r.state ?? "")} isLast={idx === releases.length - 1}>
            <HistoryRow
              release={r}
              prevReleaseId={prevIdFor(idx)}
              prevSeq={prevSeqFor(idx)}
              detail={detail}
              isOpen={openIds.has(r.id ?? "")}
              onToggle={toggle}
              onRollback={onRollback}
              onCancel={onCancel}
              onCopyId={onCopyId}
            />
          </RailNode>
        );
      })}

      {earlier.length > windowN && (
        <div className="ml-12 pt-2">
          <button onClick={() => setWindowN(earlier.length)} className="font-sans text-[12.5px] font-medium text-primary">
            Show more ({earlier.length - windowN})
          </button>
        </div>
      )}
    </div>
  );
}
