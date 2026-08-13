import { describe, it, expect } from "vitest";
import { deleteResourceAndReferences, findResourceDependents } from "../delete-references";
import type { ResourceArr, VolumeArr } from "../stack-diff";

/**
 * Siblings address a resource by name, and connections are derived from those
 * names at serialize time. A delete that leaves the references behind produces
 * a connection to a resource that does not exist, which the API only rejects
 * once it renders the release.
 */

const web = (env: unknown[] = [], depends: string[] = [], mounts: string[] = []): ResourceArr[number] =>
  ({
    name: "web",
    sourceType: "image",
    source: { image: { ref: "nginx:1.27" } },
    depends_on: depends,
    execution_config: { environment_variables: env },
    volume_mounts: mounts.map((v) => ({ source_volume_name: v, target_path: "/data" })),
  }) as never;

const redis = (mounts: string[] = []): ResourceArr[number] =>
  ({
    name: "redis",
    sourceType: "image",
    source: { image: { ref: "redis:7" } },
    volume_mounts: mounts.map((v) => ({ source_volume_name: v, target_path: "/data" })),
  }) as never;

const resourceRow = (resourceName: string) => ({
  from: "resource",
  name: "REDIS_HOST",
  resourceName,
  output: "host",
});

const templateRow = (resourceName: string) => ({
  from: "resourceTemplate",
  name: "REDIS_URL",
  resourceName,
  template: "redis://{h}:6379",
  values: { h: "host" },
});

const literalRow = (value: string) => ({ from: "stack", name: "NOTE", value });

const envOf = (r: ResourceArr[number]) =>
  (r?.execution_config?.environment_variables ?? []) as Array<{ name: string }>;

const volumes = (...names: string[]): VolumeArr => names.map((name) => ({ name }) as never);

describe("deleteResourceAndReferences", () => {
  it("removes the resource itself", () => {
    const next = deleteResourceAndReferences([web(), redis()], "redis");
    expect(next.map((r) => r?.name)).toEqual(["web"]);
  });

  it("drops a depends_on entry naming the deleted resource", () => {
    const next = deleteResourceAndReferences([web([], ["redis", "db"]), redis()], "redis");
    expect(next[0]?.depends_on).toEqual(["db"]);
  });

  it("drops a whole-value env row", () => {
    const next = deleteResourceAndReferences([web([resourceRow("redis")]), redis()], "redis");
    expect(envOf(next[0])).toEqual([]);
  });

  it("drops a templated env row", () => {
    const next = deleteResourceAndReferences([web([templateRow("redis")]), redis()], "redis");
    expect(envOf(next[0])).toEqual([]);
  });

  it("leaves rows naming a different resource alone", () => {
    const next = deleteResourceAndReferences([web([resourceRow("cache")]), redis()], "redis");
    expect(envOf(next[0])).toHaveLength(1);
  });

  it("leaves a literal value mentioning the deleted name untouched", () => {
    const row = literalRow("http://${redis.host}:6379");
    const next = deleteResourceAndReferences([web([row]), redis()], "redis");
    expect(envOf(next[0])).toEqual([row]);
  });

  it("returns untouched resources by identity so the diff keeps pairing them", () => {
    const untouched = web();
    const next = deleteResourceAndReferences([untouched, redis()], "redis");
    expect(next[0]).toBe(untouched);
  });
});

describe("findResourceDependents", () => {
  it("reports the resource holding a depends_on entry", () => {
    const found = findResourceDependents([web([], ["redis"]), redis()], [], "redis");
    expect(found.dependsOn).toEqual(["web"]);
  });

  it("reports structured env refs by their env key", () => {
    const found = findResourceDependents([web([resourceRow("redis")]), redis()], [], "redis");
    expect(found.envRefs).toEqual([{ resource: "web", envName: "REDIS_HOST" }]);
  });

  it("separates a literal mention from a structured ref", () => {
    const resources = [web([resourceRow("redis"), literalRow("http://${redis.host}:6379")]), redis()];
    const found = findResourceDependents(resources, [], "redis");
    expect(found.envRefs).toHaveLength(1);
    expect(found.literalRefs).toEqual([{ resource: "web", envName: "NOTE" }]);
  });

  it("does not count a literal that names an unrelated resource", () => {
    const found = findResourceDependents([web([literalRow("http://${cache.host}")]), redis()], [], "redis");
    expect(found.literalRefs).toEqual([]);
  });

  it("lists a volume only the deleted resource mounted", () => {
    const found = findResourceDependents([web(), redis(["cache-data"])], volumes("cache-data"), "redis");
    expect(found.orphanedVolumes).toEqual(["cache-data"]);
  });

  it("omits a volume a survivor still mounts", () => {
    const found = findResourceDependents(
      [web([], [], ["shared"]), redis(["shared"])],
      volumes("shared"),
      "redis",
    );
    expect(found.orphanedVolumes).toEqual([]);
  });

  /** Only volumes this delete detaches. One that was already sitting unmounted
   *  is not the delete's doing, and saying so makes the dialog lie. */
  it("omits a volume that was already mounted by nothing", () => {
    const found = findResourceDependents([web(), redis()], volumes("stray"), "redis");
    expect(found.orphanedVolumes).toEqual([]);
  });

  it("reports nothing when no one references the resource", () => {
    const found = findResourceDependents([web(), redis()], [], "redis");
    expect(found.dependsOn).toEqual([]);
    expect(found.envRefs).toEqual([]);
    expect(found.literalRefs).toEqual([]);
  });
});
