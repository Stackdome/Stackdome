import dagre from "dagre";
import type { CanvasGraph } from "./derive-graph";

/** Card dimensions dagre reserves per node (matches ResourceNode's box). */
const NODE_WIDTH = 216;
const NODE_HEIGHT = 88;

export interface LayoutOptions {
  direction?: "LR" | "TB";
}

/**
 * Pure auto-layout. Runs dagre over the graph and returns a NEW graph with
 * resolved positions (copy-on-write — the input is never mutated). Deterministic
 * for a given input, so unrelated re-renders never reshuffle the board.
 *
 * dagre reports node centres; React Flow positions are top-left, so we offset by
 * half the node box.
 */
export function layoutGraph(graph: CanvasGraph, options: LayoutOptions = {}): CanvasGraph {
  if (graph.nodes.length === 0) return { nodes: [], edges: graph.edges };

  const g = new dagre.graphlib.Graph();
  g.setGraph({ rankdir: options.direction ?? "LR", nodesep: 48, ranksep: 96, marginx: 24, marginy: 24 });
  g.setDefaultEdgeLabel(() => ({}));

  for (const node of graph.nodes) {
    g.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT });
  }
  for (const edge of graph.edges) {
    g.setEdge(edge.source, edge.target);
  }

  dagre.layout(g);

  const nodes = graph.nodes.map((node) => {
    const { x, y } = g.node(node.id);
    return { ...node, position: { x: x - NODE_WIDTH / 2, y: y - NODE_HEIGHT / 2 } };
  });

  return { nodes, edges: graph.edges };
}
