import { BaseEdge, EdgeLabelRenderer, getBezierPath, type EdgeProps } from "@xyflow/react";
import {
  EDGE_KIND, EDGE_SOURCE_OF_TRUTH,
  type ConnectionEdgeData, type EdgeKind,
} from "@/pages/stacks/lib/canvas/graph-from-connections";

/** Short chip label per connection kind. */
const KIND_LABEL: Record<EdgeKind, string> = {
  [EDGE_KIND.env]: "env",
  [EDGE_KIND.volumeMount]: "mount",
  [EDGE_KIND.buildArtifactSource]: "build",
  [EDGE_KIND.dependsOn]: "deps",
};

/**
 * Connection edge — an explicit StackConnection (solid brand) or a derived
 * relationship such as depends_on (dashed, muted). A mid-edge chip names the
 * connection kind; a filled dot marks the consuming end.
 */
export function ConnectionEdge({
  id, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, data,
}: EdgeProps) {
  const [path, labelX, labelY] = getBezierPath({
    sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, curvature: 0.5,
  });
  const edgeData = (data ?? { kind: EDGE_KIND.env, sourceOfTruth: EDGE_SOURCE_OF_TRUTH.connection }) as ConnectionEdgeData;
  const derived = edgeData.sourceOfTruth === EDGE_SOURCE_OF_TRUTH.derived;

  return (
    <>
      <BaseEdge
        id={id}
        path={path}
        style={
          derived
            ? { stroke: "var(--fg-muted)", strokeWidth: 1.2, strokeOpacity: 0.5, strokeDasharray: "4 4" }
            : { stroke: "var(--brand)", strokeWidth: 1.4, strokeOpacity: 0.7 }
        }
      />
      <circle cx={targetX} cy={targetY} r={3} fill={derived ? "var(--fg-muted)" : "var(--brand)"} />
      <EdgeLabelRenderer>
        <span
          style={{ transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)` }}
          className="pointer-events-none absolute rounded border border-border bg-card px-1 py-px font-mono text-[9px] uppercase tracking-[0.1em] text-fg-muted"
        >
          {KIND_LABEL[edgeData.kind] ?? edgeData.kind}
        </span>
      </EdgeLabelRenderer>
    </>
  );
}
