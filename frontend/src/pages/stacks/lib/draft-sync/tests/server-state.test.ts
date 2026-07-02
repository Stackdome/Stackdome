import { describe, it, expect } from "vitest";
import { connectionIdentityKey, serverStateFromStack } from "../server-state";
import type { Stack } from "@/api/stacks";

describe("connectionIdentityKey", () => {
  it("keys an addon connection by kind, endpoints and config", () => {
    const key = connectionIdentityKey({
      kind: "env",
      from: { type: "addon/postgres", id: "ad-1" },
      to: { type: "stack_resource", name: "web" },
      config: { database: "app", superuser: false },
    });
    expect(key).toBe("env|addon/postgres:ad-1|stack_resource:web|db:app");
  });

  it("distinguishes superuser from database-scoped addon connections", () => {
    const base = {
      kind: "env" as const,
      from: { type: "addon/postgres", id: "ad-1" },
      to: { type: "stack_resource", name: "web" },
    };
    expect(connectionIdentityKey({ ...base, config: { superuser: true } })).not.toBe(
      connectionIdentityKey({ ...base, config: { database: "app", superuser: false } }),
    );
  });

  it("keys secret and resource sources without config", () => {
    expect(
      connectionIdentityKey({ kind: "env", from: { type: "secret", id: "s-1" }, to: { type: "stack_resource", name: "web" } }),
    ).toBe("env|secret:s-1|stack_resource:web|");
  });

  it("ignores mappings — same identity with different mappings collides", () => {
    const a = { kind: "env" as const, from: { type: "secret", id: "s-1" }, to: { type: "stack_resource", name: "web" }, mappings: [{ target: { type: "env", name: "A" }, value: { output: "k" } }] };
    const b = { ...a, mappings: [{ target: { type: "env", name: "B" }, value: { output: "k2" } }] };
    expect(connectionIdentityKey(a)).toBe(connectionIdentityKey(b));
  });
});

describe("serverStateFromStack", () => {
  const stack = {
    id: "st-1",
    name: "demo",
    spec: {
      stack_resources: [
        { id: "r-1", stack_id: "st-1", revision: 3, name: "web", image_spec: { image: "nginx:1" }, status: { state: "Ready" }, volume_mounts: [{ source_volume_name: "web-data", target_path: "/data", stack_resource_id: "r-1", source_volume_type: "pvc" }] },
      ],
      volumes: [{ id: "v-1", name: "web-data", spec: { size: "1Gi", access_mode: "ReadWriteOnce" }, status: {} }],
      connections: [{ id: "c-1", kind: "env", from: { type: "secret", id: "s-1" }, to: { type: "stack_resource", name: "web" }, mappings: [{ target: { type: "env", name: "TOKEN" }, value: { output: "token" } }] }],
    },
  } as unknown as Stack;

  it("indexes resources by name with read-only fields stripped", () => {
    const s = serverStateFromStack(stack);
    const web = s.resourcesByName.get("web")!;
    expect(web).toBeDefined();
    expect((web as Record<string, unknown>).id).toBeUndefined();
    expect((web as Record<string, unknown>).status).toBeUndefined();
    // volume_mounts are stripped — they are now represented as volume_mount
    // connections, not as a field on the resource.
    expect(web.volume_mounts).toBeUndefined();
  });

  it("strips volume_mounts from server resource to prevent phantom updateResource diffs", () => {
    // Server returns volume_mounts:[] on the resource while the desired state
    // also strips them. Both sides become undefined → deepEqual returns true → no op.
    const s = serverStateFromStack(stack);
    expect(s.resourcesByName.get("web")?.volume_mounts).toBeUndefined();
  });

  it("indexes volumes by name and maps ids", () => {
    const s = serverStateFromStack(stack);
    expect(s.volumeIdByName.get("web-data")).toBe("v-1");
    expect((s.volumesByName.get("web-data") as Record<string, unknown>).id).toBeUndefined();
  });

  it("indexes connections by identity key, retaining the server id", () => {
    const s = serverStateFromStack(stack);
    const entry = s.connections.get("env|secret:s-1|stack_resource:web|")!;
    expect(entry.id).toBe("c-1");
    expect(entry.conn.mappings).toHaveLength(1);
  });
});
