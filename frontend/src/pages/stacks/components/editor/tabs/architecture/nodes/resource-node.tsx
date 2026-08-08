import { Fragment, memo } from "react";
import { Handle, Position, type NodeProps, type Node } from "@xyflow/react";
import { ExternalLink, HardDrive } from "lucide-react";
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
  // crimson + dimmed. Selection wins the border colour with an ink wash, not
  // an orange ring — brand orange stays reserved for wires/eyebrows/mark.
  const stripeColor = dirty === "removed" ? "bg-danger" : dirty ? "bg-brand" : null;
  const borderClass = selected
    ? "border-border-strong"
    : dirty === "removed"
      ? "border-danger/50"
      : dirty
        ? "border-brand/60"
        : "border-border";

  return (
    <div
      className={cn(
        "relative w-[216px] cursor-grab overflow-hidden rounded-lg border bg-surface-node transition-colors",
        borderClass,
        // Deliberate exception to "orange = wires only": a drop is about to
        // create a connection edge, so the transient drag-affordance borrows
        // the wire colour. Distinct from the steady-state ink-wash `selected`
        // treatment above — orange reads as "act now", ink reads as "at rest".
        data.dropTarget && "border-brand ring-[3px] ring-brand/30",
        dirty === "removed" && "opacity-60",
      )}
    >
      {selected && <span className="pointer-events-none absolute inset-0 bg-foreground/[0.06]" aria-hidden />}
      {stripeColor && <span className={cn("absolute inset-y-0 left-0 w-[3px]", stripeColor)} aria-hidden />}
      <Handle type="target" position={Position.Left} style={HIDDEN_HANDLE} isConnectable={false} />
      <Handle type="source" position={Position.Right} style={HIDDEN_HANDLE} isConnectable={false} />

      <div className="px-[13px] py-3">
        <div className="flex items-center gap-2.5">
          <span className={cn("size-2 shrink-0 rounded-full", DOT_CLASS[data.dotVariant])} aria-hidden />
          <NodeGlyph glyph={data.glyph} brandSlug={data.brandSlug} size={17} className="size-[17px] shrink-0 text-fg-2" />
          <span className="flex-1 truncate text-body font-medium text-foreground">{data.name}</span>
          <span className="shrink-0 font-mono text-[9px] text-fg-muted">
            {data.kindLabel}
          </span>
        </div>
        <div className="mt-1.5 pl-[18px] font-mono text-label text-muted-foreground">
          <div className="truncate">{data.summary}</div>
          {(data.details ?? []).length > 0 && (
            <div className="mt-0.5 flex flex-wrap items-center gap-x-1 gap-y-0.5">
              <span>ports:</span>
              {(data.details ?? []).map((line, i) => {
                const url = data.portUrls?.[line.port];
                const sep = i > 0 && <span className="opacity-40">·</span>;
                if (!url) {
                  return (
                    <Fragment key={`${line.port}-${i}`}>
                      {sep}
                      <span title={line.text}>{line.port}</span>
                    </Fragment>
                  );
                }
                return (
                  <Fragment key={`${line.port}-${i}`}>
                    {sep}
                    <a
                      href={url}
                      target="_blank"
                      rel="noreferrer"
                      title={`${line.text} — ${url}`}
                      // Card drag/click owns the mousedown; stop it so the link clicks.
                      onMouseDown={(e) => e.stopPropagation()}
                      onClick={(e) => e.stopPropagation()}
                      className="group/port flex items-center gap-0.5 transition-colors hover:text-foreground"
                    >
                      <span className="group-hover/port:underline">{line.port}</span>
                      <ExternalLink className="size-2.5 shrink-0 opacity-40 transition-opacity group-hover/port:opacity-100" aria-hidden />
                    </a>
                  </Fragment>
                );
              })}
            </div>
          )}
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
          <span className="truncate font-mono text-label text-fg-2">{v.name}</span>
        </div>
      ))}
    </div>
  );
}

export const ResourceNode = memo(ResourceNodeImpl);
