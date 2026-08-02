import { describe, it, expect, vi, afterEach } from "vitest";
import { deriveFailingResources, deriveRecovered, humanizeFailureType, causeLabel, formatDuration, formatReleaseTime, ResourceFailureType } from "../derive";
import { deriveStages, deriveReleaseTitle, releaseGitSha, phaseTone, toneTextClass, toneDotClass, stateTone, toneFromVariant, BUILD_READY_CONDITION, CONDITION_TRUE } from "../derive";
import { deriveHeaderHealth, latestDeployFailed, stackSummariesStale, stripUnpinnedGitRevisions } from "../derive";
import type { FailingResource, Stack, StackResource } from "../derive";
import type { StackRelease, ReleaseLiveStatus, ReleaseSummary } from "@/api/releases";

function release(partial: Partial<StackRelease>): StackRelease {
  return { id: "r1", ...partial } as StackRelease;
}

function liveStatusWith(resources: Record<string, Record<string, unknown>>): ReleaseLiveStatus {
  return { resources } as unknown as ReleaseLiveStatus;
}

function readinessLiveStatus(state: string): ReleaseLiveStatus {
  return liveStatusWith({
    web: {
      state,
      conditions: [{ type: BUILD_READY_CONDITION, status: CONDITION_TRUE }],
      last_failure: {
        type: ResourceFailureType.Readiness, container: { failure_type: "port_not_listening", reason: "PortNotListening" } },
    },
  });
}

function stackWithReleases(current?: Partial<ReleaseSummary>, latest?: Partial<ReleaseSummary>): Stack {
  return { converged_release: current, latest_release: latest } as unknown as Stack;
}

describe("deriveFailingResources", () => {
  it("joins last_failure from live_status.resources", () => {
    const liveStatus = liveStatusWith({
      tooljet: { state: "CrashLoopBackOff", last_failure: {
        type: "runtime_crash", container: { failure_type: "crash_loop", reason: "CrashLoopBackOff", message: "exit 1", exit_code: 1, restart_count: 5 } } },
      redis: { state: "Ready" },
    });
    const out = deriveFailingResources(release({}), liveStatus);
    expect(out).toHaveLength(1);
    expect(out[0]).toMatchObject({ name: "tooljet", type: "runtime_crash", stage: "runtime", reason: "CrashLoopBackOff", exitCode: 1, restartCount: 5 });
  });

  it("classifies a build failure to the build stage", () => {
    const liveStatus = liveStatusWith({
      api: { state: "Error", last_failure: {
        type: "build_failure", build: { failure_type: "image_pull_failed", reason: "ErrImagePull", message: "manifest unknown" } } },
    });
    expect(deriveFailingResources(release({}), liveStatus)[0]).toMatchObject({ name: "api", type: "build_failure", stage: "build", reason: "ErrImagePull" });
  });

  it("classifies an init-container failure to the init stage", () => {
    const liveStatus = liveStatusWith({
      migrate: { state: "Error", last_failure: {
        type: "runtime_crash", init_container: { failure_type: "exit_error", reason: "InitFailed", exit_code: 2 } } },
    });
    expect(deriveFailingResources(release({}), liveStatus)[0]).toMatchObject({ name: "migrate", stage: "init", reason: "InitFailed", exitCode: 2 });
  });

  it("maps a readiness failure without a restart count", () => {
    const liveStatus = liveStatusWith({
      web: { state: "Pending", last_failure: {
        type: ResourceFailureType.Readiness, container: { failure_type: "port_not_listening", reason: "PortNotListening", message: "readiness check failed: nothing listening on port 8080", restart_count: 0 } } },
    });
    const out = deriveFailingResources(release({}), liveStatus);
    expect(out).toHaveLength(1);
    expect(out[0]).toMatchObject({ name: "web", type: ResourceFailureType.Readiness, reason: "PortNotListening", failureType: "port_not_listening" });
    expect(out[0].restartCount).toBeUndefined();
    expect(out[0].exitCode).toBeUndefined();
  });

  it("returns nothing without a live status (not live or active)", () => {
    expect(deriveFailingResources(release({}))).toEqual([]);
  });
});

describe("deriveRecovered", () => {
  it("flags a Ready resource that still carries last_failure", () => {
    const liveStatus = liveStatusWith({
      tooljet: { state: "Ready", last_failure: {
        type: "runtime_crash", container: { reason: "CrashLoopBackOff", restart_count: 5 } } },
    });
    expect(deriveRecovered(release({}), liveStatus)).toEqual([{ name: "tooljet", reason: "CrashLoopBackOff", restartCount: 5 }]);
  });

  it("drops the restart count for a recovered readiness failure", () => {
    const liveStatus = liveStatusWith({
      web: { state: "Ready", last_failure: {
        type: ResourceFailureType.Readiness, container: { failure_type: "port_not_listening", reason: "PortNotListening", restart_count: 0 } } },
    });
    expect(deriveRecovered(release({}), liveStatus)).toEqual([{ name: "web", reason: "PortNotListening", restartCount: undefined }]);
  });

  it("does not flag a failing resource as recovered", () => {
    const liveStatus = liveStatusWith({
      tooljet: { state: "CrashLoopBackOff", last_failure: { type: "runtime_crash", container: { reason: "x" } } },
    });
    expect(deriveRecovered(release({}), liveStatus)).toEqual([]);
  });
});

describe("humanizeFailureType", () => {
  it("maps known types", () => {
    expect(humanizeFailureType("out_of_memory")).toBe("Out of memory");
    expect(humanizeFailureType("crash_loop")).toBe("Crash loop");
  });
  it("labels a port_not_listening detail", () => {
    expect(humanizeFailureType("port_not_listening")).toBe("Port not listening");
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
  it("extracts the sequence from the backend's sentence detail", () => {
    // Backend sends detail as "rollback to release #1" — must not double-prefix.
    expect(causeLabel({ kind: "rollback", detail: "rollback to release #1" })).toBe("Rollback to #1");
  });
});

describe("tones derive from statusVariant", () => {
  it("toneFromVariant collapses 5 variants to 4 tones", () => {
    expect(toneFromVariant("ready")).toBe("ok");
    expect(toneFromVariant("pending")).toBe("amber");
    expect(toneFromVariant("error")).toBe("err");
    expect(toneFromVariant("info")).toBe("muted");
    expect(toneFromVariant("neutral")).toBe("muted");
  });
});

describe("stateTone", () => {
  it("colors the rail dot by lifecycle state", () => {
    expect(stateTone("Released")).toBe("ok");
    expect(stateTone("Failed")).toBe("err");
    expect(stateTone("Pending")).toBe("amber");
    expect(stateTone("InProgress")).toBe("amber");
    expect(stateTone("Superseded")).toBe("muted");
    expect(stateTone("Cancelled")).toBe("muted");
  });
});

describe("formatReleaseTime", () => {
  afterEach(() => vi.useRealTimers());

  it("buckets by day with full 'today' / 'yesterday' words", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-06-22T10:00:00"));
    expect(formatReleaseTime("2026-06-22T00:33:00")).toBe("today 00:33");
    expect(formatReleaseTime("2026-06-21T23:56:00")).toBe("yesterday 23:56");
    expect(formatReleaseTime("2026-06-03T09:02:00")).toBe("Jun 3 09:02");
  });

  it("returns empty for missing or invalid input", () => {
    expect(formatReleaseTime(undefined)).toBe("");
    expect(formatReleaseTime("not-a-date")).toBe("");
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

  it("all done when the release is Released", () => {
    expect(deriveStages(release({ id: "r1", state: "Released", pins: imagePins }), []))
      .toEqual({ build: "done", deploy: "done", ready: "done" });
  });

  it("live status presence does NOT mark an in-flight release converged (overlay is present for active releases too)", () => {
    const liveStatus = { resources: {} } as ReleaseLiveStatus;
    expect(deriveStages(release({ id: "r1", state: "InProgress", pins: imagePins }), [], liveStatus))
      .toEqual({ build: "active", deploy: "todo", ready: "todo" });
  });

  it("build stays active while a resource is still waiting on its image build", () => {
    const liveStatus = {
      resources: {
        redis: { conditions: [{ type: "Available", status: CONDITION_TRUE }] },
        worker: { conditions: [{ type: BUILD_READY_CONDITION, status: "False" }] },
      },
    } as unknown as ReleaseLiveStatus;
    expect(deriveStages(release({ state: "InProgress", pins: imagePins }), [], liveStatus))
      .toEqual({ build: "active", deploy: "todo", ready: "todo" });
  });

  it("build stays active until the agent publishes a build condition at all", () => {
    const liveStatus = {
      resources: { redis: { conditions: [{ type: "Available", status: CONDITION_TRUE }] } },
    } as unknown as ReleaseLiveStatus;
    expect(deriveStages(release({ state: "InProgress", pins: imagePins }), [], liveStatus))
      .toEqual({ build: "active", deploy: "todo", ready: "todo" });
  });

  it("build moves on once every build reports ready", () => {
    const liveStatus = {
      resources: { worker: { conditions: [{ type: BUILD_READY_CONDITION, status: CONDITION_TRUE }] } },
    } as unknown as ReleaseLiveStatus;
    expect(deriveStages(release({ state: "InProgress", pins: imagePins }), [], liveStatus))
      .toEqual({ build: "done", deploy: "active", ready: "todo" });
  });

  it("build active while Pending with build pins", () => {
    expect(deriveStages(release({ state: "Pending", pins: imagePins }), []))
      .toEqual({ build: "active", deploy: "todo", ready: "todo" });
  });

  it("build failed when a resource reports build_failure", () => {
    const failing = [{ name: "api", type: "build_failure" as const, stage: "build" as const, reason: "x" }];
    expect(deriveStages(release({ state: "InProgress", pins: imagePins }), failing))
      .toEqual({ build: "failed", deploy: "todo", ready: "todo" });
  });

  it("deploy failed when a runtime crash occurs (build already done)", () => {
    const failing = [{ name: "api", type: "runtime_crash" as const, stage: "runtime" as const, reason: "x" }];
    expect(deriveStages(release({ state: "InProgress", pins: imagePins }), failing))
      .toEqual({ build: "done", deploy: "failed", ready: "todo" });
  });

  it("deploy failed when a condemned readiness failure occurs", () => {
    const rel = release({ state: "InProgress", pins: imagePins });
    const live = readinessLiveStatus("Failed");
    const failing = deriveFailingResources(rel, live);
    expect(deriveStages(rel, failing, live)).toEqual({ build: "done", deploy: "failed", ready: "todo" });
  });

  it("deploy still active for a provisional readiness failure mid-rollout", () => {
    const rel = release({ state: "InProgress", pins: imagePins });
    const live = readinessLiveStatus("Pending");
    const failing = deriveFailingResources(rel, live);
    expect(deriveStages(rel, failing, live)).toEqual({ build: "done", deploy: "active", ready: "todo" });
  });

  it("terminal Failed runtime crash maps to Deploy done / Ready ✕", () => {
    const failing = [{ name: "api", type: "runtime_crash" as const, stage: "runtime" as const, reason: "CrashLoopBackOff" }];
    expect(deriveStages(release({ state: "Failed", pins: imagePins }), failing))
      .toEqual({ build: "done", deploy: "done", ready: "failed" });
  });

  it("pre-cluster Failed with no resource failure maps to Build ✕", () => {
    expect(deriveStages(release({ state: "Failed", pins: imagePins }), []))
      .toEqual({ build: "failed", deploy: "todo", ready: "todo" });
  });

  it("convergence timeout (outcome recorded) maps to Deploy done / Ready ✕, Build skipped for image-only", () => {
    const rel = release({ state: "Failed", message: "timed out waiting for convergence after 15m0s",
      pins: { resources: { web: { image_digest: "sha256:..." } } },
      outcome: { resources: { web: { phase: "Ready", ready_replicas: 1, replicas: 2 } } } });
    expect(deriveStages(rel, [])).toEqual({ build: "skipped", deploy: "done", ready: "failed" });
  });

  it("pre-cluster Failed on an image-only stack lands on Deploy, never Build", () => {
    const rel = release({ state: "Failed", pins: { resources: { web: { image_digest: "sha256:..." } } } });
    expect(deriveStages(rel, [])).toEqual({ build: "skipped", deploy: "failed", ready: "todo" });
  });

  it("image-only stack (no build pins) skips Build and starts at Deploy", () => {
    expect(deriveStages(release({ state: "InProgress", pins: { resources: { api: { image_digest: "sha256:..." } } } }), []))
      .toEqual({ build: "skipped", deploy: "active", ready: "todo" });
  });

  it("Superseded returns all todo", () => {
    expect(deriveStages(release({ state: "Superseded", pins: imagePins }), []))
      .toEqual({ build: "todo", deploy: "todo", ready: "todo" });
  });

  it("build failed even while Pending", () => {
    const failing = [{ name: "api", type: "build_failure" as const, stage: "build" as const, reason: "x" }];
    expect(deriveStages(release({ state: "Pending", pins: imagePins }), failing))
      .toEqual({ build: "failed", deploy: "todo", ready: "todo" });
  });
});

describe("deriveReleaseTitle", () => {
  const buildFail: FailingResource = { name: "api", type: "build_failure", stage: "build", reason: "x" };
  const crash: FailingResource = { name: "tooljet", type: "runtime_crash", stage: "runtime", reason: "x" };
  const stages = { build: "done", deploy: "done", ready: "done" } as const;

  it("names the failing resource for build / runtime crashes", () => {
    expect(deriveReleaseTitle(release({ state: "InProgress" }), [buildFail], stages)).toBe("Build failed: api");
    expect(deriveReleaseTitle(release({ state: "InProgress" }), [crash], stages)).toBe("Runtime crash: tooljet");
  });
  it("names the port failure for a condemned readiness failure, but not a provisional one", () => {
    const rel = release({ state: "InProgress" });
    const condemned = deriveFailingResources(rel, readinessLiveStatus("Failed"));
    const provisional = deriveFailingResources(rel, readinessLiveStatus("Pending"));
    expect(deriveReleaseTitle(rel, condemned, stages)).toBe("Port not listening: web");
    expect(deriveReleaseTitle(rel, provisional, stages)).toBe("Deploying");
  });
  it("a terminal crash reads as Deploy failed", () => {
    expect(deriveReleaseTitle(release({ state: "Failed" }), [crash], stages)).toBe("Deploy failed");
  });
  it("labels lifecycle states with no failures", () => {
    expect(deriveReleaseTitle(release({ state: "Released" }), [], stages)).toBe("Released");
    expect(deriveReleaseTitle(release({ state: "Pending" }), [], { build: "active", deploy: "todo", ready: "todo" })).toBe("Build queued");
    expect(deriveReleaseTitle(release({ state: "Pending" }), [], { build: "skipped", deploy: "active", ready: "todo" })).toBe("Deploying");
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

describe("phaseTone", () => {
  it("maps phases from rollout domain via statusVariant", () => {
    expect(phaseTone("Ready")).toBe("ok");
    expect(phaseTone("Pending")).toBe("amber");
    expect(phaseTone("Degraded")).toBe("err");
    expect(phaseTone("Failed")).toBe("err");
    expect(phaseTone("Unknown")).toBe("muted");
  });
  it("maps tones to brand token classes", () => {
    expect(toneTextClass("ok")).toBe("text-success");
    expect(toneTextClass("err")).toBe("text-danger");
    expect(toneTextClass("amber")).toBe("text-warn");
    expect(toneDotClass("muted")).toBe("bg-fg-muted");
  });
});

describe("deriveHeaderHealth", () => {
  it("undefined when no releases exist (never deployed)", () => {
    expect(deriveHeaderHealth(stackWithReleases(undefined, undefined))).toBeUndefined();
  });

  it("progressing while the latest release is in flight", () => {
    expect(deriveHeaderHealth(stackWithReleases(undefined, { id: "r1", state: "InProgress" }))).toBe("progressing");
    expect(deriveHeaderHealth(stackWithReleases({ id: "r0", health: "ok" }, { id: "r1", state: "Pending" }))).toBe("progressing");
  });

  it("live release health once the latest is terminal", () => {
    expect(deriveHeaderHealth(stackWithReleases({ id: "r1", health: "ok" }, { id: "r1", state: "Released" }))).toBe("ok");
    expect(deriveHeaderHealth(stackWithReleases({ id: "r1", health: "degraded" }, { id: "r1", state: "Released" }))).toBe("degraded");
  });

  it("live health wins over a newer failed attempt (healthy current + failed latest)", () => {
    expect(deriveHeaderHealth(stackWithReleases({ id: "r1", health: "ok" }, { id: "r2", state: "Failed" }))).toBe("ok");
  });

  it("failed first deploy (no live release) falls back to 'failed'", () => {
    expect(deriveHeaderHealth(stackWithReleases(undefined, { id: "r1", state: "Failed" }))).toBe("failed");
  });

  it("cancelled/superseded-only history (nothing ever ran) stays undefined → 'Not deployed'", () => {
    expect(deriveHeaderHealth(stackWithReleases(undefined, { id: "r1", state: "Cancelled" }))).toBeUndefined();
    expect(deriveHeaderHealth(stackWithReleases(undefined, { id: "r1", state: "Superseded" }))).toBeUndefined();
  });
});

describe("latestDeployFailed", () => {
  it("true when the latest release failed while a different release stays live", () => {
    expect(latestDeployFailed(stackWithReleases({ id: "r1", health: "ok" }, { id: "r2", state: "Failed" }))).toBe(true);
  });

  it("false when nothing is live (failed first deploy — main pill already reads failed)", () => {
    expect(latestDeployFailed(stackWithReleases(undefined, { id: "r1", state: "Failed" }))).toBe(false);
  });

  it("false when the live release itself is the failed latest", () => {
    expect(latestDeployFailed(stackWithReleases({ id: "r1", health: "failed" }, { id: "r1", state: "Failed" }))).toBe(false);
  });

  it("false for non-failed latest states and never-deployed stacks", () => {
    expect(latestDeployFailed(stackWithReleases({ id: "r1", health: "ok" }, { id: "r2", state: "InProgress" }))).toBe(false);
    expect(latestDeployFailed(stackWithReleases({ id: "r1", health: "ok" }, { id: "r1", state: "Released" }))).toBe(false);
    expect(latestDeployFailed(stackWithReleases(undefined, undefined))).toBe(false);
  });

  it("false when the main pill already reads failed (no double error)", () => {
    expect(latestDeployFailed(stackWithReleases({ id: "r1", health: "failed" }, { id: "r2", state: "Failed" }))).toBe(false);
    expect(latestDeployFailed(stackWithReleases({ id: "r1" }, { id: "r2", state: "Failed" }))).toBe(false);
  });
});

describe("stackSummariesStale", () => {
  it("false with no active release or a non-terminal one", () => {
    expect(stackSummariesStale(undefined, stackWithReleases(undefined, undefined))).toBe(false);
    expect(stackSummariesStale(release({ state: "InProgress" }), stackWithReleases(undefined, undefined))).toBe(false);
  });

  it("stays true while latest_release lags a terminal release, regardless of prior observations", () => {
    // The backend updates release summaries asynchronously after convergence; a
    // single "handled" refetch that raced the pointer must not end the retries.
    for (const state of ["Released", "Failed", "Cancelled", "Superseded"] as const) {
      expect(stackSummariesStale(
        release({ id: "r2", state }),
        stackWithReleases({ id: "r1" }, { id: "r2", state: "InProgress" }),
      )).toBe(true);
    }
  });

  it("true when latest_release caught up but converged_release still lags a Released release", () => {
    expect(stackSummariesStale(
      release({ id: "r2", state: "Released" }),
      stackWithReleases({ id: "r1" }, { id: "r2", state: "Released" }),
    )).toBe(true);
  });

  it("false when fully caught up (converged matches for Released)", () => {
    expect(stackSummariesStale(
      release({ id: "r2", state: "Released" }),
      stackWithReleases({ id: "r2" }, { id: "r2", state: "Released" }),
    )).toBe(false);
  });

  it("false for a Failed release once latest_release matches — converged stays on the old live release", () => {
    expect(stackSummariesStale(
      release({ id: "r2", state: "Failed" }),
      stackWithReleases({ id: "r1" }, { id: "r2", state: "Failed" }),
    )).toBe(false);
  });

  it("false with no stack loaded yet", () => {
    expect(stackSummariesStale(release({ id: "r2", state: "Released" }), undefined)).toBe(false);
  });
});

describe("stripUnpinnedGitRevisions", () => {
  function gitResource(name: string, git: Record<string, unknown>): StackResource {
    return { name, source: { git } } as unknown as StackResource;
  }

  it("strips resolver-written branch/commit/tag when the saved spec pins nothing, keeping other git fields", () => {
    const snapshot = [gitResource("api", { repo_url: "https://x/y", branch: "main", commit: "abc123", dockerfile_path: "Dockerfile" })];
    const saved = [gitResource("api", { repo_url: "https://x/y", dockerfile_path: "Dockerfile" })];
    const [out] = stripUnpinnedGitRevisions(snapshot, saved);
    expect(out.source?.git).toEqual({ repo_url: "https://x/y", dockerfile_path: "Dockerfile" });
  });

  it("strips only the resolver-written commit when the saved spec pins a branch, keeping the branch", () => {
    const snapshot = [gitResource("api", { repo_url: "https://x/y", branch: "main", commit: "abc123" })];
    const saved = [gitResource("api", { repo_url: "https://x/y", branch: "main" })];
    const [out] = stripUnpinnedGitRevisions(snapshot, saved);
    expect(out.source?.git).toEqual({ repo_url: "https://x/y", branch: "main" });
  });

  it("passes a snapshot resource through unchanged (same object reference) when the saved spec pins every revision key present", () => {
    const pinned = gitResource("api", { repo_url: "https://x/y", branch: "main", commit: "abc123" });
    const saved = [gitResource("api", { repo_url: "https://x/y", branch: "main", commit: "abc123" })];
    const [out] = stripUnpinnedGitRevisions([pinned], saved);
    expect(out).toBe(pinned);
  });

  it("leaves a non-git-sourced resource untouched", () => {
    const image = { name: "redis", source: { image: { image: "redis:7" } } } as unknown as StackResource;
    const saved = [{ name: "redis", source: { image: { image: "redis:7" } } } as unknown as StackResource];
    const [out] = stripUnpinnedGitRevisions([image], saved);
    expect(out).toBe(image);
  });

  it("leaves a snapshot resource untouched when it has no counterpart in the saved spec", () => {
    const orphan = gitResource("gone", { repo_url: "https://x/y", branch: "main", commit: "abc123" });
    const [out] = stripUnpinnedGitRevisions([orphan], []);
    expect(out).toBe(orphan);
  });
});
