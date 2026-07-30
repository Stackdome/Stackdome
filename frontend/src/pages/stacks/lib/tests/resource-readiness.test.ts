import { describe, it, expect } from "vitest";
import { isResourceReady } from "../resource-readiness";

describe("isResourceReady", () => {
  it("accepts Ready in any casing", () => {
    expect(isResourceReady("Ready")).toBe(true);
    expect(isResourceReady("ready")).toBe(true);
  });

  it("rejects every non-ready state", () => {
    expect(isResourceReady("Pending")).toBe(false);
    expect(isResourceReady("Degraded")).toBe(false);
    expect(isResourceReady("Failed")).toBe(false);
    expect(isResourceReady("")).toBe(false);
    expect(isResourceReady(undefined)).toBe(false);
    expect(isResourceReady(null)).toBe(false);
  });
});
