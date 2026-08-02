import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";
import { canonicalFromDraft } from "./from-form";
import { diffStacks } from "./diff";
import { deepEqual } from "./equal";

export interface SessionDirt {
  /** Draft indices the user has changed, for the canvas badges and drawer dots. */
  dirtyResourceIdx: Set<number>;
  dirtyVolumeIdx: Set<number>;
}

/**
 * Which entries of the draft the user has changed, by index — the shape the
 * editor's index-keyed UI needs.
 */
export function sessionDirt(draft: EditSessionDraft, baseline: EditSessionDraft): SessionDirt {
  const draftCanonical = canonicalFromDraft(draft);
  const diff = diffStacks(canonicalFromDraft(baseline), draftCanonical);

  const dirtyResourceIdx = new Set<number>();
  for (const entry of diff.resources) {
    // A removed resource has no draft index to badge; the changes modal lists it.
    const idx = draftCanonical.indexByName.get(entry.name);
    if (idx !== undefined) dirtyResourceIdx.add(idx);
  }

  // A resource that doesn't validate yet has no canonical form, so the diff
  // cannot see it. Compare those raw.
  const baselineByName = new Map(
    baseline.resources.map((r, i) => [r?.name?.trim() || `#${i}`, r] as const),
  );
  draft.resources.forEach((r, idx) => {
    const name = r?.name?.trim();
    if (name && !draftCanonical.held.has(name)) return;
    const previous = baselineByName.get(name || `#${idx}`) ?? baseline.resources[idx];
    if (!deepEqual(r, previous)) dirtyResourceIdx.add(idx);
  });

  const volumeIdxByName = new Map(draft.volumes.map((v, i) => [v?.name?.trim() ?? "", i]));
  const dirtyVolumeIdx = new Set<number>();
  for (const entry of diff.volumes) {
    const idx = volumeIdxByName.get(entry.name);
    if (idx !== undefined) dirtyVolumeIdx.add(idx);
  }

  return { dirtyResourceIdx, dirtyVolumeIdx };
}
