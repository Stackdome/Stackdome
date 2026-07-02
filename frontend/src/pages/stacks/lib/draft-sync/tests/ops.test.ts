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
});
