import { Position } from "@xyflow/react";

/** Node rectangle in flow coordinates: top-left origin + measured size. */
export interface NodeRect {
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface FloatingEdgeGeometry {
  sourceX: number;
  sourceY: number;
  targetX: number;
  targetY: number;
  sourcePosition: Position;
  targetPosition: Position;
}

/** Perpendicular spacing between parallel edges of the same node pair. */
export const PARALLEL_EDGE_PITCH = 14;

interface FacePoint {
  x: number;
  y: number;
  position: Position;
}

/**
 * Intersection of the ray from this rect's centre (cx,cy) toward (ox,oy)
 * with the rect boundary, plus which face it lies on. Ray-based, so it is
 * well-defined even when the other centre is inside this rect.
 */
function faceIntersection(rect: NodeRect, cx: number, cy: number, ox: number, oy: number): FacePoint {
  const dx = ox - cx;
  const dy = oy - cy;
  const hw = rect.width / 2;
  const hh = rect.height / 2;
  const scaleX = dx !== 0 ? hw / Math.abs(dx) : Number.POSITIVE_INFINITY;
  const scaleY = dy !== 0 ? hh / Math.abs(dy) : Number.POSITIVE_INFINITY;
  const scale = Math.min(scaleX, scaleY);
  const position =
    scaleX < scaleY
      ? dx > 0
        ? Position.Right
        : Position.Left
      : dy > 0
        ? Position.Bottom
        : Position.Top;
  return { x: cx + dx * scale, y: cy + dy * scale, position };
}

/**
 * Floating-edge attachment: each end attaches where the (optionally offset)
 * centre-to-centre line crosses its node's rectangle, with the matching
 * React Flow Position for bezier control-point direction.
 *
 * Parallel edges between the same node pair pass their index/count; the
 * whole centre line shifts perpendicular by a symmetric multiple of
 * PARALLEL_EDGE_PITCH so the edges render as distinct parallel curves.
 */
export function floatingEdgeGeometry(
  source: NodeRect,
  target: NodeRect,
  parallelIndex = 0,
  parallelCount = 1,
): FloatingEdgeGeometry {
  let scx = source.x + source.width / 2;
  let scy = source.y + source.height / 2;
  let tcx = target.x + target.width / 2;
  let tcy = target.y + target.height / 2;

  const dx = tcx - scx;
  const dy = tcy - scy;
  const len = Math.hypot(dx, dy);

  if (len === 0) {
    // Coincident centres (mid-drag overlap): stable vertical fallback.
    return {
      sourceX: scx,
      sourceY: source.y + source.height,
      targetX: tcx,
      targetY: target.y,
      sourcePosition: Position.Bottom,
      targetPosition: Position.Top,
    };
  }

  if (parallelCount > 1) {
    // The WHOLE centre line shifts before intersecting, so offset attachment
    // points sit slightly off the exact perimeter (~half the pitch at most).
    // Deliberate: it keeps parallel curves parallel along their full length.
    const offset = (parallelIndex - (parallelCount - 1) / 2) * PARALLEL_EDGE_PITCH;
    const px = (-dy / len) * offset;
    const py = (dx / len) * offset;
    scx += px;
    scy += py;
    tcx += px;
    tcy += py;
  }

  const s = faceIntersection(source, scx, scy, tcx, tcy);
  const t = faceIntersection(target, tcx, tcy, scx, scy);
  return {
    sourceX: s.x,
    sourceY: s.y,
    targetX: t.x,
    targetY: t.y,
    sourcePosition: s.position,
    targetPosition: t.position,
  };
}
