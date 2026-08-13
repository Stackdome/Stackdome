import type {
  FormStackResourceData,
  FormVolumeExtendedData,
} from "@/pages/stacks/schemas/form-schema";

/**
 * Editing helpers for the stack draft: dot-path access, and the reverts that
 * put one field, one row, or one entity back to its baseline value.
 *
 * What counts as a change is not decided here — that lives in the canonical
 * model (`lib/stack-model`), which every surface shares.
 */

export type ResourceArr = Partial<FormStackResourceData>[];
export type VolumeArr = Partial<FormVolumeExtendedData>[];

import type { FormEnvRow } from "@/pages/stacks/lib/connection-mapping";
import { deleteResourceAndReferences } from "@/pages/stacks/lib/delete-references";
import { renameResourceReferences } from "@/pages/stacks/lib/rename-references";
import { deepEqual, pairByFingerprint } from "@/pages/stacks/lib/stack-model/equal";
import { getAtPath } from "@/pages/stacks/lib/stack-model/path";
import { resourceFingerprint, volumeFingerprint } from "@/pages/stacks/lib/stack-model/diff";
import { canonicalResourceFromForm, canonicalVolumeFromForm } from "@/pages/stacks/lib/stack-model/from-form";

/** Deep clone via JSON round-trip. Form data is plain JSON so this is safe.
 *  Passes undefined through (JSON.parse(JSON.stringify(undefined)) throws). */
export function cloneJson<T>(v: T): T {
  if (v === undefined) return v;
  return JSON.parse(JSON.stringify(v)) as T;
}

/**
 * Siblings point at whatever name the resource carried while it was renamed, so
 * a revert that restores the old name has to bring them back with it.
 * Keyed on the name actually changing, not on which field was reverted.
 */
function carryNameRevert(resources: ResourceArr, idx: number, nameBefore?: string): ResourceArr {
  const nameAfter = resources[idx]?.name;
  if (!nameBefore || nameBefore === nameAfter) return resources;
  // A draft-only resource has no baseline name to restore, so reverting the
  // field clears it. Siblings then have nothing to point at.
  if (!nameAfter) return deleteResourceAndReferences(resources, nameBefore);
  return renameResourceReferences(resources, nameBefore, nameAfter);
}

/**
 * Drop the restored entry's references to resources the draft no longer has.
 * A baseline that predates a sibling's rename names something now gone.
 */
function dropUnknownResourceRefs(resources: ResourceArr, idx: number): ResourceArr {
  const entry = resources[idx];
  if (!entry) return resources;

  const known = new Set(resources.map((r) => r.name).filter(Boolean));
  const deps = entry.depends_on ?? [];
  const keptDeps = deps.filter((d) => known.has(d));
  const rows = (entry.execution_config?.environment_variables ?? []) as FormEnvRow[];
  const keptRows = rows.filter(
    (row) =>
      (row.from !== "resource" && row.from !== "resourceTemplate") || known.has(row.resourceName),
  );

  const depsChanged = keptDeps.length !== deps.length;
  const rowsChanged = keptRows.length !== rows.length;
  if (!depsChanged && !rowsChanged) return resources;

  const next = resources.slice();
  next[idx] = {
    ...entry,
    ...(depsChanged ? { depends_on: keptDeps } : {}),
    ...(rowsChanged
      ? { execution_config: { ...entry.execution_config, environment_variables: keptRows } }
      : {}),
  };
  return next;
}

export function revertResource(
  draft: { resources: ResourceArr; volumes: VolumeArr },
  baseline: { resources: ResourceArr; volumes: VolumeArr },
  idx: number,
): { resources: ResourceArr; volumes: VolumeArr } {
  const next = { ...draft, resources: draft.resources.slice() };
  const nameBefore = draft.resources[idx]?.name;
  // A name-aligned baseline can carry nullish holes for draft-only resources
  // (see alignBaselineToDraft) — treat those the same as "past the end".
  const baselineEntry = idx < baseline.resources.length ? baseline.resources[idx] : undefined;
  if (baselineEntry != null) {
    // Keep the draft's live status: the baseline's was captured at deploy time
    // and restoring it would show stale telemetry until the next refresh.
    const liveStatus = (draft.resources[idx] as { status?: unknown } | undefined)?.status;
    const restored = {
      ...cloneJson(baselineEntry),
      ...(liveStatus !== undefined ? { status: liveStatus } : {}),
    } as (typeof next.resources)[number];
    // A restored mount may reference a volume the draft has since deleted —
    // reattaching it would create a dangling mount, so drop those rows.
    if (Array.isArray(restored.volume_mounts)) {
      const draftVolumeNames = new Set(draft.volumes.map((v) => v?.name).filter(Boolean));
      restored.volume_mounts = restored.volume_mounts.filter(
        (m) => m?.source_volume_name && draftVolumeNames.has(m.source_volume_name),
      );
    }
    next.resources[idx] = restored;
    next.resources = dropUnknownResourceRefs(carryNameRevert(next.resources, idx, nameBefore), idx);
  } else if (nameBefore) {
    // The resource only exists in the draft, so dropping it is a delete and
    // has to take its references with it.
    next.resources = deleteResourceAndReferences(next.resources, nameBefore);
  } else {
    next.resources.splice(idx, 1);
  }
  return next;
}

/**
 * Reorder a baseline array so it aligns positionally with the draft, matching
 * entries by `name`. All downstream diffing (diffStack, per-drawer baselines)
 * is positional, but the server returns stack_resources in unstable order and
 * a release snapshot's order need not match the live stack's — so the baseline
 * must be re-keyed onto the draft's order before any index-wise comparison.
 *
 * Draft entries with no baseline match get an `undefined` hole (they read as
 * "added"); baseline entries missing from the draft are appended at the end so
 * deletions still register as dirt.
 */
export function alignBaselineToDraft<T extends { name?: string }>(
  baseline: T[],
  draft: Array<{ name?: string }>,
  fingerprint?: (entry: unknown) => string,
): (T | undefined)[] {
  const byName = new Map<string, T>();
  for (const b of baseline) {
    // First occurrence wins on (malformed) duplicate names.
    if (b?.name && !byName.has(b.name)) byName.set(b.name, b);
  }
  const used = new Set<string>();
  const aligned: (T | undefined)[] = draft.map((d) => {
    if (!d?.name) return undefined;
    const match = byName.get(d.name);
    if (match) used.add(d.name);
    return match;
  });
  // Rename pass: a draft entry with no name match paired with a leftover
  // baseline entry of identical content is the same entity renamed. Slot the
  // baseline into the draft's position so positional diffing reads one changed
  // field (the name) instead of an addition plus a deletion.
  if (fingerprint) {
    const holes = aligned.flatMap((a, i) => (a === undefined && draft[i]?.name ? [i] : []));
    const leftovers = baseline.filter((b) => b?.name && !used.has(b.name));
    const pairs = pairByFingerprint(holes, leftovers, (i) => fingerprint(draft[i]), fingerprint);
    for (const [i, b] of pairs) {
      aligned[i] = b;
      used.add(b.name as string);
    }
  }
  for (const b of baseline) {
    if (!b?.name || !used.has(b.name)) aligned.push(b);
  }
  return aligned;
}

/**
 * Content identity for rename detection, in the canonical terms the diff uses.
 * Two entries pair as a rename here exactly when they pair as a rename in
 * `diffStacks`; both surfaces must reach the same verdict about the same pair.
 *
 * `alignBaselineToDraft` only fingerprints named entries, so canonicalizing
 * cannot fail on a nameless one.
 */
export function resourceRenameFingerprint(entry: unknown): string {
  return resourceFingerprint(canonicalResourceFromForm(entry as FormStackResourceData));
}

export function volumeRenameFingerprint(entry: unknown): string {
  return volumeFingerprint(canonicalVolumeFromForm(entry as FormVolumeExtendedData));
}

/**
 * Per-row status of env vars in a resource's draft, relative to baseline.
 * Rows are matched by `name` (the env KEY) for stable identity across edits;
 * unnamed rows fall back to positional matching. Baseline rows that don't
 * appear in the draft don't surface here — they live in baseline only.
 */
export type EnvRowStatus = "unchanged" | "modified" | "added";

export function envRowsDiff(
  draftRows: Array<Record<string, unknown>>,
  baselineRows: Array<Record<string, unknown>>,
): EnvRowStatus[] {
  const baselineByName = new Map<string, Record<string, unknown>>();
  baselineRows.forEach((r) => {
    const name = typeof r?.name === "string" ? r.name : "";
    if (name) baselineByName.set(name, r);
  });

  return draftRows.map((draft, i) => {
    const draftName = typeof draft?.name === "string" ? draft.name : "";
    if (draftName) {
      const base = baselineByName.get(draftName);
      if (!base) return "added";
      return deepEqual(draft, base) ? "unchanged" : "modified";
    }
    // Unnamed row — fall back to positional match.
    const base = baselineRows[i];
    if (!base) return "added";
    return deepEqual(draft, base) ? "unchanged" : "modified";
  });
}

/**
 * Revert a single env row in `draft.resources[resourceIdx]` to its baseline
 * value, matched by name (or positional fallback for unnamed rows). If the
 * row is "added" with no baseline counterpart, it's removed entirely.
 */
export function revertEnvRow(
  draft: { resources: ResourceArr; volumes: VolumeArr },
  baseline: { resources: ResourceArr; volumes: VolumeArr },
  resourceIdx: number,
  envIdx: number,
): { resources: ResourceArr; volumes: VolumeArr } {
  const draftResource = draft.resources[resourceIdx];
  if (!draftResource) return draft;
  const draftRows = (draftResource.execution_config?.environment_variables ?? []) as Array<
    Record<string, unknown>
  >;
  const draftRow = draftRows[envIdx];
  if (!draftRow) return draft;

  const baselineResource = baseline.resources[resourceIdx];
  const baselineRows = (baselineResource?.execution_config?.environment_variables ?? []) as Array<
    Record<string, unknown>
  >;

  const draftName = typeof draftRow.name === "string" ? draftRow.name : "";
  let baselineRow: Record<string, unknown> | undefined;
  if (draftName) {
    baselineRow = baselineRows.find((r) => (typeof r?.name === "string" ? r.name : "") === draftName);
  } else {
    baselineRow = baselineRows[envIdx];
  }

  const nextRows = draftRows.slice();
  if (baselineRow) {
    nextRows[envIdx] = cloneJson(baselineRow);
  } else {
    // No baseline counterpart — the row is newly added; drop it.
    nextRows.splice(envIdx, 1);
  }

  const nextResources = draft.resources.slice();
  nextResources[resourceIdx] = {
    ...draftResource,
    execution_config: {
      ...draftResource.execution_config,
      environment_variables: nextRows as never,
    },
  };
  return { ...draft, resources: nextResources };
}

// --- Generic dot-path helpers, used by per-field dirty/reset infra. ---

/** Immutably set a nested value via dot-path; returns a clone of `obj` with `path` set.
 * Preserves array-ness: `{...arr}` would coerce arrays into plain objects with
 * numeric keys, so we shallow-clone arrays with `arr.slice()` instead. */
export function setAtPath<T>(obj: T, path: string, value: unknown): T {
  if (!path) return value as T;
  const parts = path.split(".");
  const cloneNode = (node: unknown): Record<string, unknown> | unknown[] => {
    if (Array.isArray(node)) return node.slice() as unknown[];
    if (node && typeof node === "object") return { ...(node as Record<string, unknown>) };
    return {};
  };
  const root = cloneNode(obj);
  let cur: Record<string, unknown> | unknown[] = root;
  for (let i = 0; i < parts.length - 1; i++) {
    const k = parts[i];
    const idx = Array.isArray(cur) ? Number(k) : k;
    const next = (cur as Record<string | number, unknown>)[idx as string];
    const cloned = cloneNode(next);
    (cur as Record<string | number, unknown>)[idx as string] = cloned;
    cur = cloned;
  }
  const lastKey = parts[parts.length - 1];
  const lastIdx = Array.isArray(cur) ? Number(lastKey) : lastKey;
  (cur as Record<string | number, unknown>)[lastIdx as string] = value;
  return root as T;
}

/** Revert a single dot-path field on a resource to its baseline value. */
export function revertResourceField(
  draft: { resources: ResourceArr; volumes: VolumeArr },
  baseline: { resources: ResourceArr; volumes: VolumeArr },
  resourceIdx: number,
  path: string,
): { resources: ResourceArr; volumes: VolumeArr } {
  const draftResource = draft.resources[resourceIdx];
  if (!draftResource) return draft;
  const baselineResource = baseline.resources[resourceIdx];
  const baselineValue = getAtPath(baselineResource, path);
  const nextResource = setAtPath(draftResource, path, cloneJson(baselineValue));
  const nextResources = draft.resources.slice();
  nextResources[resourceIdx] = nextResource;
  return {
    ...draft,
    resources: dropUnknownResourceRefs(
      carryNameRevert(nextResources, resourceIdx, draftResource.name),
      resourceIdx,
    ),
  };
}

export function revertVolume(
  draft: { resources: ResourceArr; volumes: VolumeArr },
  baseline: { resources: ResourceArr; volumes: VolumeArr },
  idx: number,
): { resources: ResourceArr; volumes: VolumeArr } {
  const next = { ...draft, volumes: draft.volumes.slice() };
  // Nullish holes from a name-aligned baseline mean "draft-only" — drop the row.
  const baselineEntry = idx < baseline.volumes.length ? baseline.volumes[idx] : undefined;
  if (baselineEntry != null) {
    const liveStatus = (draft.volumes[idx] as { status?: unknown } | undefined)?.status;
    next.volumes[idx] = {
      ...cloneJson(baselineEntry),
      ...(liveStatus !== undefined ? { status: liveStatus } : {}),
    } as (typeof next.volumes)[number];
  } else {
    next.volumes.splice(idx, 1);
  }
  return next;
}
