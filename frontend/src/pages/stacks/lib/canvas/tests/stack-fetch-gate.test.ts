import { describe, it, expect } from "vitest";
import { createStackFetchGate } from "../stack-fetch-gate";

describe("createStackFetchGate", () => {
  it("applies responses arriving in order", () => {
    const gate = createStackFetchGate();
    const a = gate.begin();
    const b = gate.begin();
    expect(gate.shouldApply(a)).toBe(true);
    expect(gate.shouldApply(b)).toBe(true);
  });

  it("drops a stale response that resolves after a newer one", () => {
    const gate = createStackFetchGate();
    const older = gate.begin();
    const newer = gate.begin();
    expect(gate.shouldApply(newer)).toBe(true);
    expect(gate.shouldApply(older)).toBe(false);
  });

  it("never applies the same ticket twice", () => {
    const gate = createStackFetchGate();
    const t = gate.begin();
    expect(gate.shouldApply(t)).toBe(true);
    expect(gate.shouldApply(t)).toBe(false);
  });
});
