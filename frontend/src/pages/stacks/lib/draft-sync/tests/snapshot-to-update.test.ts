import { describe, it, expect } from "vitest";
import { snapshotToUpdateRequest, stackToUpdateRequest, volumesToDelete } from "../snapshot-to-update";
import type { StackReleaseSnapshot } from "@/api/releases";
import type { Stack } from "@/api/stacks";

const snap = {
  resources: [{ id: "r-1", stack_id: "st-1", revision: 2, status: {}, name: "web", source: { image: { ref: "nginx:1" } }, volume_mounts: [{ source_volume_name: "web-data", target_path: "/data", stack_resource_id: "r-1", source_volume_type: "pvc" }] }],
  volumes: [{ id: "v-1", status: {}, name: "web-data", spec: { size: "1Gi" } }],
  connections: [{ id: "c-1", kind: "env", from: { type: "secret", id: "s-1" }, to: { type: "stack_resource", name: "web" }, mappings: [] }],
} as unknown as StackReleaseSnapshot;

describe("snapshotToUpdateRequest", () => {
  it("strips read-only fields from resources and volumes, keeps connection ids", () => {
    const req = snapshotToUpdateRequest(snap, { name: "demo", labels: [] });
    expect(req.name).toBe("demo");
    const res = req.spec.stack_resources[0] as Record<string, unknown>;
    expect(res.id).toBeUndefined();
    expect(res.revision).toBeUndefined();
    expect((req.spec.volumes?.[0] as Record<string, unknown>).id).toBeUndefined();
    // ids retained so the PUT's replace-all upserts instead of delete+create
    expect(req.spec.connections?.[0].id).toBe("c-1");
  });

  it("omits connections when the snapshot has none", () => {
    const req = snapshotToUpdateRequest({ ...snap, connections: [] } as StackReleaseSnapshot, { name: "demo" });
    expect(req.spec.connections).toBeUndefined();
  });

  it("emits explicit empty arrays for execution_config fields when resource has no execution_config", () => {
    const snapNoExec = {
      resources: [{ id: "r-2", stack_id: "st-1", revision: 1, status: {}, name: "api", source: { image: { ref: "myapp:1" } } }],
      volumes: [],
      connections: [],
    } as unknown as StackReleaseSnapshot;
    const req = snapshotToUpdateRequest(snapNoExec, { name: "demo" });
    const res = req.spec.stack_resources[0] as Record<string, unknown>;
    const exec = res.execution_config as Record<string, unknown>;
    expect(exec).toBeDefined();
    expect(exec.environment_variables).toEqual([]);
    expect(exec.command).toEqual([]);
    expect(exec.args).toEqual([]);
  });

  it("emits explicit empty arrays for volume_mounts and depends_on when resource has none", () => {
    const snapNoDeps = {
      resources: [{ id: "r-3", stack_id: "st-1", revision: 1, status: {}, name: "worker", source: { image: { ref: "worker:1" } } }],
      volumes: [],
      connections: [],
    } as unknown as StackReleaseSnapshot;
    const req = snapshotToUpdateRequest(snapNoDeps, { name: "demo" });
    const res = req.spec.stack_resources[0] as Record<string, unknown>;
    expect(res.volume_mounts).toEqual([]);
    expect(res.depends_on).toEqual([]);
  });

  it("emits explicit empty arrays for ports, labels, and annotations when resource has none", () => {
    // A snapshot resource with no ports/labels/annotations (e.g. reverted after
    // a canvas edit that added a port). The PUT body must carry explicit []
    // so gorm's Updates() does not skip the column and leave the draft value.
    const snapNoPorts = {
      resources: [{ id: "r-4", stack_id: "st-1", revision: 1, status: {}, name: "api", source: { image: { ref: "api:1" } } }],
      volumes: [],
      connections: [],
    } as unknown as StackReleaseSnapshot;
    const req = snapshotToUpdateRequest(snapNoPorts, { name: "demo" });
    const res = req.spec.stack_resources[0] as Record<string, unknown>;
    expect(res.ports).toEqual([]);
    expect(res.labels).toEqual([]);
    expect(res.annotations).toEqual([]);
  });
});

describe("volumesToDelete", () => {
  it("lists stack volumes absent from the snapshot (PUT never deletes volumes)", () => {
    const stack = {
      spec: { volumes: [{ id: "v-1", name: "web-data" }, { id: "v-2", name: "scratch" }] },
    } as unknown as Stack;
    expect(volumesToDelete(stack, snap)).toEqual([{ id: "v-2", name: "scratch" }]);
  });
});

describe("stackToUpdateRequest", () => {
  const stack = {
    id: "s1",
    name: "acme",
    labels: [{ key: "user", value: "old" }],
    spec: {
      stack_resources: [
        {
          id: "r1",
          name: "web",
          revision: "3",
          outputs: [{ name: "url" }],
          ports: [{ number: 80, exposed_to_public: true }],
        },
      ],
      volumes: [{ id: "v1", project_id: "t1", status: { phase: "Ready" }, name: "data", spec: { size: "1Gi" } }],
      connections: [{ id: "c1", kind: "volume_mount" }],
    },
  } as unknown as Stack;

  it("carries the full spec INCLUDING connections (PUT is replace-all)", () => {
    const req = stackToUpdateRequest(stack, [{ key: "user", value: "prod" }]);
    expect(req.spec?.connections).toEqual([{ id: "c1", kind: "volume_mount" }]);
    expect(req.spec?.stack_resources).toHaveLength(1);
    expect(req.spec?.volumes).toHaveLength(1);
  });

  it("applies the new labels and keeps the name", () => {
    const req = stackToUpdateRequest(stack, [{ key: "user", value: "prod" }]);
    expect(req.name).toBe("acme");
    expect(req.labels).toEqual([{ key: "user", value: "prod" }]);
  });

  it("strips server-only resource fields (id/revision/status/outputs)", () => {
    const req = stackToUpdateRequest(stack, []);
    const r = req.spec!.stack_resources![0] as Record<string, unknown>;
    expect(r.id).toBeUndefined();
    expect(r.revision).toBeUndefined();
    expect(r.status).toBeUndefined();
    expect(r.outputs).toBeUndefined();
  });

  it("strips readOnly volume fields (id/project_id/status) — schema validation rejects them", () => {
    const req = stackToUpdateRequest(stack, []);
    const v = req.spec!.volumes![0] as Record<string, unknown>;
    expect(v.id).toBeUndefined();
    expect(v.project_id).toBeUndefined();
    expect(v.status).toBeUndefined();
    expect(v.name).toBe("data");
  });
});
