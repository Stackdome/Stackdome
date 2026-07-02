import { describe, it, expect } from "vitest";
import { snapshotToUpdateRequest, volumesToDelete } from "../snapshot-to-update";
import type { StackReleaseSnapshot } from "@/api/releases";
import type { Stack } from "@/api/stacks";

const snap = {
  resources: [{ id: "r-1", stack_id: "st-1", revision: 2, status: {}, name: "web", image_spec: { image: "nginx:1" }, volume_mounts: [{ source_volume_name: "web-data", target_path: "/data", stack_resource_id: "r-1", source_volume_type: "pvc" }] }],
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
});

describe("volumesToDelete", () => {
  it("lists stack volumes absent from the snapshot (PUT never deletes volumes)", () => {
    const stack = {
      spec: { volumes: [{ id: "v-1", name: "web-data" }, { id: "v-2", name: "scratch" }] },
    } as unknown as Stack;
    expect(volumesToDelete(stack, snap)).toEqual([{ id: "v-2", name: "scratch" }]);
  });
});
