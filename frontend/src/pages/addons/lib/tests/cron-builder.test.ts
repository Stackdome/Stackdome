// @vitest-environment node
import { describe, it, expect } from "vitest";
import { buildCron, parseCron, normalizeCron, isValidCronArity } from "../cron-builder";

describe("isValidCronArity", () => {
  it("accepts 6-field and 5-field (upgraded) expressions", () => {
    expect(isValidCronArity("0 0 3 * * *")).toBe(true);
    expect(isValidCronArity("0 3 * * *")).toBe(true);
  });
  it("accepts @macros", () => {
    expect(isValidCronArity("@daily")).toBe(true);
  });
  it("rejects a missed-space expression", () => {
    expect(isValidCronArity("0 0 3 1 *1")).toBe(false);
  });
  it("rejects empty and too-many fields", () => {
    expect(isValidCronArity("")).toBe(false);
    expect(isValidCronArity("0 0 3 1 1 1 1")).toBe(false);
  });
});

describe("normalizeCron", () => {
  it("trims and collapses whitespace", () => {
    expect(normalizeCron("  0   0 3 *  * * ")).toBe("0 0 3 * * *");
  });
  it("upgrades a 5-field expression to 6 by prepending seconds", () => {
    expect(normalizeCron("0 3 * * *")).toBe("0 0 3 * * *");
  });
  it("leaves a 6-field expression as-is", () => {
    expect(normalizeCron("0 30 2 * * 1")).toBe("0 30 2 * * 1");
  });
  it("passes @macros through untouched", () => {
    expect(normalizeCron("  @daily ")).toBe("@daily");
  });
});

describe("buildCron", () => {
  const base = { minute: 0, hour: 3, dayOfWeek: 0, dayOfMonth: 1, custom: "" };
  it("hourly → at minute each hour", () => {
    expect(buildCron({ ...base, frequency: "hourly", minute: 15 })).toBe("0 15 * * * *");
  });
  it("daily → at hour:minute", () => {
    expect(buildCron({ ...base, frequency: "daily", hour: 2, minute: 30 })).toBe("0 30 2 * * *");
  });
  it("weekly → day-of-week at hour:minute", () => {
    expect(buildCron({ ...base, frequency: "weekly", dayOfWeek: 1, hour: 4, minute: 0 })).toBe(
      "0 0 4 * * 1",
    );
  });
  it("monthly → day-of-month at hour:minute", () => {
    expect(buildCron({ ...base, frequency: "monthly", dayOfMonth: 15, hour: 5, minute: 0 })).toBe(
      "0 0 5 15 * *",
    );
  });
  it("custom → normalized raw", () => {
    expect(buildCron({ ...base, frequency: "custom", custom: "0 */15 * * *" })).toBe(
      "0 0 */15 * * *",
    );
  });
});

describe("parseCron round-trips", () => {
  it("parses hourly", () => {
    const p = parseCron("0 15 * * * *");
    expect(p.frequency).toBe("hourly");
    expect(p.minute).toBe(15);
  });
  it("parses daily", () => {
    const p = parseCron("0 30 2 * * *");
    expect(p.frequency).toBe("daily");
    expect(p.hour).toBe(2);
    expect(p.minute).toBe(30);
  });
  it("parses weekly", () => {
    const p = parseCron("0 0 4 * * 1");
    expect(p.frequency).toBe("weekly");
    expect(p.dayOfWeek).toBe(1);
    expect(p.hour).toBe(4);
  });
  it("parses monthly", () => {
    const p = parseCron("0 0 5 15 * *");
    expect(p.frequency).toBe("monthly");
    expect(p.dayOfMonth).toBe(15);
  });
  it("falls back to custom for exotic expressions", () => {
    const p = parseCron("0 */15 * * * *");
    expect(p.frequency).toBe("custom");
    expect(p.custom).toBe("0 */15 * * * *");
  });
  it("accepts a 5-field custom and normalizes it", () => {
    const p = parseCron("*/10 * * * *");
    expect(p.frequency).toBe("custom");
    expect(p.custom).toBe("0 */10 * * * *");
  });
});
