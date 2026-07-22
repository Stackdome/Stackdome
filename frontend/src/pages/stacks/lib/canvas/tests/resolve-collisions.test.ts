import { describe, it, expect } from "vitest";
import { resolveCollisions, type CollidableNode } from "../resolve-collisions";

const node = (id: string, x: number, y: number, w = 216, h = 104): CollidableNode => ({
  id,
  position: { x, y },
  measured: { width: w, height: h },
});

const rectsOverlap = (a: CollidableNode, b: CollidableNode, margin = 0) => {
  const aw = a.measured!.width!, ah = a.measured!.height!;
  const bw = b.measured!.width!, bh = b.measured!.height!;
  return (
    a.position.x < b.position.x + bw + margin &&
    b.position.x < a.position.x + aw + margin &&
    a.position.y < b.position.y + bh + margin &&
    b.position.y < a.position.y + ah + margin
  );
};

describe("resolveCollisions", () => {
  it("returns the same array when nothing overlaps", () => {
    const nodes = [node("a", 0, 0), node("b", 500, 0)];
    expect(resolveCollisions(nodes, { margin: 15 })).toBe(nodes);
  });

  it("separates two overlapping nodes to at least the margin", () => {
    const nodes = [node("a", 0, 0), node("b", 40, 10)];
    const out = resolveCollisions(nodes, { margin: 15 });
    expect(rectsOverlap(out[0], out[1])).toBe(false);
  });

  it("never moves locked nodes — the movable one is pushed instead", () => {
    const nodes = [node("a", 0, 0), node("b", 40, 10)];
    const out = resolveCollisions(nodes, { margin: 15, isLocked: (n) => n.id === "a" });
    expect(out[0].position).toEqual({ x: 0, y: 0 });
    expect(rectsOverlap(out[0], out[1])).toBe(false);
  });

  it("leaves a locked-locked overlap alone and terminates", () => {
    const nodes = [node("a", 0, 0), node("b", 10, 10)];
    const out = resolveCollisions(nodes, { margin: 15, isLocked: () => true });
    expect(out[0].position).toEqual({ x: 0, y: 0 });
    expect(out[1].position).toEqual({ x: 10, y: 10 });
  });

  it("resolves a pile of stacked nodes into a fully non-overlapping set", () => {
    const nodes = [node("a", 0, 0), node("b", 5, 5), node("c", 10, 10), node("d", 15, 15)];
    const out = resolveCollisions(nodes, { margin: 15 });
    for (let i = 0; i < out.length; i++) {
      for (let j = i + 1; j < out.length; j++) {
        expect(rectsOverlap(out[i], out[j])).toBe(false);
      }
    }
  });

  it("respects maxIterations as a hard stop", () => {
    const nodes = [node("a", 0, 0), node("b", 1, 1), node("c", 2, 2)];
    // One iteration may not fully resolve a pile; the call must still return.
    const out = resolveCollisions(nodes, { margin: 15, maxIterations: 1 });
    expect(out).toHaveLength(3);
  });

  it("parks a node wedged in a too-narrow locked corridor above the layout instead of returning overlap", () => {
    // Two TALL locked walls 100px apart; a 216-wide movable node between them
    // can never fit horizontally and the walls are too tall to slide past —
    // it must escape above the layout, not stay overlapping.
    const walls = [node("left", 0, 0, 216, 2000), node("right", 316, 0, 216, 2000)];
    const trapped = node("t", 100, 900);
    const out = resolveCollisions([...walls, trapped], {
      margin: 15,
      isLocked: (n) => n.id !== "t",
    });
    const t = out.find((n) => n.id === "t")!;
    for (const w of out.filter((n) => n.id !== "t")) {
      expect(rectsOverlap(t, w)).toBe(false);
    }
    expect(t.position.y).toBeLessThan(0); // parked above the frozen layout
  });

  it("uses fallbackSize for unmeasured nodes", () => {
    // Unmeasured node sitting 60px above a locked box: with the 56px-tall
    // attachment fallback its bottom edge (-4) clears the box; the default
    // 104px fallback would read as overlapping and shove it.
    const small = { id: "s", position: { x: 0, y: -60 } }; // no measured
    const big = node("b", 0, 0);
    const opts = { margin: 0, isLocked: (n: { id: string }) => n.id === "b" };
    const withFallback = resolveCollisions([big, small], {
      ...opts,
      fallbackSize: () => ({ width: 180, height: 56 }),
    });
    expect(withFallback.find((n) => n.id === "s")!.position).toEqual({ x: 0, y: -60 });
    const withoutFallback = resolveCollisions([big, small], opts);
    expect(withoutFallback.find((n) => n.id === "s")!.position).not.toEqual({ x: 0, y: -60 });
  });

  it("does not mutate the input nodes", () => {
    const nodes = [node("a", 0, 0), node("b", 40, 10)];
    resolveCollisions(nodes, { margin: 15 });
    expect(nodes[1].position).toEqual({ x: 40, y: 10 });
  });
});
