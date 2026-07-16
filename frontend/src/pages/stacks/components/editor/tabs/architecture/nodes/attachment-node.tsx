import { memo } from "react";
import { Handle, Position, type Node, type NodeProps } from "@xyflow/react";
import { Archive, HardDrive, KeyRound, type LucideIcon } from "lucide-react";
import { NODE_KIND, type AttachmentKind, type AttachmentNodeData } from "@/pages/stacks/lib/canvas/graph-from-connections";

export type AttachmentFlowNode = Node<AttachmentNodeData, "attachment">;

/** Edges are derived, not hand-drawn, so the handles are anchors only. */
const HIDDEN_HANDLE = { opacity: 0, pointerEvents: "none" as const };

const ICON: Record<AttachmentKind, LucideIcon> = {
  [NODE_KIND.secret]: KeyRound,
  [NODE_KIND.volume]: HardDrive,
  [NODE_KIND.objectStore]: Archive,
};

/**
 * Compact node for connection endpoints that aren't workloads — secrets,
 * volumes, object stores. Same card language as ResourceNode, one line,
 * visually lighter. Volume nodes open the volume drawer on click; secret/
 * object-store nodes stay display-only for now.
 */
function AttachmentNodeImpl({ data }: NodeProps<AttachmentFlowNode>) {
  const Icon = ICON[data.kind];
  return (
    <div className="w-[180px] cursor-pointer rounded-lg border border-border bg-card px-[13px] py-2.5 shadow-xs transition-colors hover:border-brand/60">
      <Handle type="target" position={Position.Left} style={HIDDEN_HANDLE} isConnectable={false} />
      <Handle type="source" position={Position.Right} style={HIDDEN_HANDLE} isConnectable={false} />
      <div className="flex items-center gap-2.5">
        <Icon className="size-[15px] shrink-0 text-fg-muted" aria-hidden />
        <span className="flex-1 truncate text-[13px] font-medium text-fg-2">{data.name}</span>
        <span className="shrink-0 font-mono text-[9px] uppercase tracking-[0.12em] text-fg-muted">{data.kindLabel}</span>
      </div>
    </div>
  );
}

export const AttachmentNode = memo(AttachmentNodeImpl);
