import { NODE_WIDTH, NODE_HEIGHT } from "./layout-graph";

/**
 * Structural subset of a React Flow node this util needs — keeps it pure and
 * unit-testable without React Flow imports (mirrors layout-graph's approach).
 */
export interface CollidableNode {
  id: string;
  position: { x: number; y: number };
  measured?: { width?: number; height?: number };
  [key: string]: unknown;
}

export interface ResolveCollisionsOptions {
  /** Minimum clear space to leave between node boxes after separation. */
  margin: number;
  /** Hard stop for the relaxation loop (a pile of N nodes needs ~N passes). */
  maxIterations?: number;
  /** Locked nodes are never moved; an overlap between two locked nodes is left
   *  alone. Use to pin the user's arranged layout while placing new nodes. */
  isLocked?: (node: CollidableNode) => boolean;
  /** Box dimensions for nodes without `measured` (fresh layout output is never
   *  measured — React Flow only measures rendered nodes). Defaults to the
   *  resource-card dims, which oversizes attachment nodes. */
  fallbackSize?: (node: CollidableNode) => { width: number; height: number };
}

const DEFAULT_MAX_ITERATIONS = 50;

/**
 * Iteratively push overlapping node boxes apart along the axis of least
 * penetration until no pair overlaps (or maxIterations). Pattern from React
 * Flow's node-collisions example, adapted with lockable nodes so collision
 * resolution can place NEW nodes without disturbing the frozen layout.
 * Copy-on-write: returns the input array untouched when nothing overlaps.
 */
export function resolveCollisions<T extends CollidableNode>(
  nodes: T[],
  { margin, maxIterations = DEFAULT_MAX_ITERATIONS, isLocked, fallbackSize }: ResolveCollisionsOptions,
): T[] {
  const boxes = nodes.map((n) => {
    const fallback = fallbackSize?.(n) ?? { width: NODE_WIDTH, height: NODE_HEIGHT };
    return {
      node: n,
      x: n.position.x,
      y: n.position.y,
      w: n.measured?.width ?? fallback.width,
      h: n.measured?.height ?? fallback.height,
      locked: isLocked?.(n) ?? false,
      moved: false,
    };
  });

  let anyMoved = false;
  for (let iter = 0; iter < maxIterations; iter++) {
    let movedThisPass = false;
    for (let i = 0; i < boxes.length; i++) {
      for (let j = i + 1; j < boxes.length; j++) {
        const a = boxes[i];
        const b = boxes[j];
        if (a.locked && b.locked) continue;

        const overlapX = Math.min(a.x + a.w, b.x + b.w) - Math.max(a.x, b.x) + margin;
        const overlapY = Math.min(a.y + a.h, b.y + b.h) - Math.max(a.y, b.y) + margin;
        if (overlapX <= 0 || overlapY <= 0) continue;

        // Separate along the axis needing the smaller shove. Direction pushes
        // the boxes' centres apart; ties push b down-right (deterministic).
        const axis: "x" | "y" = overlapX <= overlapY ? "x" : "y";
        const amount = axis === "x" ? overlapX : overlapY;
        const aCentre = axis === "x" ? a.x + a.w / 2 : a.y + a.h / 2;
        const bCentre = axis === "x" ? b.x + b.w / 2 : b.y + b.h / 2;
        const dir = bCentre >= aCentre ? 1 : -1;

        const push = (box: typeof a, delta: number) => {
          if (axis === "x") box.x += delta;
          else box.y += delta;
          box.moved = true;
        };
        if (a.locked) push(b, dir * amount);
        else if (b.locked) push(a, -dir * amount);
        else {
          push(a, -dir * (amount / 2));
          push(b, dir * (amount / 2));
        }
        movedThisPass = true;
        anyMoved = true;
      }
    }
    if (!movedThisPass) break;
  }

  // Escape hatch: a movable node wedged in a corridor narrower than its box
  // (between locked nodes) ping-pongs until maxIterations and would return
  // still-overlapping. Park any such node above the bounding box of everything
  // else, staggered — deterministic and always resolvable.
  const stillOverlapping = (i: number) =>
    boxes.some((other, j) => {
      if (i === j) return false;
      const a = boxes[i];
      return (
        Math.min(a.x + a.w, other.x + other.w) - Math.max(a.x, other.x) + margin > 0 &&
        Math.min(a.y + a.h, other.y + other.h) - Math.max(a.y, other.y) + margin > 0
      );
    });
  const wedged = boxes.map((_, i) => i).filter((i) => !boxes[i].locked && stillOverlapping(i));
  // rest empty = an all-movable pile under a tiny maxIterations budget, not a
  // locked corridor — there's no frozen layout to park against, so leave it.
  if (wedged.length > 0 && wedged.length < boxes.length) {
    const wedgedSet = new Set(wedged);
    const rest = boxes.filter((_, i) => !wedgedSet.has(i));
    const minX = Math.min(...rest.map((b) => b.x));
    const minY = Math.min(...rest.map((b) => b.y));
    wedged.forEach((i, k) => {
      const b = boxes[i];
      b.x = minX + k * (b.w + margin);
      b.y = minY - b.h - margin;
      b.moved = true;
      anyMoved = true;
    });
  }

  if (!anyMoved) return nodes;
  return boxes.map(({ node, x, y, moved }) =>
    moved ? { ...node, position: { x, y } } : node,
  );
}
