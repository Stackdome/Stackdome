import { describe, it, expect } from "vitest";
import { deriveDeployLifecycle } from "../use-deploy-lifecycle";
import type { Stack } from "@/api/stacks";
import type { StackRelease, StackReleaseSnapshot } from "@/api/releases";
import type { StackDiff } from "@/pages/stacks/lib/stack-diff";

const cleanDirty = (): StackDiff => ({
  dirtyResourceIdx: new Set(),
  dirtyVolumeIdx: new Set(),
  perResourceDirty: new Map(),
  perVolumeDirty: new Map(),
  addonLinkCount: 0,
});

const dirty = (): StackDiff => ({ ...cleanDirty(), dirtyResourceIdx: new Set([0]) });

const mkStack = (image: string, updatedAt?: string): Stack =>
  ({
    spec: { stack_resources: [{ name: "web", image_spec: { image } }] },
    status: { last_converged: { release_id: "live-1" } },
    updated_at: updatedAt,
  } as unknown as Stack);

const mkRelease = (over: Partial<StackRelease>): StackRelease =>
  ({ id: "r", sequence: 7, state: "Released", ...over } as StackRelease);

const snap = (image: string): StackReleaseSnapshot => ({ resources: [{ name: "web", image_spec: { image } }] });

describe("deriveDeployLifecycle", () => {
  it("editing — active session with dirty edits, even while a deploy is in flight", () => {
    const r = deriveDeployLifecycle({
      stack: mkStack("nginx:1.27"),
      dirty: dirty(),
      isActive: true,
      activeRelease: mkRelease({ sequence: 8, state: "InProgress" }),
      liveRelease: mkRelease({ id: "live-1", sequence: 7 }),
      activeSnapshot: snap("nginx:1.27"),
      liveSnapshot: snap("nginx:1.25"),
    });
    expect(r.phase).toBe("editing");
  });

  it("deploying — the in-flight release is shipping exactly the saved spec", () => {
    const r = deriveDeployLifecycle({
      stack: mkStack("nginx:1.27"),
      dirty: cleanDirty(),
      isActive: false,
      activeRelease: mkRelease({ sequence: 8, state: "InProgress" }),
      liveRelease: mkRelease({ id: "live-1", sequence: 7 }),
      activeSnapshot: snap("nginx:1.27"), // == saved spec → nothing new
      liveSnapshot: snap("nginx:1.25"),
    });
    expect(r.phase).toBe("deploying");
    expect(r.nextSeq).toBe(9);
    expect(r.liveSeq).toBe(7);
  });

  it("staged — a fresh edit made mid-deploy supersedes the in-flight release", () => {
    const r = deriveDeployLifecycle({
      stack: mkStack("nginx:1.28"), // newer than the in-flight #8
      dirty: cleanDirty(),
      isActive: false,
      activeRelease: mkRelease({ sequence: 8, state: "InProgress" }),
      liveRelease: mkRelease({ id: "live-1", sequence: 7 }),
      activeSnapshot: snap("nginx:1.27"), // in-flight ships 1.27, saved is 1.28
      liveSnapshot: snap("nginx:1.25"),
    });
    expect(r.phase).toBe("staged");
    expect(r.stagedDiff?.resources).toHaveLength(1);
  });

  it("staged — retrying a failed release whose spec still isn't live", () => {
    const r = deriveDeployLifecycle({
      stack: mkStack("nginx:1.27"),
      dirty: cleanDirty(),
      isActive: false,
      activeRelease: mkRelease({ sequence: 8, state: "Failed" }),
      liveRelease: mkRelease({ id: "live-1", sequence: 7 }),
      activeSnapshot: snap("nginx:1.27"), // == saved, but the attempt failed and isn't live
      liveSnapshot: snap("nginx:1.25"),
    });
    expect(r.phase).toBe("staged");
  });

  it("staged — saved spec differs from the live snapshot (no deploy in flight)", () => {
    const r = deriveDeployLifecycle({
      stack: mkStack("nginx:1.27"),
      dirty: cleanDirty(),
      isActive: false,
      activeRelease: mkRelease({ sequence: 7, state: "Released" }),
      liveRelease: mkRelease({ id: "live-1", sequence: 7 }),
      liveSnapshot: snap("nginx:1.25"),
    });
    expect(r.phase).toBe("staged");
    expect(r.stagedDiff?.resources[0]).toMatchObject({ name: "web", change: "modified" });
  });

  it("clean — saved spec matches the live snapshot", () => {
    const r = deriveDeployLifecycle({
      stack: mkStack("nginx:1.25"),
      dirty: cleanDirty(),
      isActive: false,
      activeRelease: mkRelease({ sequence: 7, state: "Released" }),
      liveRelease: mkRelease({ id: "live-1", sequence: 7 }),
      activeSnapshot: snap("nginx:1.25"),
      liveSnapshot: snap("nginx:1.25"),
    });
    expect(r.phase).toBe("clean");
    expect(r.stagedDiff).toBeUndefined();
  });

  it("staged (first deploy) — never deployed, saved resources present, no snapshot", () => {
    const r = deriveDeployLifecycle({
      stack: { spec: { stack_resources: [{ name: "web", image_spec: { image: "nginx:1.27" } }] }, status: {} } as unknown as Stack,
      dirty: cleanDirty(),
      isActive: false,
      activeRelease: undefined,
      liveRelease: undefined,
    });
    expect(r.phase).toBe("staged");
    expect(r.nextSeq).toBe(1);
  });

  it("deploying — in-flight with its snapshot not loaded yet stays a status (no flash)", () => {
    const r = deriveDeployLifecycle({
      stack: mkStack("nginx:1.27"),
      dirty: cleanDirty(),
      isActive: false,
      activeRelease: mkRelease({ sequence: 8, state: "InProgress" }),
      liveRelease: mkRelease({ id: "live-1", sequence: 7 }),
      activeSnapshot: undefined,
      liveSnapshot: undefined,
    });
    expect(r.phase).toBe("deploying");
  });

  it("clean (heuristic) — live snapshot unloaded, no drift, nothing in flight", () => {
    const r = deriveDeployLifecycle({
      stack: mkStack("nginx:1.27", "2026-06-25T00:30:00Z"),
      dirty: cleanDirty(),
      isActive: false,
      activeRelease: mkRelease({ sequence: 7, state: "Released" }),
      liveRelease: mkRelease({ id: "live-1", sequence: 7, completed_at: "2026-06-25T01:00:00Z" }),
      liveSnapshot: undefined,
    });
    expect(r.phase).toBe("clean");
  });
});
