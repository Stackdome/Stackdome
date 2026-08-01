import { describe, it, expect } from "vitest";
import type { Stack } from "@/api/stacks";
import { computeSyncOps, type SyncOp } from "../ops";
import { serverConnectionIndex, type ServerConnectionIndex } from "../server-state";
import { canonicalFromStack } from "@/pages/stacks/lib/stack-model/from-api";
import { canonicalFromDraft, type CanonicalDraft } from "@/pages/stacks/lib/stack-model/from-form";
import { formResourcesFromSpec, mapVolumeToFormData } from "@/pages/stacks/lib/spec-to-form";
import { EMPTY_CANONICAL_STACK } from "@/pages/stacks/lib/stack-model/canonical";

const kinds = (ops: SyncOp[]) => ops.map((o) => o.kind);

const web = {
  id: "r-web",
  name: "web",
  source: { image: { ref: "nginx:1" } },
  ports: [],
  volume_mounts: [],
  depends_on: [],
  execution_config: { command: [], args: [], environment_variables: [] },
};

const secretConn = (to: string, secretId = "s-1", envName = "TOKEN") => ({
  id: `c-${secretId}-${to}`,
  kind: "env",
  from: { type: "secret", id: secretId },
  to: { type: "stack_resource", name: to },
  mappings: [{ target: { type: "env", name: envName }, value: { output: "token" } }],
});

const mountConn = (mountPath: string, extra: Record<string, unknown> = {}, to = "web") => ({
  id: "vm-1",
  kind: "volume_mount",
  from: { type: "volume", name: "web-data" },
  to: { type: "stack_resource", name: to },
  config: { mount_path: mountPath, ...extra },
});

const dataVolume = { id: "v-1", name: "web-data", spec: { size: "1Gi", access_mode: "ReadWriteOnce" } };

function specOf(resources: unknown[], volumes: unknown[] = [], connections: unknown[] = []): Stack {
  return { id: "s1", name: "demo", spec: { stack_resources: resources, volumes, connections } } as unknown as Stack;
}

/** The server side: what it holds, plus the connection ids the write path needs. */
function server(resources: unknown[], volumes: unknown[] = [], connections: unknown[] = []) {
  const stack = specOf(resources, volumes, connections);
  return { stack: canonicalFromStack(stack), connections: serverConnectionIndex(stack) };
}

/** The draft side, built the way the editor builds it: server shapes → form → canonical. */
function draft(resources: unknown[], volumes: unknown[] = [], connections: unknown[] = []): CanonicalDraft {
  const stack = specOf(resources, volumes, connections);
  return canonicalFromDraft({
    resources: formResourcesFromSpec(stack.spec?.stack_resources, stack.spec?.connections),
    volumes: (stack.spec?.volumes ?? []).map(mapVolumeToFormData),
  });
}

const emptyDraft = (): CanonicalDraft => ({
  ...EMPTY_CANONICAL_STACK,
  held: new Set(),
  issues: new Map(),
  indexByName: new Map(),
});

const heldDraft = (names: string[]): CanonicalDraft => ({ ...emptyDraft(), held: new Set(names) });

const noConnections: ServerConnectionIndex = new Map();

describe("computeSyncOps", () => {
  it("returns no ops when the server already holds the draft", () => {
    const s = server([web]);
    expect(computeSyncOps(s.stack, draft([web]), s.connections)).toEqual([]);
  });

  it("creates a new resource", () => {
    expect(kinds(computeSyncOps(EMPTY_CANONICAL_STACK, draft([web]), noConnections))).toEqual([
      "createResource",
    ]);
  });

  it("updates a changed resource by name", () => {
    const s = server([web]);
    const ops = computeSyncOps(s.stack, draft([{ ...web, source: { image: { ref: "nginx:2" } } }]), s.connections);
    expect(ops).toHaveLength(1);
    expect(ops[0]).toMatchObject({ kind: "updateResource", name: "web" });
  });

  it("treats structurally-empty differences as equal", () => {
    const s = server([{ ...web, depends_on: [], labels: {} }]);
    const { depends_on, ...withoutEmpties } = web;
    void depends_on;
    expect(computeSyncOps(s.stack, draft([withoutEmpties]), s.connections)).toEqual([]);
  });

  it("deletes a resource's connections before the resource (no backend cascade)", () => {
    const s = server([web], [], [secretConn("web")]);
    const ks = kinds(computeSyncOps(s.stack, emptyDraft(), s.connections));
    expect(ks.indexOf("deleteConnection")).toBeLessThan(ks.indexOf("deleteResource"));
  });

  it("orders a rename as create-new before delete-old", () => {
    const s = server([web]);
    const ks = kinds(computeSyncOps(s.stack, draft([{ ...web, name: "web2" }]), s.connections));
    expect(ks.indexOf("createResource")).toBeLessThan(ks.indexOf("deleteResource"));
  });

  it("emits createVolume before resource ops", () => {
    const ks = kinds(computeSyncOps(EMPTY_CANONICAL_STACK, draft([web], [dataVolume]), noConnections));
    expect(ks.indexOf("createVolume")).toBeLessThan(ks.indexOf("createResource"));
  });

  it("never deletes or updates volumes (no thin endpoints; revert handles removal)", () => {
    const s = server([], [dataVolume]);
    expect(computeSyncOps(s.stack, emptyDraft(), s.connections)).toEqual([]);
  });

  it("updates a connection whose mappings changed, keyed by server id", () => {
    const s = server([web], [], [secretConn("web")]);
    const ops = computeSyncOps(s.stack, draft([web], [], [secretConn("web", "s-1", "API_TOKEN")]), s.connections);
    expect(ops).toHaveLength(1);
    expect(ops[0]).toMatchObject({ kind: "updateConnection", id: "c-s-1-web" });
  });

  it("creates a connection with a new identity and deletes the replaced one", () => {
    const s = server([web], [], [secretConn("web", "s-1")]);
    const ks = kinds(computeSyncOps(s.stack, draft([web], [], [secretConn("web", "s-2")]), s.connections));
    expect(ks).toEqual(["deleteConnection", "createConnection"]);
  });

  it("exempts held resources and their connections from deletion", () => {
    const s = server([{ ...web, name: "api" }], [], [secretConn("api")]);
    expect(computeSyncOps(s.stack, heldDraft(["api"]), s.connections)).toEqual([]);
  });

  it("does not create new connections to a held resource", () => {
    const d = draft([{ ...web, name: "api" }], [], [secretConn("api")]);
    const withHeld: CanonicalDraft = { ...d, resources: [], held: new Set(["api"]) };
    expect(computeSyncOps(EMPTY_CANONICAL_STACK, withHeld, noConnections)).toEqual([]);
  });

  it("skips a server connection without an id (heals on the next refetch)", () => {
    const { id, ...idless } = secretConn("web");
    void id;
    const s = server([web], [], [idless]);
    expect(computeSyncOps(s.stack, draft([web]), s.connections)).toEqual([]);
  });

  it("updates a mount whose path changed", () => {
    const s = server([web], [dataVolume], [mountConn("/data")]);
    const ops = computeSyncOps(s.stack, draft([web], [dataVolume], [mountConn("/mnt/data")]), s.connections);
    expect(ops).toHaveLength(1);
    expect(ops[0]).toMatchObject({ kind: "updateConnection", id: "vm-1" });
  });

  it("updates a mount that gained a sub path", () => {
    const s = server([web], [dataVolume], [mountConn("/data")]);
    const ops = computeSyncOps(
      s.stack,
      draft([web], [dataVolume], [mountConn("/data", { sub_path: "logs" })]),
      s.connections,
    );
    expect(ops).toHaveLength(1);
    expect(ops[0].kind).toBe("updateConnection");
  });

  it("emits nothing for an unchanged mount, read_only and all", () => {
    const conn = mountConn("/data", { sub_path: "logs", read_only: true });
    const s = server([web], [dataVolume], [conn]);
    expect(computeSyncOps(s.stack, draft([web], [dataVolume], [conn]), s.connections)).toEqual([]);
  });

  it("deletes a mount the draft no longer has", () => {
    const s = server([web], [dataVolume], [mountConn("/data")]);
    const ops = computeSyncOps(s.stack, draft([web], [dataVolume]), s.connections);
    expect(ops).toHaveLength(1);
    expect(ops[0]).toMatchObject({ kind: "deleteConnection", id: "vm-1" });
  });

  it("emits nothing for an unchanged addon connection with superuser config", () => {
    const addonConn = {
      id: "c-1",
      kind: "env",
      from: { type: "addon/postgres", id: "a-1" },
      to: { type: "stack_resource", name: "web" },
      config: { superuser: true },
      mappings: [{ target: { type: "env", name: "PG_URL" }, value: { output: "url" } }],
    };
    const s = server([web], [], [addonConn]);
    expect(computeSyncOps(s.stack, draft([web], [], [addonConn]), s.connections)).toEqual([]);
  });

  it("does not delete a mount belonging to a held resource", () => {
    const s = server([{ ...web, name: "api" }], [dataVolume], [mountConn("/data", {}, "api")]);
    expect(computeSyncOps(s.stack, heldDraft(["api"]), s.connections)).toEqual([]);
  });
});
