import { describe, it, expect } from "vitest";
import { volumeDotClass } from "@/pages/stacks/lib/volume-status";

describe("volumeDotClass — one map for item and detail", () => {
  it("Ready → success", () => expect(volumeDotClass("Ready")).toBe("bg-success"));
  it("Pending → warn", () => expect(volumeDotClass("Pending")).toBe("bg-warn"));
  it("missing phase → muted (not fake-pending)", () => expect(volumeDotClass(undefined)).toBe("bg-fg-muted"));
  it("unknown word (the old 'running' fight) → info", () => expect(volumeDotClass("running")).toBe("bg-info"));
});
