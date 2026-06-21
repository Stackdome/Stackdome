// @vitest-environment jsdom
import { describe, it, expect } from "vitest";
import { diffSnapshots } from "../release-snapshot-diff";

const snap = (resources: unknown[]) => ({ resources });
const web = (over: Record<string, unknown> = {}) => ({
  name: "web",
  image_spec: { image: "web:1" },
  ports: [{ number: 3000 }],
  execution_config: { command: ["node", "a.js"], environment_variables: [{ name: "LOG", value: "info" }] },
  ...over,
});

describe("diffSnapshots", () => {
  it("returns [] when nothing changed", () => {
    expect(diffSnapshots(snap([web()]), snap([web()]))).toEqual([]);
  });

  it("flags a modified image as a changed configuration row", () => {
    const out = diffSnapshots(snap([web()]), snap([web({ image_spec: { image: "web:2" } })]));
    expect(out).toHaveLength(1);
    expect(out[0]).toMatchObject({ name: "web", change: "modified" });
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows).toContainEqual({ key: "image", from: "web:1", to: "web:2", kind: "changed" });
  });

  it("splits env changes into an environment section", () => {
    const out = diffSnapshots(
      snap([web()]),
      snap([web({ execution_config: { command: ["node", "a.js"], environment_variables: [{ name: "LOG", value: "debug" }, { name: "NEW", value: "1" }] } })]),
    );
    const env = out[0].sections.find((s) => s.kind === "environment")!;
    expect(env.rows).toContainEqual({ key: "LOG", from: "info", to: "debug", kind: "changed" });
    expect(env.rows).toContainEqual({ key: "NEW", to: "1", kind: "added" });
  });

  it("marks an added resource with all rows as added", () => {
    const out = diffSnapshots(snap([]), snap([web()]));
    expect(out[0]).toMatchObject({ name: "web", change: "added" });
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows.every((r) => r.kind === "added")).toBe(true);
  });

  it("marks a removed resource with a note and no sections", () => {
    const out = diffSnapshots(snap([web()]), snap([]));
    expect(out[0]).toMatchObject({ name: "web", change: "removed", sections: [] });
    expect(out[0].note).toMatch(/removed/i);
  });

  it("returns [] when there is no previous snapshot", () => {
    expect(diffSnapshots(undefined, snap([web()]))).toEqual([]);
  });
});
