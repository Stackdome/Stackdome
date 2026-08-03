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
 * One real edit per policy rule, each of which must survive to the server —
 * the direction `invariants.test.ts` cannot see, since asserting that loading
 * a stack changes nothing catches only a rule that suppresses too little.
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
   * KNOWN LIMITATION. Against a release baseline, "this spec never pinned a
   * commit" and "the user just cleared the pin" both arrive as an absent
   * commit, so neither the release card nor the drawer tint reports a cleared
   * pin while the live release stays on the old commit. The write still
   * happens, so no work is lost. Telling the two apart needs what the spec
   * pinned when the release was cut, which no snapshot records.
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
