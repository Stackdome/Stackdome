import { describe, it, expect } from "vitest";
import { FormStackSchema } from "../form-schema";

/**
 * An env row that sources another resource's output addresses it by name. When
 * that resource is gone — deleted, or renamed without carrying its references —
 * the API only rejects it while rendering the release, so the deploy fails with
 * "references unknown resource". Catch it in the form instead, the way
 * `depends_on` already is.
 */

const resource = (name: string, env: unknown[] = []) => ({
  name,
  sourceType: "image",
  source: { image: { ref: "nginx:1.27" } },
  execution_config: { environment_variables: env },
});

const stackWith = (resources: unknown[]) => ({
  name: "demo",
  spec: { stack_resources: resources, volumes: [] },
});

const issuesFor = (resources: unknown[]) => {
  const result = FormStackSchema.safeParse(stackWith(resources));
  return result.success ? [] : result.error.issues;
};

const envIssues = (resources: unknown[]) =>
  issuesFor(resources).filter((i) => i.path.includes("resourceName"));

describe("env rows that source another resource", () => {
  const rowFor = (resourceName: string) => ({
    from: "resource",
    name: "REDIS_HOST",
    resourceName,
    output: "host",
  });

  it("rejects a row naming a resource that is not in the stack", () => {
    const issues = envIssues([resource("web", [rowFor("redis")]), resource("cache")]);
    expect(issues).toHaveLength(1);
    expect(issues[0].message).toBe('Unknown resource "redis"');
    expect(issues[0].path).toEqual([
      "spec",
      "stack_resources",
      0,
      "execution_config",
      "environment_variables",
      0,
      "resourceName",
    ]);
  });

  it("accepts a row naming a resource that exists", () => {
    expect(envIssues([resource("web", [rowFor("cache")]), resource("cache")])).toEqual([]);
  });

  it("rejects a templated row naming a missing resource", () => {
    const row = { from: "resourceTemplate", name: "URL", resourceName: "redis", template: "{h}", values: {} };
    expect(envIssues([resource("web", [row]), resource("cache")])).toHaveLength(1);
  });

  /** A row may reference the resource it sits on. */
  it("accepts a row naming its own resource", () => {
    expect(envIssues([resource("web", [rowFor("web")])])).toEqual([]);
  });
});
