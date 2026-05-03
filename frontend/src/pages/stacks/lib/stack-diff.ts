import type {
  FormStackResourceData,
  FormVolumeExtendedData,
} from "@/pages/stacks/schemas/form-schema";

/**
 * Pure diff helpers for the stack edit session.
 *
 * Compares a working `draft` against an immutable `baseline` snapshot via a
 * small recursive deep-equality helper (no external dependency). Results are
 * intended for memoization in `useStackEditSession`.
 */

export type ResourceArr = Partial<FormStackResourceData>[];
export type VolumeArr = Partial<FormVolumeExtendedData>[];

export interface PerResourceDirty {
  rowsChanged: number;
  fieldsChanged: number;
}

export interface PerVolumeDirty {
  fieldsChanged: number;
}

export interface StackDiff {
  dirtyResourceIdx: Set<number>;
  dirtyVolumeIdx: Set<number>;
  perResourceDirty: Map<number, PerResourceDirty>;
  perVolumeDirty: Map<number, PerVolumeDirty>;
  addonLinkCount: number;
}

/** Recursive structural equality that treats `undefined` like a missing key. */
function deepEqual(a: unknown, b: unknown): boolean {
  if (a === b) return true;
  if (a === null || b === null) return a === b;
  if (typeof a !== typeof b) return false;
  if (typeof a !== "object") return false;

  if (Array.isArray(a) || Array.isArray(b)) {
    if (!Array.isArray(a) || !Array.isArray(b)) return false;
    if (a.length !== b.length) return false;
    for (let i = 0; i < a.length; i++) {
      if (!deepEqual(a[i], b[i])) return false;
    }
    return true;
  }

  const ao = a as Record<string, unknown>;
  const bo = b as Record<string, unknown>;
  const aKeys = Object.keys(ao).filter((k) => ao[k] !== undefined);
  const bKeys = Object.keys(bo).filter((k) => bo[k] !== undefined);
  if (aKeys.length !== bKeys.length) return false;
  for (const k of aKeys) {
    if (!deepEqual(ao[k], bo[k])) return false;
  }
  return true;
}

/** Count top-level field changes between two objects (shallow on keys, deep on values). */
function countChangedFields(
  a: Record<string, unknown> | undefined | null,
  b: Record<string, unknown> | undefined | null,
): number {
  const ao = (a ?? {}) as Record<string, unknown>;
  const bo = (b ?? {}) as Record<string, unknown>;
  const keys = new Set<string>([...Object.keys(ao), ...Object.keys(bo)]);
  let n = 0;
  for (const k of keys) {
    if (!deepEqual(ao[k], bo[k])) n++;
  }
  return n;
}

export function isResourceDirty(
  draftResource: Partial<FormStackResourceData> | undefined,
  baselineResource: Partial<FormStackResourceData> | undefined,
): boolean {
  return !deepEqual(draftResource, baselineResource);
}

export function isVolumeDirty(
  draftVolume: Partial<FormVolumeExtendedData> | undefined,
  baselineVolume: Partial<FormVolumeExtendedData> | undefined,
): boolean {
  return !deepEqual(draftVolume, baselineVolume);
}

function getEnvVars(
  r: Partial<FormStackResourceData> | undefined,
): Array<Record<string, unknown>> {
  return ((r?.execution_config?.environment_variables ?? []) as Array<
    Record<string, unknown>
  >);
}

function countEnvRowsChanged(
  draft: Partial<FormStackResourceData> | undefined,
  base: Partial<FormStackResourceData> | undefined,
): number {
  const d = getEnvVars(draft);
  const b = getEnvVars(base);
  const max = Math.max(d.length, b.length);
  let n = 0;
  for (let i = 0; i < max; i++) {
    if (!deepEqual(d[i], b[i])) n++;
  }
  return n;
}

function uniqueAddonIds(resources: ResourceArr): Set<string> {
  const ids = new Set<string>();
  for (const r of resources) {
    const envs = getEnvVars(r);
    for (const e of envs) {
      if (e && e.from === "addon" && typeof e.addonId === "string" && e.addonId.length > 0) {
        ids.add(e.addonId);
      }
    }
  }
  return ids;
}

export function diffStack(
  draft: { resources: ResourceArr; volumes: VolumeArr },
  baseline: { resources: ResourceArr; volumes: VolumeArr },
): StackDiff {
  const dirtyResourceIdx = new Set<number>();
  const perResourceDirty = new Map<number, PerResourceDirty>();
  const maxR = Math.max(draft.resources.length, baseline.resources.length);
  for (let i = 0; i < maxR; i++) {
    const d = draft.resources[i];
    const b = baseline.resources[i];
    if (!isResourceDirty(d, b)) continue;
    dirtyResourceIdx.add(i);
    perResourceDirty.set(i, {
      rowsChanged: countEnvRowsChanged(d, b),
      fieldsChanged: countChangedFields(
        d as Record<string, unknown> | undefined,
        b as Record<string, unknown> | undefined,
      ),
    });
  }

  const dirtyVolumeIdx = new Set<number>();
  const perVolumeDirty = new Map<number, PerVolumeDirty>();
  const maxV = Math.max(draft.volumes.length, baseline.volumes.length);
  for (let i = 0; i < maxV; i++) {
    const d = draft.volumes[i];
    const b = baseline.volumes[i];
    if (!isVolumeDirty(d, b)) continue;
    dirtyVolumeIdx.add(i);
    perVolumeDirty.set(i, {
      fieldsChanged: countChangedFields(
        d as Record<string, unknown> | undefined,
        b as Record<string, unknown> | undefined,
      ),
    });
  }

  const addonLinkCount = uniqueAddonIds(draft.resources).size;

  return {
    dirtyResourceIdx,
    dirtyVolumeIdx,
    perResourceDirty,
    perVolumeDirty,
    addonLinkCount,
  };
}

/** Deep clone via JSON round-trip. Form data is plain JSON so this is safe. */
export function cloneJson<T>(v: T): T {
  return JSON.parse(JSON.stringify(v)) as T;
}

export function revertResource(
  draft: { resources: ResourceArr; volumes: VolumeArr },
  baseline: { resources: ResourceArr; volumes: VolumeArr },
  idx: number,
): { resources: ResourceArr; volumes: VolumeArr } {
  const next = { ...draft, resources: draft.resources.slice() };
  if (idx < baseline.resources.length) {
    next.resources[idx] = cloneJson(baseline.resources[idx]);
  } else {
    // The resource only exists in the draft — drop it.
    next.resources.splice(idx, 1);
  }
  return next;
}

export function revertVolume(
  draft: { resources: ResourceArr; volumes: VolumeArr },
  baseline: { resources: ResourceArr; volumes: VolumeArr },
  idx: number,
): { resources: ResourceArr; volumes: VolumeArr } {
  const next = { ...draft, volumes: draft.volumes.slice() };
  if (idx < baseline.volumes.length) {
    next.volumes[idx] = cloneJson(baseline.volumes[idx]);
  } else {
    next.volumes.splice(idx, 1);
  }
  return next;
}
