import { describe, expect, it } from "vitest";
import { deriveGraph, EDGE_KIND, NODE_KIND } from "../graph-from-connections";
import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";

const web = (envRows: unknown[] = [], extra: Partial<FormStackResourceData> = {}): Partial<FormStackResourceData> =>
  ({
    name: "web",
    image_spec: { image: "nginx:1" },
    execution_config: { environment_variables: envRows },
    ...extra,
  }) as Partial<FormStackResourceData>;

const base = {
  resources: [] as Partial<FormStackResourceData>[],
  linkedAddonIds: new Set<string>(),
  addonNameById: new Map<string, string>(),
};

describe("deriveGraph (connection projection)", () => {
  it("projects addon env rows into an addon→resource env edge (from = producer)", () => {
    const g = deriveGraph({
      ...base,
      resources: [web([{ from: "addon", name: "DB_URL", addonId: "a1", superuser: false, database: "app", credField: "url" }])],
      linkedAddonIds: new Set(["a1"]),
      addonNameById: new Map([["a1", "tooljet-db"]]),
    });
    expect(g.edges).toEqual([
      expect.objectContaining({
        id: "env:addon:a1->resource:web",
        source: "addon:a1",
        target: "resource:web",
        type: "connection",
        data: { kind: EDGE_KIND.env, sourceOfTruth: "connection" },
      }),
    ]);
  });

  it("never renders secret references as nodes or edges", () => {
    const g = deriveGraph({
      ...base,
      resources: [web([{ from: "secret", name: "TOKEN", secretId: "s1", secretKey: "token" }])],
    });
    expect(g.nodes.find((n) => n.id === "secret:s1")).toBeUndefined();
    expect(g.edges).toEqual([]);
  });

  it("renders mounted volumes as card chips only — no volume node, no volume_mount edge", () => {
    const g = deriveGraph({
      ...base,
      resources: [web([], { volume_mounts: [{ name: "data", source_volume_name: "data", target_path: "/var/data" }] })],
      volumeNames: ["data"],
    });
    expect(g.nodes.find((n) => n.id === "volume:data")).toBeUndefined();
    expect(g.edges).toHaveLength(0);
  });

  it("renders unmounted volumes as free attachment nodes with no edges", () => {
    const g = deriveGraph({
      ...base,
      resources: [web()],
      volumeNames: ["orphan"],
    });
    const volumeNode = g.nodes.find((n) => n.id === "volume:orphan");
    expect(volumeNode?.type).toBe("attachment");
    expect(volumeNode?.data).toMatchObject({ kind: NODE_KIND.volume, name: "orphan" });
    expect(g.edges).toHaveLength(0);
  });

  it("emits depends_on edges as derived and skips unknown deps", () => {
    const g = deriveGraph({
      ...base,
      resources: [web(), { name: "worker", depends_on: ["web", "ghost"] } as Partial<FormStackResourceData>],
    });
    const dep = g.edges.find((e) => e.data.kind === EDGE_KIND.dependsOn);
    expect(dep).toMatchObject({ source: "resource:web", target: "resource:worker", data: { sourceOfTruth: "derived" } });
    expect(g.edges.filter((e) => e.data.kind === EDGE_KIND.dependsOn)).toHaveLength(1);
  });

  it("dedupes edges with the same kind and endpoints (two rows, one connection group)", () => {
    const rows = [
      { from: "resource", name: "API_URL", resourceName: "api", output: "url" },
      { from: "resource", name: "API_HOST", resourceName: "api", output: "host" },
    ];
    const g = deriveGraph({ ...base, resources: [web(rows), { name: "api" } as Partial<FormStackResourceData>] });
    expect(g.edges.filter((e) => e.data.kind === EDGE_KIND.env)).toHaveLength(1);
  });

  it("skips edges whose resource endpoint does not exist and skips in-progress rows", () => {
    const g = deriveGraph({
      ...base,
      resources: [web([
        { from: "resource", name: "X", resourceName: "missing", output: "url" },
        { from: "secret", name: "Y", secretId: "", secretKey: "" }, // in-progress: dropped by splitEnvRows
      ])],
    });
    expect(g.edges).toHaveLength(0);
    expect(g.nodes.filter((n) => n.type === "attachment")).toHaveLength(0);
  });

  it("adds addon nodes from linkedAddonIds with resolved names", () => {
    const g = deriveGraph({
      ...base,
      linkedAddonIds: new Set(["a1"]),
      addonNameById: new Map([["a1", "db"]]),
    });
    const addon = g.nodes.find((n) => n.id === "addon:a1");
    expect(addon?.data).toMatchObject({ kind: NODE_KIND.addon, name: "db" });
  });

  it("falls back to the addon id when no name is known", () => {
    const g = deriveGraph({ ...base, linkedAddonIds: new Set(["a9"]) });
    expect(g.nodes.find((n) => n.id === "addon:a9")?.data.name).toBe("a9");
  });

  it("folds volume_mounts into node volume chips, keyed by source_volume_name", () => {
    const g = deriveGraph({
      ...base,
      resources: [web([], { name: "db", volume_mounts: [{ name: "data", source_volume_name: "data", target_path: "/var/lib" }] })],
    });
    const dbNode = g.nodes.find((n) => n.id === "resource:db");
    expect(dbNode?.data).toMatchObject({ volumes: [{ name: "data", mountPath: "/var/lib" }] });
  });

  it("leaves dirtyState undefined without dirty info", () => {
    const g = deriveGraph({ ...base, resources: [web()] });
    expect(g.nodes[0].data.dirtyState).toBeUndefined();
  });

  it("marks a resource beyond the baseline count as new", () => {
    const g = deriveGraph({
      ...base,
      resources: [web(), { name: "fresh" } as Partial<FormStackResourceData>],
      dirty: { baselineResourceCount: 1, dirtyResourceIdx: new Set() },
    });
    expect(g.nodes[0].data.dirtyState).toBeUndefined();
    expect(g.nodes[1].data.dirtyState).toBe("new");
  });

  it("marks a changed existing resource as edited", () => {
    const g = deriveGraph({
      ...base,
      resources: [web()],
      dirty: { baselineResourceCount: 1, dirtyResourceIdx: new Set([0]) },
    });
    expect(g.nodes[0].data.dirtyState).toBe("edited");
  });

  it("marks a newly linked addon as new and a baseline addon as unchanged", () => {
    const g = deriveGraph({
      ...base,
      linkedAddonIds: new Set(["a1", "a2"]),
      dirty: { baselineAddonIds: new Set(["a1"]) },
    });
    const a1 = g.nodes.find((n) => n.id === "addon:a1");
    const a2 = g.nodes.find((n) => n.id === "addon:a2");
    expect(a1?.data.dirtyState).toBeUndefined();
    expect(a2?.data.dirtyState).toBe("new");
  });

  it("omits chips for mounts whose volume is missing from volumeNames", () => {
    const graph = deriveGraph({
      resources: [
        {
          name: "web",
          volume_mounts: [
            { source_volume_name: "data", source_sub_path: "", target_path: "/d" },
            { source_volume_name: "ghost", source_sub_path: "", target_path: "/g" },
          ],
        },
      ],
      linkedAddonIds: new Set(),
      addonNameById: new Map(),
      volumeNames: ["data"],
    });
    const web = graph.nodes.find((n) => n.id === "resource:web");
    expect((web?.data as { volumes: { name: string }[] }).volumes.map((v) => v.name)).toEqual(["data"]);
  });
});
