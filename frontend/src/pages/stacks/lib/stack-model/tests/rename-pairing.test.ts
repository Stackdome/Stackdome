import { describe, it, expect } from "vitest";
import { canonicalFromSnapshot } from "../from-api";
import { diffStacks, resourceFingerprint } from "../diff";

/** Captured from a live release snapshot: two image resources, one volume
 *  mount connection per resource, one resource-output env row on `web`. */
const SNAPSHOT = {
  captured_at: "2026-08-02T21:02:14Z",
  stack: { name: "e2e-diff-check" },
  resources: [
    {
      execution_config: { environment_variables: [] },
      id: "res-web",
      name: "web",
      outputs: [{ name: "host", sensitive: false, type: "string" }],
      ports: [{ exposed_to_public: true, name: "http-80", number: 80, protocol: "http", subdomain_prefix: "" }],
      schedule: "",
      source: { image: { ref: "nginx:1.27" } },
      stack_id: "stack-1",
      volume_mounts: [],
      workload_type: "Service",
    },
    {
      execution_config: { environment_variables: [] },
      id: "res-redis",
      name: "redis",
      outputs: [{ name: "host", sensitive: false, type: "string" }],
      ports: [{ exposed_to_public: false, name: "tcp-6379", number: 6379, protocol: "tcp", subdomain_prefix: "" }],
      schedule: "",
      source: { image: { ref: "redis:7" } },
      stack_id: "stack-1",
      volume_mounts: [],
      workload_type: "Service",
    },
  ],
  volumes: [
    { id: "vol-data", name: "data", spec: { size: "1Gi" } },
    { id: "vol-redis", name: "redis-data", spec: { size: "1Gi" } },
  ],
  connections: [
    {
      from: { name: "redis", type: "stack_resource" },
      id: "conn-env",
      kind: "env",
      mappings: [{ target: { name: "REDIS_HOST", type: "env" }, value: { output: "host" } }],
      to: { name: "web", type: "stack_resource" },
    },
    {
      config: { mount_path: "/usr/share/nginx/html" },
      from: { name: "data", type: "volume" },
      id: "conn-mount-web",
      kind: "volume_mount",
      to: { name: "web", type: "stack_resource" },
    },
    {
      config: { mount_path: "/data" },
      from: { name: "redis-data", type: "volume" },
      id: "conn-mount-redis",
      kind: "volume_mount",
      to: { name: "redis", type: "stack_resource" },
    },
  ],
} as never;

/** The same snapshot after renaming `redis` to `cache`. Connections address
 *  resources by name, so both the env source and the mount target follow. */
function renamed() {
  const s = JSON.parse(JSON.stringify(SNAPSHOT));
  s.resources[1].name = "cache";
  s.connections[0].from.name = "cache";
  s.connections[2].to.name = "cache";
  return s as never;
}

/** What the editor actually produces: the mount follows the rename but the env
 *  reference on `web` is left pointing at a resource that no longer exists. */
function renamedLeavingEnvDangling() {
  const s = JSON.parse(JSON.stringify(SNAPSHOT));
  s.resources[1].name = "cache";
  s.connections[2].to.name = "cache";
  return s as never;
}

describe("renaming a deployed resource", () => {
  it("fingerprints the renamed resource identically to its old self", () => {
    const before = canonicalFromSnapshot(SNAPSHOT);
    const after = canonicalFromSnapshot(renamed());
    const redis = before.resources.find((r) => r.name === "redis")!;
    const cache = after.resources.find((r) => r.name === "cache")!;
    expect(resourceFingerprint(cache)).toBe(resourceFingerprint(redis));
  });

  it("reports one renamed resource, not a removal plus an addition", () => {
    const diff = diffStacks(canonicalFromSnapshot(SNAPSHOT), canonicalFromSnapshot(renamed()), {
      baselineIsRelease: true,
    });
    // `web` is modified alongside it: rewriting the reference changes the value
    // of its REDIS_HOST row from `redis · host` to `cache · host`.
    expect(diff.resources.map((r) => `${r.name}:${r.change}`)).toEqual([
      "web:modified",
      "cache:renamed",
    ]);
  });

  it("still pairs when the rename leaves an env reference dangling", () => {
    const diff = diffStacks(
      canonicalFromSnapshot(SNAPSHOT),
      canonicalFromSnapshot(renamedLeavingEnvDangling()),
      { baselineIsRelease: true },
    );
    expect(diff.resources.map((r) => `${r.name}:${r.change}`)).toContain("cache:renamed");
  });
});
