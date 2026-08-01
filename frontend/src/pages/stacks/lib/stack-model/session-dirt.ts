import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";
import { canonicalFromDraft } from "./from-form";
import { diffStacks, type FieldChange } from "./diff";
import { deepEqual } from "./equal";
import { sectionForField, type FieldSection } from "./policy";

export interface ResourceDirtyTabs {
  configuration: boolean;
  deployment: boolean;
  environment: boolean;
}

export interface SessionDirt {
  /** Draft indices with changes, for the canvas badges and the drawer. */
  dirtyResourceIdx: Set<number>;
  dirtyVolumeIdx: Set<number>;
  /** Changed canonical paths per draft index, for the per-field tints. */
  dirtyPathsByResourceIdx: Map<number, ReadonlySet<string>>;
  tabsByResourceIdx: Map<number, ResourceDirtyTabs>;
  /** Entities the baseline had and the draft doesn't. They have no draft index,
   *  so they can't be badged — but they are still undeployed changes to count. */
  removedCount: number;
}

export const NO_DIRT: SessionDirt = {
  dirtyResourceIdx: new Set(),
  dirtyVolumeIdx: new Set(),
  dirtyPathsByResourceIdx: new Map(),
  tabsByResourceIdx: new Map(),
  removedCount: 0,
};

function tabsOf(fields: FieldChange[]): ResourceDirtyTabs {
  const tabs: ResourceDirtyTabs = { configuration: false, deployment: false, environment: false };
  for (const f of fields) tabs[f.section] = true;
  return tabs;
}

/**
 * What the user has changed in this session, in the shape the editor's UI needs:
 * indices for badges, canonical paths for tints, sections for tab dots.
 *
 * The comparison itself is the same one autosave and the deploy surfaces run, so
 * a field can no longer read as edited here while reading as unchanged there.
 */
export function sessionDirt(draft: EditSessionDraft, baseline: EditSessionDraft): SessionDirt {
  const draftCanonical = canonicalFromDraft(draft);
  const diff = diffStacks(canonicalFromDraft(baseline), draftCanonical);

  const dirtyResourceIdx = new Set<number>();
  const dirtyPathsByResourceIdx = new Map<number, ReadonlySet<string>>();
  const tabsByResourceIdx = new Map<number, ResourceDirtyTabs>();
  let removedCount = 0;

  for (const entry of diff.resources) {
    if (entry.change === "removed") {
      removedCount += 1;
      continue;
    }
    const idx = draftCanonical.indexByName.get(entry.name);
    if (idx === undefined) continue;
    dirtyResourceIdx.add(idx);
    dirtyPathsByResourceIdx.set(idx, new Set(entry.fields.map((f) => f.path)));
    tabsByResourceIdx.set(idx, tabsOf(entry.fields));
  }

  // A resource that doesn't validate yet has no canonical form, so the diff
  // cannot see it — but the user is still typing into it, and an edit they can
  // see must be an edit the editor admits to. Compare those raw.
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
    if (entry.change === "removed") {
      removedCount += 1;
      continue;
    }
    const idx = volumeIdxByName.get(entry.name);
    if (idx !== undefined) dirtyVolumeIdx.add(idx);
  }

  return { dirtyResourceIdx, dirtyVolumeIdx, dirtyPathsByResourceIdx, tabsByResourceIdx, removedCount };
}

export function dirtyTotal(dirt: SessionDirt): number {
  return dirt.dirtyResourceIdx.size + dirt.dirtyVolumeIdx.size + dirt.removedCount;
}

/**
 * A drawer field names the canonical path (or paths) it edits. The git revision
 * helpers name two: which of branch/tag carries the value is itself the edit.
 */
export function isAnyPathDirty(paths: ReadonlySet<string> | undefined, fieldPaths: string): boolean {
  if (!paths?.size) return false;
  return fieldPaths.split(" ").some((p) => {
    if (paths.has(p)) return true;
    // A parent path covers its children: `source` covers `source.git.repo_url`.
    for (const dirty of paths) if (dirty.startsWith(`${p}.`)) return true;
    return false;
  });
}

export function sectionOfPath(path: string): FieldSection {
  return sectionForField(path);
}
