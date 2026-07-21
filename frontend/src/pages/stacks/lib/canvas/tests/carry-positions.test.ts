import { describe, it, expect } from "vitest";
import { carryPositions, type CarryableNode } from "../carry-positions";

const res = (id: string, idx: number, x = 0, y = 0): CarryableNode => ({
  id: `resource:${id}`,
  type: "resource",
  position: { x, y },
  data: { kind: "service", name: id, resourceIdx: idx },
});

const vol = (name: string, idx: number, x = 0, y = 0): CarryableNode => ({
  id: `volume:${name}`,
  type: "attachment",
  position: { x, y },
  data: { kind: "volume", name, volumeIdx: idx },
});

describe("carryPositions", () => {
  it("carries positions by id for unchanged nodes", () => {
    const prev = [res("web", 0, 100, 200)];
    const laid = [res("web", 0)];
    const { nodes, keptIds } = carryPositions(prev, laid);
    expect(nodes[0].position).toEqual({ x: 100, y: 200 });
    expect(keptIds.has("resource:web")).toBe(true);
  });

  it("a renamed resource inherits its old position via resourceIdx", () => {
    const prev = [res("web", 0, 100, 200), res("api", 1, 400, 200)];
    const laid = [res("websda", 0), res("api", 1)];
    const { nodes, keptIds } = carryPositions(prev, laid);
    expect(nodes[0].position).toEqual({ x: 100, y: 200 });
    expect(keptIds.has("resource:websda")).toBe(true);
  });

  it("does NOT inherit across a simultaneous delete + rename (index shift)", () => {
    // [A(0), B(1)] -> delete A, rename B→B2: B2 now has idx 0. Inheriting A's
    // position would teleport B2 onto the deleted node's spot.
    const prev = [res("a", 0, 100, 200), res("b", 1, 400, 200)];
    const laid = [res("b2", 0)];
    const { nodes, keptIds } = carryPositions(prev, laid);
    expect(keptIds.has("resource:b2")).toBe(false);
    expect(nodes[0].position).toEqual({ x: 0, y: 0 }); // fresh layout coords kept
  });

  it("a renamed volume inherits its old position via volumeIdx", () => {
    const prev = [res("web", 0, 100, 200), vol("data", 0, 700, 100)];
    const laid = [res("web", 0), vol("dat", 0)];
    const { nodes, keptIds } = carryPositions(prev, laid);
    expect(nodes[1].position).toEqual({ x: 700, y: 100 });
    expect(keptIds.has("volume:dat")).toBe(true);
  });

  it("volume parity gate: delete+rename of volumes falls back to fresh coords", () => {
    const prev = [vol("a", 0, 100, 100), vol("b", 1, 400, 100)];
    const laid = [vol("b2", 0)];
    const { keptIds } = carryPositions(prev, laid);
    expect(keptIds.has("volume:b2")).toBe(false);
  });

  it("carries measured dimensions forward for kept nodes", () => {
    const prev = [{ ...res("web", 0, 100, 200), measured: { width: 216, height: 104 } }];
    const laid = [res("web", 0)];
    const { nodes } = carryPositions(prev, laid);
    expect(nodes[0].measured).toEqual({ width: 216, height: 104 });
  });

  it("genuinely-new nodes are not in keptIds and keep their layout coords", () => {
    const prev = [res("web", 0, 100, 200)];
    const laid = [res("web", 0), res("api", 1, 33, 44)];
    const { nodes, keptIds } = carryPositions(prev, laid);
    expect(keptIds.has("resource:api")).toBe(false);
    expect(nodes[1].position).toEqual({ x: 33, y: 44 });
  });
});
