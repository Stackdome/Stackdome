import { describe, it, expect } from "vitest";
import { cloneJson, getAddonLinkCount, isPathDirty } from "../stack-diff";
import type { ResourceArr } from "../stack-diff";

describe("cloneJson", () => {
  it("passes undefined through (JSON.parse(JSON.stringify(undefined)) would throw)", () => {
    // Regression: the field-reset arrow used to crash the page with
    // `"undefined" is not valid JSON` when the path being reverted was
    // absent from baseline (e.g., resetting init_spec.command on a resource
    // that never had an init spec).
    expect(cloneJson(undefined)).toBe(undefined);
  });

  it("deep-clones plain objects, breaking referential identity", () => {
    const src = { a: 1, nested: { b: 2 } };
    const out = cloneJson(src);
    expect(out).toEqual(src);
    expect(out).not.toBe(src);
    expect(out.nested).not.toBe(src.nested);
  });

  it("deep-clones arrays", () => {
    const src = [1, [2, 3]];
    const out = cloneJson(src);
    expect(out).toEqual(src);
    expect(out).not.toBe(src);
    expect(out[1]).not.toBe(src[1]);
  });

  it("returns primitives as-is", () => {
    expect(cloneJson(42)).toBe(42);
    expect(cloneJson("hi")).toBe("hi");
    expect(cloneJson(null)).toBe(null);
    expect(cloneJson(false)).toBe(false);
  });
});

describe("isPathDirty — structurally-empty equivalence", () => {
  // Regression: clearing a comma-separated input (e.g., Init Command) leaves
  // the form with `{ init_spec: { command: [] } }` while baseline has
  // `init_spec: undefined`. These should diff as equal so the field stops
  // reading as dirty after the user reverts.

  it("treats empty array vs undefined as not-dirty", () => {
    expect(isPathDirty(
      { init_spec: { command: [] } },
      { init_spec: undefined },
      "init_spec.command",
    )).toBe(false);
  });

  it("treats empty object vs undefined as not-dirty", () => {
    expect(isPathDirty(
      { init_spec: {} },
      { init_spec: undefined },
      "init_spec",
    )).toBe(false);
  });

  it("treats empty string vs undefined as not-dirty", () => {
    expect(isPathDirty(
      { name: "" },
      { name: undefined },
      "name",
    )).toBe(false);
  });

  it("treats deeply-nested all-empty as not-dirty against undefined", () => {
    expect(isPathDirty(
      { init_spec: { command: [], args: [], image_spec: { image: "" } } },
      { init_spec: undefined },
      "init_spec",
    )).toBe(false);
  });

  it("flags as dirty when one side has actual content", () => {
    expect(isPathDirty(
      { init_spec: { command: ["sh"] } },
      { init_spec: undefined },
      "init_spec.command",
    )).toBe(true);
  });

  it("flags as dirty when value differs from a non-empty baseline", () => {
    expect(isPathDirty(
      { name: "" },
      { name: "redis" },
      "name",
    )).toBe(true);
    expect(isPathDirty(
      { ports: [] },
      { ports: [80] },
      "ports",
    )).toBe(true);
  });

  it("returns false when values are deep-equal", () => {
    expect(isPathDirty(
      { ports: [{ number: 80 }] },
      { ports: [{ number: 80 }] },
      "ports",
    )).toBe(false);
  });
});

describe("getAddonLinkCount", () => {
  // Regression: the stack create page's sticky action bar wasn't surfacing
  // any indicator when an addon was attached — only resources/volumes had
  // segments. This helper feeds that segment, so it must union both sources
  // (panel-linked ids and resource-derived ids) and dedupe.

  it("returns 0 with no linked addons and no resources", () => {
    expect(getAddonLinkCount(new Set(), [])).toBe(0);
  });

  it("counts addons explicitly linked via the panel", () => {
    expect(getAddonLinkCount(new Set(["pg-1", "pg-2"]), [])).toBe(2);
  });

  it("counts addons referenced as env-var sources on resources", () => {
    const resources: ResourceArr = [
      {
        execution_config: {
          environment_variables: [
            { from: "addon", addonId: "pg-1", name: "DATABASE_URL", value: "" },
            { from: "literal", name: "FOO", value: "bar" },
          ],
        } as never,
      },
    ];
    expect(getAddonLinkCount(new Set(), resources)).toBe(1);
  });

  it("unions panel-linked and env-var-derived ids without double counting", () => {
    const resources: ResourceArr = [
      {
        execution_config: {
          environment_variables: [
            { from: "addon", addonId: "pg-1", name: "DATABASE_URL", value: "" },
          ],
        } as never,
      },
    ];
    // pg-1 in both sources, pg-2 only in panel — total 2 distinct ids.
    expect(getAddonLinkCount(new Set(["pg-1", "pg-2"]), resources)).toBe(2);
  });

  it("ignores env vars without an addonId or with a different `from` source", () => {
    const resources: ResourceArr = [
      {
        execution_config: {
          environment_variables: [
            { from: "addon", addonId: "", name: "BAD", value: "" },
            { from: "addon", name: "MISSING_ID", value: "" },
            { from: "secret", addonId: "pg-9", name: "FROM_SECRET", value: "" },
          ],
        } as never,
      },
    ];
    expect(getAddonLinkCount(new Set(), resources)).toBe(0);
  });

  it("dedupes the same addonId referenced across multiple resources", () => {
    const resources: ResourceArr = [
      {
        execution_config: {
          environment_variables: [
            { from: "addon", addonId: "pg-1", name: "URL_A", value: "" },
          ],
        } as never,
      },
      {
        execution_config: {
          environment_variables: [
            { from: "addon", addonId: "pg-1", name: "URL_B", value: "" },
          ],
        } as never,
      },
    ];
    expect(getAddonLinkCount(new Set(), resources)).toBe(1);
  });

  it("handles resources with no execution_config", () => {
    const resources: ResourceArr = [{ name: "svc-1" }, { name: "svc-2" }];
    expect(getAddonLinkCount(new Set(["pg-1"]), resources)).toBe(1);
  });
});
