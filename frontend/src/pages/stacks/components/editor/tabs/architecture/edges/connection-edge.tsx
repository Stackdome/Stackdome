import { BaseEdge, getBezierPath, useInternalNode, type EdgeProps, type InternalNode } from "@xyflow/react";
import { floatingEdgeGeometry, type NodeRect } from "@/pages/stacks/lib/canvas/floating-edge-geometry";
import { NODE_WIDTH, NODE_HEIGHT } from "@/pages/stacks/lib/canvas/layout-graph";
import type { ConnectionEdgeData } from "@/pages/stacks/lib/canvas/graph-from-connections";

/** Measured node → flow-coordinate rect; falls back to layout constants pre-measure. */
function rectOf(node: InternalNode): NodeRect {
  return {
    x: node.internals.positionAbsolute.x,
    y: node.internals.positionAbsolute.y,
    width: node.measured?.width ?? NODE_WIDTH,
    height: node.measured?.height ?? NODE_HEIGHT,
  };
}

/**
 * Connection edge — floating attachment: each end connects where the
 * centre-to-centre line crosses the node's perimeter (not a fixed side
 * handle), so fan-ins spread across a face instead of pinching into one
 * point. Parallel same-pair edges are offset into distinct lines. A filled
 * dot marks the consuming (target) end. Kind and source_of_truth stay in
 * the edge data for future styling.
 */
export function ConnectionEdge({ id, source, target, data }: EdgeProps) {
  const sourceNode = useInternalNode(source);
  const targetNode = useInternalNode(target);
  if (!sourceNode || !targetNode) return null;

  const edgeData = data as ConnectionEdgeData | undefined;
  const geo = floatingEdgeGeometry(
    rectOf(sourceNode),
    rectOf(targetNode),
    edgeData?.parallelIndex,
    edgeData?.parallelCount,
  );
  const [path] = getBezierPath({ ...geo, curvature: 0.5 });

  return (
    <>
      <BaseEdge id={id} path={path} style={{ stroke: "var(--brand)", strokeWidth: 1.4, strokeOpacity: 0.7 }} />
      <circle cx={geo.targetX} cy={geo.targetY} r={3} fill="var(--brand)" />
    </>
  );
}
