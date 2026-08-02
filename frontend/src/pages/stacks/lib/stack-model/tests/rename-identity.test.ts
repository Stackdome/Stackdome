import { describe, expect, it } from "vitest";
import type { Stack } from "@/api/stacks";
import { formResourcesFromSpec } from "@/pages/stacks/lib/spec-to-form";
import { canonicalFromStack } from "../from-api";
import { canonicalFromDraft } from "../from-form";
import { diffStacks } from "../diff";
import { alignBaselineToDraft, resourceRenameFingerprint } from "@/pages/stacks/lib/stack-diff";

/**
 * Two surfaces decide whether a pair is a rename: `diffStacks`, which lists it
 * as one renamed entry, and `alignBaselineToDraft`, which slots the baseline
 * into the draft's position so the drawer has something to tint against. They
 * must agree — a pair the diff calls a rename but the alignment does not leaves
 * the drawer with no baseline, and every field on the resource reads edited.
 */

const base = {
  id: "r-web",
  stack_id: "s1",
  revision: 3,
  name: "web",
  workload_type: "Service",
  source: { git: { repo_url: "https://github.com/acme/demo.git", branch: "main" } },
  ports: [],
  volume_mounts: [],
  depends_on: [],
  execution_config: { command: ["node", "a.js"], args: [], environment_variables: [] },
};

const stackOf = (resources: unknown[]): Stack =>
  ({ id: "s1", name: "demo", spec: { stack_resources: resources, volumes: [], connections: [] } }) as unknown as Stack;

const formOf = (resources: unknown[]) => formResourcesFromSpec(resources as never, []);

/** Did the alignment pair the renamed draft entry with the old baseline entry? */
function alignmentPairs(baseline: unknown[], draft: unknown[]): boolean {
  const aligned = alignBaselineToDraft(
    formOf(baseline) as Array<{ name?: string }>,
    formOf(draft) as Array<{ name?: string }>,
    resourceRenameFingerprint,
  );
  return aligned[0] != null;
}

/** Did the diff call it a rename? */
function diffCallsItRename(baseline: unknown[], draft: unknown[]): boolean {
  const diff = diffStacks(
    canonicalFromStack(stackOf(baseline)),
    canonicalFromDraft({ resources: formOf(draft), volumes: [] }),
  );
  return diff.resources.length === 1 && diff.resources[0].change === "renamed";
}

describe("the diff and the drawer alignment agree about renames", () => {
  const cases: Array<[string, unknown[], unknown[]]> = [
    ["a plain rename", [base], [{ ...base, name: "web2" }]],
    [
      "a rename whose baseline carries server-written fields",
      [{ ...base, status: { state: "Running" }, outputs: [{ name: "url" }] }],
      [{ ...base, name: "web2" }],
    ],
    [
      "a rename where the baseline omits the build paths the form fills in",
      [base],
      [{ ...base, name: "web2", source: { git: { ...base.source.git, dockerfile_path: "Dockerfile", build_context: "." } } }],
    ],
  ];

  for (const [label, baseline, draft] of cases) {
    it(`${label}: both say renamed`, () => {
      expect(diffCallsItRename(baseline, draft)).toBe(true);
      expect(alignmentPairs(baseline, draft)).toBe(true);
    });
  }

  it("a rename that also changes the image is not a rename at all", () => {
    const draft = [{ ...base, name: "web2", source: { image: { ref: "web:2" } }, sourceType: "image" }];
    expect(diffCallsItRename([base], draft)).toBe(false);
    expect(alignmentPairs([base], draft)).toBe(false);
  });

  /**
   * Revision keys are excluded from the fingerprint so a resolver-written
   * commit cannot block pairing. A rename that also retargets the branch
   * therefore still pairs — and must carry the branch change with it, or the
   * edit is reported nowhere.
   */
  it("a rename that also retargets the branch pairs, and still reports the branch", () => {
    const draft = [{ ...base, name: "web2", source: { git: { ...base.source.git, branch: "next" } } }];
    const diff = diffStacks(
      canonicalFromStack(stackOf([base])),
      canonicalFromDraft({ resources: formOf(draft), volumes: [] }),
    );
    expect(diff.resources).toHaveLength(1);
    expect(diff.resources[0]).toMatchObject({ name: "web2", fromName: "web", change: "renamed" });
    expect(diff.resources[0].fields).toEqual([
      expect.objectContaining({ path: "source.git.branch", from: "main", to: "next" }),
    ]);
  });

  /**
   * Key order is not content. The alignment once fingerprinted the raw form
   * object with JSON.stringify, so two structurally identical resources whose
   * keys were built in a different order produced different fingerprints.
   */
  it("ignores the order the object's keys were built in", () => {
    const reordered = { execution_config: base.execution_config, source: base.source, ...base, name: "web2" };
    expect(alignmentPairs([base], [reordered])).toBe(true);
  });
});
