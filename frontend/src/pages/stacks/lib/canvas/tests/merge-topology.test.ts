import { describe, expect, it } from "vitest";
import { mergeTopology } from "../merge-topology";
import type { CanvasGraph } from "../graph-from-connections";
import type { StackTopology } from "@/api/topology";

const localGraph = (): CanvasGraph => ({
  nodes: [
    { id: "resource:web", type: "resource", position: { x: 0, y: 0 }, data: { kind: "service", name: "web", kindLabel: "WEB", glyph: "web", dotVariant: "neutral", summary: "", volumes: [] } },
    { id: "resource:api", type: "resource", position: { x: 0, y: 0 }, data: { kind: "service", name: "api", kindLabel: "WEB", glyph: "web", dotVariant: "neutral", summary: "", volumes: [] } },
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

  it("skips derived edges whose resource endpoint is gone locally and never materializes secret endpoints", () => {
    const server = {
      nodes: [{ ref: { type: "secret", id: "s9" }, label: "legacy-creds" }],
      edges: [
        { kind: "depends_on", source: { type: "stack_resource", name: "ghost" }, target: { type: "stack_resource", name: "web" }, source_of_truth: "derived" },
        { kind: "env", source: { type: "secret", id: "s9" }, target: { type: "stack_resource", name: "web" }, source_of_truth: "derived" },
      ],
    } as unknown as StackTopology;
    const merged = mergeTopology(localGraph(), server);
    expect(merged.edges.map((e) => e.id)).not.toContain("depends_on:resource:ghost->resource:web");
    expect(merged.nodes.find((n) => n.id === "secret:s9")).toBeUndefined();
    expect(merged.edges.map((e) => e.id)).not.toContain("env:secret:s9->resource:web");
  });

  it("overlays server node state onto matching resource node's dotVariant", () => {
    const server = {
      nodes: [{ ref: { type: "stack_resource", name: "web" }, label: "web", state: "Ready" }],
      edges: [],
    } as unknown as StackTopology;
    const merged = mergeTopology(localGraph(), server);
    expect(merged.nodes.find((n) => n.id === "resource:web")?.data.dotVariant).toBe("ready");
  });

  it("is a no-op (referentially equal node) when the derived dotVariant is unchanged", () => {
    const local = localGraph();
    const server = {
      // "unknown" resource state maps to statusVariant("resource", "unknown") === "info",
      // not "neutral" — pick a state that actually maps back to "neutral" for a true no-op.
      nodes: [{ ref: { type: "stack_resource", name: "web" }, label: "web", state: "" }],
      edges: [],
    } as unknown as StackTopology;
    const merged = mergeTopology(local, server);
    const before = local.nodes.find((n) => n.id === "resource:web");
    const after = merged.nodes.find((n) => n.id === "resource:web");
    expect(after).toBe(before);
  });

  it("leaves a node's dotVariant untouched when it is absent from server topology", () => {
    const local = localGraph();
    const server = { nodes: [], edges: [] } as unknown as StackTopology;
    const merged = mergeTopology(local, server);
    expect(merged.nodes.find((n) => n.id === "resource:api")?.data.dotVariant).toBe("neutral");
  });
});

describe("mergeTopology release-in-flight overlay", () => {
  it("forces resource dots to pending while a release is in flight, beating a stale server Ready", () => {
    const server = {
      nodes: [{ ref: { type: "stack_resource", name: "web" }, label: "web", state: "Ready" }],
      edges: [],
    } as unknown as StackTopology;
    const merged = mergeTopology(localGraph(), server, true);
    expect(merged.nodes.find((n) => n.id === "resource:web")?.data.dotVariant).toBe("pending");
    expect(merged.nodes.find((n) => n.id === "resource:api")?.data.dotVariant).toBe("pending");
  });

  it("applies in flight even without server topology", () => {
    const merged = mergeTopology(localGraph(), null, true);
    expect(merged.nodes.find((n) => n.id === "resource:web")?.data.dotVariant).toBe("pending");
  });

  it("lets a reported error win over the in-flight pending", () => {
    const server = {
      nodes: [{ ref: { type: "stack_resource", name: "web" }, label: "web", state: "Failed" }],
      edges: [],
    } as unknown as StackTopology;
    const merged = mergeTopology(localGraph(), server, true);
    expect(merged.nodes.find((n) => n.id === "resource:web")?.data.dotVariant).toBe("error");
  });

  it("does not touch dots when no release is in flight", () => {
    const merged = mergeTopology(localGraph(), null, false);
    expect(merged.nodes.find((n) => n.id === "resource:web")?.data.dotVariant).toBe("neutral");
  });
});

describe("mergeTopology live-status overlay", () => {
  const serverReady = () =>
    ({
      nodes: [{ ref: { type: "stack_resource", name: "web" }, label: "web", state: "Ready" }],
      edges: [],
    }) as unknown as StackTopology;

  it("live state wins over the topology endpoint's node state", () => {
    const live = new Map([["resource:web", "Failed"]]);
    const merged = mergeTopology(localGraph(), serverReady(), false, live);
    expect(merged.nodes.find((n) => n.id === "resource:web")?.data.dotVariant).toBe("error");
  });

  it("a live entry with an empty state suppresses the topology state (map has priority)", () => {
    const live = new Map<string, string | undefined>([["resource:web", undefined]]);
    const merged = mergeTopology(localGraph(), serverReady(), false, live);
    expect(merged.nodes.find((n) => n.id === "resource:web")?.data.dotVariant).toBe("neutral");
  });

  it("a resource node absent from a provided live map renders neutral, not the stale topology state", () => {
    const live = new Map([["resource:api", "Failed"]]);
    const merged = mergeTopology(localGraph(), serverReady(), false, live);
    expect(merged.nodes.find((n) => n.id === "resource:web")?.data.dotVariant).toBe("neutral");
  });

  it("release-in-flight pending still overrides a live Ready", () => {
    const live = new Map([["resource:web", "Ready"]]);
    const merged = mergeTopology(localGraph(), serverReady(), true, live);
    expect(merged.nodes.find((n) => n.id === "resource:web")?.data.dotVariant).toBe("pending");
  });

  it("release-in-flight pending still overrides a resource node excluded from the live map", () => {
    const live = new Map([["resource:api", "Ready"]]);
    const merged = mergeTopology(localGraph(), serverReady(), true, live);
    expect(merged.nodes.find((n) => n.id === "resource:web")?.data.dotVariant).toBe("pending");
  });
});
