import { describe, it, expect } from "vitest";
import { Position } from "@xyflow/react";
import { floatingEdgeGeometry, type NodeRect } from "../floating-edge-geometry";

const rect = (x: number, y: number, width = 216, height = 88): NodeRect => ({ x, y, width, height });

describe("floatingEdgeGeometry", () => {
  it("target directly above source: leaves source top, enters target bottom", () => {
    const source = rect(0, 300);
    const target = rect(0, 0);
    const geo = floatingEdgeGeometry(source, target);
    expect(geo.sourcePosition).toBe(Position.Top);
    expect(geo.sourceX).toBeCloseTo(108); // horizontal centre of source
    expect(geo.sourceY).toBeCloseTo(300); // source's top face
    expect(geo.targetPosition).toBe(Position.Bottom);
    expect(geo.targetY).toBeCloseTo(88); // target's bottom face
  });

  it("target to the right: leaves source right face, enters target left face", () => {
    const source = rect(0, 0);
    const target = rect(600, 0);
    const geo = floatingEdgeGeometry(source, target);
    expect(geo.sourcePosition).toBe(Position.Right);
    expect(geo.sourceX).toBeCloseTo(216);
    expect(geo.targetPosition).toBe(Position.Left);
    expect(geo.targetX).toBeCloseTo(600);
  });

  it("diagonal counterpart lands on the facing face, not a corner overshoot", () => {
    const source = rect(0, 400);
    const target = rect(400, 0); // up and to the right; steeper vertically than the card aspect
    const geo = floatingEdgeGeometry(source, target);
    // Attachment points always lie ON the rectangle boundary.
    const onBoundary = (r: NodeRect, x: number, y: number) =>
      (Math.abs(x - r.x) < 1e-6 ||
        Math.abs(x - (r.x + r.width)) < 1e-6 ||
        Math.abs(y - r.y) < 1e-6 ||
        Math.abs(y - (r.y + r.height)) < 1e-6) &&
      x >= r.x - 1e-6 &&
      x <= r.x + r.width + 1e-6 &&
      y >= r.y - 1e-6 &&
      y <= r.y + r.height + 1e-6;
    expect(onBoundary(source, geo.sourceX, geo.sourceY)).toBe(true);
    expect(onBoundary(target, geo.targetX, geo.targetY)).toBe(true);
  });

  it("keeps attachment on the boundary when one rect fully contains the other", () => {
    const outer = rect(0, 0, 1000, 1000);
    const inner = rect(400, 450, 100, 40); // inner centre inside outer
    const geo = floatingEdgeGeometry(outer, inner);
    // Ray-based intersection: both points finite and on their own rect edges.
    expect(Number.isFinite(geo.sourceX)).toBe(true);
    expect(Number.isFinite(geo.targetX)).toBe(true);
    const onEdge = (r: NodeRect, x: number, y: number) =>
      Math.abs(x - r.x) < 1e-6 || Math.abs(x - (r.x + r.width)) < 1e-6 ||
      Math.abs(y - r.y) < 1e-6 || Math.abs(y - (r.y + r.height)) < 1e-6;
    expect(onEdge(outer, geo.sourceX, geo.sourceY)).toBe(true);
    expect(onEdge(inner, geo.targetX, geo.targetY)).toBe(true);
  });

  it("coincident centres fall back to source-bottom/target-top without NaN", () => {
    const geo = floatingEdgeGeometry(rect(0, 0), rect(0, 0));
    expect(Number.isFinite(geo.sourceX)).toBe(true);
    expect(Number.isFinite(geo.targetY)).toBe(true);
    expect(geo.sourcePosition).toBe(Position.Bottom);
    expect(geo.targetPosition).toBe(Position.Top);
  });

  it("parallel edges get distinct attachment points, symmetric about the centre line", () => {
    const source = rect(0, 300);
    const target = rect(0, 0);
    const a = floatingEdgeGeometry(source, target, 0, 2);
    const b = floatingEdgeGeometry(source, target, 1, 2);
    expect(a.sourceX).not.toBeCloseTo(b.sourceX);
    // Symmetric offsets: mean of the two equals the unoffset centre.
    expect((a.sourceX + b.sourceX) / 2).toBeCloseTo(108);
  });

  it("single edge (count 1) is identical to no-offset call", () => {
    const source = rect(0, 300);
    const target = rect(0, 0);
    expect(floatingEdgeGeometry(source, target, 0, 1)).toEqual(floatingEdgeGeometry(source, target));
  });
});
