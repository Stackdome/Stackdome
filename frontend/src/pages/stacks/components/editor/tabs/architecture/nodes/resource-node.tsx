import { memo } from "react";
import { Handle, Position, type NodeProps, type Node } from "@xyflow/react";
import { HardDrive } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ResourceNodeData } from "@/pages/stacks/lib/canvas/graph-from-connections";
import type { StatusVariant } from "@/components/branded/status-variant";
import { NodeGlyph } from "./node-glyph";

export type ResourceFlowNode = Node<ResourceNodeData, "resource">;

/** Edges are derived, not hand-drawn, so the handles are anchors only. */
const HIDDEN_HANDLE = { opacity: 0, pointerEvents: "none" as const };

/** Live status dot per variant. Ready breathes (slow pulse = heartbeat);
 *  pending uses the standard motion pulse; error/info/neutral hold still. */
const DOT_CLASS: Record<StatusVariant, string> = {
  ready: "bg-success animate-breathe",
  pending: "bg-warn animate-pulse",
  error: "bg-danger",
  info: "bg-info",
  neutral: "bg-fg-muted",
};

function ResourceNodeImpl({ data, selected }: NodeProps<ResourceFlowNode>) {
  const dirty = data.dirtyState;
  // Unsaved changes read as a left accent stripe + tinted border; removal is
  // crimson + dimmed. Selection (amber border + halo) wins the border colour.
  const stripeColor = dirty === "removed" ? "bg-danger" : dirty ? "bg-brand" : null;
  const borderClass = selected
    ? "border-brand"
    : dirty === "removed"
      ? "border-danger/50"
      : dirty
        ? "border-brand/60"
        : "border-border";

  return (
    <div
      className={cn(
        "relative w-[216px] cursor-grab overflow-hidden rounded-lg border bg-card shadow-xs transition-colors",
        borderClass,
        selected && "ring-[3px] ring-brand/20",
        data.dropTarget && "border-brand ring-[3px] ring-brand/30",
        dirty === "removed" && "opacity-60",
      )}
    >
      {stripeColor && <span className={cn("absolute inset-y-0 left-0 w-[3px]", stripeColor)} aria-hidden />}
      <Handle type="target" position={Position.Left} style={HIDDEN_HANDLE} isConnectable={false} />
      <Handle type="source" position={Position.Right} style={HIDDEN_HANDLE} isConnectable={false} />

      <div className="px-[13px] py-3">
        <div className="flex items-center gap-2.5">
          <span className={cn("size-2 shrink-0 rounded-full", DOT_CLASS[data.dotVariant])} aria-hidden />
          <NodeGlyph glyph={data.glyph} brandSlug={data.brandSlug} size={17} className="size-[17px] shrink-0 text-fg-2" />
          <span className="flex-1 truncate text-sm font-medium text-foreground">{data.name}</span>
          <span className="shrink-0 font-mono text-[9px] uppercase tracking-[0.12em] text-fg-muted">
            {data.kindLabel}
          </span>
        </div>
        <div className="mt-1.5 pl-[18px] font-mono text-[11px] text-muted-foreground">
          <div className="truncate">{data.summary}</div>
          {data.detail && <div className="mt-0.5 truncate">{data.detail}</div>}
        </div>
      </div>

      {data.volumes.map((v) => (
        <div
          key={v.name}
          title={v.mountPath}
          data-volume-chip={v.name}
          className="flex cursor-pointer items-center gap-2 border-t border-border bg-background px-[13px] py-2 transition-colors hover:bg-card"
        >
          <HardDrive className="size-[13px] shrink-0 text-fg-muted" aria-hidden />
          <span className="truncate font-mono text-[10.5px] text-fg-2">{v.name}</span>
        </div>
      ))}
    </div>
  );
}

export const ResourceNode = memo(ResourceNodeImpl);
