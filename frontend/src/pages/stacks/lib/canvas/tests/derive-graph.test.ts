import { describe, it, expect } from "vitest";
import { deriveGraph, NODE_KIND } from "../derive-graph";
import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";

const r = (partial: Partial<FormStackResourceData>) => partial as Partial<FormStackResourceData>;
const base = {
  resources: [] as Partial<FormStackResourceData>[],
  linkedAddonIds: new Set<string>(),
  addonNameById: new Map<string, string>(),
};

describe("deriveGraph", () => {
  it("makes one service node per resource, preserving order", () => {
    const g = deriveGraph({ ...base, resources: [r({ name: "web" }), r({ name: "api" })] });
    expect(g.nodes.map((n) => n.id)).toEqual(["resource:web", "resource:api"]);
    expect(g.nodes[0].data.kind).toBe(NODE_KIND.service);
    expect(g.nodes[0].data.resourceIdx).toBe(0);
    expect(g.nodes[1].data.resourceIdx).toBe(1);
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

  it("derives an addon edge from an env-var reference", () => {
    const g = deriveGraph({
      ...base,
      resources: [
        r({ name: "web", execution_config: { environment_variables: [{ from: "addon", name: "DB", addonId: "a1" }] } }),
      ],
      linkedAddonIds: new Set(["a1"]),
      addonNameById: new Map([["a1", "db"]]),
    });
    expect(g.edges).toContainEqual(expect.objectContaining({ source: "resource:web", target: "addon:a1" }));
  });

  it("derives a resource→resource edge and dedupes repeats", () => {
    const g = deriveGraph({
      ...base,
      resources: [
        r({
          name: "web",
          execution_config: {
            environment_variables: [
              { from: "resource", name: "U1", resourceName: "api", output: "URL" },
              { from: "resource", name: "U2", resourceName: "api", output: "URL" },
            ],
          },
        }),
        r({ name: "api" }),
      ],
    });
    expect(g.edges.filter((e) => e.source === "resource:web" && e.target === "resource:api")).toHaveLength(1);
  });

  it("ignores edges to targets that do not exist as nodes", () => {
    const g = deriveGraph({
      ...base,
      resources: [r({ name: "web", execution_config: { environment_variables: [{ from: "addon", name: "DB", addonId: "ghost" }] } })],
    });
    expect(g.edges).toHaveLength(0);
  });

  it("folds volume_mounts into node volume chips", () => {
    const g = deriveGraph({
      ...base,
      resources: [r({ name: "db", volume_mounts: [{ name: "data", mount_path: "/var/lib" }] })],
    });
    expect(g.nodes[0].data.volumes).toEqual([{ name: "data", mountPath: "/var/lib" }]);
  });

  it("gives every node a zero position (layout assigns real ones later)", () => {
    const g = deriveGraph({ ...base, resources: [r({ name: "web" })] });
    expect(g.nodes[0].position).toEqual({ x: 0, y: 0 });
  });

  it("leaves dirtyState undefined without dirty info", () => {
    const g = deriveGraph({ ...base, resources: [r({ name: "web" })] });
    expect(g.nodes[0].data.dirtyState).toBeUndefined();
  });

  it("marks a resource beyond the baseline count as new", () => {
    const g = deriveGraph({
      ...base,
      resources: [r({ name: "web" }), r({ name: "fresh" })],
      dirty: { baselineResourceCount: 1, dirtyResourceIdx: new Set() },
    });
    expect(g.nodes[0].data.dirtyState).toBeUndefined();
    expect(g.nodes[1].data.dirtyState).toBe("new");
  });

  it("marks a changed existing resource as edited", () => {
    const g = deriveGraph({
      ...base,
      resources: [r({ name: "web" })],
      dirty: { baselineResourceCount: 1, dirtyResourceIdx: new Set([0]) },
    });
    expect(g.nodes[0].data.dirtyState).toBe("edited");
  });

  it("marks a pending-detach addon as removed and a newly linked addon as new", () => {
    const g = deriveGraph({
      ...base,
      linkedAddonIds: new Set(["a1", "a2"]),
      dirty: { baselineAddonIds: new Set(["a1"]), pendingDetach: new Set(["a1"]) },
    });
    const a1 = g.nodes.find((n) => n.id === "addon:a1");
    const a2 = g.nodes.find((n) => n.id === "addon:a2");
    expect(a1?.data.dirtyState).toBe("removed");
    expect(a2?.data.dirtyState).toBe("new");
  });
});
