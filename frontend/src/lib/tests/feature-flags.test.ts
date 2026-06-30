// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from "vitest";
import { isCanvasEnabled } from "../feature-flags";

describe("isCanvasEnabled", () => {
  beforeEach(() => localStorage.clear());

  it("is false by default", () => {
    expect(isCanvasEnabled()).toBe(false);
  });

  it("is true when the localStorage override is set", () => {
    localStorage.setItem("stackCanvas", "1");
    expect(isCanvasEnabled()).toBe(true);
  });

  it("ignores other localStorage values", () => {
    localStorage.setItem("stackCanvas", "0");
    expect(isCanvasEnabled()).toBe(false);
  });
});
