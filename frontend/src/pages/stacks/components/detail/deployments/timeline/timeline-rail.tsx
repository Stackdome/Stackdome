import { useState } from "react";
import { EmptyState } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { stateTone } from "../derive";
import { useReleaseDetail } from "../use-release-detail";
import { RailNode, type RailDotShape } from "./rail-node";
import { TimelineNode } from "./timeline-node";
import type { LogContext } from "./resource-row";

const DEPLOYING = new Set(["Pending", "InProgress"]);

export interface TimelineRailProps {
  releases: StackRelease[];
  activeRelease?: StackRelease;
  stack: Stack;
  logContext?: LogContext;
  onOpenLogs?: (name: string) => void;
  banner?: React.ReactNode;
  /** Optional draft node, rendered at the head of the rail (saved-but-undeployed). */
  draftNode?: React.ReactNode;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
  onCopyId: (id: string) => void;
  initialWindow?: number;
}

function dotShape(state: string): RailDotShape {
  if (DEPLOYING.has(state)) return "spinner";
  if (state === "Failed") return "ring";
  return "solid";
}

/**
 * One continuous deploy timeline — newest at top, no Current/Earlier split. Each
 * release is a TimelineNode (lean row + card-below); the latest deploy opens by
 * default, earlier nodes start closed. An optional draft node leads the rail.
 */
export function TimelineRail(props: TimelineRailProps) {
  const { releases, activeRelease, stack, logContext, onOpenLogs, banner, draftNode, onRollback, onCancel, onCopyId, initialWindow = 15 } = props;
  const detail = useReleaseDetail(logContext?.orgId ?? "", logContext?.teamName ?? "", logContext?.stackId ?? "");
  const liveReleaseId = stack.status?.last_converged?.release_id;
  // Open the latest deploy AND the live release by default — the live one is the
  // jump-to-live target, so it must already be expanded when scrolled into view.
  const [openIds, setOpenIds] = useState<Set<string>>(
    () => new Set([activeRelease?.id, liveReleaseId].filter((x): x is string => !!x)),
  );
  const [windowN, setWindowN] = useState(initialWindow);

  // Multiple release details can be open at once — not an accordion.
  const toggle = (id: string) =>
    setOpenIds((cur) => {
      const next = new Set(cur);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  const shown = releases.slice(0, windowN);
  const hidden = releases.length - shown.length;
  const prevIdFor = (idx: number) => releases[idx + 1]?.id; // idx is the position in `releases`
  const prevSeqFor = (idx: number) => releases[idx + 1]?.sequence;

  return (
    <div className="space-y-0">
      {banner && <div className="mb-5">{banner}</div>}

      {draftNode}

      {releases.length === 0 && !draftNode ? (
        <RailNode tone="muted" isLast>
          <EmptyState title="No deployments yet" description="Deploy this stack to create your first release." />
        </RailNode>
      ) : (
        shown.map((r, idx) => {
          const isLast = idx === shown.length - 1 && hidden <= 0;
          const state = r.state ?? "";
          return (
            <RailNode key={r.id ?? idx} id={r.id ? `deploy-node-${r.id}` : undefined} tone={stateTone(state)} shape={dotShape(state)} pulse={DEPLOYING.has(state)} isLast={isLast}>
              <TimelineNode
                release={r}
                prevReleaseId={prevIdFor(idx)}
                prevSeq={prevSeqFor(idx)}
                detail={detail}
                isOpen={openIds.has(r.id ?? "")}
                onToggle={toggle}
                onRollback={onRollback}
                onCancel={onCancel}
                onCopyId={onCopyId}
                isActive={idx === 0}
                isLive={!!liveReleaseId && r.id === liveReleaseId}
                stack={stack}
                logContext={logContext}
                onOpenLogs={onOpenLogs}
              />
            </RailNode>
          );
        })
      )}

      {hidden > 0 && (
        <div className="ml-12 pt-2">
          <button onClick={() => setWindowN(releases.length)} className="font-sans text-[12.5px] font-medium text-primary">
            Show more ({hidden})
          </button>
        </div>
      )}
    </div>
  );
}
