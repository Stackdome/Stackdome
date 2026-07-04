import { BaseEdge, getBezierPath, type EdgeProps } from "@xyflow/react";

/**
 * Connection edge — a uniform reference line between two nodes, with a filled
 * dot at the consuming end. All kinds render identically as plain unlabeled
 * lines; the connection's kind and source_of_truth stay in the edge data for
 * future use.
 */
export function ConnectionEdge({
  id, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition,
}: EdgeProps) {
  const [path] = getBezierPath({
    sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, curvature: 0.5,
  });

  return (
    <>
      <BaseEdge id={id} path={path} style={{ stroke: "var(--brand)", strokeWidth: 1.4, strokeOpacity: 0.7 }} />
      <circle cx={targetX} cy={targetY} r={3} fill="var(--brand)" />
    </>
  );
}
