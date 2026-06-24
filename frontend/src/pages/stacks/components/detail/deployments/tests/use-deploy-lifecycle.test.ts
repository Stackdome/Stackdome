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
      liveSnapshot: snap("nginx:1.25"),
    });
    expect(r.phase).toBe("editing");
  });

  it("deploying — active release is non-terminal", () => {
    const r = deriveDeployLifecycle({
      stack: mkStack("nginx:1.27"),
      dirty: cleanDirty(),
      isActive: false,
      activeRelease: mkRelease({ sequence: 8, state: "InProgress" }),
      liveRelease: mkRelease({ id: "live-1", sequence: 7 }),
      liveSnapshot: snap("nginx:1.25"),
    });
    expect(r.phase).toBe("deploying");
  });

  it("staged — saved spec differs from the live snapshot, with a per-field diff", () => {
    const r = deriveDeployLifecycle({
      stack: mkStack("nginx:1.27"),
      dirty: cleanDirty(),
      isActive: false,
      activeRelease: mkRelease({ sequence: 7, state: "Released" }),
      liveRelease: mkRelease({ id: "live-1", sequence: 7 }),
      liveSnapshot: snap("nginx:1.25"),
    });
    expect(r.phase).toBe("staged");
    expect(r.stagedDiff?.resources).toHaveLength(1);
    expect(r.stagedDiff?.resources[0]).toMatchObject({ name: "web", change: "modified" });
    expect(r.nextSeq).toBe(8);
    expect(r.liveSeq).toBe(7);
  });

  it("clean — saved spec matches the live snapshot", () => {
    const r = deriveDeployLifecycle({
      stack: mkStack("nginx:1.25"),
      dirty: cleanDirty(),
      isActive: false,
      activeRelease: mkRelease({ sequence: 7, state: "Released" }),
      liveRelease: mkRelease({ id: "live-1", sequence: 7 }),
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
      liveSnapshot: undefined,
    });
    expect(r.phase).toBe("staged");
    expect(r.nextSeq).toBe(1);
  });

  it("staged (heuristic) — live release exists but snapshot unloaded, stack updated after it", () => {
    const r = deriveDeployLifecycle({
      stack: mkStack("nginx:1.27", "2026-06-25T02:00:00Z"),
      dirty: cleanDirty(),
      isActive: false,
      activeRelease: mkRelease({ sequence: 7, state: "Released" }),
      liveRelease: mkRelease({ id: "live-1", sequence: 7, completed_at: "2026-06-25T01:00:00Z" }),
      liveSnapshot: undefined,
    });
    expect(r.phase).toBe("staged");
    expect(r.stagedDiff).toBeUndefined();
  });

  it("clean (heuristic) — live release exists, snapshot unloaded, no drift", () => {
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
