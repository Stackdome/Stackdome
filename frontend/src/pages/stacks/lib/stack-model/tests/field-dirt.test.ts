import { describe, it, expect } from "vitest";
import { isFieldDirty } from "../field-dirt";

/**
 * The form and the API disagree constantly about how an absent value is
 * spelled. None of those spellings is an edit the user made.
 */
describe("isFieldDirty — structurally-empty equivalence", () => {
  // Regression: clearing a comma-separated input (e.g., Init Command) leaves
  // the form with `{ init_spec: { command: [] } }` while baseline has
  // `init_spec: undefined`. These should diff as equal so the field stops
  // reading as dirty after the user reverts.

  it("treats empty array vs undefined as not-dirty", () => {
    expect(isFieldDirty(
      { init_spec: { command: [] } },
      { init_spec: undefined },
      "init_spec.command",
    )).toBe(false);
  });

  it("treats empty object vs undefined as not-dirty", () => {
    expect(isFieldDirty(
      { init_spec: {} },
      { init_spec: undefined },
      "init_spec",
    )).toBe(false);
  });

  it("treats empty string vs undefined as not-dirty", () => {
    expect(isFieldDirty(
      { name: "" },
      { name: undefined },
      "name",
    )).toBe(false);
  });

  it("treats deeply-nested all-empty as not-dirty against undefined", () => {
    expect(isFieldDirty(
      { init_spec: { command: [], args: [], image_spec: { image: "" } } },
      { init_spec: undefined },
      "init_spec",
    )).toBe(false);
  });

  it("flags as dirty when one side has actual content", () => {
    expect(isFieldDirty(
      { init_spec: { command: ["sh"] } },
      { init_spec: undefined },
      "init_spec.command",
    )).toBe(true);
  });

  it("flags as dirty when value differs from a non-empty baseline", () => {
    expect(isFieldDirty(
      { name: "" },
      { name: "redis" },
      "name",
    )).toBe(true);
    expect(isFieldDirty(
      { ports: [] },
      { ports: [80] },
      "ports",
    )).toBe(true);
  });

  it("returns false when values are deep-equal", () => {
    expect(isFieldDirty(
      { ports: [{ number: 80 }] },
      { ports: [{ number: 80 }] },
      "ports",
    )).toBe(false);
  });
});

