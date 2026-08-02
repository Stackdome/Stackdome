import { templates } from "@/pages/stacks/data/templates/registry";
import { templateToFormData } from "@/pages/stacks/data/templates/template-to-form";
import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";

/**
 * Every shipped template as an edit-session draft, built through the same
 * compose parse + convert the Templates Browser runs.
 *
 * The hand-written fixtures elsewhere encode one shape each, deliberately. This
 * corpus is the opposite: whatever the converter actually emits, at whatever
 * size, so a normalization rule that is a little too broad has real data to
 * trip on rather than a fixture written by the same person as the rule.
 */
export function templateDrafts(): Array<[string, EditSessionDraft]> {
  return templates.map((t) => {
    const { data } = templateToFormData(t);
    return [
      t.id,
      {
        resources: data.spec?.stack_resources ?? [],
        volumes: data.spec?.volumes ?? [],
      } as EditSessionDraft,
    ];
  });
}

/** The first image-sourced resource, which every template has at least one of. */
export function firstImageResourceName(draft: EditSessionDraft): string {
  const hit = draft.resources.find((r) => r.source?.image?.ref);
  if (!hit?.name) throw new Error("template has no image-sourced resource to bump");
  return hit.name;
}

/** Retag one resource's image — the smallest edit a user makes to a template. */
export function bumpImage(draft: EditSessionDraft, resourceName: string, ref: string): EditSessionDraft {
  return {
    ...draft,
    resources: draft.resources.map((r) =>
      r.name === resourceName ? { ...r, source: { image: { ...r.source?.image, ref } } } : r,
    ),
  } as EditSessionDraft;
}
