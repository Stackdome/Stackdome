// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from "vitest";
import { isCanvasEnabled } from "../feature-flags";

describe("isCanvasEnabled", () => {
  beforeEach(() => localStorage.clear());

  it("is true by default (canvas-first)", () => {
    expect(isCanvasEnabled()).toBe(true);
  });

  it("falls back to the legacy form only when explicitly opted out", () => {
    localStorage.setItem("stackCanvas", "0");
    expect(isCanvasEnabled()).toBe(false);
  });

  it("stays on for any other localStorage value", () => {
    localStorage.setItem("stackCanvas", "1");
    expect(isCanvasEnabled()).toBe(true);
  });
});
