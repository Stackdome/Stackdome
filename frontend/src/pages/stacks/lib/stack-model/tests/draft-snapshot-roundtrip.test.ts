import { describe, it, expect } from "vitest";
import { canonicalFromSnapshot } from "../from-api";
import { resourceFingerprint } from "../diff";
import { resourceToApi, connectionsOf } from "../to-api";

/** A live release snapshot resource, verbatim. */
const SNAPSHOT = {
  captured_at: "2026-08-02T21:02:14Z",
  stack: { name: "e2e-diff-check" },
  resources: [
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
  volumes: [{ id: "vol-redis", name: "redis-data", spec: { size: "1Gi" } }],
  connections: [
    {
      config: { mount_path: "/data" },
      from: { name: "redis-data", type: "volume" },
      id: "conn-mount-redis",
      kind: "volume_mount",
      to: { name: "redis", type: "stack_resource" },
    },
  ],
} as never;

/**
 * The staged diff sends the draft through canonical → API shape → canonical
 * again (draft-snapshot.ts), while the baseline arrives as API shape once. Any
 * loss in that extra leg makes an untouched resource fingerprint differently
 * from its own deployed self.
 */
describe("canonical → API → canonical round trip", () => {
  it("preserves the resource fingerprint", () => {
    const once = canonicalFromSnapshot(SNAPSHOT);
    const roundTripped = canonicalFromSnapshot({
      resources: once.resources.map(resourceToApi),
      volumes: once.volumes,
      connections: connectionsOf(once),
    } as never);
    expect(resourceFingerprint(roundTripped.resources[0])).toBe(
      resourceFingerprint(once.resources[0]),
    );
  });
});
