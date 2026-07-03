import { describe, expect, it } from "vitest";
import { mergeTopology } from "../merge-topology";
import type { CanvasGraph } from "../graph-from-connections";
import type { StackTopology } from "@/api/topology";

const localGraph = (): CanvasGraph => ({
  nodes: [
    { id: "resource:web", type: "resource", position: { x: 0, y: 0 }, data: { kind: "service", name: "web", kindLabel: "WEB", glyph: "web", dotState: "ok", summary: "", volumes: [] } },
    { id: "resource:api", type: "resource", position: { x: 0, y: 0 }, data: { kind: "service", name: "api", kindLabel: "WEB", glyph: "web", dotState: "ok", summary: "", volumes: [] } },
  ],
  edges: [
    { id: "env:resource:api->resource:web", source: "resource:api", target: "resource:web", type: "connection", data: { kind: "env", sourceOfTruth: "connection" } },
  ],
});

describe("mergeTopology", () => {
  it("returns local unchanged when server is null", () => {
    const local = localGraph();
    expect(mergeTopology(local, null)).toBe(local);
  });

  it("ignores server authored edges (local wins) and does not duplicate", () => {
    const server = {
      nodes: [],
      edges: [{ kind: "env", source: { type: "stack_resource", name: "api" }, target: { type: "stack_resource", name: "web" }, source_of_truth: "connection" }],
    } as unknown as StackTopology;
    expect(mergeTopology(localGraph(), server).edges).toHaveLength(1);
  });

  it("unions server derived edges, deduped against local", () => {
    const server = {
      nodes: [],
      edges: [
        { kind: "depends_on", source: { type: "stack_resource", name: "api" }, target: { type: "stack_resource", name: "web" }, source_of_truth: "derived" },
        { kind: "env", source: { type: "stack_resource", name: "api" }, target: { type: "stack_resource", name: "web" }, source_of_truth: "derived" }, // same key as local authored edge
      ],
    } as unknown as StackTopology;
    const merged = mergeTopology(localGraph(), server);
    expect(merged.edges.map((e) => e.id)).toEqual([
      "env:resource:api->resource:web",
      "depends_on:resource:api->resource:web",
    ]);
  });

  it("skips derived edges whose resource endpoint is gone locally, materializes missing secret endpoints", () => {
    const server = {
      nodes: [{ ref: { type: "secret", id: "s9" }, label: "legacy-creds" }],
      edges: [
        { kind: "depends_on", source: { type: "stack_resource", name: "ghost" }, target: { type: "stack_resource", name: "web" }, source_of_truth: "derived" },
        { kind: "env", source: { type: "secret", id: "s9" }, target: { type: "stack_resource", name: "web" }, source_of_truth: "derived" },
      ],
    } as unknown as StackTopology;
    const merged = mergeTopology(localGraph(), server);
    expect(merged.edges.map((e) => e.id)).not.toContain("depends_on:resource:ghost->resource:web");
    expect(merged.nodes.find((n) => n.id === "secret:s9")?.data).toMatchObject({ kind: "secret", name: "legacy-creds" });
    expect(merged.edges.map((e) => e.id)).toContain("env:secret:s9->resource:web");
  });

  it("overlays server node state onto matching resource nodes as status", () => {
    const server = {
      nodes: [{ ref: { type: "stack_resource", name: "web" }, label: "web", state: "Degraded" }],
      edges: [],
    } as unknown as StackTopology;
    const merged = mergeTopology(localGraph(), server);
    expect(merged.nodes.find((n) => n.id === "resource:web")?.data.status).toBe("Degraded");
  });
});
