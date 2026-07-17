import { describe, it, expect } from "vitest";
import { formResourcesFromSpec, mapVolumeToFormData } from "@/pages/stacks/lib/spec-to-form";
import type { StackResource, Volume } from "@/pages/stacks/types";
import type { StackConnection } from "@/api/connections";

// Fixtures use the same server shapes as connection-mapping.test.ts; only the
// fields asserted below matter.

describe("formResourcesFromSpec", () => {
  it("maps resources and folds connection-backed env rows into execution_config", () => {
    const resources = [
      { id: "r1", stack_id: "s1", revision: 2, name: "web" },
    ] as unknown as StackResource[];
    const connections = [
      {
        id: "c1",
        kind: "env",
        from: { type: "secret", id: "s1" },
        to: { type: "stack_resource", name: "web" },
        mappings: [{ target: { type: "env", name: "LOCKBOX_MASTER_KEY" }, value: { output: "LOCKBOX_MASTER_KEY" } }],
      },
    ] as unknown as StackConnection[];

    const out = formResourcesFromSpec(resources, connections);

    expect(out).toHaveLength(1);
    expect(out[0].name).toBe("web");
    // read-only fields must not leak into form data
    expect((out[0] as Record<string, unknown>).id).toBeUndefined();
    expect((out[0] as Record<string, unknown>).revision).toBeUndefined();
    expect(out[0].execution_config?.environment_variables).toContainEqual(
      expect.objectContaining({ from: "secret", name: "LOCKBOX_MASTER_KEY", secretId: "s1", secretKey: "LOCKBOX_MASTER_KEY" }),
    );
  });

  it("returns [] for undefined resources", () => {
    expect(formResourcesFromSpec(undefined, undefined)).toEqual([]);
  });

  it("populates volume_mounts from volume_mount connections (server returns [] on the resource)", () => {
    const resources = [
      { id: "r1", stack_id: "s1", revision: 2, name: "web", volume_mounts: [] },
    ] as unknown as StackResource[];
    const connections = [
      {
        id: "vm-1",
        kind: "volume_mount",
        from: { type: "volume", name: "web-data" },
        to: { type: "stack_resource", name: "web" },
        config: { mount_path: "/data" },
      },
    ] as unknown as StackConnection[];

    const out = formResourcesFromSpec(resources, connections);

    expect(out[0].volume_mounts).toContainEqual(
      expect.objectContaining({ source_volume_name: "web-data", target_path: "/data" }),
    );
  });
});

describe("mapVolumeToFormData", () => {
  it("strips the read-only id before converting", () => {
    const vol = { id: "v1", name: "data", size: "1Gi" } as unknown as Volume;
    const out = mapVolumeToFormData(vol);
    expect(out.name).toBe("data");
    expect((out as Record<string, unknown>).id).toBeUndefined();
  });
});
