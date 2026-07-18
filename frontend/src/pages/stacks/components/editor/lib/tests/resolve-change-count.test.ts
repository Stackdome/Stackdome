import { describe, it, expect } from "vitest";
import { resolveChangeCount } from "@/pages/stacks/components/editor/lib/resolve-change-count";
import { SYNC_STATUS } from "@/pages/stacks/lib/draft-sync/constants";

const diff = (r: number, v: number, c: number) => ({
  resources: new Array(r).fill(0),
  volumes: new Array(v).fill(0),
  connections: new Array(c).fill(0),
});

describe("resolveChangeCount", () => {
  it("trusts the staged diff when the sync is settled", () => {
    // session dirt can overcount (e.g. a mount added then removed nets to zero staged)
    expect(resolveChangeCount(diff(2, 1, 0), 99, SYNC_STATUS.saved)).toBe(3);
  });

  it("falls back to session dirt when no staged diff has resolved yet", () => {
    expect(resolveChangeCount(null, 4, SYNC_STATUS.saved)).toBe(4);
  });

  it("never lets a stale staged count hide dirt while a sync is in flight", () => {
    expect(resolveChangeCount(diff(0, 0, 0), 2, SYNC_STATUS.saving)).toBe(2);
  });

  it("never lets a stale staged count hide dirt after a sync error", () => {
    expect(resolveChangeCount(diff(1, 0, 0), 3, SYNC_STATUS.error)).toBe(3);
  });

  it("takes the staged count while unsettled when it is the larger", () => {
    expect(resolveChangeCount(diff(5, 0, 0), 2, SYNC_STATUS.saving)).toBe(5);
  });
});
