import { describe, it, expect } from "vitest";
import { computeSyncOps, type SyncOp } from "../ops";
import type { ServerStackState } from "../server-state";
import type { DesiredStackState } from "../desired-state";

const emptyServer = (): ServerStackState => ({
  resourcesByName: new Map(), volumeIdByName: new Map(), volumesByName: new Map(), connections: new Map(),
});
const emptyDesired = (): DesiredStackState => ({
  resources: new Map(), held: new Set(), volumes: new Map(), connections: new Map(), resourceIssues: new Map(),
});
const kinds = (ops: SyncOp[]) => ops.map((o) => o.kind);

const webResource = { name: "web", image_spec: { image: "nginx:1" } } as never;
const secretConn = (to: string) => ({
  kind: "env", from: { type: "secret", id: "s-1" }, to: { type: "stack_resource", name: to },
  mappings: [{ target: { type: "env", name: "TOKEN" }, value: { output: "token" } }],
}) as never;

describe("computeSyncOps", () => {
  it("returns no ops when server and desired match", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    expect(computeSyncOps(server, desired)).toEqual([]);
  });

  it("creates a new resource", () => {
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    expect(kinds(computeSyncOps(emptyServer(), desired))).toEqual(["createResource"]);
  });

  it("updates a changed resource by name", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    const desired = emptyDesired();
    desired.resources.set("web", { name: "web", image_spec: { image: "nginx:2" } } as never);
    const ops = computeSyncOps(server, desired);
    expect(ops).toEqual([{ kind: "updateResource", name: "web", resource: desired.resources.get("web") }]);
  });

  it("treats structurally-empty differences as equal (no spurious updates)", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", { name: "web", image_spec: { image: "nginx:1" }, depends_on: [] } as never);
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    expect(computeSyncOps(server, desired)).toEqual([]);
  });

  it("deletes a resource's connections before the resource (no backend cascade)", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set("k1", { id: "c-1", conn: secretConn("web") });
    const ops = computeSyncOps(server, emptyDesired());
    const ks = kinds(ops);
    expect(ks.indexOf("deleteConnection")).toBeLessThan(ks.indexOf("deleteResource"));
  });

  it("orders a rename as create-new before delete-old", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    const desired = emptyDesired();
    desired.resources.set("web2", { name: "web2", image_spec: { image: "nginx:1" } } as never);
    const ks = kinds(computeSyncOps(server, desired));
    expect(ks.indexOf("createResource")).toBeLessThan(ks.indexOf("deleteResource"));
  });

  it("emits createVolume before resource ops", () => {
    const desired = emptyDesired();
    desired.volumes.set("web-data", { name: "web-data" } as never);
    desired.resources.set("web", webResource);
    const ks = kinds(computeSyncOps(emptyServer(), desired));
    expect(ks.indexOf("createVolume")).toBeLessThan(ks.indexOf("createResource"));
  });

  it("never deletes or updates volumes (no thin endpoints; revert handles removal)", () => {
    const server = emptyServer();
    server.volumesByName.set("old-data", { name: "old-data" } as never);
    server.volumeIdByName.set("old-data", "v-9");
    const ops = computeSyncOps(server, emptyDesired());
    expect(ops).toEqual([]);
  });

  it("updates a connection whose mappings changed, keyed by server id", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set("env|secret:s-1|stack_resource:web|", { id: "c-1", conn: secretConn("web") });
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    const changed = secretConn("web") as { mappings: unknown[] };
    changed.mappings = [{ target: { type: "env", name: "API_TOKEN" }, value: { output: "token" } }];
    desired.connections.set("env|secret:s-1|stack_resource:web|", changed as never);
    expect(computeSyncOps(server, desired)).toEqual([
      { kind: "updateConnection", id: "c-1", identityKey: "env|secret:s-1|stack_resource:web|", conn: changed },
    ]);
  });

  it("creates a connection with a new identity and deletes the replaced one", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set("env|secret:s-1|stack_resource:web|", { id: "c-1", conn: secretConn("web") });
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    desired.connections.set("env|secret:s-2|stack_resource:web|", secretConn("web"));
    const ks = kinds(computeSyncOps(server, desired));
    expect(ks).toEqual(["deleteConnection", "createConnection"]);
  });

  it("exempts held resources and their connections from deletion", () => {
    const server = emptyServer();
    server.resourcesByName.set("api", { name: "api", image_spec: { image: "node:20" } } as never);
    server.connections.set("k", { id: "c-2", conn: secretConn("api") });
    const desired = emptyDesired();
    desired.held.add("api");
    expect(computeSyncOps(server, desired)).toEqual([]);
  });

  it("does not create new connections to a held resource", () => {
    const desired = emptyDesired();
    desired.held.add("api");
    desired.connections.set("k", secretConn("api")); // desired connection, but target is held
    expect(computeSyncOps(emptyServer(), desired)).toEqual([]);
  });

  it("skips a server connection without an id for update/delete (heals on next refetch)", () => {
    const server = emptyServer();
    server.connections.set("k", { id: undefined, conn: secretConn("web") });
    expect(computeSyncOps(server, emptyDesired())).toEqual([]);
  });

  // ── Config-aware update tests ─────────────────────────────────────────────

  const vmConn = (mountPath: string, subPath?: string) => ({
    kind: "volume_mount",
    from: { type: "volume", name: "web-data" },
    to: { type: "stack_resource", name: "web" },
    config: { mount_path: mountPath, ...(subPath ? { sub_path: subPath } : {}) },
  }) as never;

  const vmKey = "volume_mount|volume:web-data|stack_resource:web|db:";

  it("emits updateConnection when mount_path config changes", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set(vmKey, { id: "vm-1", conn: vmConn("/data") });
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    desired.connections.set(vmKey, vmConn("/mnt/data"));
    const ops = computeSyncOps(server, desired);
    expect(ops).toEqual([{ kind: "updateConnection", id: "vm-1", identityKey: vmKey, conn: vmConn("/mnt/data") }]);
  });

  it("emits updateConnection when sub_path is added to config", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set(vmKey, { id: "vm-1", conn: vmConn("/data") });
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    desired.connections.set(vmKey, vmConn("/data", "logs"));
    const ops = computeSyncOps(server, desired);
    expect(ops).toHaveLength(1);
    expect(ops[0].kind).toBe("updateConnection");
  });

  it("emits no op when volume_mount connection config is unchanged", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set(vmKey, { id: "vm-1", conn: vmConn("/data", "sub") });
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    desired.connections.set(vmKey, vmConn("/data", "sub"));
    expect(computeSyncOps(server, desired)).toEqual([]);
  });

  it("does not delete a server volume_mount connection that is present in desired", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set(vmKey, { id: "vm-1", conn: vmConn("/data") });
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    desired.connections.set(vmKey, vmConn("/data"));
    const ops = computeSyncOps(server, desired);
    expect(ops.some((o) => o.kind === "deleteConnection")).toBe(false);
  });

  it("deletes a server volume_mount connection absent from desired (mount row removed)", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set(vmKey, { id: "vm-1", conn: vmConn("/data") });
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    // no volume_mount connection in desired → should be deleted
    expect(computeSyncOps(server, desired)).toEqual([
      { kind: "deleteConnection", id: "vm-1", identityKey: vmKey },
    ]);
  });

  // Finding 1: read_only round-trip — server conn with read_only:true, desired from
  // round-tripped row → shapes match → NO op (no spurious updateConnection).
  it("emits no op when volume_mount connection with read_only:true is unchanged", () => {
    const vmConnRO = {
      kind: "volume_mount",
      from: { type: "volume", name: "web-data" },
      to: { type: "stack_resource", name: "web" },
      config: { mount_path: "/data", sub_path: "logs", read_only: true },
    } as never;
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set(vmKey, { id: "vm-1", conn: vmConnRO });
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    desired.connections.set(vmKey, vmConnRO);
    expect(computeSyncOps(server, desired)).toEqual([]);
  });

  // Finding 2: addon-config comparison — server and desired both carry { superuser: true };
  // deep-equal must hold so no spurious updateConnection is emitted.
  it("emits no op for unchanged addon connection with superuser config", () => {
    const addonConn = {
      kind: "env",
      from: { type: "addon/postgres", id: "a-1" },
      to: { type: "stack_resource", name: "web" },
      config: { superuser: true },
      mappings: [{ target: { type: "env", name: "PG_URL" }, value: { output: "url" } }],
    } as never;
    const addonKey = "env|addon/postgres:a-1|stack_resource:web|superuser";
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set(addonKey, { id: "c-1", conn: addonConn });
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    desired.connections.set(addonKey, addonConn);
    expect(computeSyncOps(server, desired)).toEqual([]);
  });

  // Finding 3: held-resource volume_mount — server has a volume_mount connection to "api";
  // desired omits "api" and marks it held → connection must NOT be deleted.
  it("does not delete a volume_mount connection to a held resource", () => {
    const apiVmConn = {
      kind: "volume_mount",
      from: { type: "volume", name: "api-data" },
      to: { type: "stack_resource", name: "api" },
      config: { mount_path: "/data" },
    } as never;
    const apiVmKey = "volume_mount|volume:api-data|stack_resource:api|db:";
    const server = emptyServer();
    server.connections.set(apiVmKey, { id: "vm-2", conn: apiVmConn });
    const desired = emptyDesired();
    desired.held.add("api");
    expect(computeSyncOps(server, desired)).toEqual([]);
  });
});
