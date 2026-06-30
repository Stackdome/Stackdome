import { memo } from "react";
import { Handle, Position, type NodeProps, type Node } from "@xyflow/react";
import { cn } from "@/lib/utils";
import { NODE_KIND, type ResourceNodeData, type DirtyState } from "@/pages/stacks/lib/canvas/derive-graph";
import { NodeGlyph } from "./node-glyph";

export type ResourceFlowNode = Node<ResourceNodeData, "resource">;

/** Edges are derived, not hand-drawn, so the handles are anchors only. */
const HIDDEN_HANDLE = { opacity: 0, pointerEvents: "none" as const };

const DIRTY_MARK: Record<DirtyState, { border: string; label: string; text: string }> = {
  new: { border: "border-success", label: "NEW", text: "text-success" },
  edited: { border: "border-brand", label: "EDITED", text: "text-brand" },
  removed: { border: "border-danger", label: "REMOVED", text: "text-danger" },
};

function kindLabel(data: ResourceNodeData): string {
  return data.kind === NODE_KIND.addon ? "POSTGRES" : "SERVICE";
}

function ResourceNodeImpl({ data, selected }: NodeProps<ResourceFlowNode>) {
  const mark = data.dirtyState ? DIRTY_MARK[data.dirtyState] : undefined;
  const borderClass = mark ? mark.border : selected ? "border-brand" : "border-border";

  return (
    <div
      className={cn(
        "relative w-[216px] rounded-lg border bg-card px-3 py-2.5 shadow-xs transition-colors",
        borderClass,
        data.dirtyState === "removed" && "opacity-60",
      )}
    >
      <Handle type="target" position={Position.Left} style={HIDDEN_HANDLE} isConnectable={false} />
      <Handle type="source" position={Position.Right} style={HIDDEN_HANDLE} isConnectable={false} />

      <div className="flex items-center gap-2">
        <span className="size-1.5 shrink-0 rounded-full bg-success" aria-hidden />
        <NodeGlyph kind={data.kind} className="size-3.5 shrink-0 text-fg-muted" />
        <span className="truncate text-sm font-medium text-foreground">{data.name}</span>
        <span
          className={cn(
            "ml-auto shrink-0 font-mono text-[10px] uppercase tracking-wider",
            mark ? mark.text : "text-fg-muted",
          )}
        >
          {mark ? mark.label : kindLabel(data)}
        </span>
      </div>

      <div className="mt-1 truncate font-mono text-[11px] text-muted-foreground">{data.summary}</div>

      {data.volumes.length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1">
          {data.volumes.map((v) => (
            <span
              key={v.name}
              title={v.mountPath}
              className="rounded-sm bg-muted px-1.5 py-0.5 font-mono text-[10px] text-fg-2"
            >
              {v.name}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

export const ResourceNode = memo(ResourceNodeImpl);
