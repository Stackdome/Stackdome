import { describe, expect, it } from "vitest";
import {
  buildHelloStackSeed,
  HELLO_STACK_REPO_URL,
} from "../hello-stack-seed";

describe("buildHelloStackSeed", () => {
  const seed = buildHelloStackSeed();
  const byName = Object.fromEntries(seed.resources.map((r) => [r.name, r]));

  it("seeds web, worker, and redis", () => {
    expect(Object.keys(byName).sort()).toEqual(["redis", "web", "worker"]);
  });

  it("builds web and worker from the demo repo subdirectories", () => {
    expect(byName.web.source?.git).toMatchObject({
      repo_url: HELLO_STACK_REPO_URL,
      build_context: "hello-stack/web",
    });
    expect(byName.worker.source?.git).toMatchObject({
      repo_url: HELLO_STACK_REPO_URL,
      build_context: "hello-stack/worker",
    });
  });

  it("gives web its own public URL, resolved at deploy", () => {
    expect(byName.web.execution_config?.environment_variables).toContainEqual({
      from: "resource",
      name: "PUBLIC_URL",
      resourceName: "web",
      output: "public_url",
    });
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
