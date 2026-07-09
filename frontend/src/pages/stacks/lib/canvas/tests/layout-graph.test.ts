import { describe, it, expect } from "vitest";
import { layoutGraph } from "../layout-graph";
import type { CanvasGraph } from "../graph-from-connections";

const graph: CanvasGraph = {
  nodes: [
    {
      id: "resource:web",
      type: "resource",
      data: { kind: "service", name: "web", kindLabel: "Web", glyph: "web", summary: "", volumes: [] },
      position: { x: 0, y: 0 },
    },
    {
      id: "addon:a1",
      type: "resource",
      data: { kind: "addon", name: "db", kindLabel: "Postgres", glyph: "postgres", summary: "", volumes: [] },
      position: { x: 0, y: 0 },
    },
  ],
  edges: [
    {
      id: "resource:web->addon:a1",
      source: "resource:web",
      target: "addon:a1",
      type: "connection",
      data: { kind: "env", sourceOfTruth: "connection" },
    },
  ],
};

describe("layoutGraph", () => {
  it("assigns non-origin, distinct positions", () => {
    const out = layoutGraph(graph);
    expect(out.nodes[0].position).not.toEqual({ x: 0, y: 0 });
    expect(out.nodes[0].position).not.toEqual(out.nodes[1].position);
  });

  it("is deterministic for the same input", () => {
    const a = layoutGraph(graph);
    const b = layoutGraph(graph);
    expect(a.nodes.map((n) => n.position)).toEqual(b.nodes.map((n) => n.position));
  });

  it("does not mutate the input graph", () => {
    layoutGraph(graph);
    expect(graph.nodes[0].position).toEqual({ x: 0, y: 0 });
  });

  it("returns every input node", () => {
    const out = layoutGraph(graph);
    expect(out.nodes.map((n) => n.id).sort()).toEqual(["addon:a1", "resource:web"]);
  });

  it("handles an empty graph", () => {
    expect(layoutGraph({ nodes: [], edges: [] })).toEqual({ nodes: [], edges: [] });
  });
});
