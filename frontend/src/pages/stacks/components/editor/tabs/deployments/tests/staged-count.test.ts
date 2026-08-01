import { describe, expect, it } from "vitest";
import { deriveDeployLifecycle } from "../use-deploy-lifecycle";
import type { Stack } from "@/api/stacks";
import type { StackRelease, StackReleaseSnapshot } from "@/api/releases";

/**
 * The deploy pill's count and the changes modal's rows come from one place, so
 * a number can never appear that the modal is unable to itemize — the state
 * that showed "Undeployed changes (1)" above "Loading changes…".
 */
const liveRelease = {
  id: "rel-1",
  sequence: 3,
  state: "Released",
  completed_at: "2026-01-01T00:00:00Z",
} as unknown as StackRelease;

const stack = {
  id: "s1",
  updated_at: "2026-01-02T00:00:00Z", // later than the release: the old drift heuristic
  converged_release: { id: "rel-1" },
  spec: {
    stack_resources: [
      { name: "web", source: { image: { ref: "web:1" } }, execution_config: { environment_variables: [] } },
    ],
    volumes: [],
    connections: [],
  },
} as unknown as Stack;

const snapshotOf = (ref: string) =>
  ({
    resources: [{ name: "web", source: { image: { ref } }, execution_config: { environment_variables: [] } }],
    volumes: [],
    connections: [],
  }) as unknown as StackReleaseSnapshot;

describe("staged changes", () => {
  it("offers no diff while the release snapshot it would compare against is still loading", () => {
    const lifecycle = deriveDeployLifecycle({
      stack,
      unsaved: false,
      activeRelease: liveRelease,
      liveRelease,
      liveSnapshot: undefined,
    });
    expect(lifecycle.stagedDiff).toBeUndefined();
  });

  it("counts exactly what it can itemize once the snapshot lands", () => {
    const lifecycle = deriveDeployLifecycle({
      stack,
      unsaved: false,
      activeRelease: liveRelease,
      liveRelease,
      liveSnapshot: snapshotOf("web:0"),
    });
    const diff = lifecycle.stagedDiff!;
    expect(diff.resources).toHaveLength(1);
    expect(diff.resources[0]).toMatchObject({ name: "web", change: "modified" });
  });

  it("reports nothing staged when the draft matches what is deployed", () => {
    const lifecycle = deriveDeployLifecycle({
      stack,
      unsaved: false,
      activeRelease: liveRelease,
      liveRelease,
      liveSnapshot: snapshotOf("web:1"),
    });
    expect(lifecycle.stagedDiff).toBeUndefined();
    expect(lifecycle.phase).toBe("clean");
  });
});
