import { describe, it, expect } from "vitest";
import type { Stack } from "@/api/stacks";
import { deriveFailingResources, deriveRecovered, humanizeFailureType, causeLabel, formatDuration } from "../derive";

function stackWith(resources: Array<Record<string, unknown>>): Stack {
  return { spec: { stack_resources: resources } } as unknown as Stack;
}

describe("deriveFailingResources", () => {
  it("joins last_failure from spec.stack_resources[].status", () => {
    const stack = stackWith([
      { name: "tooljet", status: { state: "CrashLoopBackOff", last_failure: {
        type: "runtime_crash", container: { failure_type: "crash_loop", reason: "CrashLoopBackOff", message: "exit 1", exit_code: 1, restart_count: 5 } } } },
      { name: "redis", status: { state: "Ready" } },
    ]);
    const out = deriveFailingResources(stack);
    expect(out).toHaveLength(1);
    expect(out[0]).toMatchObject({ name: "tooljet", type: "runtime_crash", stage: "runtime", reason: "CrashLoopBackOff", exitCode: 1, restartCount: 5 });
  });

  it("classifies a build failure to the build stage", () => {
    const stack = stackWith([
      { name: "api", status: { state: "Error", last_failure: {
        type: "build_failure", build: { failure_type: "image_pull_failed", reason: "ErrImagePull", message: "manifest unknown" } } } },
    ]);
    expect(deriveFailingResources(stack)[0]).toMatchObject({ name: "api", type: "build_failure", stage: "build", reason: "ErrImagePull" });
  });
});

describe("deriveRecovered", () => {
  it("flags a Ready resource that still carries last_failure", () => {
    const stack = stackWith([
      { name: "tooljet", status: { state: "Ready", last_failure: {
        type: "runtime_crash", container: { reason: "CrashLoopBackOff", restart_count: 5 } } } },
    ]);
    expect(deriveRecovered(stack)).toEqual([{ name: "tooljet", reason: "CrashLoopBackOff", restartCount: 5 }]);
  });

  it("does not flag a failing resource as recovered", () => {
    const stack = stackWith([
      { name: "tooljet", status: { state: "CrashLoopBackOff", last_failure: { type: "runtime_crash", container: { reason: "x" } } } },
    ]);
    expect(deriveRecovered(stack)).toEqual([]);
  });
});

describe("humanizeFailureType", () => {
  it("maps known types", () => {
    expect(humanizeFailureType("out_of_memory")).toBe("Out of memory");
    expect(humanizeFailureType("crash_loop")).toBe("Crash loop");
  });
  it("falls back to the raw value", () => {
    expect(humanizeFailureType("weird_thing")).toBe("weird_thing");
    expect(humanizeFailureType(undefined)).toBe("Unknown");
  });
});

describe("causeLabel", () => {
  it("labels manual / rollback / webhook", () => {
    expect(causeLabel({ kind: "manual" })).toBe("Manual deploy");
    expect(causeLabel({ kind: "rollback", detail: "12" })).toBe("Rollback to #12");
    expect(causeLabel({ kind: "webhook_push" })).toBe("Webhook push");
  });
});

describe("formatDuration", () => {
  it("derives from rendered_at → completed_at", () => {
    expect(formatDuration("2026-06-21T12:00:00Z", "2026-06-21T12:00:32Z")).toBe("32s");
    expect(formatDuration("2026-06-21T12:00:00Z", "2026-06-21T12:02:05Z")).toBe("2m 5s");
  });
  it("returns dash when missing", () => {
    expect(formatDuration(undefined, undefined)).toBe("—");
  });
});
