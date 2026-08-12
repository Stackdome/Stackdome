import { describe, it, expect } from "vitest";
import { renameResourceReferences, renameResourceReferencesByMap } from "../rename-references";
import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";

/**
 * Connections address resources by name, so a rename that does not carry its
 * references leaves them pointing at a resource that no longer exists. The
 * backend only catches that at render time, with the deploy already failing.
 */

const web = (env: unknown[] = [], depends: string[] = []): FormStackResourceData =>
  ({
    name: "web",
    sourceType: "image",
    source: { image: { ref: "nginx:1.27" } },
    depends_on: depends,
    execution_config: { environment_variables: env },
  }) as never;

const redis = (): FormStackResourceData =>
  ({ name: "redis", sourceType: "image", source: { image: { ref: "redis:7" } } }) as never;

const resourceRow = (resourceName: string) => ({
  from: "resource",
  name: "REDIS_HOST",
  resourceName,
  output: "host",
});

const envOf = (r: FormStackResourceData) =>
  (r.execution_config?.environment_variables ?? []) as Array<{ resourceName?: string }>;

describe("renameResourceReferences", () => {
  it("repoints an env row that names the renamed resource", () => {
    const next = renameResourceReferences([web([resourceRow("redis")]), redis()], "redis", "cache");
    expect(envOf(next[0])[0].resourceName).toBe("cache");
  });

  it("repoints a templated env row", () => {
    const row = { from: "resourceTemplate", name: "URL", resourceName: "redis", template: "{h}", values: {} };
    const next = renameResourceReferences([web([row]), redis()], "redis", "cache");
    expect(envOf(next[0])[0].resourceName).toBe("cache");
  });

  it("repoints a depends_on entry", () => {
    const next = renameResourceReferences([web([], ["redis"]), redis()], "redis", "cache");
    expect(next[0].depends_on).toEqual(["cache"]);
  });

  it("leaves rows that name a different resource alone", () => {
    const next = renameResourceReferences(
      [web([resourceRow("other")], ["other"]), redis()],
      "redis",
      "cache",
    );
    expect(envOf(next[0])[0].resourceName).toBe("other");
    expect(next[0].depends_on).toEqual(["other"]);
  });

  it("leaves env rows of other kinds untouched", () => {
    const secretRow = { from: "secret", name: "TOKEN", secretId: "s1", secretKey: "K" };
    const next = renameResourceReferences([web([secretRow]), redis()], "redis", "cache");
    expect(envOf(next[0])[0]).toEqual(secretRow);
  });

  /** A rename is one resource's edit; the other resources must keep their
   *  identity so the diff still pairs them rather than re-reporting them. */
  it("returns the same object for a resource it did not touch", () => {
    const resources = [web([resourceRow("other")]), redis()];
    const next = renameResourceReferences(resources, "redis", "cache");
    expect(next[0]).toBe(resources[0]);
  });

  /** Block de-duplication renames several resources at once. Applying the pair
   *  form in a loop would rewrite `a` to `a-2` and then catch the sibling that
   *  was already called `a-2`; every lookup here reads the original name. */
  it("carries simultaneous renames without a rename catching its own output", () => {
    const app = { ...web([resourceRow("a")], ["a", "a-2"]) };
    const next = renameResourceReferencesByMap(
      [app],
      new Map([
        ["a", "a-2"],
        ["a-2", "a-3"],
      ]),
    );
    expect(envOf(next[0])[0].resourceName).toBe("a-2");
    expect(next[0].depends_on).toEqual(["a-2", "a-3"]);
  });

  it("does nothing when given no renames", () => {
    const resources = [web([resourceRow("redis")]), redis()];
    expect(renameResourceReferencesByMap(resources, new Map())).toBe(resources);
  });

  it("does nothing when the name did not change", () => {
    const resources = [web([resourceRow("redis")]), redis()];
    expect(renameResourceReferences(resources, "redis", "redis")).toBe(resources);
  });

  /** A half-typed new name would rewrite references to garbage on every
   *  keystroke; the rename only carries once there is a name to carry to. */
  it("does nothing when either name is empty", () => {
    const resources = [web([resourceRow("redis")]), redis()];
    expect(renameResourceReferences(resources, "redis", "")).toBe(resources);
    expect(renameResourceReferences(resources, "", "cache")).toBe(resources);
  });
});
