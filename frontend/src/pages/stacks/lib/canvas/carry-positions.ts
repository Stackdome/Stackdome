import { NODE_KIND } from "./graph-from-connections";

/**
 * Structural subset of a React Flow node this helper needs — pure and
 * unit-testable without React Flow imports.
 */
export interface CarryableNode {
  id: string;
  type?: string;
  position: { x: number; y: number };
  measured?: { width?: number; height?: number };
  data: { kind?: unknown; resourceIdx?: number; volumeIdx?: number; [key: string]: unknown };
  [key: string]: unknown;
}

export interface CarryResult<T> {
  nodes: T[];
  /** Ids (from `laid`) whose position came from `prev` — i.e. everything that
   *  is NOT genuinely new. Callers lock these during collision resolution. */
  keptIds: Set<string>;
}

const isResource = (n: CarryableNode) => n.type === "resource" && n.data.resourceIdx != null;
const isVolume = (n: CarryableNode) => n.type === "attachment" && n.data.kind === NODE_KIND.volume;

/**
 * Carry user-arranged positions from the previous node set onto a freshly
 * laid-out one. Match by id first; a miss falls back to the node's stable
 * session identity (resourceIdx for resources, volumeIdx for floating
 * volumes) — node ids embed the NAME, so renaming mints a new id per
 * keystroke and id-matching alone would treat the rename as an add.
 *
 * The identity fallback is parity-gated per kind: indexes are raw array
 * positions, so a topology change that deletes AND renames in one pass
 * shifts them — node B (idx 1) renamed after deleting A (idx 0) would
 * inherit deleted A's position. When the kind's node count changed, the
 * fallback is disabled and an unmatched node keeps its fresh layout coords.
 *
 * Kept nodes also inherit `measured` (React Flow only measures rendered
 * nodes; fresh layout output carries none, and collision resolution needs
 * real boxes).
 */
export function carryPositions<T extends CarryableNode>(prev: T[], laid: T[]): CarryResult<T> {
  const prevById = new Map(prev.map((n) => [n.id, n]));

  const resourceParity = prev.filter(isResource).length === laid.filter(isResource).length;
  const volumeParity = prev.filter(isVolume).length === laid.filter(isVolume).length;

  const prevByResourceIdx = new Map<number, T>();
  const prevByVolumeIdx = new Map<number, T>();
  for (const n of prev) {
    if (isResource(n)) prevByResourceIdx.set(n.data.resourceIdx!, n);
    else if (isVolume(n) && n.data.volumeIdx != null) prevByVolumeIdx.set(n.data.volumeIdx, n);
  }

  const keptIds = new Set<string>();
  const nodes = laid.map((n) => {
    let source = prevById.get(n.id);
    if (!source) {
      if (isResource(n) && resourceParity) source = prevByResourceIdx.get(n.data.resourceIdx!);
      else if (isVolume(n) && volumeParity && n.data.volumeIdx != null) source = prevByVolumeIdx.get(n.data.volumeIdx);
    }
    if (!source) return n;
    keptIds.add(n.id);
    return {
      ...n,
      position: source.position,
      ...(source.measured ? { measured: source.measured } : {}),
    };
  });

  return { nodes, keptIds };
}
