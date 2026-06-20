import { describe, it, expect } from "vitest";
import type { Stack } from "@/api/stacks";
import { deriveFailingResources, deriveRecovered, humanizeFailureType, causeLabel, formatDuration } from "../derive";
import { deriveStages, releaseGitSha } from "../derive";
import type { StackRelease } from "@/api/releases";

function release(partial: Partial<StackRelease>): StackRelease {
  return { id: "r1", ...partial } as StackRelease;
}

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

  it("classifies an init-container failure to the init stage", () => {
    const stack = stackWith([
      { name: "migrate", status: { state: "Error", last_failure: {
        type: "runtime_crash", init_container: { failure_type: "exit_error", reason: "InitFailed", exit_code: 2 } } } },
    ]);
    expect(deriveFailingResources(stack)[0]).toMatchObject({ name: "migrate", stage: "init", reason: "InitFailed", exitCode: 2 });
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
  it("labels a rollback with no detail as plain Rollback", () => {
    expect(causeLabel({ kind: "rollback" })).toBe("Rollback");
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
  it("returns dash for a negative interval", () => {
    expect(formatDuration("2026-06-21T12:00:32Z", "2026-06-21T12:00:00Z")).toBe("—");
  });
});

describe("deriveStages", () => {
  const imagePins = { resources: { api: { git_sha: "9c69af2" } } };

  it("all done when converged to the active release", () => {
    const stack = { status: { last_converged: { release_id: "r1" } } } as unknown as import("../derive").Stack;
    expect(deriveStages(stack, release({ id: "r1", state: "Released", pins: imagePins }), []))
      .toEqual({ build: "done", deploy: "done", ready: "done" });
  });

  it("build active while Pending with build pins", () => {
    const stack = { status: {} } as unknown as import("../derive").Stack;
    expect(deriveStages(stack, release({ state: "Pending", pins: imagePins }), []))
      .toEqual({ build: "active", deploy: "todo", ready: "todo" });
  });

  it("build failed when a resource reports build_failure", () => {
    const stack = { status: {} } as unknown as import("../derive").Stack;
    const failing = [{ name: "api", type: "build_failure" as const, stage: "build" as const, reason: "x" }];
    expect(deriveStages(stack, release({ state: "InProgress", pins: imagePins }), failing))
      .toEqual({ build: "failed", deploy: "todo", ready: "todo" });
  });

  it("deploy failed when a runtime crash occurs (build already done)", () => {
    const stack = { status: {} } as unknown as import("../derive").Stack;
    const failing = [{ name: "api", type: "runtime_crash" as const, stage: "runtime" as const, reason: "x" }];
    expect(deriveStages(stack, release({ state: "InProgress", pins: imagePins }), failing))
      .toEqual({ build: "done", deploy: "failed", ready: "todo" });
  });

  it("pre-cluster Failed with no resource failure maps to Build ✕", () => {
    const stack = { status: {} } as unknown as import("../derive").Stack;
    expect(deriveStages(stack, release({ state: "Failed", pins: imagePins }), []))
      .toEqual({ build: "failed", deploy: "todo", ready: "todo" });
  });

  it("image-only stack (no build pins) starts at Deploy", () => {
    const stack = { status: {} } as unknown as import("../derive").Stack;
    expect(deriveStages(stack, release({ state: "InProgress", pins: { resources: { api: { image_digest: "sha256:..." } } } }), []))
      .toMatchObject({ build: "todo", deploy: "active" });
  });
});

describe("releaseGitSha", () => {
  it("returns the first non-empty git_sha from pins", () => {
    expect(releaseGitSha(release({ pins: { resources: { api: { git_sha: "abc1234" } } } }))).toBe("abc1234");
  });
  it("returns undefined when no git pins", () => {
    expect(releaseGitSha(release({ pins: { resources: { api: { image_digest: "x" } } } }))).toBeUndefined();
  });
});
