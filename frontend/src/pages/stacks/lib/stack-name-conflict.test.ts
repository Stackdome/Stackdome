import { describe, it, expect } from "vitest";
import { stackNameConflictError, stackNameTakenMessage } from "./stack-name-conflict";
import type { Stack } from "@/pages/stacks/types";

const stack = (name: string, project_id: string): Pick<Stack, "name" | "project_id"> => ({
  name,
  project_id,
});

const projectId = "project-1";
const existingStacks = [stack("api", projectId), stack("web", "project-2")];

describe("stackNameConflictError", () => {
  it("blocks create when the name is already taken in the target project", () => {
    const err = stackNameConflictError({
      isCreate: true,
      name: "api",
      projectId,
      existingStacks,
    });
    expect(err).toBe(stackNameTakenMessage("api"));
  });

  it("allows create when the name is free in the target project", () => {
    const err = stackNameConflictError({
      isCreate: true,
      name: "worker",
      projectId,
      existingStacks,
    });
    expect(err).toBeUndefined();
  });

  it("does not treat a name taken only in another project as a conflict", () => {
    const err = stackNameConflictError({
      isCreate: true,
      name: "web",
      projectId,
      existingStacks,
    });
    expect(err).toBeUndefined();
  });

  it("trims the candidate name before comparing", () => {
    const err = stackNameConflictError({
      isCreate: true,
      name: "  api  ",
      projectId,
      existingStacks,
    });
    expect(err).toBe(stackNameTakenMessage("api"));
  });

  it("matches case-sensitively", () => {
    const err = stackNameConflictError({
      isCreate: true,
      name: "API",
      projectId,
      existingStacks,
    });
    expect(err).toBeUndefined();
  });

  it("never blocks the update path even on a name collision", () => {
    const err = stackNameConflictError({
      isCreate: false,
      name: "api",
      projectId,
      existingStacks,
    });
    expect(err).toBeUndefined();
  });
});
