import { describe, it, expect } from "vitest";
import { cloneJson, isPathDirty } from "./stack-diff";

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
