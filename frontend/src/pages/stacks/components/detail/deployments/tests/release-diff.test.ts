import { describe, it, expect } from "vitest";
import { diffSnapshots } from "../release-diff";

describe("diffSnapshots", () => {
  it("reports changed scalar leaves with before/after", () => {
    const out = diffSnapshots({ a: 1, b: { c: "x" } }, { a: 2, b: { c: "x" } });
    expect(out).toEqual([{ path: "a", before: 1, after: 2, kind: "changed" }]);
  });

  it("reports added and removed keys", () => {
    const out = diffSnapshots({ a: 1 }, { a: 1, b: 2 });
    expect(out).toContainEqual({ path: "b", before: undefined, after: 2, kind: "added" });
    const out2 = diffSnapshots({ a: 1, b: 2 }, { a: 1 });
    expect(out2).toContainEqual({ path: "b", before: 2, after: undefined, kind: "removed" });
  });

  it("descends into nested objects with dotted paths", () => {
    const out = diffSnapshots({ b: { c: "x" } }, { b: { c: "y" } });
    expect(out).toEqual([{ path: "b.c", before: "x", after: "y", kind: "changed" }]);
  });

  it("treats arrays as leaves (whole-value change)", () => {
    const out = diffSnapshots({ env: ["A=1"] }, { env: ["A=1", "B=2"] });
    expect(out).toEqual([{ path: "env", before: ["A=1"], after: ["A=1", "B=2"], kind: "changed" }]);
  });

  it("returns [] for equal snapshots", () => {
    expect(diffSnapshots({ a: 1 }, { a: 1 })).toEqual([]);
  });
});
