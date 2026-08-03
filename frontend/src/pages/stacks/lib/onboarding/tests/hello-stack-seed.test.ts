import { describe, expect, it } from "vitest";
import {
  buildHelloStackSeed,
  HELLO_STACK_WEB_IMAGE,
  HELLO_STACK_WORKER_IMAGE,
} from "@/pages/stacks/lib/onboarding/hello-stack-seed";

describe("buildHelloStackSeed", () => {
  const seed = buildHelloStackSeed();
  const byName = Object.fromEntries(seed.resources.map((r) => [r.name, r]));

  it("seeds web, worker, and redis", () => {
    expect(Object.keys(byName).sort()).toEqual(["redis", "web", "worker"]);
  });

  it("runs web and worker from published images", () => {
    expect(byName.web.source?.image?.ref).toBe(HELLO_STACK_WEB_IMAGE);
    expect(byName.worker.source?.image?.ref).toBe(HELLO_STACK_WORKER_IMAGE);
  });

  it("wires the addresses as references, not typed-in strings", () => {
    expect(byName.web.execution_config?.environment_variables).toContainEqual({
      from: "self",
      name: "PUBLIC_URL",
      selfOutput: "public_url",
    });
    for (const name of ["web", "worker"]) {
      expect(byName[name].execution_config?.environment_variables).toContainEqual({
        from: "resource",
        name: "REDIS_URL",
        resourceName: "redis",
        output: "url",
      });
    }
  });

  it("exposes only web to the public", () => {
    expect(byName.web.ports?.[0]).toMatchObject({ number: 3000, exposed_to_public: true });
    expect(byName.worker.ports ?? []).toHaveLength(0);
    expect(byName.redis.ports?.[0]).toMatchObject({ number: 6379, exposed_to_public: false });
  });

  it("gives redis its appendonly volume", () => {
    expect(byName.redis.source?.image?.ref).toBe("redis:7-alpine");
    expect(byName.redis.volume_mounts?.[0]).toMatchObject({
      source_volume_name: "redis-data",
      target_path: "/data",
    });
    expect(seed.volumes.map((v) => v.name)).toEqual(["redis-data"]);
  });
});
