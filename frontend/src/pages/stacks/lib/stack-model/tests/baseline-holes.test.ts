import { describe, it, expect } from "vitest";
import { canonicalFromDraft } from "../from-form";
import { alignBaselineToDraft, cloneJson } from "@/pages/stacks/lib/stack-diff";
import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";

/**
 * A draft-only resource or volume leaves a hole in the aligned baseline, and
 * the JSON clone in `session.start` turns that hole into a null. Every dirt
 * calculation runs the baseline back through `canonicalFromDraft`, so a null
 * there crashed the whole editor.
 */

const asDraft = (resources: unknown[], volumes: unknown[]) =>
  ({ resources, volumes }) as EditSessionDraft;

describe("canonicalFromDraft with baseline holes", () => {
  it("skips a null volume instead of throwing", () => {
    const draft = asDraft([], [{ name: "data", size: "1Gi" }, null]);
    expect(canonicalFromDraft(draft).volumes.map((v) => v.name)).toEqual(["data"]);
  });

  it("skips a null resource instead of throwing", () => {
    const web = { name: "web", sourceType: "image", source: { image: { ref: "nginx:1.27" } } };
    const draft = asDraft([web, null], []);
    expect(canonicalFromDraft(draft).resources.map((r) => r.name)).toEqual(["web"]);
  });

  /** The exact path that reaches it: align, clone, canonicalize. */
  it("survives the align-then-clone round trip a draft-only entry produces", () => {
    const serverVolumes = [{ name: "data" }];
    const draftVolumes = [{ name: "data" }, { name: "added-in-draft" }];
    const aligned = alignBaselineToDraft(serverVolumes, draftVolumes);
    expect(aligned[1]).toBeUndefined();

    const baseline = cloneJson({ resources: [], volumes: aligned });
    expect(baseline.volumes[1]).toBeNull();
    expect(canonicalFromDraft(baseline as EditSessionDraft).volumes.map((v) => v.name)).toEqual([
      "data",
    ]);
  });
});
