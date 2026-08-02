import { describe, expect, it } from "vitest";
import type { Stack } from "@/api/stacks";
import type { StackReleaseSnapshot } from "@/api/releases";
import { canonicalFromSnapshot, canonicalFromStack } from "../from-api";
import { canonicalFromDraft } from "../from-form";
import { diffStacks } from "../diff";
import { computeSyncOps } from "@/pages/stacks/lib/draft-sync/ops";
import { serverConnectionIndex } from "@/pages/stacks/lib/draft-sync/server-state";
import { formResourcesFromSpec, mapVolumeToFormData } from "@/pages/stacks/lib/spec-to-form";

/**
 * The other direction from `invariants.test.ts`.
 *
 * Those cases all assert that loading a stack changes nothing, which catches a
 * rule that suppresses too little — a phantom change. They are blind by
 * construction to a rule that suppresses too MUCH, and that blindness shipped:
 * `dropResolvedRevisions` ran against every baseline, so clearing a git pin
 * produced no diff, no write, and no tint. The edit was discarded on reload.
 *
 * So: one real edit per policy rule, and the change must survive to the server.
 */

const pinned = {
  id: "r-worker",
  name: "worker",
  workload_type: "Service",
  source: {
    git: {
      repo_url: "https://github.com/acme/demo.git",
      branch: "main",
      commit: "abc123",
      dockerfile_path: "Dockerfile",
      build_context: ".",
    },
  },
  ports: [],
  volume_mounts: [],
  depends_on: [],
  execution_config: { command: [], args: [], environment_variables: [] },
};

const stackOf = (resources: unknown[]): Stack =>
  ({
    id: "s1",
    name: "demo",
    spec: { stack_resources: resources, volumes: [], connections: [] },
  }) as unknown as Stack;

function draftOf(stack: Stack) {
  return {
    resources: formResourcesFromSpec(stack.spec?.stack_resources, stack.spec?.connections),
    volumes: (stack.spec?.volumes ?? []).map(mapVolumeToFormData),
  };
}

/** What autosave would send for a stack whose draft differs from the server. */
function opsFor(server: Stack, draft: Stack) {
  return computeSyncOps(
    canonicalFromStack(server),
    canonicalFromDraft(draftOf(draft)),
    serverConnectionIndex(server),
  );
}

const gitEdits: Array<[string, Record<string, unknown>, string]> = [
  ["clearing the commit pin", { commit: undefined }, "source.git.commit"],
  ["changing the commit pin", { commit: "def456" }, "source.git.commit"],
  ["clearing the branch and its pin", { branch: undefined, commit: undefined }, "source.git.branch"],
  ["switching branch to tag", { branch: undefined, commit: undefined, tag: "v1" }, "source.git.tag"],
  ["changing the dockerfile path", { dockerfile_path: "docker/Dockerfile" }, "source.git.dockerfile_path"],
  ["changing the build context", { build_context: "svc" }, "source.git.build_context"],
];

describe("a real edit survives to the server", () => {
  for (const [label, patch, expectedPath] of gitEdits) {
    it(label, () => {
      const edited = stackOf([{ ...pinned, source: { git: { ...pinned.source.git, ...patch } } }]);
      const diff = diffStacks(canonicalFromStack(stackOf([pinned])), canonicalFromDraft(draftOf(edited)));

      expect(diff.resources).toHaveLength(1);
      expect(diff.resources[0].fields.map((f) => f.path)).toContain(expectedPath);
      expect(opsFor(stackOf([pinned]), edited).map((o) => o.kind)).toContain("updateResource");
    });
  }

  /**
   * KNOWN LIMITATION. Against a release baseline the diff sees only the draft,
   * and "this spec never pinned a commit" is indistinguishable there from "the
   * user just cleared the pin" — both arrive as an absent commit. So the
   * undeployed-changes card omits that one row.
   *
   * It is cosmetic, not lost work: autosave compares against the saved spec and
   * writes the change (asserted above), and the drawer tints the field, because
   * its baseline is stripped against the saved spec by stripUnpinnedGitRevisions
   * rather than against the draft. Closing it means threading the saved spec
   * into diffSnapshots.
   */
  it("hides a cleared pin from the release card, but still writes it", () => {
    const deployed = { resources: [pinned], volumes: [], connections: [] } as unknown as StackReleaseSnapshot;
    const unpinned = stackOf([
      { ...pinned, source: { git: { ...pinned.source.git, commit: undefined } } },
    ]);
    const card = diffStacks(canonicalFromSnapshot(deployed), canonicalFromDraft(draftOf(unpinned)), {
      baselineIsRelease: true,
    });
    expect(card.resources).toEqual([]);
    expect(opsFor(stackOf([pinned]), unpinned).map((o) => o.kind)).toContain("updateResource");
  });

  it("but a revision the spec never pinned stays hidden", () => {
    const spec = { ...pinned, source: { git: { ...pinned.source.git, commit: undefined } } };
    const resolved = {
      resources: [{ ...spec, source: { git: { ...spec.source.git, commit: "9f1c2b7" } } }],
      volumes: [],
      connections: [],
    } as unknown as StackReleaseSnapshot;
    const diff = diffStacks(canonicalFromSnapshot(resolved), canonicalFromDraft(draftOf(stackOf([spec]))), {
      baselineIsRelease: true,
    });
    expect(diff.resources).toEqual([]);
  });
});
