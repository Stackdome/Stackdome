import { BaseEdge, getBezierPath, type EdgeProps } from "@xyflow/react";

/**
 * Connection edge — a variable reference between two resources. Rendered as a
 * thin dashed amber Bézier with a small filled dot at the destination, per the
 * design. Curvature 0.5 puts the control handles at the horizontal midpoint
 * (symmetric curve leaving the source's right, entering the target's left).
 */
export function ConnectionEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
}: EdgeProps) {
  const [path] = getBezierPath({
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
    curvature: 0.5,
  });

  return (
    <>
      <BaseEdge
        id={id}
        path={path}
        style={{
          stroke: "var(--brand)",
          strokeWidth: 1.4,
          strokeOpacity: 0.55,
          strokeDasharray: "5 4",
        }}
      />
      <circle cx={targetX} cy={targetY} r={3} fill="var(--brand)" />
    </>
  );
}
