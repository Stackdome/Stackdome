import { describe, it, expect } from "vitest";
import type { z } from "zod";
import { formatDraftValidationIssues } from "@/pages/stacks/lib/format-draft-validation";

const issue = (path: (string | number)[], message: string) =>
  ({ path, message, code: "custom" }) as unknown as z.ZodIssue;

describe("formatDraftValidationIssues", () => {
  it("routes a name issue to nameError, not messages", () => {
    const out = formatDraftValidationIssues([issue(["name"], "Required")], [], []);
    expect(out.nameError).toBe("Required");
    expect(out.messages).toEqual([]);
  });

  it("labels resource issues with the resource name and field path", () => {
    const out = formatDraftValidationIssues(
      [issue(["spec", "stack_resources", 0, "image"], "Required")],
      [{ name: "web" }],
      [],
    );
    expect(out.messages).toEqual(["web: Required (image)"]);
  });

  it("falls back to positional labels for unnamed resources and volumes", () => {
    const out = formatDraftValidationIssues(
      [
        issue(["spec", "stack_resources", 1, "image"], "Required"),
        issue(["spec", "volumes", 0, "size"], "Required"),
      ],
      [{ name: "web" }, {}],
      [{}],
    );
    expect(out.messages).toEqual(["Resource 2: Required (image)", "Volume 1: Required (size)"]);
  });

  it("passes through issues outside resources/volumes with their joined path", () => {
    const out = formatDraftValidationIssues([issue(["spec", "labels"], "Invalid")], [], []);
    expect(out.messages).toEqual(["spec.labels: Invalid"]);
  });

  it("dedupes identical messages", () => {
    const dup = issue(["spec", "stack_resources", 0, "image"], "Required");
    const out = formatDraftValidationIssues([dup, dup], [{ name: "web" }], []);
    expect(out.messages).toEqual(["web: Required (image)"]);
  });
});
