import { describe, it, expect } from "vitest";
import { FormStackSchema } from "../form-schema";

/**
 * Resources are addressed by name everywhere: depends_on, env references, and
 * connections. Two sharing a name make every reference to it ambiguous, and a
 * delete matches by name so it removes both. The API rejects duplicates on
 * save; catch it in the form while the user can still see which field is wrong.
 */

const resource = (name: string) => ({
  name,
  sourceType: "image",
  source: { image: { ref: "nginx:1.27" } },
});

const nameIssues = (names: string[]) => {
  const result = FormStackSchema.safeParse({
    name: "demo",
    spec: { stack_resources: names.map(resource), volumes: [] },
  });
  return result.success
    ? []
    : result.error.issues.filter((i) => /already in use/i.test(i.message));
};

describe("duplicate resource names", () => {
  it("accepts distinct names", () => {
    expect(nameIssues(["web", "api", "db"])).toEqual([]);
  });

  it("rejects a repeated name", () => {
    expect(nameIssues(["web", "web"])).toHaveLength(1);
  });

  it("points at the later duplicate so the first entry stays clean", () => {
    const [issue] = nameIssues(["web", "web"]);
    expect(issue.path).toEqual(["spec", "stack_resources", 1, "name"]);
  });

  it("reports each extra occurrence", () => {
    expect(nameIssues(["web", "web", "web"])).toHaveLength(2);
  });
});
