import { describe, expect, it } from "vitest";
import { tooljet } from "@/pages/stacks/data/templates/tooljet/template";
import { templateToFormData } from "@/pages/stacks/data/templates/template-to-form";
import { draftToSnapshot } from "@/pages/stacks/lib/draft-sync/draft-snapshot";
import { diffSnapshots } from "@/pages/stacks/components/editor/tabs/deployments/release-snapshot-diff";
import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";

/**
 * The reported flow, on real template data rather than a hand-built fixture:
 * deploy a template, bump one image tag, deploy again. Six image-sourced
 * resources and two volumes, so an over-broad rule has plenty to trip on.
 */

const BUMPED_REF = "tooljet/tooljet:v3.21.0-lts";

function tooljetDraft(): EditSessionDraft {
  const { data } = templateToFormData(tooljet);
  return {
    resources: data.spec?.stack_resources ?? [],
    volumes: data.spec?.volumes ?? [],
  } as EditSessionDraft;
}

function bumpImage(draft: EditSessionDraft, resourceName: string): EditSessionDraft {
  return {
    ...draft,
    resources: draft.resources.map((r) =>
      r.name === resourceName ? { ...r, source: { image: { ...r.source?.image, ref: BUMPED_REF } } } : r,
    ),
  } as EditSessionDraft;
}

describe("deploying a template, then bumping an image tag", () => {
  it("has nothing pending once the release matches the draft", () => {
    const draft = tooljetDraft();
    expect(draft.resources).toHaveLength(6);
    const diff = diffSnapshots(draftToSnapshot(draft), draftToSnapshot(draft));
    expect(diff.resources).toEqual([]);
    expect(diff.volumes).toEqual([]);
  });

  it("stages exactly the bumped resource, with a row a reader can act on", () => {
    const draft = tooljetDraft();
    const deployed = draftToSnapshot(draft);
    const staged = diffSnapshots(deployed, draftToSnapshot(bumpImage(draft, "tooljet")));

    expect(staged.resources).toHaveLength(1);
    expect(staged.resources[0].name).toBe("tooljet");
    // A count the modal cannot itemize is the bug this guards: every staged
    // entry must carry at least one legible row.
    expect(staged.resources[0].sections.flatMap((s) => s.rows)).toEqual([
      { key: "image", from: "tooljet/tooljet:v3.20.189-lts", to: BUMPED_REF, kind: "changed" },
    ]);
  });

  it("leaves the sibling sharing that image tag alone", () => {
    const draft = tooljetDraft();
    const staged = diffSnapshots(draftToSnapshot(draft), draftToSnapshot(bumpImage(draft, "tooljet")));
    expect(staged.resources.map((r) => r.name)).not.toContain("tooljet-worker");
  });

  it("clears once the bump is deployed", () => {
    const bumped = bumpImage(tooljetDraft(), "tooljet");
    const diff = diffSnapshots(draftToSnapshot(bumped), draftToSnapshot(bumped));
    expect(diff.resources).toEqual([]);
    expect(diff.volumes).toEqual([]);
  });
});
