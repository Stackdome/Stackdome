import { describe, it, expect } from "vitest";
import { normalizeLabel } from "../labels";

describe("normalizeLabel", () => {
  it("lowercases and trims", () => {
    expect(normalizeLabel("  Prod  ")).toBe("prod");
  });
  it("collapses whitespace runs to a single dash", () => {
    expect(normalizeLabel("payments  team core")).toBe("payments-team-core");
  });
  it("returns empty string for whitespace-only input", () => {
    expect(normalizeLabel("   ")).toBe("");
  });
});
