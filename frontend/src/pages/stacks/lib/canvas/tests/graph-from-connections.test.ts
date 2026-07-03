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
  linkedAddonIds: new Set<string>(),
  addonNameById: new Map<string, string>(),
  secretNameById: new Map<string, string>(),
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

  it("creates a secret attachment node on demand and labels it from secretNameById", () => {
    const g = deriveGraph({
      ...base,
      resources: [web([{ from: "secret", name: "TOKEN", secretId: "s1", secretKey: "token" }])],
      secretNameById: new Map([["s1", "api-creds"]]),
    });
    const secretNode = g.nodes.find((n) => n.id === "secret:s1");
    expect(secretNode?.type).toBe("attachment");
    expect(secretNode?.data).toMatchObject({ kind: NODE_KIND.secret, name: "api-creds" });
    expect(g.edges.map((e) => e.id)).toContain("env:secret:s1->resource:web");
  });

  it("projects volume mounts into a volume node plus volume_mount edge", () => {
    const g = deriveGraph({
      ...base,
      resources: [web([], { volume_mounts: [{ name: "data", source_volume_name: "data", target_path: "/var/data" }] })],
    });
    expect(g.nodes.find((n) => n.id === "volume:data")?.data).toMatchObject({ kind: NODE_KIND.volume, name: "data" });
    expect(g.edges.map((e) => e.id)).toContain("volume_mount:volume:data->resource:web");
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
});
